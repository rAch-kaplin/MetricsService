package serialization_test

import (
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/stretchr/testify/assert"

	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
)

func TestMetric_SetValue(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta *int64
		Value *float64
	}
	type args struct {
		value any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "set gauge with valid float64",
			fields: fields{
				ID:    "gauge1",
				MType: models.GaugeType,
				Delta: nil,
				Value: nil,
			},
			args:    args{value: 123.45},
			wantErr: false,
		},
		{
			name: "set gauge with invalid type",
			fields: fields{
				ID:    "gauge2",
				MType: models.GaugeType,
				Delta: nil,
				Value: nil,
			},
			args:    args{value: "string_instead_of_float"},
			wantErr: true,
		},
		{
			name: "set counter with valid int64",
			fields: fields{
				ID:    "counter1",
				MType: models.CounterType,
				Delta: nil,
				Value: nil,
			},
			args:    args{value: int64(100)},
			wantErr: false,
		},
		{
			name: "set counter with invalid type",
			fields: fields{
				ID:    "counter2",
				MType: models.CounterType,
				Delta: nil,
				Value: nil,
			},
			args:    args{value: 12.34},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtr := &serialize.Metric{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
			}

			err := mtr.SetValue(tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				switch mtr.MType {
				case models.GaugeType:
					assert.NotNil(t, mtr.Value)
					assert.Equal(t, tt.args.value, *mtr.Value)

				case models.CounterType:
					assert.NotNil(t, mtr.Delta)
					assert.Equal(t, tt.args.value, *mtr.Delta)

				default:
					t.Fatalf("unexpected type %s", mtr.MType)
				}
			}
		})
	}
}

func float64Ptr(f float64) *float64 { return &f }
func int64Ptr(i int64) *int64       { return &i }

func TestMetric_GetValue(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta *int64
		Value *float64
	}
	tests := []struct {
		name    string
		fields  fields
		want    any
		wantErr bool
	}{
		{
			name: "get gauge with valid value",
			fields: fields{
				ID:    "gauge1",
				MType: models.GaugeType,
				Value: float64Ptr(123.45),
				Delta: nil,
			},
			want:    123.45,
			wantErr: false,
		},
		{
			name: "get gauge with nil value",
			fields: fields{
				ID:    "gauge2",
				MType: models.GaugeType,
				Value: nil,
				Delta: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get counter with valid value",
			fields: fields{
				ID:    "counter1",
				MType: models.CounterType,
				Delta: int64Ptr(42),
				Value: nil,
			},
			want:    int64(42),
			wantErr: false,
		},
		{
			name: "get counter with nil value",
			fields: fields{
				ID:    "counter2",
				MType: models.CounterType,
				Delta: nil,
				Value: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unknown metric type",
			fields: fields{
				ID:    "unknown",
				MType: "unknown",
				Delta: nil,
				Value: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtr := &serialize.Metric{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
			}
			got, err := mtr.GetValue()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
