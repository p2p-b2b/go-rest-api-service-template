package model

import (
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
		{
			name: "invalid format - missing separator",
			args: args{
				s: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff")),
			},
			wantSerial: 0,
			wantId:     uuid.Nil,
			wantErr:    true,
		},
		{
			name: "invalid format - too many separators",
			args: args{
				s: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;0;extra")),
			},
			wantSerial: 0,
			wantId:     uuid.Nil,
			wantErr:    true,
		},
		{
			name: "invalid uuid",
			args: args{
				s: base64.StdEncoding.EncodeToString([]byte("not-a-uuid;0")),
			},
			wantSerial: 0,
			wantId:     uuid.Nil,
			wantErr:    true,
		},
		{
			name: "invalid serial",
			args: args{
				s: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;not-a-number")),
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

func TestPaginator_GenerateToken(t *testing.T) {
	type fields struct {
		NextToken string
		NextPage  string
		PrevToken string
		PrevPage  string
		Size      int
		Limit     int
	}
	type args struct {
		id     uuid.UUID
		serial int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "success with uuid.Max",
			fields: fields{
				Limit: 10,
			},
			args: args{
				id:     uuid.Max,
				serial: 0,
			},
			want: base64.StdEncoding.EncodeToString([]byte("ffffffff-ffff-ffff-ffff-ffffffffffff;0")),
		},
		{
			name: "success with random uuid",
			fields: fields{
				Limit: 20,
			},
			args: args{
				id:     uuid.MustParse("f47ac10b-58cc-0372-8567-0e02b2c3d479"),
				serial: 1234567890,
			},
			want: base64.StdEncoding.EncodeToString([]byte("f47ac10b-58cc-0372-8567-0e02b2c3d479;1234567890")),
		},
		{
			name: "success with nil uuid",
			fields: fields{
				Limit: 5,
			},
			args: args{
				id:     uuid.Nil,
				serial: 42,
			},
			want: base64.StdEncoding.EncodeToString([]byte("00000000-0000-0000-0000-000000000000;42")),
		},
		{
			name: "success with negative serial",
			fields: fields{
				Limit: 100,
			},
			args: args{
				id:     uuid.MustParse("a1a2a3a4-b1b2-c1c2-d1d2-d3d4d5d6d7d8"),
				serial: -9876543210,
			},
			want: base64.StdEncoding.EncodeToString([]byte("a1a2a3a4-b1b2-c1c2-d1d2-d3d4d5d6d7d8;-9876543210")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := &Paginator{
				NextToken: tt.fields.NextToken,
				NextPage:  tt.fields.NextPage,
				PrevToken: tt.fields.PrevToken,
				PrevPage:  tt.fields.PrevPage,
				Size:      tt.fields.Size,
				Limit:     tt.fields.Limit,
			}
			if got := ref.GenerateToken(tt.args.id, tt.args.serial); got != tt.want {
				t.Errorf("Paginator.GenerateToken() = %v, want %v", got, tt.want)
			}

			// Verify we can decode the token correctly
			id, serial, err := DecodeToken(tt.want)
			if err != nil {
				t.Errorf("Failed to decode token: %v", err)
			}
			if id != tt.args.id {
				t.Errorf("Decoded ID = %v, want %v", id, tt.args.id)
			}
			if serial != tt.args.serial {
				t.Errorf("Decoded serial = %v, want %v", serial, tt.args.serial)
			}
		})
	}
}

func TestPaginator_Validate(t *testing.T) {
	validUUID := uuid.New()
	validToken := EncodeToken(validUUID, 123)
	invalidToken := "invalid-token-not-base64"
	invalidBase64Token := base64.StdEncoding.EncodeToString([]byte("invalid-format"))

	tests := []struct {
		name    string
		p       Paginator
		wantErr bool
		errType error
	}{
		{
			name: "valid paginator with default values",
			p: Paginator{
				Limit: PaginatorDefaultLimit,
			},
			wantErr: false,
		},
		{
			name: "valid paginator with tokens",
			p: Paginator{
				NextToken: validToken,
				PrevToken: validToken,
				Limit:     PaginatorDefaultLimit,
			},
			wantErr: false,
		},
		{
			name: "invalid limit - below minimum",
			p: Paginator{
				Limit: PaginatorMinLimit - 1,
			},
			wantErr: true,
			errType: &InvalidPaginatorLimitError{
				MinLimit: PaginatorMinLimit,
				MaxLimit: PaginatorMaxLimit,
			},
		},
		{
			name: "invalid limit - above maximum",
			p: Paginator{
				Limit: PaginatorMaxLimit + 1,
			},
			wantErr: true,
			errType: &InvalidPaginatorLimitError{
				MinLimit: PaginatorMinLimit,
				MaxLimit: PaginatorMaxLimit,
			},
		},
		{
			name: "invalid next token - not base64",
			p: Paginator{
				NextToken: invalidToken,
				Limit:     PaginatorDefaultLimit,
			},
			wantErr: true,
			errType: &InvalidPaginatorTokenError{
				Message: "next token cannot be decoded",
			},
		},
		{
			name: "invalid prev token - not base64",
			p: Paginator{
				PrevToken: invalidToken,
				Limit:     PaginatorDefaultLimit,
			},
			wantErr: true,
			errType: &InvalidPaginatorTokenError{
				Message: "prev token cannot be decoded",
			},
		},
		{
			name: "invalid next token - invalid format",
			p: Paginator{
				NextToken: invalidBase64Token,
				Limit:     PaginatorDefaultLimit,
			},
			wantErr: true,
			errType: &InvalidPaginatorCursorError{
				Message: "next token cannot be decoded",
			},
		},
		{
			name: "invalid prev token - invalid format",
			p: Paginator{
				PrevToken: invalidBase64Token,
				Limit:     PaginatorDefaultLimit,
			},
			wantErr: true,
			errType: &InvalidPaginatorCursorError{
				Message: "prev token cannot be decoded",
			},
		},
		{
			name: "valid limit at minimum",
			p: Paginator{
				Limit: PaginatorMinLimit,
			},
			wantErr: false,
		},
		{
			name: "valid limit at maximum",
			p: Paginator{
				Limit: PaginatorMaxLimit,
			},
			wantErr: false,
		},
		{
			name: "multiple validation errors - invalid limit and tokens",
			p: Paginator{
				NextToken: invalidToken,
				PrevToken: invalidToken,
				Limit:     PaginatorMinLimit - 1,
			},
			wantErr: true,
			errType: &InvalidPaginatorLimitError{
				MinLimit: PaginatorMinLimit,
				MaxLimit: PaginatorMaxLimit,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginator.Validate(): got = '%v', want = '%v'", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var invalidLimitError *InvalidPaginatorLimitError
				if errors.As(tt.errType, &invalidLimitError) {
					if invalidLimitError.MinLimit != PaginatorMinLimit || invalidLimitError.MaxLimit != PaginatorMaxLimit {
						t.Errorf("InvalidPaginatorLimitError: got = '%v', want = '%v'", invalidLimitError, tt.errType)
					}
				}
				var invalidTokenError *InvalidPaginatorTokenError
				if errors.As(tt.errType, &invalidTokenError) {
					if invalidTokenError.Message != "next token cannot be decoded" && invalidTokenError.Message != "prev token cannot be decoded" {
						t.Errorf("InvalidPaginatorTokenError: got = '%v', want = '%v'", invalidTokenError, tt.errType)
					}
				}

				var invalidCursorError *InvalidPaginatorCursorError
				if errors.As(tt.errType, &invalidCursorError) {
					if invalidCursorError.Message != "next token cannot be decoded" && invalidCursorError.Message != "prev token cannot be decoded" {
						t.Errorf("InvalidPaginatorCursorError: got = '%v', want = '%v'", invalidCursorError, tt.errType)
					}
				}
			}
		})
	}
}

func TestPaginator_GeneratePages(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name         string
		paginator    Paginator
		url          string
		wantNextPage string
		wantPrevPage string
	}{
		{
			name: "both tokens present",
			paginator: Paginator{
				NextToken: EncodeToken(validUUID, 123),
				PrevToken: EncodeToken(validUUID, 122),
				Limit:     20,
			},
			url:          "https://api.example.com/resources",
			wantNextPage: "https://api.example.com/resources?next_token=" + EncodeToken(validUUID, 123) + "&limit=20",
			wantPrevPage: "https://api.example.com/resources?prev_token=" + EncodeToken(validUUID, 122) + "&limit=20",
		},
		{
			name: "only next token present",
			paginator: Paginator{
				NextToken: EncodeToken(validUUID, 123),
				PrevToken: "",
				Limit:     10,
			},
			url:          "https://api.example.com/resources",
			wantNextPage: "https://api.example.com/resources?next_token=" + EncodeToken(validUUID, 123) + "&limit=10",
			wantPrevPage: "",
		},
		{
			name: "only prev token present",
			paginator: Paginator{
				NextToken: "",
				PrevToken: EncodeToken(validUUID, 122),
				Limit:     30,
			},
			url:          "https://api.example.com/resources",
			wantNextPage: "",
			wantPrevPage: "https://api.example.com/resources?prev_token=" + EncodeToken(validUUID, 122) + "&limit=30",
		},
		{
			name: "no tokens present",
			paginator: Paginator{
				NextToken: "",
				PrevToken: "",
				Limit:     50,
			},
			url:          "https://api.example.com/resources",
			wantNextPage: "",
			wantPrevPage: "",
		},
		{
			name: "with path parameters",
			paginator: Paginator{
				NextToken: EncodeToken(validUUID, 123),
				PrevToken: EncodeToken(validUUID, 122),
				Limit:     15,
			},
			url:          "https://api.example.com/users/42/resources",
			wantNextPage: "https://api.example.com/users/42/resources?next_token=" + EncodeToken(validUUID, 123) + "&limit=15",
			wantPrevPage: "https://api.example.com/users/42/resources?prev_token=" + EncodeToken(validUUID, 122) + "&limit=15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the paginator to avoid modifying the test case
			p := tt.paginator

			// Call the method under test
			p.GeneratePages(tt.url)

			// Check that the next page URL is correct
			if p.NextPage != tt.wantNextPage {
				t.Errorf("GeneratePages() NextPage = %v, want %v", p.NextPage, tt.wantNextPage)
			}

			// Check that the prev page URL is correct
			if p.PrevPage != tt.wantPrevPage {
				t.Errorf("GeneratePages() PrevPage = %v, want %v", p.PrevPage, tt.wantPrevPage)
			}
		})
	}
}

