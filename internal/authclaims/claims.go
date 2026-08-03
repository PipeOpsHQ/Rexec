// Package authclaims defines JWT claim names and values shared by auth
// handlers (mint) and middleware (verify).
package authclaims

import "github.com/golang-jwt/jwt/v5"

// JWT claim keys.
const (
	// ClaimTokenUse marks the intended use of a JWT.
	// Browser sessions omit this or use TokenUseSession; BFF/automation
	// tokens set TokenUseAutomation.
	ClaimTokenUse = "token_use"

	// ClaimAudience is the standard JWT "aud" claim.
	ClaimAudience = "aud"
)

// Claim values.
const (
	// TokenUseAutomation is set on PipeOps BFF / server-to-server assert JWTs.
	// Only tokens with this claim may skip Rexec screen-lock enforcement.
	TokenUseAutomation = "automation"

	// AudiencePipeOpsBFF is the audience for PipeOps controller assert JWTs.
	AudiencePipeOpsBFF = "pipeops-bff"
)

// IsAutomationToken reports whether claims identify an intentional automation
// JWT (PipeOps BFF assert). Screen-lock bypass must use this, not "no sid".
func IsAutomationToken(claims jwt.MapClaims) bool {
	if claims == nil {
		return false
	}
	use, _ := claims[ClaimTokenUse].(string)
	return use == TokenUseAutomation
}

// IsAutomationTokenFromMap is a map-based helper for tests and non-jwt/v5 maps.
func IsAutomationTokenFromMap(claims map[string]interface{}) bool {
	if claims == nil {
		return false
	}
	use, _ := claims[ClaimTokenUse].(string)
	return use == TokenUseAutomation
}
