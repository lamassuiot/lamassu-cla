package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github/githubtest"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/memory"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

const (
	portal   = "https://cla.example"
	callback = portal + "/api/auth/github/callback"
	// Synthetic credential for the fake GitHub; it never touches a real service.
	fakeClientSecret = "fake-client-secret" //nolint:gosec // synthetic test value
)

// reqCookie builds a request cookie; the attributes are irrelevant on requests but keep gosec
// from flagging the test as an insecure cookie.
func reqCookie(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
}

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

// stack is the whole API wired with in-memory stores and the fake GitHub.
type stack struct {
	srv          *httptest.Server
	gh           *githubtest.Fake
	clock        *fixedClock
	attempts     *memory.SignInAttempts
	sessions     *memory.Sessions
	contributors *memory.Contributors
	security     *memory.SecurityLog
	client       *http.Client
}

func newStack(t *testing.T, enabled bool) *stack {
	t.Helper()
	gh := githubtest.New()
	t.Cleanup(gh.Close)
	identity, err := github.NewOAuthAdapter(github.OAuthConfig{
		ClientID: "fake-client-id", ClientSecret: fakeClientSecret, WebBaseURL: gh.URL(), APIBaseURL: gh.URL(),
	}, gh.Client())
	if err != nil {
		t.Fatal(err)
	}
	s := &stack{
		gh: gh, clock: &fixedClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)},
		attempts: memory.NewSignInAttempts(), sessions: memory.NewSessions(),
		contributors: memory.NewContributors(), security: &memory.SecurityLog{},
	}
	auth, err := application.NewAuth(application.AuthConfig{
		Enabled: enabled, CallbackURL: callback,
		ReturnTo: domain.ReturnToPolicy{Allowed: []string{"/", "/agreements/"}},
		Session:  domain.DefaultSessionPolicy,
	}, application.AuthDependencies{
		Clock: s.clock, Random: memory.Random{}, Identity: identity, Attempts: s.attempts,
		Sessions: s.sessions, Contributors: s.contributors, Security: s.security,
	})
	if err != nil {
		t.Fatal(err)
	}
	h, err := New(Config{PortalOrigin: portal, SignInErrorRoute: "/sign-in/error"}, auth, s.clock)
	if err != nil {
		t.Fatal(err)
	}
	s.srv = httptest.NewServer(h)
	t.Cleanup(s.srv.Close)
	s.client = s.srv.Client()
	s.client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return s
}

// reply is a fully read response; the body is already closed.
type reply struct {
	StatusCode int
	Header     http.Header
	Body       io.Reader
	cookies    []*http.Cookie
}

func (r *reply) Cookies() []*http.Cookie { return r.cookies }

func (s *stack) do(t *testing.T, method, path string, headers map[string]string, cookies ...*http.Cookie) *reply {
	t.Helper()
	req, err := http.NewRequest(method, s.srv.URL+path, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return &reply{StatusCode: resp.StatusCode, Header: resp.Header, Body: bytes.NewReader(body), cookies: resp.Cookies()}
}

func cookieNamed(resp *reply, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func problemOf(t *testing.T, resp *reply) map[string]any {
	t.Helper()
	if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content type = %q", ct)
	}
	var p map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatal(err)
	}
	return p
}

// signIn runs start and a successful callback and returns the session cookie and CSRF token.
func (s *stack) signIn(t *testing.T, returnTo string) (*http.Cookie, string) {
	t.Helper()
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start?return_to="+url.QueryEscape(returnTo), nil)
	if start.StatusCode != http.StatusFound {
		t.Fatalf("start status = %d", start.StatusCode)
	}
	oauth := cookieNamed(start, "__Host-cla_oauth")
	if oauth == nil {
		t.Fatal("start must set the pre-authentication cookie")
	}
	loc, _ := url.Parse(start.Header.Get("Location"))
	state := loc.Query().Get("state")
	cb := s.do(t, http.MethodGet, "/api/auth/github/callback?code=code-1&state="+url.QueryEscape(state), nil, oauth)
	if cb.StatusCode != http.StatusFound {
		t.Fatalf("callback status = %d", cb.StatusCode)
	}
	session := cookieNamed(cb, "__Host-cla_session")
	if session == nil || session.Value == "" {
		t.Fatalf("callback must set the session cookie; location=%s", cb.Header.Get("Location"))
	}
	me := s.do(t, http.MethodGet, "/api/v1/session", nil, session)
	var body struct {
		CSRFToken string `json:"csrf_token"`
	}
	_ = json.NewDecoder(me.Body).Decode(&body)
	return session, body.CSRFToken
}

