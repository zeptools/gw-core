package requests

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	// Prefer X-Forwarded-For (first entry)
	if xForwaredFor := r.Header.Get("X-Forwarded-For"); xForwaredFor != "" {
		parts := strings.Split(xForwaredFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	// Fallback to X-Real-IP
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}
	// Final fallback: RemoteAddr (nginx IP)
	hostIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return hostIP
}
