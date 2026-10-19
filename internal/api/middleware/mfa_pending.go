package middleware

import (
	"net/http"
	"strings"
)

// MFAPendingRequestAllowed reports whether a token with mfa_pending may call
// method+path. Pending tokens exist only to finish the second factor. Every
// other route, including MFA enrollment, is refused.
func MFAPendingRequestAllowed(method, rawPath string) bool {
	if strings.ToUpper(method) != http.MethodPost {
		return false
	}
	return cleanRequestPath(rawPath) == "/api/mfa/complete-login"
}
