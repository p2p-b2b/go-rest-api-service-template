package model

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
	MinLimit     int = 1
	MaxLimit     int = 100
)

// DataSeparator is the separator used to separate the data in the cursor token.
var DataSeparator string = ";"

var (
	ErrModelInvalidCursor = errors.New("invalid cursor token")
	ErrModelInvalidLimit  = errors.New("invalid limit. must be greater than " + strconv.Itoa(MinLimit) + " and less than or equal to " + strconv.Itoa(MaxLimit))
)

// Paginator represents a model.
//
// @Description Paginator represents a paginator
type Paginator struct {
	NextToken string `json:"next_token" example:"ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=" format:"string"`
	NextPage  string `json:"next_page" example:"http://localhost:8080/users?next_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=&limit=10" format:"string"`
	PrevToken string `json:"prev_token" example:"ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=" format:"string"`
	PrevPage  string `json:"prev_page" example:"http://localhost:8080/users?prev_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=&limit=10" format:"string"`
	Size      int    `json:"size" example:"10" format:"int"`
	Limit     int    `json:"limit" example:"10" format:"int"`
}

// String returns the string representation of the model.
func (ref *Paginator) String() string {
	limit := fmt.Sprintf("%d", ref.Limit)

	return fmt.Sprintf("Paginator{next: %s, next_token: %s, prev: %s, prev_token: %s, size: %d, limit: %s}",
		ref.NextPage,
		ref.NextToken,
		ref.PrevPage,
		ref.PrevToken,
		ref.Size,
		limit,
	)
}

// GenerateToken generates a token for the given id and date.
func (ref *Paginator) GenerateToken(id uuid.UUID, serial int64) string {
	return EncodeToken(id, serial)
}

// Validate validates the model.
func (ref *Paginator) Validate() error {
	if ref.Limit < MinLimit || ref.Limit > MaxLimit {
		return ErrModelInvalidLimit
	}

	// next should be a base64 encoded string
	if ref.NextToken != "" {
		_, _, err := DecodeToken(ref.NextToken)
		if err != nil {
			return ErrModelInvalidCursor
		}
	}

	// previous should be a base64 encoded string
	if ref.PrevToken != "" {
		_, _, err := DecodeToken(ref.PrevToken)
		if err != nil {
			return ErrModelInvalidCursor
		}
	}

	return nil
}

// GeneratePages generates the next and previous pages.
func (ref *Paginator) GeneratePages(url string) {
	if ref.NextToken != "" {
		ref.NextPage = url + "?next_token=" + ref.NextToken + "&limit=" + strconv.Itoa(ref.Limit)
	}
	if ref.PrevToken != "" {
		ref.PrevPage = url + "?prev_token=" + ref.PrevToken + "&limit=" + strconv.Itoa(ref.Limit)
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
		return uuid.Nil, 0, ErrModelInvalidCursor
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
