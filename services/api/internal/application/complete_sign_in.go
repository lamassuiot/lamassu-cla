package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// CompleteSignInInput is the parsed callback (F-07, main flow step 4). The handler passes raw
// query values; Malformed reports duplicate parameters, which the use case cannot see.
type CompleteSignInInput struct {
	Code  string
	State string
	Error string
	// CookieState is the value of the __Host-cla_oauth cookie, or empty when absent.
	CookieState string
	// CurrentSessionToken is the value of an existing __Host-cla_session cookie, or empty.
	CurrentSessionToken string
	Malformed           bool
	RequestID           string
}

// CompleteSignInOutput is what the handler needs to set the session cookie and redirect.
type CompleteSignInOutput struct {
	// SessionToken is the encoded value for the __Host-cla_session cookie. Never logged.
	SessionToken string
	ReturnTo     string
	ExpiresAt    int64
}

// CompleteSignIn processes a callback in the order fixed by F-07 step 4: malformed input, cookie
// binding, atomic consumption, then the GitHub exchange and the session replacement. Every error
// is one of the sentinel errors so that the handler can redirect with a reason and no detail.
func (a *Auth) CompleteSignIn(ctx context.Context, in CompleteSignInInput) (CompleteSignInOutput, error) {
	if !a.cfg.Enabled {
		a.logSecurity(ctx, "auth.callback", "disabled", in.RequestID)
		return CompleteSignInOutput{}, ErrSignInDisabled
	}
	// 4.1 Malformed input: duplicates, both or neither of code and error, missing state.
	if in.Malformed || in.State == "" || (in.Code == "") == (in.Error == "") {
		a.logSecurity(ctx, "auth.callback", "malformed", in.RequestID)
		return CompleteSignInOutput{}, ErrMalformedRequest
	}
	// 4.2 Browser binding before anything is consumed.
	if in.CookieState == "" || !domain.ConstantTimeEqual(in.CookieState, in.State) {
		a.logSecurity(ctx, "auth.callback", "binding_mismatch", in.RequestID)
		return CompleteSignInOutput{}, ErrBindingMismatch
	}
	now := a.deps.Clock.Now()
	// 4.3 Atomic consumption with the expiry condition.
	attempt, err := a.deps.Attempts.Consume(ctx, domain.HashToken(in.State), now)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			a.logSecurity(ctx, "auth.callback", "state_invalid", in.RequestID)
			return CompleteSignInOutput{}, ErrStateInvalid
		}
		return CompleteSignInOutput{}, wrapStore(fmt.Errorf("consuming attempt: %w", err))
	}
	// 4.5 Error callback after consumption.
	if in.Error != "" {
		a.logSecurity(ctx, "auth.callback", "denied", in.RequestID)
		return CompleteSignInOutput{ReturnTo: attempt.ReturnTo}, ErrDenied
	}

	identity, err := a.resolveIdentity(ctx, in.Code, attempt.Verifier, in.RequestID)
	if err != nil {
		return CompleteSignInOutput{}, err
	}

	// Step 5: contributor record, status and history untouched (D-29, AC-07-12).
	if err := a.upsertContributor(ctx, identity, now); err != nil {
		return CompleteSignInOutput{}, err
	}

	// Step 6: the session is created only now, replacing the browser's previous one atomically.
	tokenSecret, err := a.deps.Random.Secret()
	if err != nil {
		return CompleteSignInOutput{}, fmt.Errorf("session token: %w", err)
	}
	csrfSecret, err := a.deps.Random.Secret()
	if err != nil {
		return CompleteSignInOutput{}, fmt.Errorf("csrf token: %w", err)
	}
	token := tokenSecret.Encode()
	session := domain.NewSession(domain.HashToken(token), csrfSecret.Encode(), identity, a.cfg.Session, now)
	previous := ""
	if in.CurrentSessionToken != "" {
		previous = domain.HashToken(in.CurrentSessionToken)
	}
	event := domain.AuditEvent{Type: domain.AuditSignedIn, ActorID: identity.UserID, OccurredAt: now, RequestID: in.RequestID}
	if err := a.deps.Sessions.Replace(ctx, session, previous, event); err != nil {
		return CompleteSignInOutput{}, wrapStore(fmt.Errorf("replacing session: %w", err))
	}

	a.logSecurity(ctx, "auth.callback", "success", in.RequestID)
	return CompleteSignInOutput{SessionToken: token, ReturnTo: attempt.ReturnTo, ExpiresAt: session.ExpiresAt.Unix()}, nil
}

// resolveIdentity performs step 4.5: exchange, token-response scope check, GET /user with the
// X-OAuth-Scopes check, identity validation, then best-effort revocation.
func (a *Auth) resolveIdentity(ctx context.Context, code, verifier, requestID string) (domain.GitHubIdentity, error) {
	exchange, err := a.deps.Identity.Exchange(ctx, code, verifier, a.cfg.CallbackURL)
	if err != nil {
		a.logSecurity(ctx, "auth.callback", "upstream_error", requestID)
		return domain.GitHubIdentity{}, fmt.Errorf("%w: exchange", ErrUpstream)
	}
	defer a.revoke(ctx, exchange.AccessToken)
	if exchange.GrantedScope == nil || *exchange.GrantedScope != "" {
		a.logSecurity(ctx, "auth.callback", "scope_unexpected", requestID)
		return domain.GitHubIdentity{}, fmt.Errorf("%w: token response", ErrScopeUnexpected)
	}
	lookup, err := a.deps.Identity.LookupIdentity(ctx, exchange.AccessToken)
	if err != nil {
		a.logSecurity(ctx, "auth.callback", "upstream_error", requestID)
		return domain.GitHubIdentity{}, fmt.Errorf("%w: user lookup", ErrUpstream)
	}
	if lookup.OAuthScopes == nil || *lookup.OAuthScopes != "" {
		a.logSecurity(ctx, "auth.callback", "scope_unexpected", requestID)
		return domain.GitHubIdentity{}, fmt.Errorf("%w: X-OAuth-Scopes", ErrScopeUnexpected)
	}
	identity, err := domain.NewGitHubIdentity(lookup.Identity.UserID, lookup.Identity.Login)
	if err != nil {
		a.logSecurity(ctx, "auth.callback", "upstream_error", requestID)
		return domain.GitHubIdentity{}, fmt.Errorf("%w: %w", ErrUpstream, err)
	}
	return identity, nil
}

func (a *Auth) revoke(ctx context.Context, accessToken string) {
	if accessToken == "" {
		return
	}
	// Best effort by specification: a failure here must not fail the sign-in.
	_ = a.deps.Identity.Revoke(ctx, accessToken)
}

func (a *Auth) upsertContributor(ctx context.Context, identity domain.GitHubIdentity, now domainTime) error {
	existing, err := a.deps.Contributors.Get(ctx, identity.UserID)
	switch {
	case err == nil:
		refreshed, changed := existing.RefreshLogin(identity, now)
		if !changed {
			return nil
		}
		if err := a.deps.Contributors.Put(ctx, refreshed); err != nil {
			return wrapStore(fmt.Errorf("refreshing contributor: %w", err))
		}
		return nil
	case errors.Is(err, ports.ErrNotFound):
		if err := a.deps.Contributors.Put(ctx, domain.NewContributor(identity, now)); err != nil {
			return wrapStore(fmt.Errorf("creating contributor: %w", err))
		}
		return nil
	default:
		return wrapStore(fmt.Errorf("reading contributor: %w", err))
	}
}
