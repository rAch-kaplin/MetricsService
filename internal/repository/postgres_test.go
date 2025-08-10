package repository_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	repo "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_GetMetric(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &repo.Database{
		DB: db,
	}

	t.Run("GetMetric_Gauge", func(t *testing.T) {
		builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			Where(sq.Eq{`"ID"`: "test_gauge", `"MType"`: "gauge"}).
			Limit(1).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			t.Fatalf("failed to build query: %v", err)
		}

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"ID", "MType", "Delta", "Value"}).
				AddRow("test_gauge", "gauge", nil, 100.0))

		metric, err := repo.GetMetric(context.Background(), "gauge", "test_gauge")
		require.NoError(t, err)

		assert.Equal(t, "test_gauge", metric.Name())
		assert.Equal(t, "gauge", metric.Type())
		assert.Equal(t, 100.0, metric.Value())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetMetric_Counter", func(t *testing.T) {
		builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			Where(sq.Eq{`"ID"`: "test_counter", `"MType"`: "counter"}).
			Limit(1).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			t.Fatalf("failed to build query: %v", err)
		}

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"ID", "MType", "Delta", "Value"}).
				AddRow("test_counter", "counter", 100, nil))

		metric, err := repo.GetMetric(context.Background(), "counter", "test_counter")

		require.NoError(t, err)

		assert.Equal(t, "test_counter", metric.Name())
		assert.Equal(t, "counter", metric.Type())
		assert.Equal(t, int64(100), metric.Value())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetMetric_InvalidType", func(t *testing.T) {
		builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			Where(sq.Eq{`"ID"`: "test_counter", `"MType"`: "counter"}).
			Limit(1).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			t.Fatalf("failed to build query: %v", err)
		}

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnError(models.ErrInvalidMetricsType)

		_, err = repo.GetMetric(context.Background(), "counter", "test_counter")

		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidMetricsType)
	})

	t.Run("GetMetric_NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT "ID", "MType", "Delta", "Value" FROM collector WHERE "ID" = \$1 AND "MType" = \$2 LIMIT 1`).
			WithArgs("unknown_metric", "gauge").
			WillReturnError(sql.ErrNoRows)

		metric, err := repo.GetMetric(context.Background(), "gauge", "unknown_metric")

		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrMetricsNotFound)
		assert.Nil(t, metric)

		require.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestDatabase_GetAllMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &repo.Database{
		DB: db,
	}

	t.Run("GetAllMetrics_Success", func(t *testing.T) {
		builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			t.Fatalf("failed to build query: %v", err)
		}

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"ID", "MType", "Delta", "Value"}).
				AddRow("test_gauge", "gauge", nil, 100.0).
				AddRow("test_counter", "counter", 100, nil))

		metrics, err := repo.GetAllMetrics(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 2, len(metrics))
		assert.Equal(t, "test_gauge", metrics[0].Name())
		assert.Equal(t, "gauge", metrics[0].Type())
		assert.Equal(t, 100.0, metrics[0].Value())

		assert.Equal(t, "test_counter", metrics[1].Name())
		assert.Equal(t, "counter", metrics[1].Type())
		assert.Equal(t, int64(100), metrics[1].Value())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetAllMetrics_Error", func(t *testing.T) {
		builder := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		if err != nil {
			t.Fatalf("failed to build query: %v", err)
		}

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnError(sql.ErrNoRows)

		metrics, err := repo.GetAllMetrics(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Empty(t, metrics)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDatabase_UpdateMetric(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &repo.Database{
		DB: db,
	}

	t.Run("UpdateMetric_Gauge", func(t *testing.T) {
		builder := sq.Insert("collector").
			Columns(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			Values("test_gauge", "gauge", nil, 100.0).
			Suffix(`ON CONFLICT ("ID") DO UPDATE SET
            "Delta" = collector."Delta" + EXCLUDED."Delta",
            "Value" = EXCLUDED."Value",
            "MType" = EXCLUDED."MType"`).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		require.NoError(t, err)

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = repo.UpdateMetric(context.Background(), "gauge", "test_gauge", 100.0)
		require.NoError(t, err)

		builderGet := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			Where(sq.Eq{`"ID"`: "test_gauge", `"MType"`: "gauge"}).
			Limit(1).
			PlaceholderFormat(sq.Dollar)
		queryGet, argsGet, err := builderGet.ToSql()
		require.NoError(t, err)

		driverArgsGet := make([]driver.Value, len(argsGet))
		for i, a := range argsGet {
			driverArgsGet[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(queryGet)).
			WithArgs(driverArgsGet...).
			WillReturnRows(sqlmock.NewRows([]string{"ID", "MType", "Delta", "Value"}).
				AddRow("test_gauge", "gauge", nil, 100.0))

		metric, err := repo.GetMetric(context.Background(), "gauge", "test_gauge")
		require.NoError(t, err)
		assert.Equal(t, "test_gauge", metric.Name())
		assert.Equal(t, "gauge", metric.Type())
		assert.Equal(t, 100.0, metric.Value())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateMetric_Counter", func(t *testing.T) {
		builder := sq.Insert("collector").
			Columns(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			Values("test_counter", "counter", 100, nil).
			Suffix(`ON CONFLICT ("ID") DO UPDATE SET
			"Delta" = collector."Delta" + EXCLUDED."Delta",
			"Value" = EXCLUDED."Value",
			"MType" = EXCLUDED."MType"`).
			PlaceholderFormat(sq.Dollar)

		query, args, err := builder.ToSql()
		require.NoError(t, err)

		driverArgs := make([]driver.Value, len(args))
		for i, a := range args {
			driverArgs[i] = a
		}

		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = repo.UpdateMetric(context.Background(), "counter", "test_counter", int64(100))
		require.NoError(t, err)

		builderGet := sq.Select(`"ID"`, `"MType"`, `"Delta"`, `"Value"`).
			From("collector").
			Where(sq.Eq{`"ID"`: "test_counter", `"MType"`: "counter"}).
			Limit(1).
			PlaceholderFormat(sq.Dollar)
		queryGet, argsGet, err := builderGet.ToSql()
		require.NoError(t, err)

		driverArgsGet := make([]driver.Value, len(argsGet))
		for i, a := range argsGet {
			driverArgsGet[i] = a
		}

		mock.ExpectQuery(regexp.QuoteMeta(queryGet)).
			WithArgs(driverArgsGet...).
			WillReturnRows(sqlmock.NewRows([]string{"ID", "MType", "Delta", "Value"}).
				AddRow("test_counter", "counter", 100, nil))

		metric, err := repo.GetMetric(context.Background(), "counter", "test_counter")
		require.NoError(t, err)
		assert.Equal(t, "test_counter", metric.Name())
		assert.Equal(t, "counter", metric.Type())
		assert.Equal(t, int64(100), metric.Value())

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
