package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func WithDataBase(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Error().Err(err).Msg("database ping failed")
			http.Error(w, fmt.Sprintf("database ping failed: %v", err), http.StatusInternalServerError)
			return
		}
		log.Info().Msg("database ping successful")

		next(w, r)
	}
}

func PingDataBase(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
