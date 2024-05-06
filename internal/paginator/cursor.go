package paginator

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

	date, err = time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	id, err = uuid.Parse(parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	return date, id, nil
}
