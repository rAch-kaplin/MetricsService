package models_test

import (
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGaugeUpdate(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	type args struct {
		mValue any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "valid update",
			fields: fields{
				name:  "TestGauge",
				value: 10.0,
			},
			args: args{
				mValue: 42.5,
			},
			want:    42.5,
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				name:  "TestGaugeInvalid",
				value: 3.14,
			},
			args: args{
				mValue: "not a float",
			},
			want:    3.14,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := models.NewGauge(tt.fields.name, tt.fields.value)
			err := g.Update(tt.args.mValue)
			if tt.wantErr {
				require.Error(t, err, "gauge.Update() error = %v, wantErr = %v", err, tt.wantErr)
			} else {
				require.NoError(t, err, "gauge.Update() error = %v, wantErr = %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, g.Value(), "gauge.value = %v, want %v", g.Value(), tt.want)
		})
	}
}
