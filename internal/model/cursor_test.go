package model

import (
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursorToken_EncodeCursorToken(t *testing.T) {
	type args struct {
		Next uuid.UUID
		Date time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				Next: uuid.Max,
				// fixed date
				Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want:    base64.StdEncoding.EncodeToString([]byte(uuid.Max.String() + DataSeparator + time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeCursorToken(tt.args.Next, tt.args.Date); got != tt.want {
				t.Errorf("EncodeCursorToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCursorToken_DecodeCursorToken(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		want1   uuid.UUID
		wantErr bool
	}{
		{
			name: "success",

			args: args{
				s: base64.StdEncoding.EncodeToString([]byte(uuid.Max.String() + DataSeparator + time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))),
			},
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			want1:   uuid.Max,
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				s: "invalid",
			},
			want:    time.Time{},
			want1:   uuid.Nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, id, err := DecodeCursorToken(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeCursorToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if date != tt.want {
				t.Errorf("DecodeCursorToken() date = %v, want %v", date, tt.want)
			}
			if id != tt.want1 {
				t.Errorf("DecodeCursorToken() id = %v, want %v", id, tt.want1)
			}
		})
	}
}

func TestCursorToken_MarshalJSON(t *testing.T) {
	type fields struct {
		Next uuid.UUID
		Date time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Next: uuid.Max,
				// fixed date
				Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want:    []byte(`"` + base64.StdEncoding.EncodeToString([]byte(uuid.Max.String()+DataSeparator+time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))) + `"`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CursorToken{
				Next: tt.fields.Next,
				Date: tt.fields.Date,
			}
			got, err := c.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("CursorToken.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != string(tt.want) {
				t.Errorf("CursorToken.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}

			// Unmarshal the token to verify it
			var c2 CursorToken
			if err := c2.UnmarshalJSON(got); err != nil {
				t.Errorf("CursorToken.UnmarshalJSON() error = %v", err)
				return
			}

			if c2.Date != c.Date {
				t.Errorf("CursorToken.UnmarshalJSON() date = %v, want %v", c2.Date, c.Date)
			}

			if c2.Next != c.Next {
				t.Errorf("CursorToken.UnmarshalJSON() id = %v, want %v", c2.Next, c.Next)
			}
		})
	}
}

func TestCursorToken_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Next uuid.UUID
		Date time.Time
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Next: uuid.Max,
				// fixed date
				Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			args: args{
				data: []byte(`"` + base64.StdEncoding.EncodeToString([]byte(uuid.Max.String()+DataSeparator+time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))) + `"`),
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				Next: uuid.Max,
				// fixed date
				Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			args: args{
				data: []byte(`"invalid"`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CursorToken{
				Next: tt.fields.Next,
				Date: tt.fields.Date,
			}
			if err := c.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("CursorToken.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if c.Date != tt.fields.Date {
					t.Errorf("CursorToken.UnmarshalJSON() date = %v, want %v", c.Date, tt.fields.Date)
				}

				if c.Next != tt.fields.Next {
					t.Errorf("CursorToken.UnmarshalJSON() id = %v, want %v", c.Next, tt.fields.Next)
				}
			}
		})
	}
}

func TestStoCursorToken(t *testing.T) {
	type args struct {
		base64 string
	}
	tests := []struct {
		name    string
		args    args
		want    *CursorToken
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				base64: base64.StdEncoding.EncodeToString([]byte(uuid.Max.String() + DataSeparator + time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))),
			},
			want: &CursorToken{
				Next: uuid.Max,
				Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				base64: "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StoCursorToken(tt.args.base64)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoCursorToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StoCursorToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
