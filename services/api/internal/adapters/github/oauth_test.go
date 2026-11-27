package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// fakeGitHub is a local stand-in for github.com and api.github.com (F-07 preconditions: adapters
// are tested against local test servers, never real GitHub). It records what it received so that
// tests can assert on redirect_uri, scope, PKCE, and headers.
type fakeGitHub struct {
	mu            sync.Mutex
	exchangeForms []url.Values
	userAuth      []string
	revokes       []struct{ auth, body string }

	tokenStatus  int
	tokenBody    map[string]any
	userStatus   int
	userBody     map[string]any
	userScopes   *string
	revokeStatus int
}

// Synthetic credentials for the fake GitHub; they never touch a real service.
const fakeClientSecret = "fake-client-secret" //nolint:gosec // synthetic test value

func newFakeGitHub() *fakeGitHub {
	empty := ""
	return &fakeGitHub{
		tokenStatus:  http.StatusOK,
		tokenBody:    map[string]any{"access_token": "gho_fake", "scope": "", "token_type": "bearer"},
		userStatus:   http.StatusOK,
		userBody:     map[string]any{"id": 1234567, "login": "synthetic-contributor", "email": "public@example.org", "name": "Synthetic"},
		userScopes:   &empty,
		revokeStatus: http.StatusNoContent,
	}
}

func (f *fakeGitHub) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		f.exchangeForms = append(f.exchangeForms, r.PostForm)
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.tokenStatus)
		_ = json.NewEncoder(w).Encode(f.tokenBody)
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.userAuth = append(f.userAuth, r.Header.Get("Authorization"))
		f.mu.Unlock()
		if f.userScopes != nil {
			w.Header().Set("X-OAuth-Scopes", *f.userScopes)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.userStatus)
		_ = json.NewEncoder(w).Encode(f.userBody)
	})
	mux.HandleFunc("DELETE /applications/{client_id}/token", func(w http.ResponseWriter, r *http.Request) {
		var body strings.Builder
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		body.Write(buf[:n])
		f.mu.Lock()
		f.revokes = append(f.revokes, struct{ auth, body string }{r.Header.Get("Authorization"), body.String()})
		f.mu.Unlock()
		w.WriteHeader(f.revokeStatus)
	})
	return mux
}

