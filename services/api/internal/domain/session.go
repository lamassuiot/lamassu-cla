package domain

import (
	"errors"
	"time"
)

// SignInAttempt is the OAUTH#{state_hash} item: one pending GitHub authorization (F-07, main
// flow step 2). State is stored only as StateHash; the PKCE verifier is kept in the original
// form because the code exchange needs it (security considerations).
type SignInAttempt struct {
	StateHash string
	Verifier  string
	ReturnTo  string
	ExpiresAt time.Time
}

// Expired reports whether the attempt can no longer be consumed. Expiry is enforced here and in
// the store's conditional operation, never by TTL deletion alone.
func (a SignInAttempt) Expired(now time.Time) bool { return !now.Before(a.ExpiresAt) }

// Session is the SESSION#{session_token_hash} item (F-07, main flow step 6). The token itself
// is never stored; CSRFToken is stored as is so that it can be returned repeatedly.
type Session struct {
	TokenHash     string
	GitHubUserID  string
	GitHubLogin   string
	CSRFToken     string
	CreatedAt     time.Time
	LastSeenAt    time.Time
	IdleExpiresAt time.Time
	ExpiresAt     time.Time
}

// NewSession creates a session for identity under policy at now.
func NewSession(tokenHash, csrfToken string, identity GitHubIdentity, policy SessionPolicy, now time.Time) Session {
	absolute := now.Add(policy.Absolute)
	return Session{
		TokenHash:     tokenHash,
		GitHubUserID:  identity.UserID,
		GitHubLogin:   identity.Login,
		CSRFToken:     csrfToken,
		CreatedAt:     now,
		LastSeenAt:    now,
		IdleExpiresAt: minTime(now.Add(policy.Idle), absolute),
		ExpiresAt:     absolute,
	}
}

// ErrSessionExpired is returned when a stored session is past its idle or absolute expiry.
var ErrSessionExpired = errors.New("session: expired")

// Valid returns ErrSessionExpired when now is not before both expiries (AC-07-6).
func (s Session) Valid(now time.Time) error {
	if !now.Before(s.IdleExpiresAt) || !now.Before(s.ExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}

// Touch computes the sliding idle expiry for an accepted request (F-07, main flow step 8):
// idle_expires_at = min(now + idle, expires_at), throttled to one write per minute. It returns
// the updated session and whether a store write is due. Reading the session must not call it.
func (s Session) Touch(policy SessionPolicy, now time.Time) (Session, bool) {
	if s.Valid(now) != nil {
		return s, false
	}
	if now.Sub(s.LastSeenAt) < IdleRenewalThrottle {
		return s, false
	}
	s.LastSeenAt = now
	s.IdleExpiresAt = minTime(now.Add(policy.Idle), s.ExpiresAt)
	return s, true
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
