package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func IpValidation(trustedSubnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				zap.L().Error("X-Real-IP header is missing")
				http.Error(w, "X-Real-IP header is missing", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				zap.L().Error("Invalid IP address in X-Real-IP header")
				http.Error(w, "Invalid IP address in X-Real-IP header", http.StatusForbidden)
				return
			}

			if !trustedSubnet.Contains(ip) {
				zap.L().Error("Forbidden: IP not in trusted subnet", zap.String("subnet", trustedSubnet.String()), zap.String("ip", ip.String()))
				http.Error(w, "Forbidden: IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
