package application

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// F07-T1: start builds the GitHub redirect, stores the hashed state with the verifier and
// return_to, and leaves sessions alone (AC-07-1).
func TestStartSignIn(t *testing.T) {
	h := newHarness(true)
	h.sessions.items["existing"] = domain.Session{TokenHash: "existing"}

	out, err := h.start("/agreements/icla")
	if err != nil {
		t.Fatalf("StartSignIn: %v", err)
	}
	u, err := url.Parse(out.RedirectURL)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("state") != out.State || q.Get("redirect_uri") != testCallback || q.Get("code_challenge_method") != "S256" {
		t.Fatalf("redirect query = %v", q)
	}
	if _, present := q["scope"]; !present || q.Get("scope") != "" {
		t.Fatalf("scope must be present and explicitly empty: %v", q)
	}
	stored, ok := h.attempts.items[domain.HashToken(out.State)]
	if !ok {
		t.Fatal("attempt must be stored under the hash of state")
	}
	if domain.PKCEChallenge(stored.Verifier) != q.Get("code_challenge") {
		t.Fatal("stored verifier must match the challenge sent to GitHub")
	}
	if stored.ReturnTo != "/agreements/icla" || !stored.ExpiresAt.Equal(h.clock.now.Add(10*time.Minute)) {
		t.Fatalf("stored attempt = %+v", stored)
	}
	if _, still := h.sessions.items["existing"]; !still {
		t.Fatal("starting sign-in must not touch an existing session")
	}
	if strings.Contains(h.everythingObservable(), out.State) || strings.Contains(h.everythingObservable(), stored.Verifier) {
		t.Fatal("state or verifier leaked into logs or audit")
	}
}

func TestStartSignInSupersedesPreviousAttempt(t *testing.T) {
	h := newHarness(true)
	first, _ := h.start("/")
	second, err := h.auth.StartSignIn(context.Background(), StartSignInInput{PreviousState: first.State})
	if err != nil {
		t.Fatal(err)
	}
	if _, old := h.attempts.items[domain.HashToken(first.State)]; old {
		t.Fatal("the previous attempt must be deleted")
	}
	if _, cur := h.attempts.items[domain.HashToken(second.State)]; !cur {
		t.Fatal("the new attempt must exist")
	}
}

func TestStartSignInRejectedReturnToFallsBack(t *testing.T) {
	h := newHarness(true)
	out, _ := h.start("//evil.example/")
	if h.attempts.items[domain.HashToken(out.State)].ReturnTo != "/" {
		t.Fatal("rejected return_to must fall back to /")
	}
}

// F07-T15: the sign-in flag (AC-07-15).
func TestSignInDisabled(t *testing.T) {
	h := newHarness(false)
	if _, err := h.start("/"); !errors.Is(err, ErrSignInDisabled) {
		t.Fatalf("start: %v", err)
	}
	if len(h.attempts.items) != 0 || len(h.identity.authRequests) != 0 {
		t.Fatal("disabled start must not store or contact GitHub")
	}
	if _, err := h.callback("s", "s", ""); !errors.Is(err, ErrSignInDisabled) {
		t.Fatalf("callback: %v", err)
	}
}

