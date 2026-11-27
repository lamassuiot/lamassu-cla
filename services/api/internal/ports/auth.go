package ports

import (
	"context"
	"errors"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
)

// ErrUnavailable is returned by stores when the backing service cannot be reached. Callers map
// it to 503 and never interpret it as an absent item (F-07, AC-07-16).
var ErrUnavailable = errors.New("store: unavailable")

// ErrUncertain is returned when a write may or may not have committed (timeout after send).
// Callers treat it as a failure for the response while assuming the write may have happened.
var ErrUncertain = errors.New("store: outcome uncertain")

// ErrNotFound is returned when a conditional operation finds no usable item.
var ErrNotFound = errors.New("store: not found")

// Clock supplies the current time so that use cases are testable with fakes.
type Clock interface {
	Now() time.Time
}

// RandomSource supplies cryptographically random secrets.
type RandomSource interface {
	Secret() (domain.Secret, error)
}

// AuthorizationRequest carries what the GitHub authorization URL needs (F-07, step 2).
type AuthorizationRequest struct {
	State         string
	CodeChallenge string
	RedirectURI   string
}

// TokenExchange is the result of exchanging an authorization code (F-07, step 4.5).
type TokenExchange struct {
	AccessToken string
	// GrantedScope is the token response's scope field; GitHub reports an empty string for a
	// scopeless token. Nil means the field was absent, which fails sign-in.
	GrantedScope *string
	// RefreshToken is discarded by the use case; adapters return it only so that tests can
	// assert it never reaches storage.
	RefreshToken string
}

// IdentityLookup is the result of the authenticated GET /user call.
type IdentityLookup struct {
	Identity domain.GitHubIdentity
	// OAuthScopes is the X-OAuth-Scopes response header. Nil means the header was absent,
	// which fails sign-in (AC-07-13).
	OAuthScopes *string
}

// IdentityProvider is the GitHub OAuth App boundary. Adapters must not log tokens or codes.
type IdentityProvider interface {
	AuthorizationURL(req AuthorizationRequest) string
	Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (TokenExchange, error)
	LookupIdentity(ctx context.Context, accessToken string) (IdentityLookup, error)
	// Revoke is best effort; its error never fails a sign-in (F-07, step 4.5).
	Revoke(ctx context.Context, accessToken string) error
}

// SignInAttemptStore keeps OAUTH#{state_hash} items.
type SignInAttemptStore interface {
	Put(ctx context.Context, attempt domain.SignInAttempt) error
	// Consume atomically deletes the item for stateHash if it exists and expires after now, and
	// returns it. It returns ErrNotFound when the item is missing, expired, or already consumed
	// (F-07, step 4.3). Two concurrent calls succeed at most once.
	Consume(ctx context.Context, stateHash string, now time.Time) (domain.SignInAttempt, error)
	// Delete removes a superseded attempt; a missing item is not an error (F-07, step 2).
	Delete(ctx context.Context, stateHash string) error
}

// SessionStore keeps SESSION#{session_token_hash} items and writes audit events in the same
// transaction as the session change.
type SessionStore interface {
	Get(ctx context.Context, tokenHash string) (domain.Session, error)
	// Replace writes session, deletes the session identified by previousTokenHash when it is not
	// empty, and writes event, all in one transaction (F-07, step 6).
	Replace(ctx context.Context, session domain.Session, previousTokenHash string, event domain.AuditEvent) error
	// Touch persists a slid idle expiry, conditional on the session still existing and being
	// unexpired (F-07, step 8). A session deleted meanwhile returns ErrNotFound.
	Touch(ctx context.Context, session domain.Session) error
	// Delete removes the session and writes event in one transaction (F-07, step 9).
	Delete(ctx context.Context, tokenHash string, event domain.AuditEvent) error
}

// ContributorRepository keeps CONTRIB#{github_user_id} profile items.
type ContributorRepository interface {
	Get(ctx context.Context, githubUserID string) (domain.Contributor, error)
	Put(ctx context.Context, contributor domain.Contributor) error
}

// SecurityEvent is a structured log event without personal data (F-07, observability).
type SecurityEvent struct {
	Name      string
	Outcome   string
	RequestID string
}

// SecurityLogger records security-relevant outcomes. Implementations must never receive
// tokens, codes, state values, or personal data; the use cases pass none.
type SecurityLogger interface {
	Log(ctx context.Context, event SecurityEvent)
}
