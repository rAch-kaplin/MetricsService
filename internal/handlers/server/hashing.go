package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/hash"
	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

type hashResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
	key  []byte
}

func (hsw *hashResponseWriter) Write(b []byte) (int, error) {
	if size, err := hsw.body.Write(b); err != nil {
		return size, fmt.Errorf("failed write body %v", err)
	}

	if hsw.key != nil && hsw.body.Len() > 0 {
		newHash, err := hash.GetHash(hsw.key, hsw.body.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("failed to get hash")
		}

		hsw.Header().Set("HashSHA256", hex.EncodeToString(newHash))
	}

	return hsw.ResponseWriter.Write(b)
}

func WithHashing(key []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Error().Err(err).Msg("failed read body")
					http.Error(w, fmt.Sprintf("failed read body %v", err), http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body))

				h := r.Header.Get("HashSHA256")
				if h != "" {
					decoded, err := hex.DecodeString(h)
					if err != nil {
						log.Error().Err(err).Msg("failed to decode hash")
						http.Error(w, "invalid hash format", http.StatusBadRequest)
						return
					}
					valid := hash.CheckHash(key, body, decoded)
					if !valid {
						log.Error().Msg("invalid hash message")
						http.Error(w, "invalid hash message", http.StatusBadRequest)

						return
					}
				}

				hw := &hashResponseWriter{
					ResponseWriter: w,
					body:           bytes.NewBuffer(nil),
					key:            key,
				}

				next.ServeHTTP(hw, r)
			}
		})
	}
}
