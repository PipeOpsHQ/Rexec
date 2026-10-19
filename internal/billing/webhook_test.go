package billing

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleWebhookRejectsUnusableSecret(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"\t",
		"whsec_your_webhook_secret",
		"whsec_...",
		"whsec_short",
		"not-a-stripe-secret",
	}
	for _, secret := range tests {
		svc := NewServiceWithConfig(Config{WebhookSecret: secret})
		req := httptest.NewRequest("POST", "/api/billing/webhook", strings.NewReader(`{}`))
		req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
		_, err := svc.HandleWebhook(req)
		if !errors.Is(err, ErrWebhookNotConfigured) {
			t.Fatalf("secret %q: got %v, want ErrWebhookNotConfigured", secret, err)
		}
	}
}

func TestHandleWebhookVerifiesARealSecret(t *testing.T) {
	svc := NewServiceWithConfig(Config{WebhookSecret: "  whsec_" + strings.Repeat("a", 32) + "  "})
	req := httptest.NewRequest("POST", "/api/billing/webhook", strings.NewReader(`{}`))
	req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
	_, err := svc.HandleWebhook(req)
	if !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("got %v, want signature rejection", err)
	}
}
