package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	"github.com/rs/zerolog/log"
)

type Database struct {
	mutex sync.RWMutex
	DB    *sql.DB
}

func NewDatabase(ctx context.Context, dataBaseDSN string) (col.Collector, error) {
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
		mutex: sync.RWMutex{},
		DB:    db,
	}, nil
}

func (db *Database) getMetric(ctx context.Context, mType, mName string) (any, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT "ID", "MType", "Delta", "Value" FROM collector `+
			`WHERE "ID" = $1 AND "MType" = $2 LIMIT 1`, mName, mType)

	var (
		id    string
		Type  string
		delta sql.NullInt64
		value sql.NullFloat64
	)

	err := row.Scan(&id, &Type, &delta, &value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, mtr.ErrMetricsNotFound
	} else if err != nil {
		return nil, fmt.Errorf("row.Scan can't read: %w", err)
	}

	switch {
	case value.Valid:
		if Type != mtr.GaugeType {
			return nil, mtr.ErrInvalidMetricsType
		}

		return value.Float64, nil

	case delta.Valid:
		if Type != mtr.CounterType {
			return nil, mtr.ErrInvalidMetricsType
		}

		return delta.Int64, nil
	default:
		log.Error().Msg("not valid value")
		return nil, mtr.ErrInvalidValueType
	}
}

func (db *Database) GetMetric(ctx context.Context, mType, mName string) (any, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	return db.getMetric(ctx, mType, mName)
}

func (db *Database) GetAllMetrics(ctx context.Context) []mtr.Metric {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	metrics := make(serialize.MetricsList, 0)

	rows, err := db.DB.QueryContext(ctx,
		"SELECT ID, MType, Delta, Value FROM collector")

	if err != nil {
		log.Error().Err(err).Msg("The request was not processed")
		return nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var (
		delta sql.NullInt64
		value sql.NullFloat64
	)

	for rows.Next() {
		var metric serialize.Metric

		err = rows.Scan(&metric.ID, &metric.MType, &delta, &value)
		if err != nil {
			log.Error().Err(err).Msgf("failed scan row: ID = %s, MType = %s",
				metric.ID, metric.MType)
			return nil
		}

		switch {
		case value.Valid:
			if metric.MType != mtr.GaugeType {
				return nil
			}

			metric.Value = &value.Float64

		case delta.Valid:
			if metric.MType != mtr.CounterType {
				return nil
			}

			metric.Delta = &delta.Int64
		default:
			log.Error().Msg("not valid value")
			return nil
		}

		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("have rows error")
		return nil
	}

	convertedMetrics, err := converter.ConvertMetrics(metrics)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert metrics")
		return nil
	}

	return convertedMetrics
}

func (db *Database) UpdateMetric(ctx context.Context, mType, mName string, mValue any) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	oldValue, err := db.getMetric(ctx, mType, mName)

	var m mtr.Metric

	switch newVal := mValue.(type) {
	case float64:
		if mType != mtr.GaugeType {
			return mtr.ErrInvalidMetricsType
		}
		m = mtr.NewGauge(mName, newVal)

	case int64:
		if mType != mtr.CounterType {
			return mtr.ErrInvalidMetricsType
		}
		if errors.Is(err, mtr.ErrMetricsNotFound) {
			m = mtr.NewCounter(mName, newVal)
		} else {
			m = mtr.NewCounter(mName, oldValue.(int64))
			if err := m.Update(newVal); err != nil {
				return fmt.Errorf("failed update metric %v", err)
			}
		}

	default:
		return fmt.Errorf("unsupported metric type %T", mValue)
	}

	metric := serialize.Metric{
		ID:    mName,
		MType: mType,
	}
	//TODO - make method maybe for Metrics type
	switch metric.MType {
	case mtr.GaugeType:
		val, ok := m.Value().(float64)
		if !ok {
			return fmt.Errorf("expected float64 for gauge, got %T", m.Value())
		}
		metric.Value = &val

	case mtr.CounterType:
		val, ok := m.Value().(int64)
		if !ok {
			return fmt.Errorf("expected int64 for counter, got %T", m.Value())
		}
		metric.Delta = &val

	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed begin transaction")
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO collector ("ID", "MType", "Delta", "Value")
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT ("ID") DO UPDATE
		 SET "Delta" = EXCLUDED."Delta",
		     "Value" = EXCLUDED."Value",
		     "MType" = EXCLUDED."MType";`,
		metric.ID, metric.MType, metric.Delta, metric.Value)

	if err != nil {
		log.Error().Err(err).Msg("failed update insert into collector")

		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}

		return fmt.Errorf("failed update insert into collector: %w", err)
	}

	return tx.Commit()
}

func (db *Database) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func (db *Database) Close() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.DB != nil {
		if err := db.DB.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database")
			return err
		}
	}

	return nil
}
