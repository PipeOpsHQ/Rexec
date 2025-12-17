package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rexec/rexec/internal/auth"
	"github.com/rexec/rexec/internal/models"
	"github.com/rexec/rexec/internal/storage"
)

const wsTokenProtocolPrefix = "rexec.token."

func tokenFromWebSocketSubprotocolHeader(headerVal string) string {
	if headerVal == "" {
		return ""
	}
	for _, part := range strings.Split(headerVal, ",") {
		proto := strings.TrimSpace(part)
		if strings.HasPrefix(proto, wsTokenProtocolPrefix) {
			token := strings.TrimPrefix(proto, wsTokenProtocolPrefix)
			if token != "" {
				return token
			}
		}
	}
	return ""
}

// AuthMiddleware validates JWT or API tokens and extracts user info, enforcing MFA if enabled.
// jwtSecret must be the server's signing key.
func AuthMiddleware(store *storage.PostgresStore, mfaService *auth.MFAService, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authStart := time.Now()
		path := c.Request.URL.Path
		isWebSocket := websocket.IsWebSocketUpgrade(c.Request)

		// Log WebSocket connection attempts for debugging
		if isWebSocket {
			log.Printf("[Auth] WebSocket auth request for %s from %s", path, c.ClientIP())
		}

		// Create a context with timeout for all DB operations in auth
		// This prevents hanging if DB is slow or connection pool is exhausted
		dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer dbCancel()

		// Get token from Authorization header or query params
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		if authHeader != "" {
			parts := strings.Fields(authHeader)
			if len(parts) >= 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" && isWebSocket {
			// For browser WebSockets, pass the token in a header instead of the URL:
			// Sec-WebSocket-Protocol: rexec.v1, rexec.token.<token>
			tokenString = tokenFromWebSocketSubprotocolHeader(c.GetHeader("Sec-WebSocket-Protocol"))

			// Backwards compatibility: allow token in query params for legacy WebSocket clients.
			if tokenString == "" {
				tokenQuery := c.Query("token")
				if tokenQuery != "" {
					tokenString = tokenQuery
				}
			}
		}

		if tokenString == "" {
			if isWebSocket {
				log.Printf("[Auth] WebSocket auth failed for %s: no token provided", path)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Log token type for debugging (don't log actual token)
		tokenType := "JWT"
		if strings.HasPrefix(tokenString, "rexec_") {
			tokenType = "API"
		}
		if isWebSocket {
			log.Printf("[Auth] WebSocket auth with %s token for %s", tokenType, path)
		}

		// Check if this is an API token (starts with rexec_)
		if strings.HasPrefix(tokenString, "rexec_") {
			// Validate API token
			apiToken, err := store.ValidateAPIToken(dbCtx, tokenString)
			if err != nil {
				if isWebSocket {
					log.Printf("[Auth] WebSocket API token validation failed for %s after %v: %v", path, time.Since(authStart), err)
				}
				// Don't block on audit log - fire and forget
				go func() {
					auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					store.CreateAuditLog(auditCtx, &models.AuditLog{
						ID:        uuid.New().String(),
						UserID:    nil,
						Action:    "api_token_auth_failed",
						IPAddress: c.ClientIP(),
						UserAgent: c.Request.UserAgent(),
						Details:   fmt.Sprintf("Invalid API token: %v", err),
						CreatedAt: time.Now(),
					})
				}()
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API token"})
				c.Abort()
				return
			}

			// Get user info
			user, err := store.GetUserByID(dbCtx, apiToken.UserID)
			if err != nil {
				if isWebSocket {
					log.Printf("[Auth] WebSocket user lookup failed for %s after %v: %v", path, time.Since(authStart), err)
				}
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				c.Abort()
				return
			}

			// Set user info in context
			c.Set("userID", user.ID)
			c.Set("email", user.Email)
			c.Set("username", user.Username)
			c.Set("tier", user.Tier)
			c.Set("guest", false)
			c.Set("subscription_active", user.SubscriptionActive)
			c.Set("api_token", true)
			c.Set("api_token_scopes", apiToken.Scopes)

			if isWebSocket {
				log.Printf("[Auth] WebSocket API token auth successful for %s in %v", path, time.Since(authStart))
			}
			c.Next()
			return
		}

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || token == nil {
			if isWebSocket {
				log.Printf("[Auth] WebSocket JWT parse failed for %s after %v: %v", path, time.Since(authStart), err)
			}
			// Don't block on audit log
			go func() {
				auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				store.CreateAuditLog(auditCtx, &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    nil,
					Action:    "authentication_failed",
					IPAddress: c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
					Details:   fmt.Sprintf("Failed to parse token: %v", err),
					CreatedAt: time.Now(),
				})
			}()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			if isWebSocket {
				log.Printf("[Auth] WebSocket JWT claims invalid for %s after %v", path, time.Since(authStart))
			}
			// Don't block on audit log
			go func() {
				auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				store.CreateAuditLog(auditCtx, &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    nil, // No user ID yet
					Action:    "authentication_failed",
					IPAddress: c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
					Details:   fmt.Sprintf("Invalid or expired token: %v", err),
					CreatedAt: time.Now(),
				})
			}()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims and set user info in context
		userID, userID_ok := claims["user_id"].(string)
		email, email_ok := claims["email"].(string)
		exp, exp_ok := claims["exp"].(float64)
		iat, _ := claims["iat"].(float64)
		sessionID, _ := claims["sid"].(string)

		// Check if this is an MFA pending token
		mfaPending, _ := claims["mfa_pending"].(bool)
		if mfaPending {
			// MFA pending tokens only have user_id, email, exp, and mfa_pending
			if !userID_ok || !email_ok || !exp_ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA token claims"})
				c.Abort()
				return
			}
			c.Set("userID", userID)
			c.Set("email", email)
			c.Set("tokenExp", int64(exp))
			c.Set("mfa_pending", true)
			c.Next()
			return
		}

		// Regular token - needs all claims
		username, username_ok := claims["username"].(string)
		tier, tier_ok := claims["tier"].(string)
		guest, guest_ok := claims["guest"].(bool)
		subActive, subActive_ok := claims["subscription_active"].(bool)

		if !userID_ok || !email_ok || !username_ok || !tier_ok || !guest_ok || !subActive_ok || !exp_ok {
			if isWebSocket {
				log.Printf("[Auth] WebSocket JWT claims structure invalid for %s after %v", path, time.Since(authStart))
			}
			// Don't block on audit log
			go func() {
				auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				store.CreateAuditLog(auditCtx, &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    &userID,
					Action:    "authentication_failed",
					IPAddress: c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
					Details:   "Invalid token claims structure",
					CreatedAt: time.Now(),
				})
			}()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("email", email)
		c.Set("username", username)
		c.Set("tier", tier)
		c.Set("guest", guest)
		c.Set("subscription_active", subActive)
		c.Set("tokenExp", int64(exp))

		// Fetch user from DB to check MFA status (with timeout)
		userLookupStart := time.Now()
		user, err := store.GetUserByID(dbCtx, userID)
		if err != nil || user == nil {
			if isWebSocket {
				log.Printf("[Auth] WebSocket user lookup failed for %s after %v (DB: %v): %v", path, time.Since(authStart), time.Since(userLookupStart), err)
			}
			// Don't block on audit log
			go func() {
				auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				store.CreateAuditLog(auditCtx, &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    &userID,
					Action:    "authentication_failed",
					IPAddress: c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
					Details:   fmt.Sprintf("User not found in DB: %v", err),
					CreatedAt: time.Now(),
				})
			}()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Store full user object in context for later use (e.g., in AdminOnly middleware)
		c.Set("user", user)

		// --- Session revocation enforcement ---
		if sessionID != "" {
			sessionLookupStart := time.Now()
			srec, err := store.GetUserSession(dbCtx, sessionID)
			if err != nil || srec == nil || srec.UserID != userID || srec.RevokedAt != nil {
				if isWebSocket {
					log.Printf("[Auth] WebSocket session lookup failed for %s after %v (session check: %v): %v", path, time.Since(authStart), time.Since(sessionLookupStart), err)
				}
				c.JSON(http.StatusUnauthorized, gin.H{"error": "session_revoked"})
				c.Abort()
				return
			}
			// Touch last seen (best effort, fire and forget to avoid blocking)
			go func() {
				touchCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = store.TouchUserSession(touchCtx, sessionID, c.ClientIP(), c.Request.UserAgent())
			}()
			c.Set("sessionID", sessionID)
		}

		// --- Server-enforced screen lock ---
		// If the account is locked after this token was issued, block access with 423.
		if user.ScreenLockEnabled && user.ScreenLockHash != "" && user.LockRequiredSince != nil {
			// Allow unlock endpoint to proceed even with a locked token.
			if c.Request.URL.Path != "/api/security/unlock" {
				tokenIat := time.Unix(int64(iat), 0)
				if tokenIat.Before(*user.LockRequiredSince) {
					c.JSON(http.StatusLocked, gin.H{"error": "session_locked"})
					c.Abort()
					return
				}
			}
		}

		// --- IP Whitelist Enforcement ---
		if len(user.AllowedIPs) > 0 {
			clientIP := c.ClientIP()
			if !checkIPWhitelist(clientIP, user.AllowedIPs) {
				if isWebSocket {
					log.Printf("[Auth] WebSocket IP blocked for %s after %v: IP %s not in allowed list", path, time.Since(authStart), clientIP)
				}
				// Don't block on audit log
				go func() {
					auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					store.CreateAuditLog(auditCtx, &models.AuditLog{
						ID:        uuid.New().String(),
						UserID:    &userID,
						Action:    "ip_blocked",
						IPAddress: clientIP,
						UserAgent: c.Request.UserAgent(),
						Details:   fmt.Sprintf("IP %s not in allowed list", clientIP),
						CreatedAt: time.Now(),
					})
				}()
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied from this IP address"})
				c.Abort()
				return
			}
		}

		// Log successful authentication (don't block on this)
		go func() {
			auditCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			store.CreateAuditLog(auditCtx, &models.AuditLog{
				ID:        uuid.New().String(),
				UserID:    &userID,
				Action:    "authentication_success",
				IPAddress: c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Details:   fmt.Sprintf("User '%s' authenticated successfully.", username),
				CreatedAt: time.Now(),
			})
		}()

		if isWebSocket {
			log.Printf("[Auth] WebSocket auth successful for %s in %v (user: %s)", path, time.Since(authStart), userID)
		}

		c.Next()
	}
}

// checkIPWhitelist checks if a client IP is allowed
func checkIPWhitelist(clientIP string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		return true
	}

	client := net.ParseIP(clientIP)
	if client == nil {
		return false // Invalid client IP
	}

	for _, ipStr := range allowedIPs {
		ipStr = strings.TrimSpace(ipStr)
		if ipStr == "" {
			continue
		}

		// Check for CIDR
		if strings.Contains(ipStr, "/") {
			_, subnet, err := net.ParseCIDR(ipStr)
			if err == nil && subnet.Contains(client) {
				return true
			}
		} else {
			// Exact match
			ip := net.ParseIP(ipStr)
			if ip != nil && ip.Equal(client) {
				return true
			}
		}
	}

	return false
}
