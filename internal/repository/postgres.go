package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	errH "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/errors-handlers"
	"github.com/rs/zerolog/log"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(ctx context.Context, dataBaseDSN string) (*Database, error) {
	log.Info().Msgf("DSN: %s", dataBaseDSN)
	db, err := sql.Open("pgx", dataBaseDSN)
	if err != nil {
		log.Error().Err(err).Msg("sql.Open error")
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.ExecContext(ctx,
		"CREATE TABLE IF NOT EXISTS collector ("+
			"\"ID\" VARCHAR(250) PRIMARY KEY,"+
			"\"MType\" TEXT,"+
			"\"Delta\" BIGINT,"+
			"\"Value\" DOUBLE PRECISION"+
			");")

	if err != nil {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database")
		}
		log.Error().Err(err).Msg("failed create table for database")
		return nil, fmt.Errorf("failed create table for database %w", err)
	}

	return &Database{
		DB: db,
	}, nil
}

func (db *Database) GetMetric(ctx context.Context, mType, mName string) (models.Metric, error) {
	var (
		id    string
		Type  string
		delta sql.NullInt64
		value sql.NullFloat64
	)

	getMtr := func() error {
		builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			Where(sq.Eq{`"ID"`: mName, `"MType"`: mType}).
			Limit(1).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}
		row := db.DB.QueryRowContext(ctx, query, args...)

		return row.Scan(&id, &Type, &delta, &value)
	}

	err := errH.WithRetry(getMtr, errH.IsPostgresRetriableError)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrMetricsNotFound
	} else if err != nil {
		return nil, fmt.Errorf("row.Scan can't read: %w", err)
	}

	if Type != mType {
		return nil, models.ErrInvalidMetricsType
	}

	switch {
	case value.Valid:
		if Type != models.GaugeType {
			return nil, models.ErrInvalidMetricsType
		}

		return models.NewGauge(mName, value.Float64), nil

	case delta.Valid:
		if Type != models.CounterType {
			return nil, models.ErrInvalidMetricsType
		}

		return models.NewCounter(mName, delta.Int64), nil
	default:
		log.Error().Msg("not valid value")
		return nil, models.ErrInvalidValueType
	}
}

func (db *Database) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
		From("collector")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := db.DB.QueryContext(ctx, query, args...)

	if err != nil {
		log.Error().Err(err).Msg("The request was not processed")
		return nil, fmt.Errorf("failed to query all metrics: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	metrics := make([]models.Metric, 0)

	var (
		id    string
		mType string
		delta sql.NullInt64
		value sql.NullFloat64
	)

	for rows.Next() {
		err = rows.Scan(&id, &mType, &delta, &value)
		if err != nil {
			log.Error().Err(err).Msgf("failed scan row: ID = %s, MType = %s", id, mType)
			return nil, fmt.Errorf("failed to scan metric row: %v", err)
		}

		switch mType {
		case models.GaugeType:
			if value.Valid {
				metrics = append(metrics, models.NewGauge(id, value.Float64))
			}

		case models.CounterType:
			if delta.Valid {
				metrics = append(metrics, models.NewCounter(id, delta.Int64))
			}

		default:
			return nil, fmt.Errorf("incorrectly metric type %v", models.ErrInvalidMetricsType)
		}

	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("have rows error")
		return nil, rows.Err()
	}

	return metrics, nil
}

func (db *Database) UpdateMetric(ctx context.Context, mType, mName string, mValue any) error {
	var delta *int64
	var value *float64

	switch v := mValue.(type) {
	case float64:
		if mType != models.GaugeType {
			return fmt.Errorf("metric type mismatch: got float64 with type %q", mType)
		}
		value = &v

	case int64:
		if mType != models.CounterType {
			return fmt.Errorf("metric type mismatch: got int64 with type %q", mType)
		}
		delta = &v

	default:
		return fmt.Errorf("unsupported metric value type: %T", v)
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	exec := func() error {
		builder := sq.Insert("collector").
			Columns(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			Values(mName, mType, delta, value).
			Suffix(`ON CONFLICT ("ID") DO UPDATE SET
			"Delta" = collector."Delta" + EXCLUDED."Delta",
			"Value" = EXCLUDED."Value",
			"MType" = EXCLUDED."MType"`).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, query, args...)
		return err
	}

	if err := errH.WithRetry(exec, errH.IsPostgresRetriableError); err != nil {
		log.Error().Err(err).Msg("failed to insert/update metric")
		return fmt.Errorf("update metric: %w", err)
	}

	return tx.Commit()
}

func (db *Database) UpdateMetricList(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := db.UpdateMetric(ctx, metric.Type(), metric.Name(), metric.Value()); err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) Ping(ctx context.Context) error {
	if err := db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed ping database: %w", err)
	}
	return nil
}

func (db *Database) Close() error {
	if db.DB != nil {
		if err := db.DB.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database")
			return err
		}
	}

	return nil
}
