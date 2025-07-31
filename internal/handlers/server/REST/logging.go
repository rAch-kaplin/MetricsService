package rest

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	loggingResponseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *loggingResponseData
	}
)

func (res *loggingResponseWriter) Write(body []byte) (int, error) {
	size, err := res.ResponseWriter.Write(body)
	res.responseData.size += size

	return size, err
}

func WithLogging(h http.Handler) http.Handler {
	logfn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &loggingResponseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		log.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Int("status", responseData.status).
			Dur("duration", duration).
			Int("size", responseData.size).
			Msg("new request")

	}

	return http.HandlerFunc(logfn)
}
