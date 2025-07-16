package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	col "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/collector"
	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
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
			"\"Delta\" INTEGER,"+
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

func (db *Database) getMetric(ctx context.Context, mType, mName string) (any, bool) {
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
	if err != nil {
		log.Error().Err(err).Msg("failed scan row")
		return nil, false
	}

	switch {
	case value.Valid:
		if Type != mtr.GaugeType {
			return nil, false
		}

		return value.Float64, true

	case delta.Valid:
		if Type != mtr.CounterType {
			return nil, false
		}

		return delta.Int64, true
	default:
		log.Error().Msg("not valid value")
		return nil, false
	}
}

func (db *Database) GetMetric(ctx context.Context, mType, mName string) (any, bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

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
	if err != nil {
		log.Error().Err(err).Msg("failed scan row")
		return nil, false
	}

	switch {
	case value.Valid:
		if Type != mtr.GaugeType {
			return nil, false
		}

		return value.Float64, true

	case delta.Valid:
		if Type != mtr.CounterType {
			return nil, false
		}

		return delta.Int64, true
	default:
		log.Error().Msg("not valid value")
		return nil, false
	}
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

	oldValue, ok := db.getMetric(ctx, mType, mName)

	var m mtr.Metric

	if !ok {
		switch v := mValue.(type) {
		case float64:
			m = mtr.NewGauge(mName, v)
		case int64:
			m = mtr.NewCounter(mName, v)
		default:
			return fmt.Errorf("unsupported type for metric value: %T", mValue)
		}
	} else {
		switch v := oldValue.(type) {
		case float64:
			m = mtr.NewGauge(mName, v)
		case int64:
			m = mtr.NewCounter(mName, v)
		default:
			return fmt.Errorf("unsupported type for existing metric: %T", oldValue)
		}

		m.Update(mValue)
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

	_, err := db.DB.ExecContext(ctx,
		`INSERT INTO collector ("ID", "MType", "Delta", "Value")
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT ("ID") DO UPDATE
		 SET "Delta" = EXCLUDED."Delta",
		     "Value" = EXCLUDED."Value",
		     "MType" = EXCLUDED."MType";`,
		metric.ID, metric.MType, metric.Delta, metric.Value)

	if err != nil {
		log.Error().Err(err).Msg("failed update insert into collector")
		return fmt.Errorf("failed update insert into collector: %w", err)
	}

	return nil
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
