package rest

import (
	"net"
	"net/http"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

// WithTrustedSubnet is an HTTP middleware that checks if the incoming request
// is from a trusted subnet.
//
// It parses the trusted subnet from the configuration, creates a subnet object,
// and returns a http.Handler that checks if the incoming request is from a trusted subnet.
//
// If the trusted subnet is set, it checks if the incoming request is from a trusted subnet.
// If the incoming request is not from a trusted subnet, it returns a 403 Forbidden response.
// If the incoming request is from a trusted subnet, it returns next Handler.
func WithTrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	// Create a subnet object.
	var subnet *net.IPNet

	// If the trusted subnet is set, parse it.
	if trustedSubnet != "" {
		_, s, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse trusted subnet")
			return nil
		}

		subnet = s
	}

	// Return a http.Handler that checks if the incoming request is from a trusted subnet.
	return func(next http.Handler) http.Handler {
		// Create a http.HandlerFunc that checks if the incoming request is from a trusted subnet.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If the trusted subnet is not set, return next Handler.
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Get the IP address from the request header.
			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			if ip == nil {
				log.Error().Msg("failed to parse ip")
				http.Error(w, "failed to parse ip", http.StatusBadRequest)
				return
			}

			// Check if the IP address is in the trusted subnet.
			if !subnet.Contains(ip) {
				log.Error().Msg("ip is not in trusted subnet")
				http.Error(w, "ip is not in trusted subnet", http.StatusForbidden)
				return
			}

			// If the IP address is in the trusted subnet, return next Handler.
			next.ServeHTTP(w, r)
		})
	}
}
