package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"

	sdkauth "github.com/modelcontextprotocol/go-sdk/auth"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClientIPHeader is set by the API server to the caller address before the
// MCP handler runs. Tools forward it so loopback API calls keep that address
// for rate limits and audit logs.
const ClientIPHeader = "X-Rexec-Client-IP"

// HTTPHandler serves the streamable HTTP MCP transport at a single path.
// apiBase is the Rexec API origin the tools call, usually http://127.0.0.1:PORT
// when this handler is mounted on the API process itself.
//
// Every request must send Authorization: Bearer. The token is checked by the
// API on each tool call. The session is bound to a hash of the token so a
// later request cannot reuse another caller's session.
func HTTPHandler(apiBase, version string) http.Handler {
	apiBase = strings.TrimRight(strings.TrimSpace(apiBase), "/")
	if version == "" {
		version = "dev"
	}
	inner := sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		token := bearerToken(r)
		if token == "" {
			return nil
		}
		client, err := NewClient(apiBase, token, "rexec-mcp/"+version)
		if err != nil {
			log.Printf("rexec-mcp: %v", err)
			return nil
		}
		client.ForwardedFor = strings.TrimSpace(r.Header.Get(ClientIPHeader))
		return NewServer(client, version)
	}, nil)
	return sdkauth.RequireBearerToken(func(_ context.Context, token string, _ *http.Request) (*sdkauth.TokenInfo, error) {
		sum := sha256.Sum256([]byte(token))
		return &sdkauth.TokenInfo{UserID: hex.EncodeToString(sum[:])}, nil
	}, &sdkauth.RequireBearerTokenOptions{AllowMissingExpiration: true})(inner)
}

func bearerToken(r *http.Request) string {
	fields := strings.Fields(r.Header.Get("Authorization"))
	if len(fields) == 2 && strings.EqualFold(fields[0], "bearer") && fields[1] != "" {
		return fields[1]
	}
	return ""
}
