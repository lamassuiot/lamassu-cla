package domain

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

// Token sizes and lifetimes fixed by F-07 and ADR-0014.
const (
	// SecretBytes is the size of session, CSRF, state, and PKCE verifier secrets (256 bits).
	SecretBytes = 32
	// SignInAttemptLifetime bounds an OAuth state and its PKCE verifier (F-07 main flow, step 2).
	SignInAttemptLifetime = 10 * time.Minute
	// IdleRenewalThrottle bounds how often an accepted request may slide the idle expiry.
	IdleRenewalThrottle = time.Minute
)

// SessionPolicy holds the configurable timeouts (security specification, section 3.1).
type SessionPolicy struct {
	Idle     time.Duration
	Absolute time.Duration
}

// DefaultSessionPolicy is the decided default (D-06): 30 minutes idle, 8 hours absolute.
var DefaultSessionPolicy = SessionPolicy{Idle: 30 * time.Minute, Absolute: 8 * time.Hour}

// Validate reports whether the policy is usable.
func (p SessionPolicy) Validate() error {
	if p.Idle <= 0 || p.Absolute <= 0 {
		return errors.New("session policy: timeouts must be positive")
	}
	if p.Idle > p.Absolute {
		return errors.New("session policy: idle timeout exceeds the absolute timeout")
	}
	return nil
}

// Secret is a raw 256-bit random value. It never leaves memory except as a cookie or header
// value, and it is stored only through Hash, except for the PKCE verifier and the CSRF token,
// which the flow needs in the original form (F-07, security considerations).
type Secret []byte

// Encode returns the URL-safe, unpadded representation used in cookies, headers, and URLs.
func (s Secret) Encode() string { return base64.RawURLEncoding.EncodeToString(s) }

// Hash returns the lowercase hex SHA-256 of the encoded secret, the only form persisted for
// session tokens and OAuth state.
func (s Secret) Hash() string { return HashToken(s.Encode()) }

// HashToken hashes an encoded token as received from a cookie or query parameter.
func HashToken(encoded string) string {
	sum := sha256.Sum256([]byte(encoded))
	return hex.EncodeToString(sum[:])
}

// PKCEChallenge returns the S256 code challenge for a verifier (RFC 7636, section 4.2).
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ConstantTimeEqual compares two secrets without leaking their length difference position.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
