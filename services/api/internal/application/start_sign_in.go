package application

import (
	"context"
	"fmt"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// StartSignInInput is what the start handler passes (F-07, main flow steps 1 and 2).
type StartSignInInput struct {
	ReturnTo string
	// PreviousState is the value of an existing __Host-cla_oauth cookie, if any. The attempt it
	// refers to is deleted: a new attempt supersedes the previous one.
	PreviousState string
	RequestID     string
}

// StartSignInOutput is what the handler needs to redirect and set the pre-authentication cookie.
type StartSignInOutput struct {
	// State is the encoded value for the cookie and the GitHub request. Never logged.
	State string
	// RedirectURL is GitHub's authorization page with client_id, redirect_uri, state,
	// code_challenge, code_challenge_method=S256, and an explicitly empty scope.
	RedirectURL string
	ExpiresAt   int64
}

// StartSignIn validates return_to, creates the state and PKCE verifier, stores the attempt, and
// returns the GitHub redirect (AC-07-1). It never touches an existing session.
func (a *Auth) StartSignIn(ctx context.Context, in StartSignInInput) (StartSignInOutput, error) {
	if !a.cfg.Enabled {
		a.logSecurity(ctx, "auth.start", "disabled", in.RequestID)
		return StartSignInOutput{}, ErrSignInDisabled
	}
	now := a.deps.Clock.Now()
	returnTo := a.cfg.ReturnTo.Resolve(in.ReturnTo)

	if in.PreviousState != "" {
		if err := a.deps.Attempts.Delete(ctx, domain.HashToken(in.PreviousState)); err != nil {
			return StartSignInOutput{}, wrapStore(fmt.Errorf("superseding attempt: %w", err))
		}
	}

	stateSecret, err := a.deps.Random.Secret()
	if err != nil {
		return StartSignInOutput{}, fmt.Errorf("state: %w", err)
	}
	verifierSecret, err := a.deps.Random.Secret()
	if err != nil {
		return StartSignInOutput{}, fmt.Errorf("verifier: %w", err)
	}
	state := stateSecret.Encode()
	verifier := verifierSecret.Encode()

	attempt := domain.SignInAttempt{
		StateHash: domain.HashToken(state),
		Verifier:  verifier,
		ReturnTo:  returnTo,
		ExpiresAt: now.Add(domain.SignInAttemptLifetime),
	}
	if err := a.deps.Attempts.Put(ctx, attempt); err != nil {
		return StartSignInOutput{}, wrapStore(fmt.Errorf("storing attempt: %w", err))
	}

	a.logSecurity(ctx, "auth.start", "redirected", in.RequestID)
	return StartSignInOutput{
		State: state,
		RedirectURL: a.deps.Identity.AuthorizationURL(ports.AuthorizationRequest{
			State:         state,
			CodeChallenge: domain.PKCEChallenge(verifier),
			RedirectURI:   a.cfg.CallbackURL,
		}),
		ExpiresAt: attempt.ExpiresAt.Unix(),
	}, nil
}
