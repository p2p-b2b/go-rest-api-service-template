package model

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestEncodeToken(t *testing.T) {
	type args struct {
		id     uuid.UUID
		serial int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				id: uuid.Max,
				// fixed date
				serial: 0,
			},
			want: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;0")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeToken(tt.args.id, tt.args.serial); got != tt.want {
				t.Errorf("EncodeToken() = %v, want %v", got, tt.want)
			}

			// Decode the token to verify the encoding
			id, date, err := DecodeToken(tt.want)
			if err != nil {
				t.Errorf("DecodeToken() error = %v", err)
			}

			if date != tt.args.serial {
				t.Errorf("DecodeToken() date = %v, want %v", date, tt.args.serial)
			}

			if id != tt.args.id {
				t.Errorf("DecodeToken() id = %v, want %v", id, tt.args.id)
			}
		})
	}
}

func TestDecodeToken(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name       string
		args       args
		wantSerial int64
		wantId     uuid.UUID
		wantErr    bool
	}{
		{
			name: "success",
			args: args{
				s: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;0")),
			},
			wantSerial: 0,
			wantId:     uuid.Max,
			wantErr:    false,
		},
		{
			name: "invalid token",
			args: args{
				s: "invalid",
			},
			wantSerial: 0,
			wantId:     uuid.Nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, gotDate, err := DecodeToken(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDate, tt.wantSerial) {
				t.Errorf("DecodeToken() gotDate = %v, want %v", gotDate, tt.wantSerial)
			}
			if !reflect.DeepEqual(gotId, tt.wantId) {
				t.Errorf("DecodeToken() gotId = %v, want %v", gotId, tt.wantId)
			}

			if !tt.wantErr {
				// Encode the token to verify the decoding
				token := EncodeToken(gotId, gotDate)
				if token != tt.args.s {
					t.Errorf("EncodeToken() = %v, want %v", token, tt.args.s)
				}
			}
		})
	}
}
