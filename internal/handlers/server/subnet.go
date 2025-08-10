package server

import (
	"net"
	"net/http"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func WithTrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	var subnet *net.IPNet
	if trustedSubnet != "" {
		_, s, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse trusted subnet")
			return nil
		}

		subnet = s
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			if ip == nil {
				log.Error().Msg("failed to parse ip")
				http.Error(w, "failed to parse ip", http.StatusBadRequest)
				return
			}

			if !subnet.Contains(ip) {
				log.Error().Msg("ip is not in trusted subnet")
				http.Error(w, "ip is not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
