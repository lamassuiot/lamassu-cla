package handlers

import (
	"net/http"
	"time"
)

// Cookie names fixed by F-07 (main flow steps 2 and 6). The __Host- prefix requires Secure,
// Path=/, and no Domain; browsers reject the cookie otherwise, so these attributes are not
// configurable.
const (
	SessionCookieName = "__Host-cla_session"
	OAuthCookieName   = "__Host-cla_oauth"
)

func setOAuthCookie(w http.ResponseWriter, state string, expiresAt time.Time, now time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: OAuthCookieName, Value: state, Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: int(expiresAt.Sub(now).Seconds()),
	})
}

func expireOAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: OAuthCookieName, Value: "", Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time, now time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName, Value: token, Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: int(expiresAt.Sub(now).Seconds()),
	})
}

func expireSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName, Value: "", Path: "/", Secure: true, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}
