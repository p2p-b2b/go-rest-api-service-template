package paginator

import (
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEncodeToken(t *testing.T) {
	type args struct {
		id   uuid.UUID
		date time.Time
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
				date: time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
			},
			want: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;2021-09-01T00:00:00Z")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeToken(tt.args.id, tt.args.date); got != tt.want {
				t.Errorf("EncodeToken() = %v, want %v", got, tt.want)
			}

			// Decode the token to verify the encoding
			date, id, err := DecodeToken(tt.want)
			if err != nil {
				t.Errorf("DecodeToken() error = %v", err)
			}

			if date != tt.args.date {
				t.Errorf("DecodeToken() date = %v, want %v", date, tt.args.date)
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
		name     string
		args     args
		wantDate time.Time
		wantId   uuid.UUID
		wantErr  bool
	}{
		{
			name: "success",
			args: args{
				s: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;2021-09-01T00:00:00Z")),
			},
			wantDate: time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
			wantId:   uuid.Max,
			wantErr:  false,
		},
		{
			name: "invalid token",
			args: args{
				s: "invalid",
			},
			wantDate: time.Time{},
			wantId:   uuid.Nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, gotId, err := DecodeToken(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDate, tt.wantDate) {
				t.Errorf("DecodeToken() gotDate = %v, want %v", gotDate, tt.wantDate)
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
