// Package middleware provides correlation IDs, authentication, authorization, logging, recovery,
// and body limits. It never imports adapters or SDKs.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
)

type contextKey int

const requestIDKey contextKey = iota

// requestIDPattern accepts client-supplied ULIDs or UUIDs (api/conventions.md, section 8).
var requestIDPattern = regexp.MustCompile(`^[0-9A-Za-z-]{26,36}$`)

// RequestID accepts a valid client X-Request-Id or generates one, stores it in the context, and
// echoes it on the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

// RequestIDFrom returns the request ID stored by RequestID, or an empty string.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
