package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func Subnet(cidr string, log *zap.Logger) func(next http.Handler) http.Handler {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Error("failed to parse cidr string: %w", zap.Error(err))
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				rawIP := r.Header.Get("X-Real-IP")
				if rawIP == "" {
					log.Debug("X-Real-IP header expected, found nil instead")
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ip := net.ParseIP(rawIP)
				if ip == nil {
					log.Debug("X-Real-IP contains invalid IP format")
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if !subnet.Contains(ip) {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
