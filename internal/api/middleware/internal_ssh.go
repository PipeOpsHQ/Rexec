package middleware

import (
	"crypto/subtle"
	"net"
	"strings"
)

// InternalSSHAuthorized allows the SSH gateway's internal API only when a
// key is configured and presented and the TCP peer is loopback or a private
// address. A missing key fails closed. X-Forwarded-For is ignored; remoteAddr
// is the connection's RemoteAddr so a client cannot spoof the peer.
func InternalSSHAuthorized(remoteAddr, headerKey, configuredKey string) bool {
	configuredKey = strings.TrimSpace(configuredKey)
	if configuredKey == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(headerKey), []byte(configuredKey)) != 1 {
		return false
	}
	return remoteAddrIsPrivate(remoteAddr)
}

func remoteAddrIsPrivate(remoteAddr string) bool {
	host := strings.TrimSpace(remoteAddr)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}