// F07-T3 (use-case part): successful callback creates the contributor, a hashed session, and
// the audit event, and redirects to return_to (AC-07-3).
func TestCompleteSignInSuccess(t *testing.T) {
	h := newHarness(true)
	start, _ := h.start("/agreements/")
	out, err := h.callback(start.State, start.State, "")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if out.ReturnTo != "/agreements/" || out.SessionToken == "" {
		t.Fatalf("out = %+v", out)
	}
	sess, ok := h.sessions.items[domain.HashToken(out.SessionToken)]
	if !ok {
		t.Fatal("session must be stored under the hash of the token")
	}
	if sess.GitHubUserID != "1234567" || sess.GitHubLogin != "synthetic-contributor" || len(sess.CSRFToken) != 43 {
		t.Fatalf("session = %+v", sess)
	}
	if c := h.contributors.items["1234567"]; c.Status != domain.ContributorPending || c.GitHubLogin != "synthetic-contributor" {
		t.Fatalf("contributor = %+v", c)
	}
	if len(h.sessions.audit) != 1 || h.sessions.audit[0].Type != domain.AuditSignedIn || h.sessions.audit[0].ActorID != "1234567" {
		t.Fatalf("audit = %+v", h.sessions.audit)
	}
	if len(h.attempts.items) != 0 {
		t.Fatal("the attempt must be consumed")
	}
	ex := h.identity.exchanges[0]
	if ex.code != "code-1" || ex.redirect != testCallback || domain.PKCEChallenge(ex.verifier) == "" {
		t.Fatalf("exchange = %+v", ex)
	}
	if len(h.identity.revoked) != 1 || h.identity.revoked[0] != fakeAccessToken {
		t.Fatal("the access token must be revoked best-effort")
	}
}

// F07-T2 and F07-T17: malformed input, binding mismatch, and invalid state, on success and error
// callbacks; nothing is consumed on a binding mismatch (AC-07-2).
func TestCompleteSignInRejections(t *testing.T) {
	type tc struct {
		name    string
		in      func(state string) CompleteSignInInput
		wantErr error
		kept    bool // the attempt must still exist afterwards
	}
	cases := []tc{
		{"missing state", func(s string) CompleteSignInInput { return CompleteSignInInput{Code: "c", CookieState: s} }, ErrMalformedRequest, true},
		{"code and error", func(s string) CompleteSignInInput {
			return CompleteSignInInput{Code: "c", Error: "access_denied", State: s, CookieState: s}
		}, ErrMalformedRequest, true},
		{"neither code nor error", func(s string) CompleteSignInInput { return CompleteSignInInput{State: s, CookieState: s} }, ErrMalformedRequest, true},
		{"duplicate parameters", func(s string) CompleteSignInInput {
			return CompleteSignInInput{Code: "c", State: s, CookieState: s, Malformed: true}
		}, ErrMalformedRequest, true},
		{"missing cookie", func(s string) CompleteSignInInput { return CompleteSignInInput{Code: "c", State: s} }, ErrBindingMismatch, true},
		{"mismatched cookie", func(s string) CompleteSignInInput {
			return CompleteSignInInput{Code: "c", State: s, CookieState: "other"}
		}, ErrBindingMismatch, true},
		{"mismatched cookie on error callback", func(s string) CompleteSignInInput {
			return CompleteSignInInput{Error: "access_denied", State: s, CookieState: "other"}
		}, ErrBindingMismatch, true},
		{"unknown state", func(_ string) CompleteSignInInput {
			return CompleteSignInInput{Code: "c", State: "unknown", CookieState: "unknown"}
		}, ErrStateInvalid, true},
		{"unknown state on error callback", func(_ string) CompleteSignInInput {
			return CompleteSignInInput{Error: "access_denied", State: "unknown", CookieState: "unknown"}
		}, ErrStateInvalid, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(true)
			start, _ := h.start("/")
			_, err := h.auth.CompleteSignIn(context.Background(), c.in(start.State))
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("err = %v, want %v", err, c.wantErr)
			}
			if _, kept := h.attempts.items[domain.HashToken(start.State)]; kept != c.kept {
				t.Fatalf("attempt kept = %v, want %v", kept, c.kept)
			}
			if len(h.sessions.items) != 0 || len(h.sessions.audit) != 0 || len(h.identity.exchanges) != 0 {
				t.Fatal("no exchange, session, or audit event may result")
			}
		})
	}
}

