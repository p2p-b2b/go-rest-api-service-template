package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/qfv"
)

var (
	// ResourcesFilterFields is a list of valid fields for filtering permissions.
	ResourcesFilterFields = []string{"id", "name", "action", "resource", "system", "created_at", "updated_at"}

	// ResourcesSortFields is a list of valid fields for sorting permissions.
	ResourcesSortFields = []string{"id", "name", "action", "resource", "system", "created_at", "updated_at"}

	// ResourcesPartialFields is a list of valid fields for partial responses.
	ResourcesPartialFields = []string{"id", "name", "description", "action", "resource", "system", "created_at", "updated_at"}
)

// Resource represents a permission.
//
//	@Description	Resource represents a permission.
type Resource struct {
	ID          uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7aaa-a09c-d46077eff792" format:"uuid"`
	Name        string    `json:"name,omitempty" example:"Read Users" format:"string"`
	Description string    `json:"description,omitempty" example:"Allows reading of users" format:"string"`
	Action      string    `json:"action,omitempty" example:"GET" format:"string"`
	Resource    string    `json:"resource,omitempty" example:"users" format:"string"`
	System      *bool     `json:"system,omitempty,omitzero" example:"false" format:"bool"`
	CreatedAt   time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	SerialID    int64     `json:"-"`
}

type SelectResourcesInput struct {
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectResourcesInput) Validate() error {
	var errs ValidationErrors

	// Validate paginator
	if err := ref.Paginator.Validate(); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			errs.Errors = append(errs.Errors, *validationErr)
		} else if validationErrs, ok := err.(*ValidationErrors); ok {
			errs.Errors = append(errs.Errors, validationErrs.Errors...)
		} else {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "paginator",
				Message: err.Error(),
				Code:    "INVALID",
			})
		}
	}

	// Validate sort parameter
	if ref.Sort != "" {
		if _, err := qfv.NewSortParser(ResourcesSortFields).Parse(ref.Sort); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "sort",
				Message: fmt.Sprintf("invalid sort expression: %v", err),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	// Validate filter parameter
	if ref.Filter != "" {
		if _, err := qfv.NewFilterParser(ResourcesFilterFields).Parse(ref.Filter); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "filter",
				Message: fmt.Sprintf("invalid filter expression: %v", err),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	// Validate fields parameter
	if ref.Fields != "" {
		if _, err := qfv.NewFieldsParser(ResourcesPartialFields).Parse(ref.Fields); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "fields",
				Message: fmt.Sprintf("invalid fields expression: %v", err),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

// UniqueID generates a unique ID based on the Paginator's field values.
// It uses SHA-256 hashing of a formatted string representation of the fields
// to ensure a consistent and collision-resistant ID.
func (ref *SelectResourcesInput) UniqueID() string {
	// 1. Create a new SHA-256 hash instance.
	//    The hash.Hash interface implements io.Writer.
	h := sha256.New()

	// 2. Write the fields to the hash function in a deterministic order.
	//    Using a separator (like a null byte '\x00' or another unambiguous character)
	//    prevents collisions like ("ab", "c") vs ("a", "bc").
	//    fmt.Fprintf is convenient as it writes formatted data directly to the io.Writer (the hasher).
	//    It automatically handles the conversion of integers to their string representation.
	fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s",
		ref.Sort,
		ref.Filter,
		ref.Fields,
		ref.Paginator.UniqueID(),
	)

	// 3. Get the hash value and convert it to a hexadecimal string.
	//    h.Sum(nil) appends the hash to a new nil slice and returns it.
	hashBytes := h.Sum(nil)

	// 4. Encode the byte slice into a hexadecimal string.
	//    This provides a standard, readable string representation of the hash.
	return hex.EncodeToString(hashBytes)
}

type ListResourcesInput = SelectResourcesInput

type SelectResourcesOutput struct {
	Items     []Resource `json:"items"`
	Paginator Paginator  `json:"paginator"`
}

type ListResourcesOutput = SelectResourcesOutput

// ListResourcesResponse represents a list of users.
//
//	@Description	ListResourcesResponse represents a list of users.
type ListResourcesResponse = SelectResourcesOutput
