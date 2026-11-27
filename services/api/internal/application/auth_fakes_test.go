package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// Fakes for the F-07 ports. They are deterministic, in memory, and record every value they are
// given so that tests can scan for leaks (F07-T5).

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

// fakeRandom hands out predictable, distinct 32-byte secrets.
type fakeRandom struct {
	mu   sync.Mutex
	n    byte
	fail bool
}

func (r *fakeRandom) Secret() (domain.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fail {
		return nil, errors.New("entropy unavailable")
	}
	r.n++
	s := make(domain.Secret, domain.SecretBytes)
	for i := range s {
		s[i] = r.n
	}
	return s, nil
}

type fakeIdentity struct {
	mu            sync.Mutex
	authRequests  []ports.AuthorizationRequest
	exchanges     []struct{ code, verifier, redirect string }
	exchangeErr   error
	exchangeScope *string
	refreshToken  string
	lookupErr     error
	lookupScopes  *string
	lookupID      string
	lookupLogin   string
	revoked       []string
	revokeErr     error
	revokeHangs   bool
}

func scope(s string) *string { return &s }

// Synthetic credentials for the fakes; they never touch a real service.
const (
	fakeAccessToken  = "gho_fake_access_token" //nolint:gosec // synthetic test value
	fakeRefreshToken = "ghr_fake_refresh"      //nolint:gosec // synthetic test value
)

func newFakeIdentity() *fakeIdentity {
	return &fakeIdentity{exchangeScope: scope(""), lookupScopes: scope(""), lookupID: "1234567", lookupLogin: "synthetic-contributor"}
}

func (f *fakeIdentity) AuthorizationURL(req ports.AuthorizationRequest) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.authRequests = append(f.authRequests, req)
	q := url.Values{}
	q.Set("client_id", "fake-client-id")
	q.Set("redirect_uri", req.RedirectURI)
	q.Set("state", req.State)
	q.Set("code_challenge", req.CodeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("scope", "")
	return "https://github.example/login/oauth/authorize?" + q.Encode()
}

func (f *fakeIdentity) Exchange(_ context.Context, code, verifier, redirect string) (ports.TokenExchange, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.exchanges = append(f.exchanges, struct{ code, verifier, redirect string }{code, verifier, redirect})
	if f.exchangeErr != nil {
		return ports.TokenExchange{}, f.exchangeErr
	}
	return ports.TokenExchange{AccessToken: fakeAccessToken, GrantedScope: f.exchangeScope, RefreshToken: f.refreshToken}, nil
}

func (f *fakeIdentity) LookupIdentity(_ context.Context, token string) (ports.IdentityLookup, error) {
	if f.lookupErr != nil {
		return ports.IdentityLookup{}, f.lookupErr
	}
	if token != fakeAccessToken {
		return ports.IdentityLookup{}, errors.New("unknown token")
	}
	return ports.IdentityLookup{Identity: domain.GitHubIdentity{UserID: f.lookupID, Login: f.lookupLogin}, OAuthScopes: f.lookupScopes}, nil
}

func (f *fakeIdentity) Revoke(ctx context.Context, token string) error {
	if f.revokeHangs {
		<-ctx.Done()
		return ctx.Err()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.revoked = append(f.revoked, token)
	return f.revokeErr
}

// fakeAttempts implements the conditional consume with a mutex, like DynamoDB's conditional
// delete: at most one of several concurrent Consume calls succeeds.
type fakeAttempts struct {
	mu       sync.Mutex
	items    map[string]domain.SignInAttempt
	consumed int
	err      error
}

func newFakeAttempts() *fakeAttempts { return &fakeAttempts{items: map[string]domain.SignInAttempt{}} }

func (s *fakeAttempts) Put(_ context.Context, a domain.SignInAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.items[a.StateHash] = a
	return nil
}

func (s *fakeAttempts) Consume(_ context.Context, stateHash string, now time.Time) (domain.SignInAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return domain.SignInAttempt{}, s.err
	}
	a, ok := s.items[stateHash]
	if !ok || a.Expired(now) {
		return domain.SignInAttempt{}, ports.ErrNotFound
	}
	delete(s.items, stateHash)
	s.consumed++
	return a, nil
}

func (s *fakeAttempts) Delete(_ context.Context, stateHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	delete(s.items, stateHash)
	return nil
}

type fakeSessions struct {
	mu       sync.Mutex
	items    map[string]domain.Session
	audit    []domain.AuditEvent
	touches  int
	getErr   error
	writeErr error
	// commitThenFail simulates a transaction that committed before its response was lost.
	commitThenFail bool
}

func newFakeSessions() *fakeSessions { return &fakeSessions{items: map[string]domain.Session{}} }