func TestCompleteSignInExpiredAndReusedState(t *testing.T) {
	h := newHarness(true)
	start, _ := h.start("/")
	h.clock.now = h.clock.now.Add(10 * time.Minute)
	if _, err := h.callback(start.State, start.State, ""); !errors.Is(err, ErrStateInvalid) {
		t.Fatalf("expired state: %v", err)
	}
	h = newHarness(true)
	start, _ = h.start("/")
	if _, err := h.callback(start.State, start.State, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := h.callback(start.State, start.State, ""); !errors.Is(err, ErrStateInvalid) {
		t.Fatalf("reused state: %v", err)
	}
	if len(h.sessions.items) != 1 || h.attempts.consumed != 1 {
		t.Fatal("exactly one sign-in must have completed")
	}
}

// F07-T2 (concurrency): two concurrent callbacks with the same state complete at most one sign-in.
func TestCompleteSignInConcurrentCallbacks(t *testing.T) {
	h := newHarness(true)
	start, _ := h.start("/")
	const n = 16
	var wg sync.WaitGroup
	results := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = h.callback(start.State, start.State, "")
		}(i)
	}
	wg.Wait()
	successes := 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrStateInvalid):
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || h.attempts.consumed != 1 || len(h.sessions.items) != 1 {
		t.Fatalf("successes=%d consumed=%d sessions=%d", successes, h.attempts.consumed, len(h.sessions.items))
	}
}

func TestCompleteSignInDeclined(t *testing.T) {
	h := newHarness(true)
	h.sessions.items["old"] = domain.Session{TokenHash: "old"}
	start, _ := h.start("/agreements/")
	out, err := h.auth.CompleteSignIn(context.Background(), CompleteSignInInput{
		Error: "access_denied", State: start.State, CookieState: start.State, CurrentSessionToken: "old-token",
	})
	if !errors.Is(err, ErrDenied) || out.ReturnTo != "/agreements/" {
		t.Fatalf("declined: err=%v out=%+v", err, out)
	}
	if len(h.attempts.items) != 0 {
		t.Fatal("state must be consumed on a declined callback")
	}
	if _, kept := h.sessions.items["old"]; !kept || len(h.sessions.audit) != 0 {
		t.Fatal("cancellation must preserve the existing session and write no audit event")
	}
}

// F07-T13 and F07-T14 (use-case part): scope verification in order, identity validation, and
// token hygiene (AC-07-13, AC-07-14).
func TestCompleteSignInScopeAndIdentityChecks(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(h *harness)
		wantErr error
		lookups bool
	}{
		{"token scope not empty", func(h *harness) { h.identity.exchangeScope = scope("repo") }, ErrScopeUnexpected, false},
		{"token scope absent", func(h *harness) { h.identity.exchangeScope = nil }, ErrScopeUnexpected, false},
		{"header scope not empty", func(h *harness) { h.identity.lookupScopes = scope("user:email") }, ErrScopeUnexpected, true},
		{"header scope absent", func(h *harness) { h.identity.lookupScopes = nil }, ErrScopeUnexpected, true},
		{"exchange fails", func(h *harness) { h.identity.exchangeErr = errors.New("502") }, ErrUpstream, false},
		{"lookup fails", func(h *harness) { h.identity.lookupErr = errors.New("502") }, ErrUpstream, true},
		{"missing id", func(h *harness) { h.identity.lookupID = "" }, ErrUpstream, true},
		{"non-numeric id", func(h *harness) { h.identity.lookupID = "12ab" }, ErrUpstream, true},
		{"invalid login", func(h *harness) { h.identity.lookupLogin = "-bad" }, ErrUpstream, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(true)
			h.identity.refreshToken = fakeRefreshToken
			c.setup(h)
			start, _ := h.start("/")
			_, err := h.callback(start.State, start.State, "")
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("err = %v, want %v", err, c.wantErr)
			}
			if len(h.sessions.items) != 0 || len(h.sessions.audit) != 0 || len(h.contributors.items) != 0 {
				t.Fatal("no session, audit event, or contributor may result")
			}
			if c.name != "exchange fails" && len(h.identity.revoked) != 1 {
				t.Fatal("a received access token must be revoked best-effort even on failure")
			}
			if strings.Contains(h.everythingObservable(), fakeRefreshToken) {
				t.Fatal("refresh token leaked")
			}
		})
	}
}

