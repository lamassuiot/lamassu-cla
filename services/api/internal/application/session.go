package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// domainTime keeps the application signatures readable.
type domainTime = time.Time

// AuthenticateInput is what the session middleware passes (F-07, main flow step 8).
type AuthenticateInput struct {
	// SessionToken is the __Host-cla_session cookie value, or empty when absent.
	SessionToken string
	// Extend is true for accepted requests other than reading the session; reading never
	// extends the idle expiry.
	Extend    bool
	RequestID string
}

// Authenticate validates the session and, when Extend is set, slides the idle expiry under the
// renewal rule. It returns ErrSessionRequired for a missing, unknown, or expired session and
// ErrStoreUnavailable when the store cannot answer; the latter is never reported as 401.
func (a *Auth) Authenticate(ctx context.Context, in AuthenticateInput) (domain.Session, error) {
	if in.SessionToken == "" {
		a.logSecurity(ctx, "session.rejected", "unknown", in.RequestID)
		return domain.Session{}, ErrSessionRequired
	}
	now := a.deps.Clock.Now()
	session, err := a.deps.Sessions.Get(ctx, domain.HashToken(in.SessionToken))
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			a.logSecurity(ctx, "session.rejected", "unknown", in.RequestID)
			return domain.Session{}, ErrSessionRequired
		}
		a.logSecurity(ctx, "session.rejected", "store_unavailable", in.RequestID)
		return domain.Session{}, wrapStore(fmt.Errorf("reading session: %w", err))
	}
	if err := session.Valid(now); err != nil {
		outcome := "idle_expired"
		if !now.Before(session.ExpiresAt) {
			outcome = "expired"
		}
		a.logSecurity(ctx, "session.rejected", outcome, in.RequestID)
		return domain.Session{}, ErrSessionRequired
	}
	if !in.Extend {
		return session, nil
	}
	touched, due := session.Touch(a.cfg.Session, now)
	if !due {
		return session, nil
	}
	if err := a.deps.Sessions.Touch(ctx, touched); err != nil {
		// A session deleted between the read and the update stays rejected on the next request;
		// an unavailable store must not fail an already authenticated request.
		if errors.Is(err, ports.ErrNotFound) {
			a.logSecurity(ctx, "session.rejected", "unknown", in.RequestID)
			return domain.Session{}, ErrSessionRequired
		}
		return session, nil
	}
	return touched, nil
}

// VerifyCSRF compares the X-CSRF-Token header with the session's token in constant time
// (F-07, step 8). The Origin check is a transport concern handled in middleware.
func VerifyCSRF(session domain.Session, header string) bool {
	return header != "" && domain.ConstantTimeEqual(session.CSRFToken, header)
}

// SignOutInput is what the sign-out handler passes after the Origin and CSRF checks.
type SignOutInput struct {
	SessionToken string
	RequestID    string
}

// SignOut deletes the session and writes auth.signed_out in one transaction (F-07, step 9).
// Without a session it does nothing and returns nil (idempotent 204). A failed or uncertain
// deletion is returned as an error so that the handler answers 503, never 204 (AC-07-8).
func (a *Auth) SignOut(ctx context.Context, in SignOutInput) error {
	if in.SessionToken == "" {
		return nil
	}
	now := a.deps.Clock.Now()
	tokenHash := domain.HashToken(in.SessionToken)
	session, err := a.deps.Sessions.Get(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil
		}
		return wrapStore(fmt.Errorf("reading session: %w", err))
	}
	event := domain.AuditEvent{Type: domain.AuditSignedOut, ActorID: session.GitHubUserID, OccurredAt: now, RequestID: in.RequestID}
	if err := a.deps.Sessions.Delete(ctx, tokenHash, event); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil
		}
		return wrapStore(fmt.Errorf("deleting session: %w", err))
	}
	a.logSecurity(ctx, "auth.sign_out", "success", in.RequestID)
	return nil
}
