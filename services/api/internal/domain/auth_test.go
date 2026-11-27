package domain

import (
	"testing"
	"time"
)

// F07-T4: return_to canonicalization and allowlist (AC-07-4).
func TestReturnToPolicyResolve(t *testing.T) {
	policy := ReturnToPolicy{Allowed: []string{"/", "/agreements/", "/profile"}}
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty falls back", "", "/"},
		{"root", "/", "/"},
		{"allowed exact", "/profile", "/profile"},
		{"allowed descendant", "/agreements/icla/v2", "/agreements/icla/v2"},
		{"not allowed path", "/admin", "/"},
		{"absolute URL", "https://evil.example/", "/"},
		{"protocol-relative", "//evil.example/agreements/", "/"},
		{"backslash protocol-relative", "/\\evil.example", "/"},
		{"backslash inside", "/agreements/\\x", "/"},
		{"encoded protocol-relative", "%2F%2Fevil.example", "/"},
		{"double encoded slash", "/agreements/%252F", "/"},
		{"encoded backslash", "/agreements/%5Cx", "/"},
		{"control character", "/agreements/\x01", "/"},
		{"encoded control character", "/agreements/%0A", "/"},
		{"no leading slash", "agreements/", "/"},
		{"scheme without slashes", "javascript:alert(1)", "/"},
		{"dot segments", "/agreements/../admin", "/"},
		{"query stripped", "/profile?x=1", "/profile"},
		{"fragment stripped", "/profile#top", "/profile"},
		{"encoded allowed path", "/agreements/icla%2Dv2", "/agreements/icla-v2"},
		{"encoded space rejected", "/agreements/icla%20v2", "/"},
		{"malformed escape", "/agreements/%zz", "/"},
		{"trailing dot segment", "/agreements/..", "/"},
		{"too long", "/" + string(make([]byte, 2100)), "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policy.Resolve(tt.in); got != tt.want {
				t.Fatalf("Resolve(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNewGitHubIdentity(t *testing.T) {
	tests := []struct {
		id, login string
		ok        bool
	}{
		{"1234567", "synthetic-contributor", true},
		{"1", "a", true},
		{"0123", "a", false},
		{"", "a", false},
		{"12a", "a", false},
		{"123456789012345678901", "a", false},
		{"1", "", false},
		{"1", "-starts", false},
		{"1", "ends-", false},
		{"1", "double--hyphen", false},
		{"1", "has space", false},
		{"1", string(make([]byte, 40)), false},
	}
	for _, tt := range tests {
		_, err := NewGitHubIdentity(tt.id, tt.login)
		if (err == nil) != tt.ok {
			t.Errorf("NewGitHubIdentity(%q, %q) ok=%v, want %v", tt.id, tt.login, err == nil, tt.ok)
		}
	}
}

// Supports F07-T12: a login refresh changes nothing but the login (AC-07-12).
func TestContributorRefreshLogin(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	identity, _ := NewGitHubIdentity("42", "old-login")
	c := NewContributor(identity, t0)
	c.Status = ContributorSuspended

	same, changed := c.RefreshLogin(identity, t0.Add(time.Hour))
	if changed || same != c {
		t.Fatalf("same login must not change the record")
	}
	renamed, _ := NewGitHubIdentity("42", "new-login")
	got, changed := c.RefreshLogin(renamed, t0.Add(time.Hour))
	if !changed || got.GitHubLogin != "new-login" || got.Status != ContributorSuspended ||
		!got.CreatedAt.Equal(t0) || !got.UpdatedAt.Equal(t0.Add(time.Hour)) {
		t.Fatalf("RefreshLogin = %+v, want only login and UpdatedAt changed", got)
	}
	other, _ := NewGitHubIdentity("43", "old-login")
	if _, changed := c.RefreshLogin(other, t0); changed {
		t.Fatalf("a different user ID must never refresh this record")
	}
}

// F07-T6 (domain part): idle and absolute expiry, sliding renewal, throttle, cap (AC-07-6).
func TestSessionValidAndTouch(t *testing.T) {
	policy := DefaultSessionPolicy
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	identity, _ := NewGitHubIdentity("42", "login")
	s := NewSession("hash", "csrf", identity, policy, t0)

	if s.Valid(t0) != nil || s.Valid(t0.Add(29*time.Minute)) != nil {
		t.Fatalf("fresh session must be valid")
	}
	if s.Valid(t0.Add(30*time.Minute)) == nil {
		t.Fatalf("idle expiry must be exclusive")
	}
	if _, due := s.Touch(policy, t0.Add(30*time.Second)); due {
		t.Fatalf("touch within the throttle window must not write")
	}
	touched, due := s.Touch(policy, t0.Add(time.Minute))
	if !due || !touched.IdleExpiresAt.Equal(t0.Add(31*time.Minute)) || !touched.LastSeenAt.Equal(t0.Add(time.Minute)) {
		t.Fatalf("touch after one minute must slide idle expiry: %+v", touched)
	}
	if _, due := s.Touch(policy, t0.Add(31*time.Minute)); due {
		t.Fatalf("an expired session must not be renewed")
	}
	late := s
	late.LastSeenAt = t0.Add(7*time.Hour + 40*time.Minute)
	late.IdleExpiresAt = late.LastSeenAt.Add(policy.Idle)
	capped, due := late.Touch(policy, t0.Add(7*time.Hour+45*time.Minute))
	if !due || !capped.IdleExpiresAt.Equal(s.ExpiresAt) {
		t.Fatalf("idle expiry must be capped by the absolute expiry: %v", capped.IdleExpiresAt)
	}
	if s.Valid(t0.Add(8*time.Hour)) == nil {
		t.Fatalf("absolute expiry must apply")
	}
}

func TestSessionPolicyValidate(t *testing.T) {
	if DefaultSessionPolicy.Validate() != nil {
		t.Fatal("default policy must be valid")
	}
	if (SessionPolicy{Idle: time.Hour, Absolute: time.Minute}).Validate() == nil {
		t.Fatal("idle longer than absolute must be rejected")
	}
	if (SessionPolicy{}).Validate() == nil {
		t.Fatal("zero timeouts must be rejected")
	}
}

func TestSecretHelpers(t *testing.T) {
	s := Secret(make([]byte, SecretBytes))
	if len(s.Encode()) != 43 {
		t.Fatalf("32 bytes must encode to 43 URL-safe characters, got %d", len(s.Encode()))
	}
	if s.Hash() != HashToken(s.Encode()) || len(s.Hash()) != 64 {
		t.Fatal("Hash must be the hex SHA-256 of the encoded form")
	}
	// RFC 7636 appendix B.
	if got := PKCEChallenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("PKCEChallenge = %q", got)
	}
	if !ConstantTimeEqual("a", "a") || ConstantTimeEqual("a", "b") || ConstantTimeEqual("a", "aa") {
		t.Fatal("ConstantTimeEqual")
	}
}
