package lambda

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func loadEvent(t *testing.T, name string) events.APIGatewayV2HTTPRequest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // fixed test fixture directory
	if err != nil {
		t.Fatal(err)
	}
	var e events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(raw, &e); err != nil {
		t.Fatal(err)
	}
	return e
}

// echo describes the request exactly as a handler sees it.
type echo struct {
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	RawPath    string            `json:"raw_path"`
	RawQuery   string            `json:"raw_query"`
	RequestURI string            `json:"request_uri"`
	Query      url.Values        `json:"query"`
	Host       string            `json:"host"`
	RemoteAddr string            `json:"remote_addr"`
	Cookies    map[string]string `json:"cookies"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	RequestID  string            `json:"request_id"`
	SourceIP   string            `json:"source_ip"`
	UserAgent  string            `json:"user_agent"`
	Stage      string            `json:"stage"`
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	cookies := map[string]string{}
	for _, c := range r.Cookies() {
		cookies[c.Name] = c.Value
	}
	headers := map[string]string{}
	for k := range r.Header {
		if k == "Cookie" {
			continue
		}
		headers[strings.ToLower(k)] = r.Header.Get(k)
	}
	rc, _ := RequestContextFrom(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(echo{
		Method: r.Method, Path: r.URL.Path, RawPath: r.URL.RawPath, RawQuery: r.URL.RawQuery, RequestURI: r.RequestURI,
		Query: r.URL.Query(), Host: r.Host, RemoteAddr: r.RemoteAddr, Cookies: cookies, Headers: headers,
		Body:      base64.StdEncoding.EncodeToString(body),
		RequestID: rc.RequestID, SourceIP: rc.SourceIP, UserAgent: rc.UserAgent, Stage: rc.Stage,
	})
}

func handle(t *testing.T, h http.HandlerFunc, name string) (events.APIGatewayV2HTTPResponse, echo) {
	t.Helper()
	resp, err := New(h).Handle(context.Background(), loadEvent(t, name))
	if err != nil {
		t.Fatal(err)
	}
	var e echo
	if resp.Body != "" && resp.Headers["Content-Type"] == "application/json" {
		if err := json.Unmarshal([]byte(resp.Body), &e); err != nil {
			t.Fatalf("decoding echo: %v", err)
		}
	}
	return resp, e
}

// Raw query preservation, including duplicate parameters that the callback must detect, and
// request cookies folded from the event's cookies array.
func TestCallbackDuplicateCodeAndCookies(t *testing.T) {
	_, e := handle(t, echoHandler, "callback-duplicate-code.json")
	if e.Method != http.MethodGet || e.Path != "/api/auth/github/callback" || e.RawQuery != "code=first&code=second&state=abc%3D%3D" {
		t.Fatalf("request = %+v", e)
	}
	if !reflect.DeepEqual(e.Query["code"], []string{"first", "second"}) || e.Query.Get("state") != "abc==" {
		t.Fatalf("query = %v", e.Query)
	}
	if e.Cookies["__Host-cla_oauth"] != "abc%3D%3D" || e.Cookies["other"] != "value" {
		t.Fatalf("cookies = %v", e.Cookies)
	}
	if _, leaked := e.Headers["cookie"]; leaked {
		t.Fatal("the raw cookie header must not be duplicated into Header")
	}
	if e.Host != "cla.example" || e.RemoteAddr != "192.0.2.11:0" {
		t.Fatalf("host=%q remote=%q", e.Host, e.RemoteAddr)
	}
}

// Encoded paths keep their raw form under the /api prefix; the decoded path is what the mux
// matches on.
func TestEncodedPath(t *testing.T) {
	_, e := handle(t, echoHandler, "encoded-path.json")
	if e.Path != "/api/v1/things/a/b c" || e.RawPath != "/api/v1/things/a%2Fb%20c" || e.RequestURI != "/api/v1/things/a%2Fb%20c" {
		t.Fatalf("path=%q raw=%q uri=%q", e.Path, e.RawPath, e.RequestURI)
	}
}

// Encoded query values are decoded by the handler, not by the adapter.
func TestStartReturnToQuery(t *testing.T) {
	_, e := handle(t, echoHandler, "start-with-return-to.json")
	if e.RawQuery != "return_to=%2Fagreements%2Ficla%20v2" || e.Query.Get("return_to") != "/agreements/icla v2" {
		t.Fatalf("raw=%q decoded=%q", e.RawQuery, e.Query.Get("return_to"))
	}
	if e.Headers["x-request-id"] != "01J9Z3K4Q8X7M2N5P6R7S8T9V0" || e.Headers["user-agent"] != "synthetic-browser/1.0" {
		t.Fatalf("headers = %v", e.Headers)
	}
}

// Invocation context propagation: request ID, source IP, user agent, stage.
func TestRequestContextPropagation(t *testing.T) {
	_, e := handle(t, echoHandler, "start-with-return-to.json")
	if e.RequestID != "synthetic-request-1" || e.SourceIP != "192.0.2.10" || e.UserAgent != "synthetic-browser/1.0" || e.Stage != "$default" {
		t.Fatalf("context = %+v", e)
	}
	if _, ok := RequestContextFrom(context.Background()); ok {
		t.Fatal("a bare context has no request context")
	}
}

// Base64 request bodies are decoded before the handler reads them.
func TestBase64RequestBody(t *testing.T) {
	_, e := handle(t, echoHandler, "post-base64-body.json")
	if e.Method != http.MethodPost || e.Body != "AAEC/w==" || e.Headers["content-type"] != "application/octet-stream" {
		t.Fatalf("request = %+v", e)
	}
	if e.Cookies["__Host-cla_session"] != "tok" || e.Headers["origin"] != "https://cla.example" || e.Headers["x-csrf-token"] != "csrf-1" {
		t.Fatalf("cookies=%v headers=%v", e.Cookies, e.Headers)
	}
}

// Multiple response cookies go to the cookies array, never into headers; a redirect has an
// empty body and no base64 flag.
func TestRedirectWithMultipleCookies(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "__Host-cla_session", Value: "tok", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 28800})
		http.SetCookie(w, &http.Cookie{Name: "__Host-cla_oauth", Value: "", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1})
		w.Header().Set("Cache-Control", "no-store")
		http.Redirect(w, r, "https://cla.example/agreements/", http.StatusFound)
	}
	resp, _ := handle(t, h, "callback-duplicate-code.json")
	want := events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Cache-Control": "no-store",
			"Location":      "https://cla.example/agreements/",
			"Content-Type":  "text/html; charset=utf-8",
		},
		Body: "<a href=\"https://cla.example/agreements/\">Found</a>.\n\n",
		Cookies: []string{
			"__Host-cla_session=tok; Path=/; Max-Age=28800; HttpOnly; Secure; SameSite=Lax",
			"__Host-cla_oauth=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Lax",
		},
	}
	sort.Strings(resp.Cookies)
	sort.Strings(want.Cookies)
	if !reflect.DeepEqual(resp, want) {
		t.Fatalf("response = %+v\nwant %+v", resp, want)
	}
	if _, inHeaders := resp.Headers["Set-Cookie"]; inHeaders {
		t.Fatal("Set-Cookie must not appear in headers")
	}
}

// Empty responses: 204 with no body, no content type, and an empty (not null) cookies array.
func TestEmptyResponse(t *testing.T) {
	h := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNoContent)
	}
	resp, _ := handle(t, h, "post-base64-body.json")
	want := events.APIGatewayV2HTTPResponse{StatusCode: 204, Headers: map[string]string{"Cache-Control": "no-store"}, Cookies: []string{}}
	if !reflect.DeepEqual(resp, want) {
		t.Fatalf("response = %+v, want %+v", resp, want)
	}
	raw, _ := json.Marshal(resp)
	if !strings.Contains(string(raw), `"cookies":[]`) || strings.Contains(string(raw), "isBase64Encoded") {
		t.Fatalf("json = %s", raw)
	}
}

// Binary bodies are base64-encoded; JSON and problem+json stay text; multi-valued headers other
// than Set-Cookie are joined.
func TestBodyEncodingAndHeaderJoining(t *testing.T) {
	bin := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Add("Vary", "Accept")
		w.Header().Add("Vary", "Cookie")
		_, _ = w.Write([]byte{0, 1, 2, 255})
	}
	resp, _ := handle(t, bin, "encoded-path.json")
	if !resp.IsBase64Encoded || resp.Body != "AAEC/w==" || resp.Headers["Vary"] != "Accept, Cookie" {
		t.Fatalf("binary response = %+v", resp)
	}
	problem := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"ORIGIN_NOT_ALLOWED"}`))
	}
	resp, _ = handle(t, problem, "encoded-path.json")
	if resp.IsBase64Encoded || resp.StatusCode != 403 || resp.Body != `{"code":"ORIGIN_NOT_ALLOWED"}` {
		t.Fatalf("problem response = %+v", resp)
	}
	// Implicit 200 when the handler writes without WriteHeader.
	implicit := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	}
	resp, _ = handle(t, implicit, "encoded-path.json")
	if resp.StatusCode != 200 || resp.Body != "ok" || resp.IsBase64Encoded {
		t.Fatalf("implicit response = %+v", resp)
	}
}

