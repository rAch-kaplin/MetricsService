package converter_test

import (
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertByType(t *testing.T) {
	type args struct {
		mType  string
		mValue string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "valid gauge",
			args: args{
				mType:  models.GaugeType,
				mValue: "123.45",
			},
			want:    123.45,
			wantErr: false,
		},
		{
			name: "invalid gauge value",
			args: args{
				mType:  models.GaugeType,
				mValue: "not_a_number",
			},
			wantErr: true,
		},
		{
			name: "valid counter",
			args: args{
				mType:  models.CounterType,
				mValue: "42",
			},
			want:    int64(42),
			wantErr: false,
		},
		{
			name: "invalid counter value",
			args: args{
				mType:  models.CounterType,
				mValue: "42.5",
			},
			wantErr: true,
		},
		{
			name: "unknown type",
			args: args{
				mType:  "unknown",
				mValue: "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.ConvertByType(tt.args.mType, tt.args.mValue)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			switch tt.args.mType {
			case models.GaugeType:
				assert.IsType(t, float64(0), got)
				assert.Equal(t, tt.want, got)

			case models.CounterType:
				require.IsType(t, int64(0), got)
				require.Equal(t, tt.want, got)

			default:
				t.Fatalf("unexpected type %s", tt.args.mType)
			}
		})
	}
}
