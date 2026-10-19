package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rexec/rexec/internal/models"
)

func TestGuestRestoreMatches(t *testing.T) {
	guest := &models.User{ID: "g1", Tier: "guest"}
	member := &models.User{ID: "u1", Tier: "free"}
	tests := []struct {
		name   string
		user   *models.User
		cookie string
		want   bool
	}{
		{name: "matching guest", user: guest, cookie: "g1", want: true},
		{name: "wrong cookie", user: guest, cookie: "other", want: false},
		{name: "non-guest", user: member, cookie: "u1", want: false},
		{name: "nil user", user: nil, cookie: "g1", want: false},
		{name: "empty cookie", user: guest, cookie: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := guestRestoreMatches(tt.user, tt.cookie); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGuestRestoreCookieRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret-test-secret-test-secret")
	h := NewAuthHandler(nil, nil, secret)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/api/auth/guest", nil)
	h.setGuestRestoreCookie(c, "guest-123")

	res := w.Result()
	defer res.Body.Close()
	var found *http.Cookie
	for _, ck := range res.Cookies() {
		if ck.Name == guestRestoreCookie {
			found = ck
			break
		}
	}
	if found == nil {
		t.Fatal("restore cookie was not set")
	}
	if !found.HttpOnly {
		t.Fatal("cookie is not httpOnly")
	}
	if found.Path != "/api/auth" {
		t.Fatalf("cookie path %q", found.Path)
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/api/auth/guest", nil)
	req.AddCookie(found)
	c2.Request = req
	id, ok := h.guestRestoreUserID(c2)
	if !ok || id != "guest-123" {
		t.Fatalf("restore id = %q ok=%v", id, ok)
	}

	other := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"purpose": "access",
		"user_id": "guest-123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := other.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Del("Cookie")
	req.AddCookie(&http.Cookie{Name: guestRestoreCookie, Value: signed})
	if _, ok := h.guestRestoreUserID(c2); ok {
		t.Fatal("non-restore token was accepted")
	}
}