func TestCompleteSignInRevokeFailureIsTolerated(t *testing.T) {
	h := newHarness(true)
	h.identity.revokeErr = errors.New("revocation endpoint down")
	start, _ := h.start("/")
	if _, err := h.callback(start.State, start.State, ""); err != nil {
		t.Fatalf("revocation failure must not fail sign-in: %v", err)
	}
}

// A-32: revocation has its own deadline, so an unresponsive GitHub cannot hold the sign-in
// past the configured budget.
func TestCompleteSignInRevokeIsBoundedByTimeout(t *testing.T) {
	h := newHarness(true)
	h.identity.revokeHangs = true
	h.auth.cfg.RevokeTimeout = 20 * time.Millisecond
	start, _ := h.start("/")
	began := time.Now()
	if _, err := h.callback(start.State, start.State, ""); err != nil {
		t.Fatalf("revocation timeout must not fail sign-in: %v", err)
	}
	if elapsed := time.Since(began); elapsed > time.Second {
		t.Fatalf("sign-in waited %v for revocation; the budget is 20ms", elapsed)
	}
}

// F07-T12: suspended and revoked contributors sign in; login changes and reuse (AC-07-11, 12).
func TestCompleteSignInContributorStatusAndLogin(t *testing.T) {
	h := newHarness(true)
	t0 := h.clock.now
	h.contributors.items["1234567"] = domain.Contributor{GitHubUserID: "1234567", GitHubLogin: "old-login", Status: domain.ContributorRevoked, CreatedAt: t0.Add(-time.Hour), UpdatedAt: t0.Add(-time.Hour)}
	h.contributors.items["99"] = domain.Contributor{GitHubUserID: "99", GitHubLogin: "synthetic-contributor", Status: domain.ContributorActive, CreatedAt: t0, UpdatedAt: t0}

	start, _ := h.start("/")
	if _, err := h.callback(start.State, start.State, ""); err != nil {
		t.Fatalf("revoked contributor must still sign in: %v", err)
	}
	c := h.contributors.items["1234567"]
	if c.Status != domain.ContributorRevoked || c.GitHubLogin != "synthetic-contributor" || !c.CreatedAt.Equal(t0.Add(-time.Hour)) {
		t.Fatalf("status and history must be preserved, only the login refreshed: %+v", c)
	}
	if other := h.contributors.items["99"]; other.GitHubLogin != "synthetic-contributor" || other.Status != domain.ContributorActive {
		t.Fatal("a record whose former login matches must never be matched by login")
	}
	if len(h.contributors.items) != 2 {
		t.Fatal("no record may be created or merged by login")
	}
}

// F07-T3 (replacement part) and AC-07-9: the old session is replaced only on success.
func TestCompleteSignInReplacesSessionOnlyOnSuccess(t *testing.T) {
	h := newHarness(true)
	oldToken := "old-token"
	h.sessions.items[domain.HashToken(oldToken)] = domain.Session{TokenHash: domain.HashToken(oldToken)}

	start, _ := h.start("/")
	h.identity.exchangeErr = errors.New("502")
	if _, err := h.callback(start.State, start.State, oldToken); !errors.Is(err, ErrUpstream) {
		t.Fatal(err)
	}
	if _, kept := h.sessions.items[domain.HashToken(oldToken)]; !kept {
		t.Fatal("a failed attempt must preserve the old session")
	}

	h.identity.exchangeErr = nil
	start, _ = h.start("/")
	out, err := h.callback(start.State, start.State, oldToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, kept := h.sessions.items[domain.HashToken(oldToken)]; kept {
		t.Fatal("success must delete the old session in the same transaction")
	}
	if _, ok := h.sessions.items[domain.HashToken(out.SessionToken)]; !ok {
		t.Fatal("new session missing")
	}
}

func TestCompleteSignInStoreFailures(t *testing.T) {
	h := newHarness(true)
	start, _ := h.start("/")
	h.sessions.writeErr = ports.ErrUnavailable
	if _, err := h.callback(start.State, start.State, ""); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("unavailable store: %v", err)
	}
	h = newHarness(true)
	start, _ = h.start("/")
	h.sessions.commitThenFail = true
	_, err := h.callback(start.State, start.State, "")
	if !errors.Is(err, ErrStoreUncertain) {
		t.Fatalf("uncertain outcome: %v", err)
	}
	if len(h.sessions.items) != 1 || len(h.sessions.audit) != 1 {
		t.Fatal("a committed transaction leaves the session and audit event in place")
	}
	h = newHarness(true)
	h.attempts.err = ports.ErrUnavailable
	if _, err := h.start("/"); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("start with unavailable store: %v", err)
	}
}

