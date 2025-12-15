package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Default allowed origins for Rexec across all domains
var defaultAllowedOrigins = []string{
	"https://rexec.pipeops.app",
	"https://rexec.pipeops.io",
	"https://rexec.pipeops.sh",
	"https://rexec.io",
	"https://rexec.sh",
	"https://rexec.cloud",
}

// CORSMiddleware applies a conservative CORS policy.
// - In development (non-release) with no ALLOWED_ORIGINS set, all origins are allowed.
// - In release mode, only origins explicitly listed in ALLOWED_ORIGINS (plus default Rexec domains) are allowed.
// ALLOWED_ORIGINS is a comma-separated list like "https://app.rexec.com,https://staging.rexec.com".
func CORSMiddleware() gin.HandlerFunc {
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := make(map[string]struct{})

	// Add default Rexec domains
	for _, origin := range defaultAllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}

	// Add any additional origins from environment
	if allowedOriginsStr != "" {
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			o := strings.TrimSpace(origin)
			if o != "" {
				allowedOrigins[o] = struct{}{}
			}
		}
	}
	isRelease := os.Getenv("GIN_MODE") == "release"

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
			} else if !isRelease {
				// In development, allow any origin
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
