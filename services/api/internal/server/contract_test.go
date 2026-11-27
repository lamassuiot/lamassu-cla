package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// F07-T9: every response of the F-07 operations validates against specs/api/openapi.yaml
// (status declared, headers and body matching the schema).

func loadContract(t *testing.T) routers.Router {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "specs", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		t.Fatalf("loading contract: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("contract invalid: %v", err)
	}
	// Requests reach the router with the /api base already in the path; the contract's servers
	// entry is "/api", which gorillamux honours when matching.
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	return router
}

// validateExchange asserts that a request and its response are both described by the contract.
func validateExchange(t *testing.T, router routers.Router, req *http.Request, resp *reply, body []byte) {
	t.Helper()
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("%s %s not in contract: %v", req.Method, req.URL.Path, err)
	}
	if _, ok := route.Operation.Responses.Map()[strconv.Itoa(resp.StatusCode)]; !ok {
		t.Fatalf("%s %s: status %d not declared", req.Method, req.URL.Path, resp.StatusCode)
	}
	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req, PathParams: pathParams, Route: route,
			Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
		},
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   io.NopCloser(bytes.NewReader(body)),
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
		},
	}
	if err := openapi3filter.ValidateResponse(context.Background(), input); err != nil {
		t.Fatalf("%s %s -> %d does not match the contract: %v\nbody: %s", req.Method, req.URL.Path, resp.StatusCode, err, body)
	}
}

func (s *stack) request(method, path string, headers map[string]string, cookies ...*http.Cookie) *http.Request {
	req, _ := http.NewRequest(method, s.srv.URL+path, http.NoBody)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	// The contract router matches on the URL only; give it the path as the client saw it.
	req.URL, _ = url.Parse(path)
	req.URL.Scheme, req.URL.Host = "https", "cla.example"
	return req
}

func TestContract(t *testing.T) {
	router := loadContract(t)
	s := newStack(t, true)
	session, csrf := s.signIn(t, "/")
	start := s.do(t, http.MethodGet, "/api/v1/auth/github/start?return_to=%2F", nil)
	oauth := cookieNamed(start, "__Host-cla_oauth")

	cases := []struct {
		name, method, path string
		headers            map[string]string
		cookies            []*http.Cookie
		want               int
	}{
		{"health", http.MethodGet, "/api/v1/health", nil, nil, 200},
		{"start", http.MethodGet, "/api/v1/auth/github/start?return_to=%2Fagreements%2F", nil, nil, 302},
		{"start malformed", http.MethodGet, "/api/v1/auth/github/start?return_to=a&return_to=b", nil, nil, 400},
		{"callback cancelled", http.MethodGet, "/api/auth/github/callback?error=access_denied&state=" + url.QueryEscape(oauth.Value), nil, []*http.Cookie{oauth}, 302},
		{"callback malformed", http.MethodGet, "/api/auth/github/callback?code=x", nil, nil, 302},
		{"session", http.MethodGet, "/api/v1/session", nil, []*http.Cookie{session}, 200},
		{"session unauthorized", http.MethodGet, "/api/v1/session", nil, []*http.Cookie{reqCookie("__Host-cla_session", "x")}, 401},
		{"sign-out forbidden origin", http.MethodPost, "/api/v1/session/sign-out", map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{session}, 403},
		{"sign-out forbidden csrf", http.MethodPost, "/api/v1/session/sign-out", map[string]string{"Origin": portal}, []*http.Cookie{session}, 403},
		{"sign-out", http.MethodPost, "/api/v1/session/sign-out", map[string]string{"Origin": portal, "X-CSRF-Token": csrf}, []*http.Cookie{session}, 204},
		{"sign-out without session", http.MethodPost, "/api/v1/session/sign-out", map[string]string{"Origin": portal}, nil, 204},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := s.do(t, c.method, c.path, c.headers, c.cookies...)
			if resp.StatusCode != c.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, c.want)
			}
			body, _ := io.ReadAll(resp.Body)
			validateExchange(t, router, s.request(c.method, c.path, c.headers, c.cookies...), resp, body)
		})
	}

	// 503 responses: store outage on start and on the session read; disabled sign-in on start.
	s.attempts.Fail = ports.ErrUnavailable
	resp := s.do(t, http.MethodGet, "/api/v1/auth/github/start", nil)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 503 {
		t.Fatalf("start during outage = %d", resp.StatusCode)
	}
	validateExchange(t, router, s.request(http.MethodGet, "/api/v1/auth/github/start", nil), resp, body)
	s.attempts.Fail = nil

	disabled := newStack(t, false)
	resp = disabled.do(t, http.MethodGet, "/api/v1/auth/github/start", nil)
	body, _ = io.ReadAll(resp.Body)
	validateExchange(t, router, disabled.request(http.MethodGet, "/api/v1/auth/github/start", nil), resp, body)
}
