package server

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

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")

	if _, ok := supportedContentType[contentType]; !ok {
		w.Header().Del("Content-Encoding")
		return w.ResponseWriter.Write(b)
	}

	w.Header().Set("Content-Encoding", "gzip")
	return w.Writer.Write(b)
}

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