func TestHealthUnderAPIBase(t *testing.T) {
	s := newStack(t, true)
	resp := s.do(t, http.MethodGet, "/api/v1/health", nil)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("X-Request-Id") == "" {
		t.Fatalf("health: %d %v", resp.StatusCode, resp.Header)
	}
	if s.do(t, http.MethodGet, "/v1/health", nil).StatusCode != http.StatusNotFound {
		t.Fatal("operations must live under /api only")
	}
}

// F07-T1 and F07-T3 (handler part): redirect parameters and cookie attributes (AC-07-1, AC-07-3).
func TestStartAndCallbackCookies(t *testing.T) {
	s := newStack(t, true)
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start?return_to=%2Fagreements%2Ficla", nil)
	if start.StatusCode != http.StatusFound || start.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("start: %d %v", start.StatusCode, start.Header)
	}
	loc, _ := url.Parse(start.Header.Get("Location"))
	q := loc.Query()
	if !strings.HasPrefix(start.Header.Get("Location"), s.gh.URL()+"/login/oauth/authorize?") ||
		q.Get("redirect_uri") != callback || q.Get("code_challenge_method") != "S256" || q.Get("scope") != "" {
		t.Fatalf("location = %s", start.Header.Get("Location"))
	}
	oauth := cookieNamed(start, "__Host-cla_oauth")
	if oauth.Value != q.Get("state") || !oauth.HttpOnly || !oauth.Secure || oauth.Path != "/" ||
		oauth.SameSite != http.SameSiteLaxMode || oauth.MaxAge != 600 || oauth.Domain != "" {
		t.Fatalf("oauth cookie = %+v", oauth)
	}

	cb := s.do(t, http.MethodGet, "/api/auth/github/callback?code=code-1&state="+url.QueryEscape(oauth.Value), nil, oauth)
	if cb.StatusCode != http.StatusFound || cb.Header.Get("Location") != portal+"/agreements/icla" {
		t.Fatalf("callback: %d %s", cb.StatusCode, cb.Header.Get("Location"))
	}
	session := cookieNamed(cb, "__Host-cla_session")
	if session == nil || !session.HttpOnly || !session.Secure || session.Path != "/" ||
		session.SameSite != http.SameSiteLaxMode || session.MaxAge != 8*60*60 || session.Domain != "" {
		t.Fatalf("session cookie = %+v", session)
	}
	if cleared := cookieNamed(cb, "__Host-cla_oauth"); cleared == nil || cleared.MaxAge >= 0 {
		t.Fatal("the pre-authentication cookie must be cleared on success")
	}
	if s.sessions.Len() != 1 || s.contributors.Len() != 1 || len(s.sessions.Audit()) != 1 {
		t.Fatal("session, contributor, and audit event expected")
	}
	// The exchange carried redirect_uri and the verifier; the token was revoked.
	form := s.gh.ExchangeForms[0]
	if form.Get("redirect_uri") != callback || form.Get("code_verifier") == "" {
		t.Fatalf("exchange form = %v", form)
	}
	if len(s.gh.Revocations) != 1 {
		t.Fatal("best-effort revocation expected")
	}
}

