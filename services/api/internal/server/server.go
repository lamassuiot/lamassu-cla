// Package server composes middleware and handlers into the http.Handler that the Lambda entry
// point and the local development server serve. It imports no adapters: the wiring layer passes
// the use cases in.
package server

import (
	"net/http"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/handlers"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/middleware"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// BasePath is the prefix CloudFront routes to API Gateway (D-02). Every operation lives under it.
const BasePath = "/api"

// Config is the HTTP configuration.
type Config struct {
	// PortalOrigin is the portal's origin, for the Origin check and sign-in redirects (D-02).
	PortalOrigin string
	// SignInErrorRoute is the portal route that renders sign-in outcomes.
	SignInErrorRoute string
}

// New builds the handler: request IDs, the Origin check on state-changing requests, and the
// F-07 operations plus health, all under BasePath.
func New(cfg Config, auth *application.Auth, clock ports.Clock) (http.Handler, error) {
	authHandlers, err := handlers.NewAuth(handlers.AuthConfig{PortalOrigin: cfg.PortalOrigin, ErrorRoute: cfg.SignInErrorRoute}, auth, clock)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", handlers.Health)
	authHandlers.Register(mux)

	var h http.Handler = mux
	h = middleware.OriginCheck(cfg.PortalOrigin, handlers.RejectOrigin)(h)
	h = middleware.RequestID(h)

	root := http.NewServeMux()
	root.Handle(BasePath+"/", http.StripPrefix(BasePath, h))
	return root, nil
}
