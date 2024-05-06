package model

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DataSeparator is the separator used to separate the data in the cursor token.
const DataSeparator string = ";"

var DateFormat string = time.RFC3339

// ErrInvalidCursorToken is an error that is returned when the cursor token is invalid.
var ErrInvalidCursorToken = errors.New("invalid cursor token")

// String returns the string representation of the cursor token.
// func (c *CursorToken) String() string {
// 	payload := c.Next.String() + DataSeparator + c.Date.Format(time.RFC3339)
// 	out := base64.StdEncoding.EncodeToString([]byte(payload))

// 	return out
// }

// Encode encodes the date and id into a string.
func EncodeCursorToken(next uuid.UUID, date time.Time) string {
	return next.String() + DataSeparator + date.Format(DateFormat)
}

// Decode decodes the string into a date and id.
func DecodeCursorToken(s string) (date time.Time, id uuid.UUID, err error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	parts := strings.Split(string(data), DataSeparator)
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