func newAdapter(t *testing.T, f *fakeGitHub) *OAuthAdapter {
	t.Helper()
	srv := httptest.NewServer(f.handler())
	t.Cleanup(srv.Close)
	a, err := NewOAuthAdapter(OAuthConfig{
		ClientID: "fake-client-id", ClientSecret: fakeClientSecret,
		WebBaseURL: srv.URL, APIBaseURL: srv.URL, Timeout: 2 * time.Second,
	}, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestAuthorizationURL(t *testing.T) {
	a := newAdapter(t, newFakeGitHub())
	raw := a.AuthorizationURL(ports.AuthorizationRequest{State: "st", CodeChallenge: "ch", RedirectURI: "https://cla.example/api/auth/github/callback"})
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Path != "/login/oauth/authorize" || q.Get("client_id") != "fake-client-id" || q.Get("state") != "st" ||
		q.Get("code_challenge") != "ch" || q.Get("code_challenge_method") != "S256" ||
		q.Get("redirect_uri") != "https://cla.example/api/auth/github/callback" || q.Get("allow_signup") != "false" {
		t.Fatalf("authorization URL = %s", raw)
	}
	if _, present := q["scope"]; !present || q.Get("scope") != "" {
		t.Fatalf("scope must be present and empty: %s", raw)
	}
}

// F07-T3 and F07-T13 (adapter part): the exchange sends redirect_uri, the verifier, and the
// client secret; the scope field is passed through as a pointer.
func TestExchange(t *testing.T) {
	f := newFakeGitHub()
	a := newAdapter(t, f)
	out, err := a.Exchange(context.Background(), "code-1", "verifier-1", "https://cla.example/api/auth/github/callback")
	if err != nil {
		t.Fatal(err)
	}
	if out.AccessToken != "gho_fake" || out.GrantedScope == nil || *out.GrantedScope != "" {
		t.Fatalf("exchange = %+v", out)
	}
	form := f.exchangeForms[0]
	if form.Get("code") != "code-1" || form.Get("code_verifier") != "verifier-1" ||
		form.Get("redirect_uri") != "https://cla.example/api/auth/github/callback" ||
		form.Get("client_id") != "fake-client-id" || form.Get("client_secret") != "fake-client-secret" {
		t.Fatalf("exchange form = %v", form)
	}
}

func TestExchangeScopeVariants(t *testing.T) {
	cases := []struct {
		name  string
		body  map[string]any
		scope *string
		err   bool
	}{
		{"non-empty scope", map[string]any{"access_token": "gho_fake", "scope": "repo"}, strPtr("repo"), false},
		{"absent scope", map[string]any{"access_token": "gho_fake"}, nil, false},
		{"refresh token passed through", map[string]any{"access_token": "gho_fake", "scope": "", "refresh_token": "ghr_fake"}, strPtr(""), false},
		{"oauth error with 200", map[string]any{"error": "bad_verification_code"}, nil, true},
		{"no access token", map[string]any{"scope": ""}, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeGitHub()
			f.tokenBody = c.body
			a := newAdapter(t, f)
			out, err := a.Exchange(context.Background(), "c", "v", "r")
			if (err != nil) != c.err {
				t.Fatalf("err = %v", err)
			}
			if err != nil {
				return
			}
			if (out.GrantedScope == nil) != (c.scope == nil) || (c.scope != nil && *out.GrantedScope != *c.scope) {
				t.Fatalf("scope = %v, want %v", out.GrantedScope, c.scope)
			}
			if c.name == "refresh token passed through" && out.RefreshToken != "ghr_fake" {
				t.Fatal("refresh token must be visible to the use case so that it can be discarded")
			}
		})
	}
}

func TestExchangeUpstreamFailures(t *testing.T) {
	f := newFakeGitHub()
	f.tokenStatus = http.StatusBadGateway
	a := newAdapter(t, f)
	if _, err := a.Exchange(context.Background(), "c", "v", "r"); err == nil {
		t.Fatal("non-200 must fail")
	}
	// Timeout: a server that never answers until the test releases it.
	release := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(func() { close(release); slow.Close() })
	a2, _ := NewOAuthAdapter(OAuthConfig{ClientID: "id", ClientSecret: fakeClientSecret, WebBaseURL: slow.URL, APIBaseURL: slow.URL, Timeout: 50 * time.Millisecond}, slow.Client())
	_, err := a2.Exchange(context.Background(), "c", "v", "r")
	if err == nil || strings.Contains(err.Error(), slow.URL) {
		t.Fatalf("timeout must fail without embedding the URL: %v", err)
	}
}

// F07-T14 (adapter part): only id and login are extracted; the public email and other fields
// are never surfaced; the X-OAuth-Scopes header is passed through as a pointer.
func TestLookupIdentity(t *testing.T) {
	f := newFakeGitHub()
	a := newAdapter(t, f)
	out, err := a.LookupIdentity(context.Background(), "gho_fake")
	if err != nil {
		t.Fatal(err)
	}
	if out.Identity.UserID != "1234567" || out.Identity.Login != "synthetic-contributor" {
		t.Fatalf("identity = %+v", out.Identity)
	}
	if out.OAuthScopes == nil || *out.OAuthScopes != "" {
		t.Fatalf("scopes = %v", out.OAuthScopes)
	}
	if f.userAuth[0] != "Bearer gho_fake" {
		t.Fatalf("authorization header = %q", f.userAuth[0])
	}
	// The lookup result type has no room for an email; this guards against future widening.
	if strings.Contains(strings.ToLower(jsonOf(out)), "example.org") {
		t.Fatal("public email leaked through the lookup result")
	}
}

func TestLookupIdentityVariants(t *testing.T) {
	cases := []struct {
		name   string
		body   map[string]any
		scopes *string
		status int
		wantID string
		err    bool
	}{
		{"scopes header absent", map[string]any{"id": 1, "login": "a"}, nil, 200, "1", false},
		{"scopes header non-empty", map[string]any{"id": 1, "login": "a"}, strPtr("user:email"), 200, "1", false},
		{"missing id", map[string]any{"login": "a"}, strPtr(""), 200, "", false},
		{"string id is a decoding failure", map[string]any{"id": "12ab", "login": "a"}, strPtr(""), 200, "", true},
		{"negative id is left to the domain", map[string]any{"id": -5, "login": "a"}, strPtr(""), 200, "", false},
		{"large id keeps precision", map[string]any{"id": json.Number("9007199254740993"), "login": "a"}, strPtr(""), 200, "9007199254740993", false},
		{"401", map[string]any{"message": "Bad credentials"}, strPtr(""), 401, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeGitHub()
			f.userBody, f.userScopes, f.userStatus = c.body, c.scopes, c.status
			a := newAdapter(t, f)
			out, err := a.LookupIdentity(context.Background(), "gho_fake")
			if (err != nil) != c.err {
				t.Fatalf("err = %v", err)
			}
			if err != nil {
				return
			}
			if out.Identity.UserID != c.wantID {
				t.Fatalf("id = %q, want %q", out.Identity.UserID, c.wantID)
			}
			if (out.OAuthScopes == nil) != (c.scopes == nil) || (c.scopes != nil && *out.OAuthScopes != *c.scopes) {
				t.Fatalf("scopes = %v, want %v", out.OAuthScopes, c.scopes)
			}
		})
	}
}

func TestRevoke(t *testing.T) {
	f := newFakeGitHub()
	a := newAdapter(t, f)
	if err := a.Revoke(context.Background(), "gho_fake"); err != nil {
		t.Fatal(err)
	}
	rev := f.revokes[0]
	if !strings.HasPrefix(rev.auth, "Basic ") || !strings.Contains(rev.body, `"access_token":"gho_fake"`) {
		t.Fatalf("revoke = %+v", rev)
	}
	f.revokeStatus = http.StatusUnprocessableEntity
	if err := a.Revoke(context.Background(), "gho_fake"); err == nil {
		t.Fatal("non-204 must be reported so that the use case can log it")
	}
}

func TestNewOAuthAdapterValidation(t *testing.T) {
	if _, err := NewOAuthAdapter(OAuthConfig{}, nil); err == nil {
		t.Fatal("missing credentials must be rejected")
	}
	if _, err := NewOAuthAdapter(OAuthConfig{ClientID: "a", ClientSecret: fakeClientSecret, WebBaseURL: "github.com", APIBaseURL: "https://api.github.com"}, nil); err == nil {
		t.Fatal("a base URL without a scheme must be rejected")
	}
}

func TestAdapterDoesNotFollowRedirects(t *testing.T) {
	redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.example/", http.StatusFound)
	}))
	t.Cleanup(redirecting.Close)
	a, _ := NewOAuthAdapter(OAuthConfig{ClientID: "id", ClientSecret: fakeClientSecret, WebBaseURL: redirecting.URL, APIBaseURL: redirecting.URL}, redirecting.Client())
	if _, err := a.Exchange(context.Background(), "c", "v", "r"); err == nil {
		t.Fatal("a redirecting token endpoint must fail, not be followed")
	}
}

func strPtr(s string) *string { return &s }

func jsonOf(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
