package web

import (
	"net/http"
	"regexp"
	"strings"
)

var ipRegex = regexp.MustCompile(`^([\.0-9]+):\d+$`)

// ClientIP attempts to grab an IP address from the `X-Forwarded-For` header
// or the remote address. It is possible that an empty string can be returned.
func ClientIP(r *http.Request) string {
	var ip string
	// Try to set via forward header
	fwd := r.Header.Get("X-Forwarded-For")
	if len(fwd) > 0 {
		split := strings.Split(fwd, ",")
		ip = strings.TrimSpace(split[0])
	}
	// Fallback to remote address
	if len(ip) == 0 {
		if m := ipRegex.FindStringSubmatch(r.RemoteAddr); len(m) > 1 {
			ip = m[1]
		}
	}
	return ip
}
