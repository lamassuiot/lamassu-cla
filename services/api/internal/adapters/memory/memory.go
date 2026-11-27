// Package memory implements the F-07 store ports in process memory. It backs handler tests and
// the local development server; it is never deployed.
package memory

import (
	"context"
	"crypto/rand"
	"sync"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// Clock is the real clock.
type Clock struct{}

// Now returns the current UTC time.
func (Clock) Now() time.Time { return time.Now().UTC() }

// Random draws secrets from crypto/rand.
type Random struct{}

// Secret returns a fresh 256-bit secret.
func (Random) Secret() (domain.Secret, error) {
	s := make(domain.Secret, domain.SecretBytes)
	if _, err := rand.Read(s); err != nil {
		return nil, err
	}
	return s, nil
}

// SignInAttempts is an in-memory ports.SignInAttemptStore with the conditional semantics of the
// DynamoDB adapter: Consume succeeds at most once per item and enforces expiry itself.
type SignInAttempts struct {
	mu    sync.Mutex
	items map[string]domain.SignInAttempt
	// Fail, when set, is returned by every operation; tests use it to simulate outages.
	Fail error
}

// NewSignInAttempts returns an empty store.
func NewSignInAttempts() *SignInAttempts {
	return &SignInAttempts{items: map[string]domain.SignInAttempt{}}
}

// Put stores the attempt.
func (s *SignInAttempts) Put(_ context.Context, a domain.SignInAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Fail != nil {
		return s.Fail
	}
	s.items[a.StateHash] = a
	return nil
}

// Consume atomically removes and returns an unexpired attempt.
func (s *SignInAttempts) Consume(_ context.Context, stateHash string, now time.Time) (domain.SignInAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Fail != nil {
		return domain.SignInAttempt{}, s.Fail
	}
	a, ok := s.items[stateHash]
	if !ok || a.Expired(now) {
		return domain.SignInAttempt{}, ports.ErrNotFound
	}
	delete(s.items, stateHash)
	return a, nil
}

// Delete removes an attempt; a missing item is not an error.
func (s *SignInAttempts) Delete(_ context.Context, stateHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Fail != nil {
		return s.Fail
	}
	delete(s.items, stateHash)
	return nil
}

// Len reports the number of pending attempts (tests).
func (s *SignInAttempts) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

// Sessions is an in-memory ports.SessionStore. Replace and Delete apply the session change and
// the audit event together, like the DynamoDB transaction.
type Sessions struct {
	mu    sync.Mutex
	items map[string]domain.Session
	audit []domain.AuditEvent
	// Fail, when set, is returned by every write; FailReads by Get.
	Fail      error
	FailReads error
}

// NewSessions returns an empty store.
func NewSessions() *Sessions { return &Sessions{items: map[string]domain.Session{}} }

// Get returns the session for tokenHash.
func (s *Sessions) Get(_ context.Context, tokenHash string) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.FailReads != nil {
		return domain.Session{}, s.FailReads
	}
	sess, ok := s.items[tokenHash]
	if !ok {
		return domain.Session{}, ports.ErrNotFound
	}
	return sess, nil
}

// Replace writes the session, deletes the previous one, and appends the audit event.
func (s *Sessions) Replace(_ context.Context, sess domain.Session, previousTokenHash string, event domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Fail != nil {
		return s.Fail
	}
	if previousTokenHash != "" {
		delete(s.items, previousTokenHash)
	}
	s.items[sess.TokenHash] = sess
	s.audit = append(s.audit, event)
	return nil
}

// Touch persists a slid idle expiry if the session still exists.
func (s *Sessions) Touch(_ context.Context, sess domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Fail != nil {
		return s.Fail
	}
	if _, ok := s.items[sess.TokenHash]; !ok {
		return ports.ErrNotFound
	}
	s.items[sess.TokenHash] = sess
	return nil
}

// Delete removes the session and appends the audit event.
func (s *Sessions) Delete(_ context.Context, tokenHash string, event domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Fail != nil {
		return s.Fail
	}
	if _, ok := s.items[tokenHash]; !ok {
		return ports.ErrNotFound
	}
	delete(s.items, tokenHash)
	s.audit = append(s.audit, event)
	return nil
}

// Len reports the number of sessions (tests).
func (s *Sessions) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

// Audit returns a copy of the audit events (tests).
func (s *Sessions) Audit() []domain.AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.AuditEvent(nil), s.audit...)
}

// Contributors is an in-memory ports.ContributorRepository.
type Contributors struct {
	mu    sync.Mutex
	items map[string]domain.Contributor
	Fail  error
}

// NewContributors returns an empty repository.
func NewContributors() *Contributors { return &Contributors{items: map[string]domain.Contributor{}} }

// Get returns the contributor for githubUserID.
func (r *Contributors) Get(_ context.Context, githubUserID string) (domain.Contributor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Fail != nil {
		return domain.Contributor{}, r.Fail
	}
	c, ok := r.items[githubUserID]
	if !ok {
		return domain.Contributor{}, ports.ErrNotFound
	}
	return c, nil
}

// Put stores the contributor.
func (r *Contributors) Put(_ context.Context, c domain.Contributor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Fail != nil {
		return r.Fail
	}
	r.items[c.GitHubUserID] = c
	return nil
}

// Len reports the number of contributors (tests).
func (r *Contributors) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.items)
}

// SecurityLog collects security events in memory.
type SecurityLog struct {
	mu     sync.Mutex
	events []ports.SecurityEvent
}

// Log appends the event.
func (l *SecurityLog) Log(_ context.Context, e ports.SecurityEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, e)
}

// Events returns a copy of the events.
func (l *SecurityLog) Events() []ports.SecurityEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]ports.SecurityEvent(nil), l.events...)
}
