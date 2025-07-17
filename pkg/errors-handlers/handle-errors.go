package errors

import (
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

func WithRetry(f func() error, searchError func(error) bool) error {
	backoffSchedule := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	var err error

	for _, backoff := range backoffSchedule {
		err = f()
		if err != nil {
			return err
		}

		if !searchError(err) {
			return err
		}

		time.Sleep(backoff)
	}

	return err
}

func IsPostgresRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case
			// Class 08 — Connection Exception
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.ProtocolViolation,

			// Class 40 — Transaction Rollback
			pgerrcode.TransactionRollback,
			pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,

			// Class 53 — Insufficient Resources
			pgerrcode.TooManyConnections,

			// Class 57 — Operator Intervention
			pgerrcode.AdminShutdown,
			pgerrcode.CrashShutdown,
			pgerrcode.CannotConnectNow:

			return true
		}
	}

	return false
}
