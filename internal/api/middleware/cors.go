package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware applies a conservative CORS policy.
// - In development (non-release) with no ALLOWED_ORIGINS set, all origins are allowed.
// - In release mode, only origins explicitly listed in ALLOWED_ORIGINS are allowed.
// ALLOWED_ORIGINS is a comma-separated list like "https://app.rexec.com,https://staging.rexec.com".
func CORSMiddleware() gin.HandlerFunc {
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := make(map[string]struct{})
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
			if len(allowedOrigins) == 0 {
				if !isRelease {
					c.Header("Access-Control-Allow-Origin", "*")
				}
			} else if _, ok := allowedOrigins[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
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