func (s *fakeSessions) Get(_ context.Context, hash string) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.getErr != nil {
		return domain.Session{}, s.getErr
	}
	sess, ok := s.items[hash]
	if !ok {
		return domain.Session{}, ports.ErrNotFound
	}
	return sess, nil
}

func (s *fakeSessions) Replace(_ context.Context, sess domain.Session, previous string, event domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil && !s.commitThenFail {
		return s.writeErr
	}
	if previous != "" {
		delete(s.items, previous)
	}
	s.items[sess.TokenHash] = sess
	s.audit = append(s.audit, event)
	if s.commitThenFail {
		return ports.ErrUncertain
	}
	return nil
}

func (s *fakeSessions) Touch(_ context.Context, sess domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil {
		return s.writeErr
	}
	if _, ok := s.items[sess.TokenHash]; !ok {
		return ports.ErrNotFound
	}
	s.items[sess.TokenHash] = sess
	s.touches++
	return nil
}

func (s *fakeSessions) Delete(_ context.Context, hash string, event domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil && !s.commitThenFail {
		return s.writeErr
	}
	if _, ok := s.items[hash]; !ok {
		return ports.ErrNotFound
	}
	delete(s.items, hash)
	s.audit = append(s.audit, event)
	if s.commitThenFail {
		return ports.ErrUncertain
	}
	return nil
}

type fakeContributors struct {
	mu    sync.Mutex
	items map[string]domain.Contributor
	err   error
}

func newFakeContributors() *fakeContributors {
	return &fakeContributors{items: map[string]domain.Contributor{}}
}

func (r *fakeContributors) Get(_ context.Context, id string) (domain.Contributor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return domain.Contributor{}, r.err
	}
	c, ok := r.items[id]
	if !ok {
		return domain.Contributor{}, ports.ErrNotFound
	}
	return c, nil
}

func (r *fakeContributors) Put(_ context.Context, c domain.Contributor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.items[c.GitHubUserID] = c
	return nil
}

type fakeSecurity struct {
	mu     sync.Mutex
	events []ports.SecurityEvent
}

func (l *fakeSecurity) Log(_ context.Context, e ports.SecurityEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, e)
}

// harness wires an Auth with fakes and exposes them.
type harness struct {
	auth         *Auth
	clock        *fakeClock
	random       *fakeRandom
	identity     *fakeIdentity
	attempts     *fakeAttempts
	sessions     *fakeSessions
	contributors *fakeContributors
	security     *fakeSecurity
}

const testCallback = "https://cla.example/api/auth/github/callback"

func newHarness(enabled bool) *harness {
	h := &harness{
		clock:        &fakeClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)},
		random:       &fakeRandom{},
		identity:     newFakeIdentity(),
		attempts:     newFakeAttempts(),
		sessions:     newFakeSessions(),
		contributors: newFakeContributors(),
		security:     &fakeSecurity{},
	}
	auth, err := NewAuth(AuthConfig{
		Enabled:     enabled,
		CallbackURL: testCallback,
		ReturnTo:    domain.ReturnToPolicy{Allowed: []string{"/", "/agreements/"}},
		Session:     domain.DefaultSessionPolicy,
	}, AuthDependencies{
		Clock: h.clock, Random: h.random, Identity: h.identity, Attempts: h.attempts,
		Sessions: h.sessions, Contributors: h.contributors, Security: h.security,
	})
	if err != nil {
		panic(err)
	}
	h.auth = auth
	return h
}

// start runs StartSignIn and returns the state (cookie value) and redirect URL.
func (h *harness) start(returnTo string) (StartSignInOutput, error) {
	return h.auth.StartSignIn(context.Background(), StartSignInInput{ReturnTo: returnTo, RequestID: "req-start"})
}

// callback runs a success callback bound to the given state.
func (h *harness) callback(state, cookieState, currentSession string) (CompleteSignInOutput, error) {
	return h.auth.CompleteSignIn(context.Background(), CompleteSignInInput{
		Code: "code-1", State: state, CookieState: cookieState, CurrentSessionToken: currentSession, RequestID: "req-cb",
	})
}

// everythingObservable concatenates every value the fakes received outside the stores where the
// specification allows a secret to live, for the F07-T5 leak scan.
func (h *harness) everythingObservable() string {
	var b strings.Builder
	for _, e := range h.security.events {
		fmt.Fprintf(&b, "%s %s %s\n", e.Name, e.Outcome, e.RequestID)
	}
	for _, e := range h.sessions.audit {
		fmt.Fprintf(&b, "%s %s %s\n", e.Type, e.ActorID, e.RequestID)
	}
	for _, c := range h.contributors.items {
		fmt.Fprintf(&b, "%s %s %s\n", c.GitHubUserID, c.GitHubLogin, c.Status)
	}
	return b.String()
}
