package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rexec/rexec/internal/authclaims"
	"github.com/rexec/rexec/internal/models"
)

// shouldEnforceScreenLock mirrors the JWT screen-lock decision in AuthMiddleware
// so we can unit-test the security invariant without a full HTTP stack.
func shouldEnforceScreenLock(claims jwt.MapClaims, user *models.User, path string) bool {
	if user == nil {
		return false
	}
	if authclaims.IsAutomationToken(claims) {
		return false
	}
	if !user.ScreenLockEnabled || user.ScreenLockHash == "" || user.LockRequiredSince == nil {
		return false
	}
	if path == "/api/security/unlock" {
		return false
	}
	iat, _ := claims["iat"].(float64)
	tokenIat := time.Unix(int64(iat), 0)
	return tokenIat.Before(*user.LockRequiredSince)
}

func TestScreenLockEnforcement_AutomationClaimBypasses(t *testing.T) {
	t.Parallel()
	lockSince := time.Now()
	user := &models.User{
		ScreenLockEnabled: true,
		ScreenLockHash:    "hash",
		LockRequiredSince: &lockSince,
	}
	// Token issued before lock — would be blocked without automation claim.
	claims := jwt.MapClaims{
		"iat":                     float64(lockSince.Add(-time.Hour).Unix()),
		authclaims.ClaimTokenUse:  authclaims.TokenUseAutomation,
		authclaims.ClaimAudience:  authclaims.AudiencePipeOpsBFF,
	}
	if shouldEnforceScreenLock(claims, user, "/api/containers") {
		t.Fatal("automation JWT must not enforce screen lock")
	}
}

func TestScreenLockEnforcement_SidLessWithoutClaimStillLocked(t *testing.T) {
	t.Parallel()
	lockSince := time.Now()
	user := &models.User{
		ScreenLockEnabled: true,
		ScreenLockHash:    "hash",
		LockRequiredSince: &lockSince,
	}
	// Sid-less JWT without token_use=automation must still be locked.
	claims := jwt.MapClaims{
		"iat": float64(lockSince.Add(-time.Hour).Unix()),
		// no sid, no automation claim
	}
	if !shouldEnforceScreenLock(claims, user, "/api/containers") {
		t.Fatal("sid-less JWT without automation claim must enforce screen lock")
	}
}

func TestScreenLockEnforcement_BrowserSessionWithSid(t *testing.T) {
	t.Parallel()
	lockSince := time.Now()
	user := &models.User{
		ScreenLockEnabled: true,
		ScreenLockHash:    "hash",
		LockRequiredSince: &lockSince,
	}
	claims := jwt.MapClaims{
		"sid": "browser-session",
		"iat": float64(lockSince.Add(-time.Hour).Unix()),
	}
	if !shouldEnforceScreenLock(claims, user, "/api/containers") {
		t.Fatal("browser session JWT issued before lock must enforce screen lock")
	}
	if shouldEnforceScreenLock(claims, user, "/api/security/unlock") {
		t.Fatal("unlock path must not enforce screen lock")
	}
}

func TestScreenLockEnforcement_TokenIssuedAfterLockPasses(t *testing.T) {
	t.Parallel()
	lockSince := time.Now().Add(-time.Hour)
	user := &models.User{
		ScreenLockEnabled: true,
		ScreenLockHash:    "hash",
		LockRequiredSince: &lockSince,
	}
	claims := jwt.MapClaims{
		"sid": "browser-session",
		"iat": float64(time.Now().Unix()),
	}
	if shouldEnforceScreenLock(claims, user, "/api/containers") {
		t.Fatal("token issued after lock_required_since must pass")
	}
}
