package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// AuthConfig is the configuration of the sign-in and session use cases (F-07).
type AuthConfig struct {
	// Enabled is the sign-in flag, off by default until D-20 and deployment (F-07, step 1).
	Enabled bool
	// CallbackURL is the exact registered callback URL, sent as redirect_uri (D-02).
	CallbackURL string
	ReturnTo    domain.ReturnToPolicy
	Session     domain.SessionPolicy
}

// Validate reports configuration errors at wiring time.
func (c AuthConfig) Validate() error {
	if c.CallbackURL == "" {
		return errors.New("auth config: callback URL is required")
	}
	return c.Session.Validate()
}

// AuthDependencies groups the ports the sign-in use cases need.
type AuthDependencies struct {
	Clock        ports.Clock
	Random       ports.RandomSource
	Identity     ports.IdentityProvider
	Attempts     ports.SignInAttemptStore
	Sessions     ports.SessionStore
	Contributors ports.ContributorRepository
	Security     ports.SecurityLogger
}

// Auth implements the F-07 use cases.
type Auth struct {
	cfg  AuthConfig
	deps AuthDependencies
}

// NewAuth wires the use cases.
func NewAuth(cfg AuthConfig, deps AuthDependencies) (*Auth, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Auth{cfg: cfg, deps: deps}, nil
}

// Sentinel errors mapped by handlers to problem responses and redirect reasons.
var (
	ErrSignInDisabled   = errors.New("auth: sign-in is disabled")
	ErrMalformedRequest = errors.New("auth: malformed request")
	ErrBindingMismatch  = errors.New("auth: pre-authentication cookie missing or mismatched")
	ErrStateInvalid     = errors.New("auth: state missing, expired, or already used")
	ErrDenied           = errors.New("auth: the person declined on GitHub")
	ErrUpstream         = errors.New("auth: GitHub authorization or lookup failed")
	ErrScopeUnexpected  = errors.New("auth: granted scope is not empty")
	ErrSessionRequired  = errors.New("auth: session missing, unknown, or expired")
	ErrStoreUnavailable = errors.New("auth: store unavailable")
	ErrStoreUncertain   = errors.New("auth: store outcome uncertain")
)

func (a *Auth) logSecurity(ctx context.Context, name, outcome, requestID string) {
	if a.deps.Security != nil {
		a.deps.Security.Log(ctx, ports.SecurityEvent{Name: name, Outcome: outcome, RequestID: requestID})
	}
}

// wrapStore maps store sentinel errors to use-case errors, preserving the chain.
func wrapStore(err error) error {
	switch {
	case errors.Is(err, ports.ErrUnavailable):
		return fmt.Errorf("%w: %w", ErrStoreUnavailable, err)
	case errors.Is(err, ports.ErrUncertain):
		return fmt.Errorf("%w: %w", ErrStoreUncertain, err)
	default:
		return err
	}
}
