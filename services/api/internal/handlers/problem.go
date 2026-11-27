// Package handlers translates HTTP and queue events into use-case calls and maps errors to the
// API error format. It never imports adapters or SDKs.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/middleware"
)

// Problem is RFC 9457 Problem Details with the platform extensions (api/conventions.md,
// section 6). It never carries stack traces, store details, or personal data.
type Problem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Code      string `json:"code"`
	Detail    string `json:"detail,omitempty"`
	RequestID string `json:"request_id"`
}

// Problem codes documented in specs/api/openapi.yaml for the F-07 operations.
const (
	CodeValidationFailed  = "VALIDATION_FAILED"
	CodeSessionRequired   = "SESSION_REQUIRED"
	CodeCSRFTokenInvalid  = "CSRF_TOKEN_INVALID" //nolint:gosec // problem code, not a credential
	CodeOriginNotAllowed  = "ORIGIN_NOT_ALLOWED"
	CodeSignInDisabled    = "SIGN_IN_DISABLED"
	CodeStoreUnavailable  = "STORE_UNAVAILABLE"
	CodeInternal          = "INTERNAL_ERROR"
	problemTypeAboutBlank = "about:blank"
)

// WriteProblem writes a problem response with no-store caching and the request ID header.
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, code, title string) {
	requestID := middleware.RequestIDFrom(r.Context())
	h := w.Header()
	h.Set("Content-Type", "application/problem+json")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Request-Id", requestID)
	if status == http.StatusServiceUnavailable {
		h.Set("Retry-After", strconv.Itoa(retryAfterSeconds))
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type: problemTypeAboutBlank, Title: title, Status: status, Code: code, RequestID: requestID,
	})
}

const retryAfterSeconds = 5

// writeAuthError maps the application's sentinel errors to problem responses for JSON endpoints.
func writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrSessionRequired):
		expireSessionCookie(w)
		WriteProblem(w, r, http.StatusUnauthorized, CodeSessionRequired, "Session required")
	case errors.Is(err, application.ErrStoreUnavailable), errors.Is(err, application.ErrStoreUncertain):
		WriteProblem(w, r, http.StatusServiceUnavailable, CodeStoreUnavailable, "Service unavailable")
	case errors.Is(err, application.ErrSignInDisabled):
		WriteProblem(w, r, http.StatusServiceUnavailable, CodeSignInDisabled, "Sign-in is not yet available")
	default:
		WriteProblem(w, r, http.StatusInternalServerError, CodeInternal, "Internal error")
	}
}
