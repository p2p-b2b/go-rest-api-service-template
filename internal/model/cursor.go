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

// ErrInvalidCursorToken is an error that is returned when the cursor token is invalid.
var ErrInvalidCursorToken = errors.New("invalid cursor token")

// CursorToken represents a cursor token.
type CursorToken struct {
	// Next is the token to the next page of users.
	Next uuid.UUID `json:"next"`

	// Date is the date of the cursor.
	Date time.Time `json:"date"`
}

// NewCursorToken creates a new CursorToken.
func NewCursorToken(next uuid.UUID, date time.Time) *CursorToken {
	return &CursorToken{
		Next: next,
		Date: date,
	}
}

// String returns the string representation of the cursor token.
func (c *CursorToken) String() string {
	payload := c.Next.String() + DataSeparator + c.Date.Format(time.RFC3339)
	out := base64.StdEncoding.EncodeToString([]byte(payload))

	return out
}

// Encode encodes the date and id into a string.
func (c *CursorToken) Encode() string {
	return c.String()
}

// Decode decodes the string into a date and id.
func (c *CursorToken) Decode(s string) (date time.Time, id uuid.UUID, err error) {
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

// MarshalJSON marshals the cursor token into a JSON string.
func (c *CursorToken) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.String() + `"`), nil
}

// UnmarshalJSON unmarshals the cursor token from a JSON string.
func (c *CursorToken) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	date, id, err := c.Decode(s)
	if err != nil {
		return err
	}

	c.Date = date
	c.Next = id

	return nil
}
