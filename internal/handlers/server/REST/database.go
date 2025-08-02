package rest

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

// WithDataBase is HTTP middleware that checks if the database is reachable.
//
// It creates a context with a timeout of 1 second and pings the database.
// If the database is not reachable, it returns an Internal error.
// If the database is reachable, it returns next handler.
func WithDataBase(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
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

// PingDataBase is a function that returns a 200 OK response.
//
// It is used to check if the database is reachable.
func PingDataBase(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
