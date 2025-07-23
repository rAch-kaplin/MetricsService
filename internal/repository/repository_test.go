package repository_test

import (
	"context"
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_UpdateMetric(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		storage map[string]map[string]models.Metric
	}
	type args struct {
		mType  string
		mName  string
		mValue any
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantResult any
	}{
		{
			name: "Add new gauge",
			fields: fields{
				storage: make(map[string]map[string]models.Metric),
			},
			args: args{
				mType:  models.GaugeType,
				mName:  "g1",
				mValue: 3.14,
			},
			wantErr:    false,
			wantResult: 3.14,
		},
		{
			name: "Add new counter",
			fields: fields{
				storage: make(map[string]map[string]models.Metric),
			},
			args: args{
				mType:  models.CounterType,
				mName:  "c1",
				mValue: int64(10),
			},
			wantErr:    false,
			wantResult: int64(10),
		},
		{
			name: "Invalid gauge value",
			fields: fields{
				storage: make(map[string]map[string]models.Metric),
			},
			args: args{
				mType:  models.GaugeType,
				mName:  "g2",
				mValue: "not a float",
			},
			wantErr: true,
		},
		{
			name: "Update existing counter",
			fields: fields{
				storage: map[string]map[string]models.Metric{
					models.CounterType: {
						"c2": models.NewCounter("c2", 5),
					},
				},
			},
			args: args{
				mType:  models.CounterType,
				mName:  "c2",
				mValue: int64(7),
			},
			wantErr:    false,
			wantResult: int64(12),
		},
		{
			name: "Unknown metric type",
			fields: fields{
				storage: make(map[string]map[string]models.Metric),
			},
			args: args{
				mType:  "unknown",
				mName:  "whatever",
				mValue: 1,
			},
			wantErr: true,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ms := repository.NewMemStorage()

			for metricType, metrics := range tt.fields.storage {
				for name, metric := range metrics {
					_ = ms.UpdateMetric(ctx, metricType, name, metric.Value())
				}
			}

			err := ms.UpdateMetric(ctx, tt.args.mType, tt.args.mName, tt.args.mValue)
			if tt.wantErr {
				require.Error(t, err, "MemStorage.UpdateMetric() error = %v, wantErr = %v", err, tt.wantErr)
			} else {
				require.NoError(t, err, "MemStorage.UpdateMetric() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				metric, err := ms.GetMetric(ctx, tt.args.mType, tt.args.mName)
				require.NoError(t, err, "metric not found")
				assert.Equal(t, tt.wantResult, metric.Value(), "got = %v, want = %v", metric.Value(), tt.wantResult)
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	ctx := context.Background()

	counter := models.NewCounter("requests", 42)
	gauge := models.NewGauge("temperature", 36.6)

	tests := []struct {
		name    string
		fields  map[string]map[string]models.Metric
		mType   string
		mName   string
		wantVal any
		wantOk  bool
	}{
		{
			name: "existing counter",
			fields: map[string]map[string]models.Metric{
				models.CounterType: {"requests": counter},
			},
			mType:   models.CounterType,
			mName:   "requests",
			wantVal: int64(42),
			wantOk:  true,
		},
		{
			name: "existing gauge",
			fields: map[string]map[string]models.Metric{
				models.GaugeType: {"temperature": gauge},
			},
			mType:   models.GaugeType,
			mName:   "temperature",
			wantVal: float64(36.6),
			wantOk:  true,
		},
		{
			name: "missing metric name",
			fields: map[string]map[string]models.Metric{
				models.CounterType: {"requests": counter},
			},
			mType:   models.CounterType,
			mName:   "nonexistent",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "missing metric type",
			fields:  map[string]map[string]models.Metric{},
			mType:   "unknown_type",
			mName:   "anything",
			wantVal: nil,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := repository.NewMemStorage()

			for metricType, metrics := range tt.fields {
				for name, metric := range metrics {
					_ = ms.UpdateMetric(ctx, metricType, name, metric.Value())
				}
			}

			metric, err := ms.GetMetric(ctx, tt.mType, tt.mName)

			if tt.wantOk {
				require.NoError(t, err, "GetMetric() error = %v, wantOk = %v", err, tt.wantOk)
				assert.Equal(t, tt.wantVal, metric.Value(), "GetMetric() gotVal = %v, want %v", metric.Value(), tt.wantVal)
			} else {
				require.Error(t, err, "GetMetric() error = %v, wantOk = %v", err, tt.wantOk)
			}
		})
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	ctx := context.Background()

	counter := models.NewCounter("requests", 100)
	gauge := models.NewGauge("temperature", 25.5)

	tests := []struct {
		name    string
		storage map[string]map[string]models.Metric
		want    map[string]models.Metric
	}{
		{
			name:    "empty storage",
			storage: map[string]map[string]models.Metric{},
			want:    map[string]models.Metric{},
		},
		{
			name: "storage with one counter",
			storage: map[string]map[string]models.Metric{
				models.CounterType: {
					"requests": counter,
				},
			},
			want: map[string]models.Metric{
				"requests": counter,
			},
		},
		{
			name: "storage with gauge and counter",
			storage: map[string]map[string]models.Metric{
				models.CounterType: {
					"requests": counter,
				},
				models.GaugeType: {
					"temperature": gauge,
				},
			},
			want: map[string]models.Metric{
				"requests":    counter,
				"temperature": gauge,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := repository.NewMemStorage()

			for metricType, metrics := range tt.storage {
				for name, metric := range metrics {
					_ = ms.UpdateMetric(ctx, metricType, name, metric.Value())
				}
			}

			got, err := ms.GetAllMetrics(ctx)
			require.NoError(t, err, "unexpected error: %v", err)

			gotMap := make(map[string]models.Metric)
			for _, m := range got {
				gotMap[m.Name()] = m
			}

			assert.Equal(t, len(tt.want), len(gotMap), "wrong number of metrics: got %d, want %d", len(gotMap), len(tt.want))

			for name, wantMetric := range tt.want {
				gotMetric, ok := gotMap[name]
				assert.True(t, ok, "missing metric: %s", name)
				if !ok {
					continue
				}

				assert.Equal(t, wantMetric.Type(), gotMetric.Type(), "metric %s type mismatch: got %s, want %s", name, gotMetric.Type(), wantMetric.Type())

				switch wantMetric.Type() {
				case models.CounterType:
					assert.NotNil(t, gotMetric.Value(), "metric %s: expected counter delta, got nil", name)
				case models.GaugeType:
					assert.NotNil(t, gotMetric.Value(), "metric %s: expected gauge value, got nil", name)
				default:
					assert.Fail(t, "unknown metric type", "metric %s: unknown type %s", name, wantMetric.Type())
				}
			}
		})
	}
}
