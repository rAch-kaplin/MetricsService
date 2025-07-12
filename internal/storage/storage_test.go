package storage

import (
	"context"
	"sync"
	"testing"

	mtr "github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/metrics"
)

func TestMemStorage_UpdateMetric(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		storage map[string]map[string]mtr.Metric
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
				storage: make(map[string]map[string]mtr.Metric),
			},
			args: args{
				mType:  mtr.GaugeType,
				mName:  "g1",
				mValue: 3.14,
			},
			wantErr:    false,
			wantResult: 3.14,
		},
		{
			name: "Add new counter",
			fields: fields{
				storage: make(map[string]map[string]mtr.Metric),
			},
			args: args{
				mType:  mtr.CounterType,
				mName:  "c1",
				mValue: int64(10),
			},
			wantErr:    false,
			wantResult: int64(10),
		},
		{
			name: "Invalid gauge value",
			fields: fields{
				storage: make(map[string]map[string]mtr.Metric),
			},
			args: args{
				mType:  mtr.GaugeType,
				mName:  "g2",
				mValue: "not a float",
			},
			wantErr: true,
		},
		{
			name: "Update existing counter",
			fields: fields{
				storage: map[string]map[string]mtr.Metric{
					mtr.CounterType: {
						"c2": mtr.NewCounter("c2", 5),
					},
				},
			},
			args: args{
				mType:  mtr.CounterType,
				mName:  "c2",
				mValue: int64(7),
			},
			wantErr:    false,
			wantResult: int64(12),
		},
		{
			name: "Unknown metric type",
			fields: fields{
				storage: make(map[string]map[string]mtr.Metric),
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
			ms := &MemStorage{
				mutex:   sync.RWMutex{},
				storage: tt.fields.storage,
			}
			if err := ms.UpdateMetric(ctx, tt.args.mType, tt.args.mName, tt.args.mValue); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				val, ok := ms.GetMetric(ctx, tt.args.mType, tt.args.mName)
				if !ok {
					t.Errorf("metric not found")
					return
				}
				if val != tt.wantResult {
					t.Errorf("got = %v, want = %v", val, tt.wantResult)
				}
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	ctx := context.Background()
	
	counter := mtr.NewCounter("requests", 42)
	gauge := mtr.NewGauge("temperature", 36.6)

	tests := []struct {
		name    string
		fields  map[string]map[string]mtr.Metric
		mType   string
		mName   string
		wantVal any
		wantOk  bool
	}{
		{
			name: "existing counter",
			fields: map[string]map[string]mtr.Metric{
				mtr.CounterType: {"requests": counter},
			},
			mType:   mtr.CounterType,
			mName:   "requests",
			wantVal: int64(42),
			wantOk:  true,
		},
		{
			name: "existing gauge",
			fields: map[string]map[string]mtr.Metric{
				mtr.GaugeType: {"temperature": gauge},
			},
			mType:   mtr.GaugeType,
			mName:   "temperature",
			wantVal: float64(36.6),
			wantOk:  true,
		},
		{
			name: "missing metric name",
			fields: map[string]map[string]mtr.Metric{
				mtr.CounterType: {"requests": counter},
			},
			mType:   mtr.CounterType,
			mName:   "nonexistent",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "missing metric type",
			fields:  map[string]map[string]mtr.Metric{},
			mType:   "unknown_type",
			mName:   "anything",
			wantVal: nil,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				mutex:   sync.RWMutex{},
				storage: tt.fields,
			}
			gotVal, gotOk := ms.GetMetric(ctx, tt.mType, tt.mName)
			if gotVal != tt.wantVal {
				t.Errorf("GetMetric() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetMetric() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

// func TestMemStorage_GetAllMetrics(t *testing.T) {
// 	counter := mtr.NewCounter("requests", 100)
// 	gauge := mtr.NewGauge("temperature", 25.5)
// 	type fields struct {
// 		mutex   sync.RWMutex
// 		storage map[string]map[string]mtr.Metric
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   map[string]map[string]mtr.Metric
// 	}{
// 		{
// 			name: "empty storage",
// 			fields: fields{
// 				mutex:   sync.RWMutex{},
// 				storage: map[string]map[string]mtr.Metric{},
// 			},
// 			want: map[string]map[string]mtr.Metric{},
// 		},
// 		{
// 			name: "storage with one counter",
// 			fields: fields{
// 				mutex: sync.RWMutex{},
// 				storage: map[string]map[string]mtr.Metric{
// 					mtr.CounterType: {
// 						"requests": counter,
// 					},
// 				},
// 			},
// 			want: map[string]map[string]mtr.Metric{
// 				mtr.CounterType: {
// 					"requests": counter,
// 				},
// 			},
// 		},
// 		{
// 			name: "storage with gauge and counter",
// 			fields: fields{
// 				mutex: sync.RWMutex{},
// 				storage: map[string]map[string]mtr.Metric{
// 					mtr.CounterType: {
// 						"requests": counter,
// 					},
// 					mtr.GaugeType: {
// 						"temperature": gauge,
// 					},
// 				},
// 			},
// 			want: map[string]map[string]mtr.Metric{
// 				mtr.CounterType: {
// 					"requests": counter,
// 				},
// 				mtr.GaugeType: {
// 					"temperature": gauge,
// 				},
// 			},
// 		},
// 	}
//
// 	for i := range tests {
// 		tt := &tests[i]
// 		t.Run(tt.name, func(t *testing.T) {
// 			ms := &MemStorage{
// 				mutex:   sync.RWMutex{},
// 				storage: tt.fields.storage,
// 			}
// 			if got := ms.GetAllMetrics(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("MemStorage.GetAllMetrics() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