func TestMalformedEvents(t *testing.T) {
	a := New(http.HandlerFunc(echoHandler))
	cases := map[string]events.APIGatewayV2HTTPRequest{
		"no method":  {RawPath: "/api/v1/health"},
		"bad path":   {RawPath: "api", RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: "GET"}}},
		"bad escape": {RawPath: "/api/%zz", RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: "GET"}}},
		"bad base64": {RawPath: "/api/x", Body: "!!", IsBase64Encoded: true, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: "POST"}}},
	}
	for name, e := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := a.Handle(context.Background(), e); err == nil {
				t.Fatal("malformed event must be an error")
			}
		})
	}
}

// The adapter is transparent for a real server: the same handler answers identically through
// httptest and through an event.
func TestEquivalenceWithHTTPTest(t *testing.T) {
	h := http.HandlerFunc(echoHandler)
	e := loadEvent(t, "callback-duplicate-code.json")
	viaEvent, _ := New(h).Handle(context.Background(), e)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?code=first&code=second&state=abc%3D%3D", http.NoBody)
	req.Host = "cla.example"
	req.Header.Set("Cookie", "__Host-cla_oauth=abc%3D%3D; other=value")
	req.RemoteAddr = "192.0.2.11:0"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var a, b echo
	_ = json.Unmarshal([]byte(viaEvent.Body), &a)
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	// The request context fields exist only through the adapter.
	b.RequestID, b.SourceIP, b.UserAgent, b.Stage = a.RequestID, a.SourceIP, a.UserAgent, a.Stage
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("event view = %+v\nhttptest view = %+v", a, b)
	}
}
