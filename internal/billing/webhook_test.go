package billing

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleWebhookRejectsEmptySecret(t *testing.T) {
	tests := []string{"", "   ", "\t"}
	for _, secret := range tests {
		svc := NewServiceWithConfig(Config{WebhookSecret: secret})
		req := httptest.NewRequest("POST", "/api/billing/webhook", strings.NewReader(`{}`))
		req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
		if _, err := svc.HandleWebhook(req); err == nil {
			t.Fatalf("secret %q was accepted", secret)
		}
	}
}
