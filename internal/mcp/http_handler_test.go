package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTPHandlerRequiresBearer(t *testing.T) {
	handler := HTTPHandler("http://127.0.0.1:1", "test")
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestHTTPHandlerForwardsTokenAndClientIP(t *testing.T) {
	fake := newFakeAPI()
	apiSrv := httptest.NewServer(fake.handler())
	t.Cleanup(apiSrv.Close)

	mcpSrv := httptest.NewServer(HTTPHandler(apiSrv.URL, "test"))
	t.Cleanup(mcpSrv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v0"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{
		Endpoint:             mcpSrv.URL,
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
		HTTPClient: &http.Client{Transport: headerRoundTripper{
			base: http.DefaultTransport,
			set: http.Header{
				"Authorization": []string{"Bearer test-token"},
				ClientIPHeader:  []string{"203.0.113.9"},
			},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "whoami"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatal(textOf(res))
	}
	if !strings.Contains(textOf(res), `"username": "ada"`) {
		t.Fatalf("whoami: %s", textOf(res))
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.lastAuth != "Bearer test-token" {
		t.Fatalf("api auth %q", fake.lastAuth)
	}
	if fake.lastForwardedFor != "203.0.113.9" {
		t.Fatalf("forwarded for %q", fake.lastForwardedFor)
	}
}

type headerRoundTripper struct {
	base http.RoundTripper
	set  http.Header
}

func (h headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	for key, values := range h.set {
		clone.Header[key] = append([]string(nil), values...)
	}
	base := h.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}
