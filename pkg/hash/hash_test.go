package hash_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"reflect"
	"testing"

	"github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/hash"
)

func TestGetHash(t *testing.T) {
	type args struct {
		key  []byte
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid HMAC SHA256 hash",
			args: args{
				key:  []byte("secret"),
				data: []byte("hello world"),
			},
			want: func() []byte {
				h := hmac.New(sha256.New, []byte("secret"))
				h.Write([]byte("hello world"))
				return h.Sum(nil)
			}(),
			wantErr: false,
		},
		{
			name: "Empty key and data",
			args: args{
				key:  []byte{},
				data: []byte{},
			},
			want: func() []byte {
				h := hmac.New(sha256.New, []byte{})
				h.Write([]byte{})
				return h.Sum(nil)
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hash.GetHash(tt.args.key, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckHash(t *testing.T) {
	type args struct {
		key      []byte
		data     []byte
		expdHash []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Correct hash comparison",
			args: func() args {
				key := []byte("key123")
				data := []byte("message")
				h := hmac.New(sha256.New, key)
				h.Write(data)
				expected := h.Sum(nil)
				return args{key: key, data: data, expdHash: expected}
			}(),
			want: true,
		},
		{
			name: "Incorrect hash",
			args: func() args {
				key := []byte("key123")
				data := []byte("message")
				wrongHash := []byte("invalidhash")
				return args{key: key, data: data, expdHash: wrongHash}
			}(),
			want: false,
		},
		{
			name: "Empty key/data but correct expected hash",
			args: func() args {
				key := []byte{}
				data := []byte{}
				h := hmac.New(sha256.New, key)
				h.Write(data)
				expected := h.Sum(nil)
				return args{key: key, data: data, expdHash: expected}
			}(),
			want: true,
		},
		{
			name: "Nil expected hash",
			args: args{
				key:      []byte("somekey"),
				data:     []byte("somedata"),
				expdHash: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hash.CheckHash(tt.args.key, tt.args.data, tt.args.expdHash); got != tt.want {
				t.Errorf("CheckHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
