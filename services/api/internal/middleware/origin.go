package middleware

import (
	"net/http"
	"strings"
)

// OriginCheck rejects state-changing requests whose Origin header is missing, "null", or not the
// portal origin (F-07, step 8 and AC-07-7; security specification, section 3.1). Safe methods pass
// through. The rejection is written by reject so that this package stays free of the problem
// format, which lives in handlers.
func OriginCheck(allowedOrigin string, reject func(w http.ResponseWriter, r *http.Request)) func(http.Handler) http.Handler {
	allowed := strings.TrimRight(allowedOrigin, "/")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			origin := r.Header.Get("Origin")
			if origin == "" || origin == "null" || strings.TrimRight(origin, "/") != allowed {
				reject(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isSafeMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}
