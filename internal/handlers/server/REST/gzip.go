package rest

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

var supportedContentType = map[string]struct{}{
	"application/json":         {},
	"text/html; charset=utf-8": {},
}

// gzipWriter wraps http.ResponseWriter and writes the response body
// through a gzip.Writer if the response content type is supported.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write compresses the response with gzip if the Content-Type is supported.
// If not supported, it writes the response without compression.
func (w *gzipWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")

	if _, ok := supportedContentType[contentType]; !ok {
		w.Header().Del("Content-Encoding")
		return w.ResponseWriter.Write(b)
	}

	w.Header().Set("Content-Encoding", "gzip")
	return w.Writer.Write(b)
}

// WithGzipCompress is an HTTP middleware that compresses server responses using gzip.
//
// If the client includes "gzip" in the Accept-Encoding header, and the response
// has a supported Content-Type, the response body will be compressed using gzip.
// Otherwise, the response is passed through uncompressed.
func WithGzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		encodingValues := req.Header.Values("Accept-Encoding")

		if !slices.Contains(encodingValues, "gzip") {
			next.ServeHTTP(w, req)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, "Failed to create gzip writer", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := gz.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close gzip writer")
			}
		}()

		gw := &gzipWriter{
			ResponseWriter: w,
			Writer:         gz,
		}

		next.ServeHTTP(gw, req)
	})
}
