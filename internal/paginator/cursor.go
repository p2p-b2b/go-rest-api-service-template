package paginator

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultLimit int = 10
	MinLimit     int = 2
	MaxLimit     int = 100
)

// DataSeparator is the separator used to separate the data in the cursor token.
var DataSeparator string = ";"

var (
	ErrInvalidCursor      = errors.New("invalid cursor token")
	ErrMustBeOneOrGreater = errors.New("limit must be one or greater")
)

// Paginator represents a paginator.
type Paginator struct {
	NextToken string `json:"next_token"`
	NextPage  string `json:"next_page"`
	PrevToken string `json:"prev_token"`
	PrevPage  string `json:"prev_page"`
	Size      int    `json:"size"`
	Limit     int    `json:"limit"`
}

// String returns the string representation of the paginator.
func (p *Paginator) String() string {
	limit := fmt.Sprintf("%d", p.Limit)

	return fmt.Sprintf("Paginator{next: %s, next_token: %s, prev: %s, prev_token: %s, size: %d, limit: %s}",
		p.NextPage,
		p.NextToken,
		p.PrevPage,
		p.PrevToken,
		p.Size,
		limit,
	)
}

// GenerateToken generates a token for the given id and date.
func (p *Paginator) GenerateToken(id uuid.UUID, serial int64) string {
	return EncodeToken(id, serial)
}

// Validate validates the paginator.
func (p *Paginator) Validate() error {
	if p.Limit <= 0 {
		return ErrMustBeOneOrGreater
	}

	// next should be a base64 encoded string
	if p.NextToken != "" {
		_, _, err := DecodeToken(p.NextToken)
		if err != nil {
			return ErrInvalidCursor
		}
	}

	// previous should be a base64 encoded string
	if p.PrevToken != "" {
		_, _, err := DecodeToken(p.PrevToken)
		if err != nil {
			return ErrInvalidCursor
		}
	}

	return nil
}

// GeneratePages generates the next and previous pages.
func (p *Paginator) GeneratePages(url string) {
	if p.NextToken != "" {
		p.NextPage = url + "?next_token=" + p.NextToken + "&limit=" + strconv.Itoa(p.Limit)
	}
	if p.PrevToken != "" {
		p.PrevPage = url + "?prev_token=" + p.PrevToken + "&limit=" + strconv.Itoa(p.Limit)
	}
}

// EncodeToken encodes the date and id
// into a base64 string after joining them with a separator.
// use the package variables DataSeparator and DateFormat
// to set the separator and the date format.
func EncodeToken(id uuid.UUID, serial int64) string {
	payload := id.String() + DataSeparator + fmt.Sprintf("%d", serial)
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

// DecodeToken decodes the string into a date and id.
func DecodeToken(s string) (id uuid.UUID, serial int64, err error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, 0, err
	}

	parts := strings.Split(string(data), DataSeparator)
	if len(parts) != 2 {
		return uuid.Nil, 0, ErrInvalidCursor
	}

	serial, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return uuid.Nil, 0, err
	}

	id, err = uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, 0, err
	}

	return id, serial, nil
}

// GetTokens returns the next and previous tokens based on the length and limit conditions.
// size is the number of elements in the current page.
// limit is the maximum number of elements to return.
// firstID is the ID of the first element in the current page.
// firstSerial is the serial number of the first element in the current page.
// lastID is the ID of the last element in the current page.
// lastSerial  is the serial number of the last element in the current page.
func GetTokens(size int, limit int, firstID uuid.UUID, firstSerial int64, lastID uuid.UUID, lastSerial int64) (next string, prev string) {
	if size == 0 || size < limit {
		next = ""
		prev = ""
	}

	if size >= limit {
		next = EncodeToken(lastID, lastSerial)
		prev = EncodeToken(firstID, firstSerial)
	}

	return
}