// F07-T17 and F07-T2 (handler part): malformed and unbound callbacks redirect with a reason and
// set no session; the binding mismatch leaves the newer attempt's cookie alone (AC-07-2).
func TestCallbackRejections(t *testing.T) {
	s := newStack(t, true)
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil)
	oauth := cookieNamed(start, "__Host-cla_oauth")
	state := url.QueryEscape(oauth.Value)
	cases := []struct {
		name, query string
		cookie      *http.Cookie
		reason      string
		cookieKept  bool
	}{
		{"no state", "code=c", oauth, "malformed", true},
		{"duplicate code", "code=a&code=b&state=" + state, oauth, "malformed", true},
		{"code and error", "code=c&error=access_denied&state=" + state, oauth, "malformed", true},
		{"no cookie", "code=c&state=" + state, nil, "attempt_replaced", true},
		{"other cookie", "code=c&state=" + state, reqCookie("__Host-cla_oauth", "other"), "attempt_replaced", true},
		{"unknown state", "code=c&state=unknown", reqCookie("__Host-cla_oauth", "unknown"), "state_invalid", false},
		{"declined", "error=access_denied&state=" + state, oauth, "cancelled", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var resp *reply
			if c.cookie != nil {
				resp = s.do(t, http.MethodGet, "/api/auth/github/callback?"+c.query, nil, c.cookie)
			} else {
				resp = s.do(t, http.MethodGet, "/api/auth/github/callback?"+c.query, nil)
			}
			if resp.StatusCode != http.StatusFound {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			loc, _ := url.Parse(resp.Header.Get("Location"))
			if loc.Scheme+"://"+loc.Host != portal || loc.Path != "/sign-in/error" || loc.Query().Get("reason") != c.reason {
				t.Fatalf("location = %s", resp.Header.Get("Location"))
			}
			if cookieNamed(resp, "__Host-cla_session") != nil {
				t.Fatal("no session cookie may be set")
			}
			cleared := cookieNamed(resp, "__Host-cla_oauth")
			if c.cookieKept && cleared != nil {
				t.Fatal("the pre-authentication cookie must not be cleared before the attempt is consumed")
			}
			if !c.cookieKept && (cleared == nil || cleared.MaxAge >= 0) {
				t.Fatal("the pre-authentication cookie must be cleared once the attempt is consumed")
			}
		})
	}
	if s.sessions.Len() != 0 {
		t.Fatal("no session may exist")
	}
}

// F07-T13 (handler part): a non-empty granted scope redirects with scope_unexpected (AC-07-13).
func TestCallbackScopeUnexpected(t *testing.T) {
	s := newStack(t, true)
	s.gh.Mu.Lock()
	s.gh.UserScopes = githubtest.String("user:email")
	s.gh.Mu.Unlock()
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil)
	oauth := cookieNamed(start, "__Host-cla_oauth")
	cb := s.do(t, http.MethodGet, "/api/auth/github/callback?code=c&state="+url.QueryEscape(oauth.Value), nil, oauth)
	loc, _ := url.Parse(cb.Header.Get("Location"))
	if loc.Query().Get("reason") != "scope_unexpected" || cookieNamed(cb, "__Host-cla_session") != nil {
		t.Fatalf("location = %s", cb.Header.Get("Location"))
	}
	if len(s.gh.Revocations) != 1 {
		t.Fatal("the token must still be revoked")
	}
}

// F07-T15: the flag (AC-07-15).
func TestSignInDisabled(t *testing.T) {
	s := newStack(t, false)
	resp := s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil)
	if resp.StatusCode != http.StatusServiceUnavailable || problemOf(t, resp)["code"] != "SIGN_IN_DISABLED" {
		t.Fatalf("start: %d", resp.StatusCode)
	}
	if len(s.gh.ExchangeForms) != 0 || s.attempts.Len() != 0 {
		t.Fatal("disabled start must not store or contact GitHub")
	}
	cb := s.do(t, http.MethodGet, "/api/auth/github/callback?code=c&state=s", nil, reqCookie("__Host-cla_oauth", "s"))
	loc, _ := url.Parse(cb.Header.Get("Location"))
	if cb.StatusCode != http.StatusFound || loc.Query().Get("reason") != "disabled" {
		t.Fatalf("callback: %d %s", cb.StatusCode, cb.Header.Get("Location"))
	}
}

