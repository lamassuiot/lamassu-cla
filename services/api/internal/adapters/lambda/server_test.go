package lambda_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github/githubtest"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/lambda"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/memory"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/server"
)

const (
	portal   = "https://cla.example"
	callback = portal + "/api/auth/github/callback"
	// Synthetic credential for the fake GitHub; it never touches a real service.
	fakeClientSecret = "fake-client-secret" //nolint:gosec // synthetic test value
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// newAPI composes the real F-07 server with in-memory stores and the fake GitHub, exactly as
// cmd/api does with the AWS adapters.
func newAPI(t *testing.T) http.Handler {
	t.Helper()
	gh := githubtest.New()
	t.Cleanup(gh.Close)
	identity, err := github.NewOAuthAdapter(github.OAuthConfig{
		ClientID: "fake-client-id", ClientSecret: fakeClientSecret, WebBaseURL: gh.URL(), APIBaseURL: gh.URL(),
	}, gh.Client())
	if err != nil {
		t.Fatal(err)
	}
	clock := fixedClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	auth, err := application.NewAuth(application.AuthConfig{
		Enabled: true, CallbackURL: callback,
		ReturnTo: domain.ReturnToPolicy{Allowed: []string{"/", "/agreements/"}},
		Session:  domain.DefaultSessionPolicy,
	}, application.AuthDependencies{
		Clock: clock, Random: memory.Random{}, Identity: identity, Attempts: memory.NewSignInAttempts(),
		Sessions: memory.NewSessions(), Contributors: memory.NewContributors(), Security: &memory.SecurityLog{},
	})
	if err != nil {
		t.Fatal(err)
	}
	h, err := server.New(server.Config{PortalOrigin: portal, SignInErrorRoute: "/sign-in/error"}, auth, clock)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func event(method, rawPath, rawQuery string, headers map[string]string, cookies ...string) events.APIGatewayV2HTTPRequest {
	if headers == nil {
		headers = map[string]string{}
	}
	headers["host"] = "cla.example"
	return events.APIGatewayV2HTTPRequest{
		Version: "2.0", RawPath: rawPath, RawQueryString: rawQuery, Headers: headers, Cookies: cookies,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			RequestID: "synthetic-request", Stage: "$default",
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: method, Path: rawPath, Protocol: "HTTP/1.1", SourceIP: "192.0.2.1", UserAgent: "synthetic",
			},
		},
	}
}

// The whole F-07 API behind the API Gateway adapter (A-31): redirects, cookies in the cookies
// array, problem documents, and the Origin check all survive the translation.
func TestServerThroughAdapter(t *testing.T) {
	a := lambda.New(newAPI(t))
	ctx := context.Background()

	start, err := a.Handle(ctx, event(http.MethodGet, "/api/v1/auth/github/start", "return_to=%2Fagreements%2F", nil))
	if err != nil {
		t.Fatal(err)
	}
	if start.StatusCode != http.StatusFound || start.Headers["Cache-Control"] != "no-store" {
		t.Fatalf("start = %+v", start)
	}
	if len(start.Cookies) != 1 || !strings.HasPrefix(start.Cookies[0], "__Host-cla_oauth=") ||
		!strings.Contains(start.Cookies[0], "HttpOnly; Secure; SameSite=Lax") {
		t.Fatalf("cookies = %v", start.Cookies)
	}
	if _, inHeaders := start.Headers["Set-Cookie"]; inHeaders {
		t.Fatal("Set-Cookie must travel in the cookies array")
	}
	loc, err := url.Parse(start.Headers["Location"])
	if err != nil || loc.Query().Get("code_challenge_method") != "S256" || loc.Query().Get("redirect_uri") != callback {
		t.Fatalf("location = %s", start.Headers["Location"])
	}

	// A callback carrying a duplicated code arrives verbatim and is rejected as malformed.
	cookie := strings.SplitN(start.Cookies[0], ";", 2)[0]
	state := strings.TrimPrefix(cookie, "__Host-cla_oauth=")
	cb, err := a.Handle(ctx, event(http.MethodGet, "/api/auth/github/callback", "code=a&code=b&state="+url.QueryEscape(state), nil, cookie))
	if err != nil {
		t.Fatal(err)
	}
	cbLoc, err := url.Parse(cb.Headers["Location"])
	if err != nil || cb.StatusCode != http.StatusFound || cbLoc.Query().Get("reason") != "malformed" {
		t.Fatalf("callback = %+v", cb)
	}

	// Problem documents stay as text.
	me, err := a.Handle(ctx, event(http.MethodGet, "/api/v1/session", "", nil))
	if err != nil {
		t.Fatal(err)
	}
	var problem map[string]any
	if err := json.Unmarshal([]byte(me.Body), &problem); err != nil || me.StatusCode != http.StatusUnauthorized ||
		problem["code"] != "SESSION_REQUIRED" || me.IsBase64Encoded {
		t.Fatalf("session = %+v (%v)", me, err)
	}
	if me.Headers["Content-Type"] != "application/problem+json" || me.Headers["X-Request-Id"] == "" {
		t.Fatalf("headers = %v", me.Headers)
	}

	// The Origin check on a state-changing request.
	denied, err := a.Handle(ctx, event(http.MethodPost, "/api/v1/session/sign-out", "", map[string]string{"origin": "https://evil.example"}))
	if err != nil {
		t.Fatal(err)
	}
	if denied.StatusCode != http.StatusForbidden || !strings.Contains(denied.Body, "ORIGIN_NOT_ALLOWED") {
		t.Fatalf("sign-out = %+v", denied)
	}
	ok, err := a.Handle(ctx, event(http.MethodPost, "/api/v1/session/sign-out", "", map[string]string{"origin": portal}))
	if err != nil {
		t.Fatal(err)
	}
	if ok.StatusCode != http.StatusNoContent || ok.Body != "" || len(ok.Cookies) != 1 {
		t.Fatalf("idempotent sign-out = %+v", ok)
	}
}
