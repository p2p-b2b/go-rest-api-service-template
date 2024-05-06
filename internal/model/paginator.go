package model

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TokenDataSeparator is the separator used to separate the data in the cursor token.
var TokenDataSeparator string = ";"

// TokenDateFormat is the date format used in the cursor token.
var TokenDateFormat string = time.RFC3339

// ErrInvalidCursorToken is an error that is returned when the cursor token is invalid.
var ErrInvalidCursorToken = errors.New("invalid cursor token")

// Paginator represents a paginator.
//
// size: number of users returned
// next: token to the next page encoded in base64 from the last user id and created_at date
// previous: token to the previous page encoded in base64 from the first user id and created_at date
// limit: maximum number of users to return
type Paginator struct {
	// Size is the number of users returned.
	Size int `json:"size,omitempty"`

	// Next is the token to the next page of users.
	Next string `json:"next,omitempty"`

	// Previous is the token to the previous page of users.
	Previous string `json:"previous,omitempty"`

	// Limit is the maximum number of users to return.
	Limit int `json:"limit,omitempty"`
}

// EncodeToken encodes the date and id into a base64 string.
func EncodeToken(id uuid.UUID, date time.Time) string {
	payload := id.String() + TokenDataSeparator + date.Format(TokenDateFormat)
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

// DecodeToken decodes the string into a date and id.
func DecodeToken(s string) (date time.Time, id uuid.UUID, err error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	parts := strings.Split(string(data), TokenDataSeparator)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, ErrInvalidCursorToken
	}

	date, err = time.Parse(TokenDateFormat, parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	id, err = uuid.Parse(parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	return date, id, nil
}
