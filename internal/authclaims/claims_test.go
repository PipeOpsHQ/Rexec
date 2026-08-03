package authclaims

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestIsAutomationToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		claims jwt.MapClaims
		want   bool
	}{
		{
			name:   "nil",
			claims: nil,
			want:   false,
		},
		{
			name:   "empty",
			claims: jwt.MapClaims{},
			want:   false,
		},
		{
			name: "browser session with sid only",
			claims: jwt.MapClaims{
				"sid": "sess-123",
			},
			want: false,
		},
		{
			name: "sid-less without automation claim — must NOT bypass",
			claims: jwt.MapClaims{
				"user_id": "u1",
			},
			want: false,
		},
		{
			name: "explicit automation claim",
			claims: jwt.MapClaims{
				ClaimTokenUse: TokenUseAutomation,
				ClaimAudience: AudiencePipeOpsBFF,
			},
			want: true,
		},
		{
			name: "wrong token_use value",
			claims: jwt.MapClaims{
				ClaimTokenUse: "session",
			},
			want: false,
		},
		{
			name: "aud alone is not enough",
			claims: jwt.MapClaims{
				ClaimAudience: AudiencePipeOpsBFF,
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsAutomationToken(tc.claims); got != tc.want {
				t.Fatalf("IsAutomationToken() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsAutomationTokenFromMap(t *testing.T) {
	t.Parallel()
	if !IsAutomationTokenFromMap(map[string]interface{}{
		ClaimTokenUse: TokenUseAutomation,
	}) {
		t.Fatal("expected automation map to be true")
	}
	if IsAutomationTokenFromMap(map[string]interface{}{"sid": "x"}) {
		t.Fatal("sid-only map must not be automation")
	}
}