// F07-T5: nothing secret reaches logs, audit events, or contributor records (AC-07-5).
func TestNoSecretLeaks(t *testing.T) {
	h := newHarness(true)
	h.identity.refreshToken = fakeRefreshToken
	start, _ := h.start("/agreements/")
	out, err := h.callback(start.State, start.State, "")
	if err != nil {
		t.Fatal(err)
	}
	verifier := h.identity.exchanges[0].verifier
	observable := h.everythingObservable()
	for name, secret := range map[string]string{
		"state": start.State, "verifier": verifier, "code": "code-1", "access token": fakeAccessToken,
		"refresh token": fakeRefreshToken, "session token": out.SessionToken,
		"csrf token": h.sessions.items[domain.HashToken(out.SessionToken)].CSRFToken,
	} {
		if secret == "" {
			t.Fatalf("%s is empty", name)
		}
		if strings.Contains(observable, secret) {
			t.Fatalf("%s leaked into logs, audit, or contributor records", name)
		}
	}
	for hash := range h.sessions.items {
		if hash == out.SessionToken {
			t.Fatal("session stored under the raw token")
		}
	}
}

// F07-T6 (use-case part) and F07-T16: authentication, renewal, and store outages.
func TestAuthenticate(t *testing.T) {
	h := newHarness(true)
	start, _ := h.start("/")
	out, _ := h.callback(start.State, start.State, "")
	ctx := context.Background()

	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{}); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("missing cookie: %v", err)
	}
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: "unknown"}); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("unknown token: %v", err)
	}
	sess, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken, Extend: false})
	if err != nil || sess.GitHubUserID != "1234567" {
		t.Fatalf("valid session: %v %+v", err, sess)
	}
	// Reading the session never extends it, even after the throttle window.
	h.clock.now = h.clock.now.Add(2 * time.Minute)
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken}); err != nil || h.sessions.touches != 0 {
		t.Fatalf("read must not touch: err=%v touches=%d", err, h.sessions.touches)
	}
	// An accepted state-changing request slides the idle expiry, once per minute.
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken, Extend: true}); err != nil || h.sessions.touches != 1 {
		t.Fatalf("extend: err=%v touches=%d", err, h.sessions.touches)
	}
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken, Extend: true}); err != nil || h.sessions.touches != 1 {
		t.Fatalf("throttle: touches=%d", h.sessions.touches)
	}
	// Idle expiry.
	h.clock.now = h.clock.now.Add(31 * time.Minute)
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken}); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("idle expired: %v", err)
	}
	// Store outage is 503, never 401.
	h.sessions.getErr = ports.ErrUnavailable
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken}); !errors.Is(err, ErrStoreUnavailable) || errors.Is(err, ErrSessionRequired) {
		t.Fatalf("store outage: %v", err)
	}
}