// GET /session: shape, no-store, 401 with cookie expiry, and no idle extension (AC-07-6).
func TestGetSession(t *testing.T) {
	s := newStack(t, true)
	session, csrf := s.signIn(t, "/")
	resp := s.do(t, http.MethodGet, "/api/v1/session", nil, session)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("session: %d %v", resp.StatusCode, resp.Header)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["github_user_id"] != "1234567" || body["github_login"] != "synthetic-contributor" || body["csrf_token"] != csrf ||
		body["expires_at"] != "2026-01-01T20:00:00Z" {
		t.Fatalf("body = %v", body)
	}
	if _, leaked := body["session_token"]; leaked || strings.Contains(mustJSON(body), session.Value) {
		t.Fatal("the session token must never be in the body")
	}

	unauth := s.do(t, http.MethodGet, "/api/v1/session", nil, reqCookie("__Host-cla_session", "unknown"))
	if unauth.StatusCode != http.StatusUnauthorized || problemOf(t, unauth)["code"] != "SESSION_REQUIRED" {
		t.Fatalf("unknown session: %d", unauth.StatusCode)
	}
	if c := cookieNamed(unauth, "__Host-cla_session"); c == nil || c.MaxAge >= 0 {
		t.Fatal("a stale session cookie must be expired on 401")
	}
	// Reading never extends: after 31 minutes of only reads the session is idle-expired.
	s.clock.now = s.clock.now.Add(20 * time.Minute)
	if s.do(t, http.MethodGet, "/api/v1/session", nil, session).StatusCode != http.StatusOK {
		t.Fatal("still valid at 20 minutes")
	}
	s.clock.now = s.clock.now.Add(11 * time.Minute)
	if s.do(t, http.MethodGet, "/api/v1/session", nil, session).StatusCode != http.StatusUnauthorized {
		t.Fatal("reads must not have extended the idle expiry")
	}
}

// F07-T7: Origin and CSRF on a state-changing route (AC-07-7).
func TestSignOutOriginAndCSRF(t *testing.T) {
	s := newStack(t, true)
	session, csrf := s.signIn(t, "/")
	cases := []struct {
		name    string
		headers map[string]string
		status  int
		code    string
	}{
		{"missing origin", map[string]string{"X-CSRF-Token": csrf}, 403, "ORIGIN_NOT_ALLOWED"},
		{"null origin", map[string]string{"Origin": "null", "X-CSRF-Token": csrf}, 403, "ORIGIN_NOT_ALLOWED"},
		{"other origin", map[string]string{"Origin": "https://evil.example", "X-CSRF-Token": csrf}, 403, "ORIGIN_NOT_ALLOWED"},
		{"missing csrf", map[string]string{"Origin": portal}, 403, "CSRF_TOKEN_INVALID"},
		{"wrong csrf", map[string]string{"Origin": portal, "X-CSRF-Token": "wrong"}, 403, "CSRF_TOKEN_INVALID"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := s.do(t, http.MethodPost, "/api/v1/session/sign-out", c.headers, session)
			if resp.StatusCode != c.status || problemOf(t, resp)["code"] != c.code {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			if s.sessions.Len() != 1 {
				t.Fatal("a rejected request must change nothing")
			}
		})
	}
	// Safe methods are not subject to the Origin check.
	if s.do(t, http.MethodGet, "/api/v1/session", nil, session).StatusCode != http.StatusOK {
		t.Fatal("GET without Origin must pass")
	}
}

// F07-T8 (handler part): sign-out outcomes (AC-07-8).
func TestSignOut(t *testing.T) {
	s := newStack(t, true)
	session, csrf := s.signIn(t, "/")
	ok := s.do(t, http.MethodPost, "/api/v1/session/sign-out", map[string]string{"Origin": portal, "X-CSRF-Token": csrf}, session)
	if ok.StatusCode != http.StatusNoContent || s.sessions.Len() != 0 || len(s.sessions.Audit()) != 2 {
		t.Fatalf("sign-out: %d sessions=%d audit=%d", ok.StatusCode, s.sessions.Len(), len(s.sessions.Audit()))
	}
	if c := cookieNamed(ok, "__Host-cla_session"); c == nil || c.MaxAge >= 0 {
		t.Fatal("sign-out must expire the cookie")
	}
	if s.do(t, http.MethodGet, "/api/v1/session", nil, session).StatusCode != http.StatusUnauthorized {
		t.Fatal("the old token must no longer authenticate")
	}
	// Idempotent without a session; Origin still enforced.
	if s.do(t, http.MethodPost, "/api/v1/session/sign-out", map[string]string{"Origin": portal}).StatusCode != http.StatusNoContent {
		t.Fatal("sign-out without a session must be 204")
	}
	if s.do(t, http.MethodPost, "/api/v1/session/sign-out", nil).StatusCode != http.StatusForbidden {
		t.Fatal("sign-out without Origin must be 403 even without a session")
	}
	// Confirmed failure: 503, session preserved, no event.
	session, csrf = s.signIn(t, "/")
	s.sessions.Fail = ports.ErrUnavailable
	fail := s.do(t, http.MethodPost, "/api/v1/session/sign-out", map[string]string{"Origin": portal, "X-CSRF-Token": csrf}, session)
	if fail.StatusCode != http.StatusServiceUnavailable || problemOf(t, fail)["code"] != "STORE_UNAVAILABLE" || fail.Header.Get("Retry-After") == "" {
		t.Fatalf("failed deletion: %d", fail.StatusCode)
	}
	if cookieNamed(fail, "__Host-cla_session") != nil || s.sessions.Len() != 1 {
		t.Fatal("a failed deletion must not expire the cookie or drop the session")
	}
}

