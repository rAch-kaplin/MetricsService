package models

import "testing"

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
			g := &gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			err := g.Update(tt.args.mValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("gauge.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && g.value != tt.want {
				t.Errorf("gauge.value = %v, want %v", g.value, tt.want)
			}
		})
	}
}
