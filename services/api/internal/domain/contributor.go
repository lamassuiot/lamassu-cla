package domain

import (
	"errors"
	"strings"
	"time"
)

// GitHubIdentity is what sign-in extracts from GitHub's GET /user response and nothing more
// (F-07, main flow step 4.5; D-29). UserID is the stable identity; Login is display data.
type GitHubIdentity struct {
	UserID string
	Login  string
}

// ErrInvalidIdentity is returned when GET /user lacks a usable numeric id or login.
var ErrInvalidIdentity = errors.New("identity: missing or invalid id or login")

// NewGitHubIdentity validates the two fields the platform keeps. The id is a positive integer
// kept as a decimal string to avoid precision loss; the login follows GitHub's username rules.
func NewGitHubIdentity(userID, login string) (GitHubIdentity, error) {
	if !isNumericID(userID) || !isGitHubLogin(login) {
		return GitHubIdentity{}, ErrInvalidIdentity
	}
	return GitHubIdentity{UserID: userID, Login: login}, nil
}

func isNumericID(s string) bool {
	if s == "" || len(s) > 20 || s[0] == '0' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isGitHubLogin accepts 1 to 39 alphanumeric characters or single hyphens, not at the ends.
func isGitHubLogin(s string) bool {
	if s == "" || len(s) > 39 || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") ||
		strings.Contains(s, "--") {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-':
		default:
			return false
		}
	}
	return true
}

// ContributorStatus is the contributor lifecycle (domain model, section 2.2).
type ContributorStatus string

const (
	ContributorPending   ContributorStatus = "PENDING"
	ContributorActive    ContributorStatus = "ACTIVE"
	ContributorSuspended ContributorStatus = "SUSPENDED"
	ContributorRevoked   ContributorStatus = "REVOKED"
)

// Contributor is the minimal record that sign-in creates or refreshes (D-29, FR-CON-01).
// Sign-in never changes Status or agreement history; it only refreshes GitHubLogin.
type Contributor struct {
	GitHubUserID string
	GitHubLogin  string
	Status       ContributorStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewContributor creates the PENDING record for a first sign-in.
func NewContributor(identity GitHubIdentity, now time.Time) Contributor {
	return Contributor{
		GitHubUserID: identity.UserID,
		GitHubLogin:  identity.Login,
		Status:       ContributorPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// RefreshLogin returns the contributor with the current login. Status and everything else
// are untouched (AC-07-12). It reports whether anything changed.
func (c Contributor) RefreshLogin(identity GitHubIdentity, now time.Time) (Contributor, bool) {
	if c.GitHubUserID != identity.UserID {
		return c, false
	}
	if c.GitHubLogin == identity.Login {
		return c, false
	}
	c.GitHubLogin = identity.Login
	c.UpdatedAt = now
	return c, true
}