// F07-T16: store outages are 503, never 401 (AC-07-16).
func TestStoreUnavailable(t *testing.T) {
	s := newStack(t, true)
	session, _ := s.signIn(t, "/")
	s.sessions.FailReads = ports.ErrUnavailable
	resp := s.do(t, http.MethodGet, "/api/v1/session", nil, session)
	if resp.StatusCode != http.StatusServiceUnavailable || problemOf(t, resp)["code"] != "STORE_UNAVAILABLE" {
		t.Fatalf("session read during outage: %d", resp.StatusCode)
	}
	if cookieNamed(resp, "__Host-cla_session") != nil {
		t.Fatal("an outage must not expire the cookie")
	}
	s.sessions.FailReads = nil
	s.attempts.Fail = ports.ErrUnavailable
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil)
	if start.StatusCode != http.StatusServiceUnavailable || problemOf(t, start)["code"] != "STORE_UNAVAILABLE" {
		t.Fatalf("start during outage: %d", start.StatusCode)
	}
}

// AC-07-9 at the HTTP level: a signed-in browser's new sign-in replaces the session.
func TestSessionReplacement(t *testing.T) {
	s := newStack(t, true)
	first, _ := s.signIn(t, "/")
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil, first)
	oauth := cookieNamed(start, "__Host-cla_oauth")
	if s.sessions.Len() != 1 {
		t.Fatal("starting sign-in must not touch the current session")
	}
	cb := s.do(t, http.MethodGet, "/api/auth/github/callback?code=c&state="+url.QueryEscape(oauth.Value), nil, oauth, first)
	second := cookieNamed(cb, "__Host-cla_session")
	if second == nil || second.Value == first.Value || s.sessions.Len() != 1 {
		t.Fatal("success must replace the old session with a new token")
	}
	if s.do(t, http.MethodGet, "/api/v1/session", nil, first).StatusCode != http.StatusUnauthorized {
		t.Fatal("the old token must stop working")
	}
}

// F07-T5 (HTTP part): no secret in security logs or audit events.
func TestNoSecretsInLogs(t *testing.T) {
	s := newStack(t, true)
	session, csrf := s.signIn(t, "/")
	var b strings.Builder
	for _, e := range s.security.Events() {
		b.WriteString(e.Name + e.Outcome + e.RequestID)
	}
	for _, e := range s.sessions.Audit() {
		b.WriteString(string(e.Type) + e.ActorID + e.RequestID)
	}
	state := cookieNamed(s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil), "__Host-cla_oauth").Value
	for name, secret := range map[string]string{"session": session.Value, "csrf": csrf, "state": state, "token": "gho_fake", "code": "code-1"} {
		if strings.Contains(b.String(), secret) {
			t.Fatalf("%s leaked into logs or audit", name)
		}
	}
}

func TestRequestIDEcho(t *testing.T) {
	s := newStack(t, true)
	resp := s.do(t, http.MethodGet, "/api/v1/health", map[string]string{"X-Request-Id": "01J9Z3K4Q8X7M2N5P6R7S8T9V0"})
	if resp.Header.Get("X-Request-Id") != "01J9Z3K4Q8X7M2N5P6R7S8T9V0" {
		t.Fatal("a valid client request ID must be echoed")
	}
	bad := s.do(t, http.MethodGet, "/api/v1/health", map[string]string{"X-Request-Id": "<script>"})
	if id := bad.Header.Get("X-Request-Id"); id == "<script>" || id == "" {
		t.Fatalf("an invalid request ID must be replaced, got %q", id)
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
