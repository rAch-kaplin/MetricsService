package converter_test

import (
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/converter"
	serialize "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/serialization"
	modelsMocks "github.com/rAch-kaplin/mipt-golang-course/MetricsService/test/mocks/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

func TestConvertToMetricTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		src []models.Metric
	}
	tests := []struct {
		name    string
		args    args
		want    []models.MetricTable
		wantErr bool
	}{
		{
			name: "success - single gauge metric",
			args: args{
				src: func() []models.Metric {
					m := modelsMocks.NewMockMetric(ctrl)
					m.EXPECT().Type().Return("gauge")
					m.EXPECT().Name().Return("alloc")
					m.EXPECT().Value().Return(123.45)
					return []models.Metric{m}
				}(),
			},
			want: []models.MetricTable{
				{
					Type:  "gauge",
					Name:  "alloc",
					Value: "123.45",
				},
			},
			wantErr: false,
		},
		{
			name: "success - multiple metrics",
			args: args{
				src: func() []models.Metric {
					m1 := modelsMocks.NewMockMetric(ctrl)
					m1.EXPECT().Type().Return("gauge")
					m1.EXPECT().Name().Return("alloc")
					m1.EXPECT().Value().Return(123.45)

					m2 := modelsMocks.NewMockMetric(ctrl)
					m2.EXPECT().Type().Return("counter")
					m2.EXPECT().Name().Return("poll")
					m2.EXPECT().Value().Return(int64(5))

					return []models.Metric{m1, m2}
				}(),
			},
			want: []models.MetricTable{
				{
					Type:  "gauge",
					Name:  "alloc",
					Value: "123.45",
				},
				{
					Type:  "counter",
					Name:  "poll",
					Value: "5",
				},
			},
			wantErr: false,
		},
		{
			name: "error - unsupported metric type",
			args: args{
				src: func() []models.Metric {
					m := modelsMocks.NewMockMetric(ctrl)
					m.EXPECT().Type().Return("unknown")
					m.EXPECT().Name().Return("test")
					return []models.Metric{m}
				}(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.ConvertToMetricTable(tt.args.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertMetrics(t *testing.T) {
    value := float64(123.45)
    delta := int64(5)

    type args struct {
        src serialize.MetricsList
    }
    tests := []struct {
        name        string
        args        args
        wantTypes   []string
        wantNames   []string
        wantValues  []any
        wantErr     bool
    }{
        {
            name: "success - single metric",
            args: args{
                src: serialize.MetricsList{
                    {MType: "gauge", ID: "Alloc", Value: &value},
                    {MType: "counter", ID: "PollCount", Delta: &delta},
                },
            },
            wantTypes:  []string{"gauge", "counter"},
            wantNames:  []string{"Alloc", "PollCount"},
            wantValues: []any{123.45, int64(5)},
            wantErr:    false,
        },
        {
            name: "error - invalid metric type",
            args: args{
                src: serialize.MetricsList{
                    {MType: "unknown", ID: "Alloc", Value: &value},
                },
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := converter.ConvertMetrics(tt.args.src)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            require.Equal(t, len(tt.wantTypes), len(got))

            for i, metric := range got {
                assert.Equal(t, tt.wantTypes[i], metric.Type())
                assert.Equal(t, tt.wantNames[i], metric.Name())
                assert.Equal(t, tt.wantValues[i], metric.Value())
            }
        })
    }
}
