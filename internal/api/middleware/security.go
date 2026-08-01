package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-XSS-Protection is deprecated and unsafe in legacy browsers; explicitly disable.
		c.Header("X-XSS-Protection", "0")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// HSTS (Strict-Transport-Security) - Enforce HTTPS
		// max-age=31536000 (1 year), includeSubDomains
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy (CSP).
		// 'unsafe-inline' on script-src is needed for inline JSON-LD blocks in index.html;
		// removing it requires migrating to nonces/hashes.
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://js.stripe.com https://cdnjs.cloudflare.com https://eu.i.posthog.com https://*.posthog.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com; " +
			"img-src 'self' data: blob: https:; " +
			"font-src 'self' data: https://fonts.gstatic.com https://cdnjs.cloudflare.com; " +
			"connect-src 'self' wss: https://eu.i.posthog.com https://*.posthog.com https://api.stripe.com; " +
			"frame-src 'self' https://js.stripe.com https://www.youtube.com https://youtube.com https://player.vimeo.com https://screen.studio https://www.loom.com; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
