// Package github implements ports with the GitHub API. Adapters are the only layer that imports SDKs.
//
// The OAuth adapter uses net/http directly: the three calls F-07 needs (authorization URL, code
// exchange, GET /user, plus best-effort revocation) are simpler than any SDK surface, and the
// adapter must guarantee that no token, code, or state ever reaches a log.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// OAuthConfig configures the sign-in OAuth App adapter (D-06, D-30). Secrets are supplied by the
// wiring layer from Secrets Manager; nothing here is committed.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	// WebBaseURL is GitHub's web origin, normally https://github.com; tests point it at a fake.
	WebBaseURL string
	// APIBaseURL is GitHub's REST API origin, normally https://api.github.com.
	APIBaseURL string
	// Timeout bounds each upstream call (F-07, error cases: no retry loop).
	Timeout time.Duration
}

// OAuthAdapter implements ports.IdentityProvider for a GitHub OAuth App.
type OAuthAdapter struct {
	cfg    OAuthConfig
	client *http.Client
}

// maxBody bounds upstream responses; GitHub's token and user responses are small.
const maxBody = 1 << 20

// NewOAuthAdapter validates the configuration and returns an adapter.
func NewOAuthAdapter(cfg OAuthConfig, client *http.Client) (*OAuthAdapter, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("github oauth: client ID and secret are required")
	}
	for _, u := range []string{cfg.WebBaseURL, cfg.APIBaseURL} {
		parsed, err := url.Parse(u)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("github oauth: invalid base URL %q", u)
		}
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if client == nil {
		client = &http.Client{}
	}
	// Redirects from GitHub's token or user endpoints are never expected; following one could
	// send the client secret elsewhere.
	c := *client
	c.Timeout = cfg.Timeout
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &OAuthAdapter{cfg: cfg, client: &c}, nil
}

// AuthorizationURL builds the GitHub authorization URL with PKCE S256 and an explicitly empty
// scope (F-07, main flow step 2).
func (a *OAuthAdapter) AuthorizationURL(req ports.AuthorizationRequest) string {
	q := url.Values{}
	q.Set("client_id", a.cfg.ClientID)
	q.Set("redirect_uri", req.RedirectURI)
	q.Set("state", req.State)
	q.Set("code_challenge", req.CodeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("scope", "")
	q.Set("allow_signup", "false")
	return strings.TrimRight(a.cfg.WebBaseURL, "/") + "/login/oauth/authorize?" + q.Encode()
}

type tokenResponse struct {
	AccessToken      string  `json:"access_token"`
	Scope            *string `json:"scope"`
	TokenType        string  `json:"token_type"`
	RefreshToken     string  `json:"refresh_token"`
	Error            string  `json:"error"`
	ErrorDescription string  `json:"error_description"`
}

// Exchange trades the authorization code for a user access token (F-07, step 4.5). The scope
// field is returned as a pointer so that the use case can distinguish "absent" from "empty".
func (a *OAuthAdapter) Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (ports.TokenExchange, error) {
	form := url.Values{}
	form.Set("client_id", a.cfg.ClientID)
	form.Set("client_secret", a.cfg.ClientSecret)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("redirect_uri", redirectURI)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(a.cfg.WebBaseURL, "/")+"/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return ports.TokenExchange{}, fmt.Errorf("github oauth: building exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return ports.TokenExchange{}, fmt.Errorf("github oauth: exchange: %w", sanitize(err))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return ports.TokenExchange{}, fmt.Errorf("github oauth: exchange returned status %d", resp.StatusCode)
	}
	var body tokenResponse
	if err := decodeJSON(resp.Body, &body); err != nil {
		return ports.TokenExchange{}, fmt.Errorf("github oauth: exchange response: %w", err)
	}
	// GitHub reports OAuth errors with status 200 and an error field.
	if body.Error != "" {
		return ports.TokenExchange{}, fmt.Errorf("github oauth: exchange error %q", body.Error)
	}
	if body.AccessToken == "" {
		return ports.TokenExchange{}, errors.New("github oauth: exchange response has no access token")
	}
	return ports.TokenExchange{AccessToken: body.AccessToken, GrantedScope: body.Scope, RefreshToken: body.RefreshToken}, nil
}

type userResponse struct {
	// ID is decoded as a raw number to keep 64-bit precision; GitHub IDs are integers.
	ID    json.Number `json:"id"`
	Login string      `json:"login"`
}

// LookupIdentity calls GET /user and extracts only id and login (F-07, step 4.5; AC-07-14). The
// X-OAuth-Scopes header is returned as a pointer: nil means absent.
func (a *OAuthAdapter) LookupIdentity(ctx context.Context, accessToken string) (ports.IdentityLookup, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(a.cfg.APIBaseURL, "/")+"/user", http.NoBody)
	if err != nil {
		return ports.IdentityLookup{}, fmt.Errorf("github oauth: building user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := a.client.Do(req)
	if err != nil {
		return ports.IdentityLookup{}, fmt.Errorf("github oauth: user lookup: %w", sanitize(err))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return ports.IdentityLookup{}, fmt.Errorf("github oauth: user lookup returned status %d", resp.StatusCode)
	}
	var scopes *string
	if values, ok := resp.Header["X-Oauth-Scopes"]; ok && len(values) > 0 {
		s := strings.TrimSpace(values[0])
		scopes = &s
	}
	var body userResponse
	if err := decodeJSON(resp.Body, &body); err != nil {
		return ports.IdentityLookup{}, fmt.Errorf("github oauth: user response: %w", err)
	}
	id := body.ID.String()
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		id = "" // leave validation to the domain, which rejects it
	}
	return ports.IdentityLookup{Identity: domain.GitHubIdentity{UserID: id, Login: body.Login}, OAuthScopes: scopes}, nil
}

// Revoke deletes the single token (DELETE /applications/{client_id}/token). Best effort: the
// use case ignores the error (F-07, step 4.5).
func (a *OAuthAdapter) Revoke(ctx context.Context, accessToken string) error {
	payload, _ := json.Marshal(map[string]string{"access_token": accessToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		strings.TrimRight(a.cfg.APIBaseURL, "/")+"/applications/"+url.PathEscape(a.cfg.ClientID)+"/token", strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("github oauth: building revoke request: %w", err)
	}
	req.SetBasicAuth(a.cfg.ClientID, a.cfg.ClientSecret)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("github oauth: revoke: %w", sanitize(err))
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBody))
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("github oauth: revoke returned status %d", resp.StatusCode)
	}
	return nil
}

func decodeJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(io.LimitReader(r, maxBody))
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decoding: %w", err)
	}
	return nil
}

// sanitize strips request URLs from transport errors: GitHub URLs never carry secrets in this
// adapter, but url.Error strings would otherwise embed the full request target.
func sanitize(err error) error {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		return fmt.Errorf("%s request failed: %w", uerr.Op, unwrapNet(uerr.Err))
	}
	return err
}

func unwrapNet(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return errors.New("transport error")
}
