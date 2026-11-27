// Package githubtest provides an in-process stand-in for github.com and api.github.com. F-07
// adapters and handlers are tested against it; nothing in the repository talks to real GitHub
// before deployment.
package githubtest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
)

// Revocation records one DELETE /applications/{client_id}/token call.
type Revocation struct {
	Authorization string
	Body          string
}

// Fake is a configurable fake GitHub. Fields are read when each request is served, so tests may
// change them between requests; use the mutex when a test and a request may run concurrently.
type Fake struct {
	Mu sync.Mutex

	ExchangeForms []url.Values
	UserAuth      []string
	Revocations   []Revocation

	TokenStatus  int
	TokenBody    map[string]any
	UserStatus   int
	UserBody     map[string]any
	UserScopes   *string
	RevokeStatus int

	Server *httptest.Server
}

// New starts a fake that answers a scopeless sign-in for a synthetic contributor.
func New() *Fake {
	empty := ""
	f := &Fake{
		TokenStatus:  http.StatusOK,
		TokenBody:    map[string]any{"access_token": "gho_fake", "scope": "", "token_type": "bearer"},
		UserStatus:   http.StatusOK,
		UserBody:     map[string]any{"id": 1234567, "login": "synthetic-contributor", "email": "public@example.org", "name": "Synthetic"},
		UserScopes:   &empty,
		RevokeStatus: http.StatusNoContent,
	}
	f.Server = httptest.NewServer(f.handler())
	return f
}

// Close stops the server.
func (f *Fake) Close() { f.Server.Close() }

// URL is the base URL for both the web and the API origin.
func (f *Fake) URL() string { return f.Server.URL }

// Client is an HTTP client for the fake.
func (f *Fake) Client() *http.Client { return f.Server.Client() }

func (f *Fake) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.Mu.Lock()
		f.ExchangeForms = append(f.ExchangeForms, r.PostForm)
		status, body := f.TokenStatus, f.TokenBody
		f.Mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		f.Mu.Lock()
		f.UserAuth = append(f.UserAuth, r.Header.Get("Authorization"))
		status, body, scopes := f.UserStatus, f.UserBody, f.UserScopes
		f.Mu.Unlock()
		if scopes != nil {
			w.Header().Set("X-OAuth-Scopes", *scopes)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("DELETE /applications/{client_id}/token", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
		f.Mu.Lock()
		f.Revocations = append(f.Revocations, Revocation{Authorization: r.Header.Get("Authorization"), Body: string(body)})
		status := f.RevokeStatus
		f.Mu.Unlock()
		w.WriteHeader(status)
	})
	return mux
}

// String returns a pointer to s, for the scope fields.
func String(s string) *string { return &s }
