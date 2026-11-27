// Command api is the Lambda entry point for portal, public, and administrative APIs. It contains
// wiring only: configuration from the environment, AWS clients, the GitHub adapter, the HTTP
// server, and the API Gateway HTTP API adapter (A-31). Deployment stays blocked until F-03 step 5;
// sign-in is disabled unless CLA_SIGN_IN_ENABLED=true (F-07, D-20).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	lambdaruntime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	awsadapters "github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/aws"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github"
	lambdaadapter "github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/lambda"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/memory"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/server"
)

// config is read from the environment that Terraform sets on the function (F-07 unit 3).
type config struct {
	PortalOrigin      string
	CallbackURL       string
	SignInErrorRoute  string
	SignInEnabled     bool
	ReturnToAllowlist []string
	SessionIdle       time.Duration
	SessionAbsolute   time.Duration
	Tables            awsadapters.Tables
	GitHubClientID    string
	GitHubSecretID    string
	GitHubWebBaseURL  string
	GitHubAPIBaseURL  string
	DynamoDBEndpoint  string
}

// loadConfig reads the environment. Secrets are never in the environment: only the Secrets
// Manager ID of the OAuth client secret is.
func loadConfig(getenv func(string) string) (config, error) {
	or := func(name, fallback string) string {
		if v := getenv(name); v != "" {
			return v
		}
		return fallback
	}
	c := config{
		PortalOrigin:     getenv("CLA_PORTAL_ORIGIN"),
		CallbackURL:      getenv("CLA_GITHUB_CALLBACK_URL"),
		SignInErrorRoute: or("CLA_SIGN_IN_ERROR_ROUTE", "/sign-in/error"),
		SignInEnabled:    getenv("CLA_SIGN_IN_ENABLED") == "true",
		Tables: awsadapters.Tables{
			Ephemeral: getenv("CLA_TABLE_EPHEMERAL"),
			Core:      getenv("CLA_TABLE_CORE"),
			Audit:     getenv("CLA_TABLE_AUDIT"),
		},
		GitHubClientID:   getenv("CLA_GITHUB_CLIENT_ID"),
		GitHubSecretID:   getenv("CLA_GITHUB_CLIENT_SECRET_ID"),
		GitHubWebBaseURL: or("CLA_GITHUB_WEB_BASE_URL", "https://github.com"),
		GitHubAPIBaseURL: or("CLA_GITHUB_API_BASE_URL", "https://api.github.com"),
		DynamoDBEndpoint: getenv("CLA_DYNAMODB_ENDPOINT"),
	}
	var err error
	if c.SessionIdle, err = durationSeconds("CLA_SESSION_IDLE", getenv("CLA_SESSION_IDLE"), domain.DefaultSessionPolicy.Idle); err != nil {
		return c, err
	}
	if c.SessionAbsolute, err = durationSeconds("CLA_SESSION_ABSOLUTE", getenv("CLA_SESSION_ABSOLUTE"), domain.DefaultSessionPolicy.Absolute); err != nil {
		return c, err
	}
	for _, p := range strings.Split(or("CLA_RETURN_TO_ALLOWLIST", "/"), ",") {
		if p = strings.TrimSpace(p); p != "" {
			c.ReturnToAllowlist = append(c.ReturnToAllowlist, p)
		}
	}
	required := []struct{ name, value string }{
		{"CLA_PORTAL_ORIGIN", c.PortalOrigin}, {"CLA_GITHUB_CALLBACK_URL", c.CallbackURL},
		{"CLA_TABLE_EPHEMERAL", c.Tables.Ephemeral}, {"CLA_TABLE_CORE", c.Tables.Core}, {"CLA_TABLE_AUDIT", c.Tables.Audit},
		{"CLA_GITHUB_CLIENT_ID", c.GitHubClientID}, {"CLA_GITHUB_CLIENT_SECRET_ID", c.GitHubSecretID},
	}
	var missing []string
	for _, r := range required {
		if r.value == "" {
			missing = append(missing, r.name)
		}
	}
	if len(missing) > 0 {
		return c, fmt.Errorf("missing configuration: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

func durationSeconds(name, value string, fallback time.Duration) (time.Duration, error) {
	if value == "" {
		return fallback, nil
	}
	secs, err := strconv.Atoi(value)
	if err != nil || secs <= 0 {
		return 0, fmt.Errorf("%s must be a positive number of seconds", name)
	}
	return time.Duration(secs) * time.Second, nil
}

// slogSecurity writes security events as structured JSON. Events carry no personal data.
type slogSecurity struct{ log *slog.Logger }

func (l slogSecurity) Log(_ context.Context, e ports.SecurityEvent) {
	l.log.Info("security", "event", e.Name, "outcome", e.Outcome, "request_id", e.RequestID)
}

func build(ctx context.Context, c config, log *slog.Logger) (http.Handler, error) {
	awsCfg, err := awsadapters.LoadConfig(ctx)
	if err != nil {
		return nil, err
	}
	clientSecret, err := awsadapters.SecretString(ctx, secretsmanager.NewFromConfig(awsCfg), c.GitHubSecretID)
	if err != nil {
		return nil, err
	}
	identity, err := github.NewOAuthAdapter(github.OAuthConfig{
		ClientID: c.GitHubClientID, ClientSecret: clientSecret,
		WebBaseURL: c.GitHubWebBaseURL, APIBaseURL: c.GitHubAPIBaseURL, Timeout: 10 * time.Second,
	}, &http.Client{})
	if err != nil {
		return nil, err
	}
	random := memory.Random{}
	stores, err := awsadapters.NewDynamoDB(awsadapters.NewDynamoDBClient(awsCfg, c.DynamoDBEndpoint), c.Tables, random)
	if err != nil {
		return nil, err
	}
	clock := memory.Clock{}
	auth, err := application.NewAuth(application.AuthConfig{
		Enabled:     c.SignInEnabled,
		CallbackURL: c.CallbackURL,
		ReturnTo:    domain.ReturnToPolicy{Allowed: c.ReturnToAllowlist},
		Session:     domain.SessionPolicy{Idle: c.SessionIdle, Absolute: c.SessionAbsolute},
	}, application.AuthDependencies{
		Clock: clock, Random: random, Identity: identity, Attempts: stores,
		Sessions: stores.Sessions(), Contributors: stores.Contributors(), Security: slogSecurity{log},
	})
	if err != nil {
		return nil, err
	}
	return server.New(server.Config{PortalOrigin: c.PortalOrigin, SignInErrorRoute: c.SignInErrorRoute}, auth, clock)
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c, err := loadConfig(os.Getenv)
	if err != nil {
		log.Error("configuration", "error", err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	handler, err := build(ctx, c, log)
	cancel()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("start-up timed out")
		} else {
			log.Error("start-up", "error", err.Error())
		}
		os.Exit(1)
	}
	log.Info("api starting", "sign_in_enabled", c.SignInEnabled)
	lambdaruntime.Start(lambdaadapter.New(handler).Handle)
}
