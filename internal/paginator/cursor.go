package paginator

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DataSeparator is the separator used to separate the data in the cursor token.
var DataSeparator string = ";"

// DateFormat is the date format used in the cursor token.
var DateFormat string = time.RFC3339

var (
	// ErrInvalidCursor is an error that is returned when the cursor token is invalid.
	ErrInvalidCursor = errors.New("invalid cursor token")

	// ErrMustBeOneOrGreater is an error that is returned when the value is less than one.
	ErrMustBeOneOrGreater = errors.New("limit must be one or greater")
)

// Paginator represents a paginator.
type Paginator struct {
	// Next is the cursor token to the next page.
	Next string `json:"next,omitempty"`

	// Prev is the cursor token to the previous page.
	Prev string `json:"previous,omitempty"`

	// Limit is the maximum number of elements to return.
	Limit int `json:"limit,omitempty"`
}

func NewPaginator(next, prev string, limit int) *Paginator {
	return &Paginator{
		Next:  next,
		Prev:  prev,
		Limit: limit,
	}
}

// String returns the string representation of the paginator.
func (p *Paginator) String() string {
	limit := fmt.Sprintf("%d", p.Limit)
	return "Paginator{Next: " + p.Next + ", Prev: " + p.Prev + ", Limit: " + limit + "}"
}

// GenerateToken generates a token for the given id and date.
func (p *Paginator) GenerateToken(id uuid.UUID, date time.Time) string {
	return EncodeToken(id, date)
}

func (p *Paginator) Validate() error {
	if p.Limit <= 0 {
		return ErrMustBeOneOrGreater
	}

	// next should be a base64 encoded string
	if p.Next != "" {
		_, _, err := DecodeToken(p.Next)
		if err != nil {
			return ErrInvalidCursor
		}
	}

	// previous should be a base64 encoded string
	if p.Prev != "" {
		_, _, err := DecodeToken(p.Prev)
		if err != nil {
			return ErrInvalidCursor
		}
	}

	return nil
}

// EncodeToken encodes the date and id into a base64 string.
func EncodeToken(id uuid.UUID, date time.Time) string {
	payload := id.String() + DataSeparator + date.Format(DateFormat)
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

// DecodeToken decodes the string into a date and id.
func DecodeToken(s string) (id uuid.UUID, date time.Time, err error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}

	parts := strings.Split(string(data), DataSeparator)
	if len(parts) != 2 {
		return uuid.Nil, time.Time{}, ErrInvalidCursor
	}

	date, err = time.Parse(DateFormat, parts[1])
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}

	id, err = uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}

	return id, date, nil
}
