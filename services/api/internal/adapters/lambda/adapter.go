// Package lambda adapts API Gateway HTTP API events (payload format 2.0) to the net/http handler
// in internal/server and back. It is owned and maintained by this repository (A-31): the archived
// upstream library it replaces is not used, and the handlers and use cases know nothing about
// Lambda.
//
// Supported scope:
//   - API Gateway HTTP API, payload format 2.0 only (events.APIGatewayV2HTTPRequest/Response).
//     REST API (payload 1.0), ALB, and Lambda function URLs are out of scope.
//   - The raw path and raw query string are used verbatim, so percent-encoding and duplicate
//     query parameters (which the callback must detect) reach the handler unchanged.
//   - Request cookies arrive in the event's cookies array and are folded into one Cookie header;
//     response Set-Cookie headers are returned in the response cookies array, one per cookie.
//   - Request bodies are decoded when isBase64Encoded; response bodies are returned as text for
//     JSON, problem+json, and text/* content types and base64-encoded otherwise.
//   - Headers other than Set-Cookie with several values are joined with ", " (RFC 9110, 5.3).
//   - The API Gateway request context (request ID, source IP, user agent) is available to
//     handlers through RequestContextFrom; the Lambda invocation context is the request context.
package lambda

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// RequestContext is the API Gateway request context that handlers may read.
type RequestContext struct {
	RequestID string
	SourceIP  string
	UserAgent string
	Stage     string
}

type contextKey int

const requestContextKey contextKey = iota

// RequestContextFrom returns the API Gateway request context stored by the adapter.
func RequestContextFrom(ctx context.Context) (RequestContext, bool) {
	rc, ok := ctx.Value(requestContextKey).(RequestContext)
	return rc, ok
}

// Adapter serves events with an http.Handler.
type Adapter struct {
	handler http.Handler
}

// New returns an adapter for handler.
func New(handler http.Handler) *Adapter { return &Adapter{handler: handler} }

// Handle is the Lambda handler signature for API Gateway HTTP API integrations.
func (a *Adapter) Handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	req, err := ToRequest(ctx, event)
	if err != nil {
		// A malformed event is a configuration or platform fault, not a client error.
		return events.APIGatewayV2HTTPResponse{}, err
	}
	rec := newRecorder()
	a.handler.ServeHTTP(rec, req)
	return rec.response(), nil
}

// ToRequest builds the *http.Request the handler sees. Exported for tests of other packages.
func ToRequest(ctx context.Context, e events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	method := e.RequestContext.HTTP.Method
	if method == "" {
		return nil, fmt.Errorf("lambda: event has no HTTP method")
	}
	rawPath := e.RawPath
	if rawPath == "" {
		rawPath = e.RequestContext.HTTP.Path
	}
	if !strings.HasPrefix(rawPath, "/") {
		return nil, fmt.Errorf("lambda: event path %q is not absolute", rawPath)
	}
	decodedPath, err := url.PathUnescape(rawPath)
	if err != nil {
		return nil, fmt.Errorf("lambda: event path is not valid percent-encoding: %w", err)
	}
	u := &url.URL{Path: decodedPath, RawQuery: e.RawQueryString}
	if decodedPath != rawPath {
		u.RawPath = rawPath
	}

	var body io.Reader = http.NoBody
	if e.Body != "" {
		if e.IsBase64Encoded {
			decoded, err := base64.StdEncoding.DecodeString(e.Body)
			if err != nil {
				return nil, fmt.Errorf("lambda: event body is not valid base64: %w", err)
			}
			body = bytes.NewReader(decoded)
		} else {
			body = strings.NewReader(e.Body)
		}
	}

	rc := RequestContext{
		RequestID: e.RequestContext.RequestID,
		SourceIP:  e.RequestContext.HTTP.SourceIP,
		UserAgent: e.RequestContext.HTTP.UserAgent,
		Stage:     e.RequestContext.Stage,
	}
	req, err := http.NewRequestWithContext(context.WithValue(ctx, requestContextKey, rc), method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("lambda: building request: %w", err)
	}
	// url.String() re-encodes; keep the exact raw forms the client sent.
	req.URL = u
	req.RequestURI = u.RequestURI()
	for k, v := range e.Headers {
		// Like net/http's server: cookies come from the dedicated array and Host lives in req.Host.
		if strings.EqualFold(k, "cookie") || strings.EqualFold(k, "host") {
			continue
		}
		req.Header.Set(k, v)
	}
	if len(e.Cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(e.Cookies, "; "))
	}
	if host := e.Headers["host"]; host != "" {
		req.Host = host
	} else if e.RequestContext.DomainName != "" {
		req.Host = e.RequestContext.DomainName
	}
	if rc.SourceIP != "" {
		req.RemoteAddr = rc.SourceIP + ":0"
	}
	if e.RequestContext.HTTP.Protocol != "" {
		req.Proto = e.RequestContext.HTTP.Protocol
	}
	return req, nil
}

// recorder captures the handler's response.
type recorder struct {
	status int
	header http.Header
	body   bytes.Buffer
	wrote  bool
}

func newRecorder() *recorder { return &recorder{status: http.StatusOK, header: http.Header{}} }

func (r *recorder) Header() http.Header { return r.header }

func (r *recorder) WriteHeader(status int) {
	if r.wrote {
		return
	}
	r.wrote = true
	r.status = status
}

func (r *recorder) Write(p []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(p)
}

func (r *recorder) response() events.APIGatewayV2HTTPResponse {
	resp := events.APIGatewayV2HTTPResponse{StatusCode: r.status, Headers: map[string]string{}}
	for name, values := range r.header {
		if strings.EqualFold(name, "Set-Cookie") {
			resp.Cookies = append(resp.Cookies, values...)
			continue
		}
		resp.Headers[name] = strings.Join(values, ", ")
	}
	if resp.Cookies == nil {
		resp.Cookies = []string{}
	}
	if r.body.Len() == 0 {
		return resp
	}
	if isText(r.header.Get("Content-Type")) {
		resp.Body = r.body.String()
	} else {
		resp.Body = base64.StdEncoding.EncodeToString(r.body.Bytes())
		resp.IsBase64Encoded = true
	}
	return resp
}

// isText reports whether a content type is returned as text rather than base64.
func isText(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	switch {
	case ct == "application/json", ct == "application/problem+json":
		return true
	case strings.HasPrefix(ct, "text/"):
		return true
	case strings.HasSuffix(ct, "+json"):
		return true
	}
	return false
}
