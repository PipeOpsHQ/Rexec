package handlers

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rexec/rexec/internal/authclaims"
	"github.com/rexec/rexec/internal/models"
)

func TestGenerateTokenOpts_AutomationClaims(t *testing.T) {
	t.Parallel()
	h := &AuthHandler{jwtSecret: []byte("test-secret-for-automation-claims")}
	user := &models.User{
		ID:                 "user-1",
		Email:              "a@example.com",
		Username:           "alice",
		Tier:               "pro",
		SubscriptionActive: true,
	}

	tok, err := h.generateTokenOpts(user, "should-be-ignored", true)
	if err != nil {
		t.Fatalf("generateTokenOpts: %v", err)
	}

	parsed, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("parse automation token: %v", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}
	if !authclaims.IsAutomationToken(claims) {
		t.Fatalf("expected automation claims, got %#v", claims)
	}
	if claims["sid"] != nil && claims["sid"] != "" {
		t.Fatalf("automation token must not carry sid, got %v", claims["sid"])
	}
	if claims[authclaims.ClaimAudience] != authclaims.AudiencePipeOpsBFF {
		t.Fatalf("aud = %v, want %s", claims[authclaims.ClaimAudience], authclaims.AudiencePipeOpsBFF)
	}
}

func TestGenerateToken_BrowserHasNoAutomationClaim(t *testing.T) {
	t.Parallel()
	h := &AuthHandler{jwtSecret: []byte("test-secret-for-browser-claims")}
	user := &models.User{
		ID:                 "user-2",
		Email:              "b@example.com",
		Username:           "bob",
		Tier:               "pro",
		SubscriptionActive: true,
	}

	tok, err := h.generateToken(user, "sess-abc")
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}
	parsed, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("parse browser token: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if authclaims.IsAutomationToken(claims) {
		t.Fatal("browser token must not be automation")
	}
	if claims["sid"] != "sess-abc" {
		t.Fatalf("sid = %v, want sess-abc", claims["sid"])
	}
}
