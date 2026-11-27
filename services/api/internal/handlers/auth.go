package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/middleware"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// AuthConfig configures the HTTP layer of F-07.
type AuthConfig struct {
	// PortalOrigin is the portal's origin (D-02), used for the Origin check and error redirects.
	PortalOrigin string
	// ErrorRoute is the portal route that renders sign-in outcomes (F-07, UI changes). The handler
	// appends ?reason=<code>; the portal decides the wording.
	ErrorRoute string
}

// Auth serves the four F-07 operations.
type Auth struct {
	cfg   AuthConfig
	uc    *application.Auth
	clock ports.Clock
}

// NewAuth wires the handlers.
func NewAuth(cfg AuthConfig, uc *application.Auth, clock ports.Clock) (*Auth, error) {
	if cfg.PortalOrigin == "" {
		return nil, errors.New("handlers: portal origin is required")
	}
	if cfg.ErrorRoute == "" {
		cfg.ErrorRoute = "/sign-in/error"
	}
	return &Auth{cfg: cfg, uc: uc, clock: clock}, nil
}

// Register mounts the operations on mux with paths relative to the /api base (D-02).
func (h *Auth) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/auth/github/start", h.start)
	mux.HandleFunc("GET /auth/github/callback", h.callback)
	mux.HandleFunc("GET /v1/session", h.getSession)
	mux.HandleFunc("POST /v1/session/sign-out", h.signOut)
}

// Reason codes appended to the portal error route. They carry no personal data.
const (
	reasonDisabled   = "disabled"
	reasonMalformed  = "malformed"
	reasonBinding    = "attempt_replaced"
	reasonState      = "state_invalid"
	reasonCancelled  = "cancelled"
	reasonUpstream   = "upstream_error"
	reasonScope      = "scope_unexpected"
	reasonStoreError = "service_unavailable"
)

func (h *Auth) start(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	q := r.URL.Query()
	if len(q["return_to"]) > 1 || len(q.Get("return_to")) > 2048 {
		WriteProblem(w, r, http.StatusBadRequest, CodeValidationFailed, "Malformed request")
		return
	}
	out, err := h.uc.StartSignIn(r.Context(), application.StartSignInInput{
		ReturnTo:      q.Get("return_to"),
		PreviousState: cookieValue(r, OAuthCookieName),
		RequestID:     middleware.RequestIDFrom(r.Context()),
	})
	if err != nil {
		writeAuthError(w, r, err)
		return
	}
	setOAuthCookie(w, out.State, time.Unix(out.ExpiresAt, 0), h.clock.Now())
	http.Redirect(w, r, out.RedirectURL, http.StatusFound)
}

func (h *Auth) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	q := r.URL.Query()
	malformed := len(q["code"]) > 1 || len(q["state"]) > 1 || len(q["error"]) > 1 ||
		len(q.Get("code")) > 256 || len(q.Get("state")) > 256 || len(q.Get("error")) > 128
	out, err := h.uc.CompleteSignIn(r.Context(), application.CompleteSignInInput{
		Code:                q.Get("code"),
		State:               q.Get("state"),
		Error:               q.Get("error"),
		CookieState:         cookieValue(r, OAuthCookieName),
		CurrentSessionToken: cookieValue(r, SessionCookieName),
		Malformed:           malformed,
		RequestID:           middleware.RequestIDFrom(r.Context()),
	})
	// The pre-authentication cookie is cleared on every outcome once the attempt was looked at;
	// on a binding mismatch the cookie belongs to a newer attempt and must survive (step 4.4).
	if !errors.Is(err, application.ErrBindingMismatch) && !errors.Is(err, application.ErrMalformedRequest) &&
		!errors.Is(err, application.ErrSignInDisabled) {
		expireOAuthCookie(w)
	}
	if err != nil {
		h.redirectError(w, r, err)
		return
	}
	setSessionCookie(w, out.SessionToken, time.Unix(out.ExpiresAt, 0), h.clock.Now())
	http.Redirect(w, r, h.cfg.PortalOrigin+out.ReturnTo, http.StatusFound)
}

func (h *Auth) redirectError(w http.ResponseWriter, r *http.Request, err error) {
	reason := reasonUpstream
	switch {
	case errors.Is(err, application.ErrSignInDisabled):
		reason = reasonDisabled
	case errors.Is(err, application.ErrMalformedRequest):
		reason = reasonMalformed
	case errors.Is(err, application.ErrBindingMismatch):
		reason = reasonBinding
	case errors.Is(err, application.ErrStateInvalid):
		reason = reasonState
	case errors.Is(err, application.ErrDenied):
		reason = reasonCancelled
	case errors.Is(err, application.ErrScopeUnexpected):
		reason = reasonScope
	case errors.Is(err, application.ErrStoreUnavailable), errors.Is(err, application.ErrStoreUncertain):
		reason = reasonStoreError
	}
	target := h.cfg.PortalOrigin + h.cfg.ErrorRoute + "?" + url.Values{"reason": {reason}}.Encode()
	http.Redirect(w, r, target, http.StatusFound)
}

// sessionResponse mirrors the Session schema in specs/api/openapi.yaml.
type sessionResponse struct {
	GitHubUserID string `json:"github_user_id"`
	GitHubLogin  string `json:"github_login"`
	ExpiresAt    string `json:"expires_at"`
	CSRFToken    string `json:"csrf_token"`
}

func (h *Auth) getSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	session, err := h.uc.Authenticate(r.Context(), application.AuthenticateInput{
		SessionToken: cookieValue(r, SessionCookieName),
		Extend:       false,
		RequestID:    middleware.RequestIDFrom(r.Context()),
	})
	if err != nil {
		writeAuthError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sessionResponse{
		GitHubUserID: session.GitHubUserID,
		GitHubLogin:  session.GitHubLogin,
		ExpiresAt:    session.ExpiresAt.UTC().Format(time.RFC3339),
		CSRFToken:    session.CSRFToken,
	})
}

// signOut runs behind the Origin middleware. With a valid session it requires the CSRF token;
// without a session it is idempotent (F-07, alternative flows).
func (h *Auth) signOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	token := cookieValue(r, SessionCookieName)
	if token != "" {
		session, err := h.uc.Authenticate(r.Context(), application.AuthenticateInput{
			SessionToken: token, Extend: false, RequestID: middleware.RequestIDFrom(r.Context()),
		})
		switch {
		case err == nil:
			if !application.VerifyCSRF(session, strings.TrimSpace(r.Header.Get("X-CSRF-Token"))) {
				WriteProblem(w, r, http.StatusForbidden, CodeCSRFTokenInvalid, "CSRF token invalid")
				return
			}
		case errors.Is(err, application.ErrSessionRequired):
			// Unknown or expired token: nothing to delete; fall through to the idempotent 204.
		default:
			writeAuthError(w, r, err)
			return
		}
	}
	if err := h.uc.SignOut(r.Context(), application.SignOutInput{
		SessionToken: token, RequestID: middleware.RequestIDFrom(r.Context()),
	}); err != nil {
		// A failed or uncertain deletion is never reported as success (AC-07-8).
		writeAuthError(w, r, err)
		return
	}
	expireSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// RejectOrigin is the Origin-middleware rejection in the API's problem format.
func RejectOrigin(w http.ResponseWriter, r *http.Request) {
	WriteProblem(w, r, http.StatusForbidden, CodeOriginNotAllowed, "Origin not allowed")
}
