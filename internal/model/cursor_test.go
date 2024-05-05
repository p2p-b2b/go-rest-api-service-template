package model

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursorToken_Encode(t *testing.T) {
	type fields struct {
		Next uuid.UUID
		Date time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
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
			c := &CursorToken{
				Next: tt.fields.Next,
				Date: tt.fields.Date,
			}
			got := c.Encode()
			if got != tt.want {
				t.Errorf("CursorToken.Encode() = %v, want %v", got, tt.want)
			}

			// Decode the token to verify it
			date, id, err := c.Decode(got)
			if err != nil {
				t.Errorf("CursorToken.Decode() error = %v", err)
				return
			}

			if date != c.Date {
				t.Errorf("CursorToken.Decode() date = %v, want %v", date, c.Date)
			}

			if id != c.Next {
				t.Errorf("CursorToken.Decode() id = %v, want %v", id, c.Next)
			}
		})
	}
}

func TestCursorToken_Decode(t *testing.T) {
	type fields struct {
		Next uuid.UUID
		Date time.Time
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    time.Time
		want1   uuid.UUID
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
				s: base64.StdEncoding.EncodeToString([]byte(uuid.Max.String() + DataSeparator + time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))),
			},
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			want1:   uuid.Max,
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
				s: "invalid",
			},
			want:    time.Time{},
			want1:   uuid.Nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CursorToken{
				Next: tt.fields.Next,
				Date: tt.fields.Date,
			}
			got, got1, err := c.Decode(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("CursorToken.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CursorToken.Decode() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CursorToken.Decode() got1 = %v, want %v", got1, tt.want1)
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