func TestGetTokens(t *testing.T) {
	// Setup test data
	firstID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	firstSerial := int64(1000)
	lastID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	lastSerial := int64(9000)

	tests := []struct {
		name        string
		size        int
		limit       int
		firstID     uuid.UUID
		firstSerial int64
		lastID      uuid.UUID
		lastSerial  int64
		wantNext    bool
		wantPrev    bool
	}{
		{
			name:        "empty result",
			size:        0,
			limit:       10,
			firstID:     firstID,
			firstSerial: firstSerial,
			lastID:      lastID,
			lastSerial:  lastSerial,
			wantNext:    false,
			wantPrev:    false,
		},
		{
			name:        "partial result",
			size:        5,
			limit:       10,
			firstID:     firstID,
			firstSerial: firstSerial,
			lastID:      lastID,
			lastSerial:  lastSerial,
			wantNext:    false,
			wantPrev:    false,
		},
		{
			name:        "full result",
			size:        10,
			limit:       10,
			firstID:     firstID,
			firstSerial: firstSerial,
			lastID:      lastID,
			lastSerial:  lastSerial,
			wantNext:    true,
			wantPrev:    true,
		},
		{
			name:        "more than limit",
			size:        15,
			limit:       10,
			firstID:     firstID,
			firstSerial: firstSerial,
			lastID:      lastID,
			lastSerial:  lastSerial,
			wantNext:    true,
			wantPrev:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, prev := GetTokens(tt.size, tt.limit, tt.firstID, tt.firstSerial, tt.lastID, tt.lastSerial)

			if tt.wantNext {
				assert.NotEmpty(t, next, "next token should not be empty")
				// Decode and verify token contents
				decodedID, decodedSerial, err := DecodeToken(next)
				assert.NoError(t, err)
				assert.Equal(t, tt.lastID, decodedID)
				assert.Equal(t, tt.lastSerial, decodedSerial)
			} else {
				assert.Empty(t, next, "next token should be empty")
			}

			if tt.wantPrev {
				assert.NotEmpty(t, prev, "prev token should not be empty")
				// Decode and verify token contents
				decodedID, decodedSerial, err := DecodeToken(prev)
				assert.NoError(t, err)
				assert.Equal(t, tt.firstID, decodedID)
				assert.Equal(t, tt.firstSerial, decodedSerial)
			} else {
				assert.Empty(t, prev, "prev token should be empty")
			}
		})
	}
}
