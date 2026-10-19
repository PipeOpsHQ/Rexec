package middleware

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// APITokenAllowed reports whether an rexec_ API token with scopes may call
// method+path. Browser JWTs never go through this check.
//
//   - /api/admin and /ws/admin require the admin scope
//   - agent websocket, agent list/delete, and token validate allow the agent scope
//   - every other route requires read, write, or admin
//
// An agent-scoped host token therefore cannot open the admin UI or call user APIs.
func APITokenAllowed(method, rawPath string, scopes []string) bool {
	method = strings.ToUpper(method)
	p := cleanRequestPath(rawPath)
	have := map[string]struct{}{}
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if scope != "" {
			have[scope] = struct{}{}
		}
	}
	if isAdminPath(p) {
		_, ok := have["admin"]
		return ok
	}
	if _, ok := have["agent"]; ok && agentScopePath(method, p) {
		return true
	}
	_, read := have["read"]
	_, write := have["write"]
	_, admin := have["admin"]
	// Mutating methods need write or admin. read is safe methods only.
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		return write || admin
	}
	return read || write || admin
}

// AllowedAPITokenScopes is the closed set accepted at token creation.
var AllowedAPITokenScopes = map[string]struct{}{
	"read":  {},
	"write": {},
	"admin": {},
	"agent": {},
}

// NormalizeAPITokenScopes lowercases, de-duplicates, and rejects unknown scopes.
// An empty input defaults to read and write.
func NormalizeAPITokenScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return []string{"read", "write"}, nil
	}
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if scope == "" {
			continue
		}
		if _, ok := AllowedAPITokenScopes[scope]; !ok {
			return nil, fmt.Errorf("invalid scope %q", scope)
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid scopes")
	}
	return out, nil
}

func cleanRequestPath(raw string) string {
	if raw == "" {
		return "/"
	}
	p := path.Clean(raw)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func isAdminPath(p string) bool {
	return p == "/api/admin" || strings.HasPrefix(p, "/api/admin/") ||
		p == "/ws/admin" || strings.HasPrefix(p, "/ws/admin/")
}

func agentScopePath(method, p string) bool {
	switch {
	case (method == http.MethodGet || method == http.MethodHead) && p == "/api/tokens/validate":
		return true
	case (method == http.MethodGet || method == http.MethodHead) && p == "/api/agents":
		return true
	case strings.HasPrefix(p, "/api/agents/"):
		return method == http.MethodGet || method == http.MethodHead || method == http.MethodDelete
	case method == http.MethodGet && strings.HasPrefix(p, "/ws/agent/") && !strings.Contains(p, "/terminal"):
		return true
	default:
		return false
	}
}

func apiTokenScopes(c *gin.Context) []string {
	v, ok := c.Get("api_token_scopes")
	if !ok || v == nil {
		return nil
	}
	scopes, _ := v.([]string)
	return scopes
}

// rejectAPITokenWithoutAdmin stops an API token that belongs to an admin user
// from calling admin routes unless the token itself has the admin scope.
// Browser JWTs are unchanged. Returns true when the request was aborted.
func rejectAPITokenWithoutAdmin(c *gin.Context) bool {
	if !c.GetBool("api_token") {
		return false
	}
	for _, scope := range apiTokenScopes(c) {
		if strings.EqualFold(strings.TrimSpace(scope), "admin") {
			return false
		}
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
	return true
}