func TestAuthenticateTouchAfterDelete(t *testing.T) {
	h := newHarness(true)
	start, _ := h.start("/")
	out, _ := h.callback(start.State, start.State, "")
	h.clock.now = h.clock.now.Add(2 * time.Minute)
	// Simulate a sign-out racing the renewal: the item is gone when the touch lands.
	delete(h.sessions.items, domain.HashToken(out.SessionToken))
	h.sessions.items["decoy"] = domain.Session{TokenHash: "decoy"}
	h.sessions.getErr = nil
	// Get must still find the session for the read; emulate by re-adding then removing in Touch.
	s := domain.NewSession(domain.HashToken(out.SessionToken), "csrf", domain.GitHubIdentity{UserID: "1", Login: "a"}, domain.DefaultSessionPolicy, h.clock.now.Add(-2*time.Minute))
	h.sessions.items[s.TokenHash] = s
	h.sessions.writeErr = ports.ErrNotFound
	if _, err := h.auth.Authenticate(context.Background(), AuthenticateInput{SessionToken: out.SessionToken, Extend: true}); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("a session deleted before the touch must be rejected: %v", err)
	}
}

func TestVerifyCSRF(t *testing.T) {
	s := domain.Session{CSRFToken: "abc"}
	if !VerifyCSRF(s, "abc") || VerifyCSRF(s, "abd") || VerifyCSRF(s, "") {
		t.Fatal("VerifyCSRF")
	}
}

// F07-T8 (use-case part): sign-out semantics (AC-07-8).
func TestSignOut(t *testing.T) {
	ctx := context.Background()
	h := newHarness(true)
	start, _ := h.start("/")
	out, _ := h.callback(start.State, start.State, "")
	if err := h.auth.SignOut(ctx, SignOutInput{SessionToken: out.SessionToken, RequestID: "r"}); err != nil {
		t.Fatal(err)
	}
	if len(h.sessions.items) != 0 || len(h.sessions.audit) != 2 || h.sessions.audit[1].Type != domain.AuditSignedOut {
		t.Fatalf("sign-out must delete the session and create the audit event: %+v", h.sessions.audit)
	}
	if _, err := h.auth.Authenticate(ctx, AuthenticateInput{SessionToken: out.SessionToken}); !errors.Is(err, ErrSessionRequired) {
		t.Fatal("old token must no longer authenticate")
	}
	// Idempotent without a session or with an unknown token.
	if err := h.auth.SignOut(ctx, SignOutInput{}); err != nil {
		t.Fatal(err)
	}
	if err := h.auth.SignOut(ctx, SignOutInput{SessionToken: "unknown"}); err != nil {
		t.Fatal(err)
	}
	// Confirmed failure: session preserved, no audit event.
	h = newHarness(true)
	start, _ = h.start("/")
	out, _ = h.callback(start.State, start.State, "")
	h.sessions.writeErr = ports.ErrUnavailable
	if err := h.auth.SignOut(ctx, SignOutInput{SessionToken: out.SessionToken}); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("confirmed failure: %v", err)
	}
	if len(h.sessions.items) != 1 || len(h.sessions.audit) != 1 {
		t.Fatal("confirmed failure must preserve the session and write no sign-out event")
	}
	// Committed but the response was lost: session deleted and event present, still an error.
	h.sessions.writeErr = nil
	h.sessions.commitThenFail = true
	if err := h.auth.SignOut(ctx, SignOutInput{SessionToken: out.SessionToken}); !errors.Is(err, ErrStoreUncertain) {
		t.Fatalf("uncertain outcome: %v", err)
	}
	if len(h.sessions.items) != 0 || len(h.sessions.audit) != 2 {
		t.Fatal("a committed deletion leaves the session gone and the audit event present")
	}
}

func TestNewAuthValidatesConfig(t *testing.T) {
	if _, err := NewAuth(AuthConfig{}, AuthDependencies{}); err == nil {
		t.Fatal("missing callback URL must be rejected")
	}
	if _, err := NewAuth(AuthConfig{CallbackURL: testCallback, Session: domain.SessionPolicy{Idle: time.Hour, Absolute: time.Minute}}, AuthDependencies{}); err == nil {
		t.Fatal("invalid session policy must be rejected")
	}
}
