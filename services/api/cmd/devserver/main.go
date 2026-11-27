// Command devserver runs the API locally for portal development (F-07 unit 2). It wires the use
// cases to in-memory stores and an in-process fake GitHub, so it needs no AWS account and no
// GitHub OAuth App; every sign-in completes for a synthetic contributor. It is never deployed.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/github/githubtest"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/adapters/memory"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/application"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/server"
)

type stdoutSecurityLog struct{}

func (stdoutSecurityLog) Log(_ context.Context, e ports.SecurityEvent) {
	// Security events carry no personal data or secrets by construction.
	log.Printf("security event=%s outcome=%s request_id=%s", e.Name, e.Outcome, e.RequestID)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	portal := flag.String("portal", "http://127.0.0.1:5173", "portal origin (Vite dev server)")
	flag.Parse()

	fake := githubtest.New()
	defer fake.Close()
	// The fake GitHub accepts any client credentials; these never leave the process.
	const devClientSecret = "dev-client-secret" //nolint:gosec // synthetic, fake GitHub only
	identity, err := github.NewOAuthAdapter(github.OAuthConfig{
		ClientID: "dev-client-id", ClientSecret: devClientSecret, WebBaseURL: fake.URL(), APIBaseURL: fake.URL(),
	}, fake.Client())
	if err != nil {
		log.Fatal(err)
	}
	clock := memory.Clock{}
	auth, err := application.NewAuth(application.AuthConfig{
		Enabled:     true,
		CallbackURL: *portal + server.BasePath + "/auth/github/callback",
		ReturnTo:    domain.ReturnToPolicy{Allowed: []string{"/", "/agreements/", "/profile"}},
		Session:     domain.DefaultSessionPolicy,
	}, application.AuthDependencies{
		Clock: clock, Random: memory.Random{}, Identity: identity,
		Attempts: memory.NewSignInAttempts(), Sessions: memory.NewSessions(),
		Contributors: memory.NewContributors(), Security: stdoutSecurityLog{},
	})
	if err != nil {
		log.Fatal(err)
	}
	handler, err := server.New(server.Config{PortalOrigin: *portal, SignInErrorRoute: "/sign-in/error"}, auth, clock)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr: *addr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("devserver listening on http://%s%s (portal origin %s, fake GitHub %s)\n", *addr, server.BasePath, *portal, fake.URL())
	fmt.Println("note: cookies are Secure; use the portal's HTTPS dev proxy or a browser that treats 127.0.0.1 as secure")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
