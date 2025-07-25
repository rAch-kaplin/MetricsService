package models_test

import (
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterUpdate(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	type args struct {
		mValue any
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantValue int64
	}{
		{
			name: "valid increment",
			fields: fields{
				name:  "test_counter",
				value: 10,
			},
			args: args{
				mValue: int64(5),
			},
			wantErr:   false,
			wantValue: 15,
		},
		{
			name: "zero increment",
			fields: fields{
				name:  "test_counter",
				value: 100,
			},
			args: args{
				mValue: int64(0),
			},
			wantErr:   false,
			wantValue: 100,
		},
		{
			name: "negative increment",
			fields: fields{
				name:  "test_counter",
				value: 20,
			},
			args: args{
				mValue: int64(-5),
			},
			wantErr:   false,
			wantValue: 15,
		},
		{
			name: "invalid type (string)",
			fields: fields{
				name:  "test_counter",
				value: 42,
			},
			args: args{
				mValue: "not an int",
			},
			wantErr:   true,
			wantValue: 42,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := models.NewCounter(tt.fields.name, tt.fields.value)
			err := c.Update(tt.args.mValue)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantValue, c.Value())
		})
	}
}
