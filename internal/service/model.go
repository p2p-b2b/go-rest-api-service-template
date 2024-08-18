package service

import (
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
)

// User represents a user entity used to model the data stored in the database.
type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserInput represents the common input for the user entity.
type UserInput struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}

// Validate validates the user input.
func (ui *UserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrInvalidID
	}

	if len(ui.FirstName) < 2 {
		return ErrInvalidFirstName
	}

	if len(ui.LastName) < 2 {
		return ErrInvalidLastName
	}

	// minimal email validation
	if len(ui.Email) < 6 {
		return ErrInvalidEmail
	}

	_, err := mail.ParseAddress(ui.Email)
	if err != nil {
		return ErrInvalidEmail
	}

	return nil
}

// CreateUserInput represents the input for the CreateUser method.
type CreateUserInput UserInput

// Validate validates the CreateUserInput.
func (ui *CreateUserInput) Validate() error {
	return (*UserInput)(ui).Validate()
}

// UpdateUserInput represents the input for the UpdateUser method.
type UpdateUserInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
}

// Validate validates the UpdateUserInput.
func (ui *UpdateUserInput) Validate() error {
	if reflect.DeepEqual(ui, &UpdateUserInput{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if ui.ID == uuid.Nil {
		return ErrInvalidID
	}
	if ui.FirstName != nil && len(*ui.FirstName) < 2 {
		return ErrInvalidFirstName
	}

	if ui.LastName != nil && len(*ui.LastName) < 2 {
		return ErrInvalidLastName
	}

	if ui.Email != nil && *ui.Email != "" {
		if len(*ui.Email) < 6 {
			return ErrInvalidEmail
		}

		_, err := mail.ParseAddress(*ui.Email)
		if err != nil {
			return ErrInvalidEmail
		}
	}

	return nil
}

// DeleteUserInput represents the input for the DeleteUser method.
type DeleteUserInput struct {
	ID uuid.UUID `json:"id"`
}

// Validate validates the DeleteUserInput.
func (ui *DeleteUserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrInvalidID
	}
	return nil
}

// ListUserInput represents the input for the ListUser method.
type ListUserInput struct {
	Sort      string
	Filter    string
	Fields    []string
	Paginator paginator.Paginator
}

// ListUserOutput represents the output for the ListUser method.
type ListUserOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}

// ------------------------------------------------------------
// Status is an enumeration of health statuses.
type Status bool

// Health statuses enumeration.
const (
	StatusUp   Status = true
	StatusDown Status = false
)

// String returns the string representation of the status.
func (s Status) String() string {
	if s == StatusUp {
		return "UP"
	}
	return "DOWN"
}

// MarshalJSON marshals the status to JSON.
func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// UnmarshalJSON unmarshals the status from JSON.
func (s *Status) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"UP"`:
		*s = StatusUp
	case `"DOWN"`:
		*s = StatusDown
	default:
		*s = StatusDown
	}
	return nil
}

// Check represents a health check.
type Check struct {
	Name   string
	Kind   string
	Status Status
	Data   map[string]interface{}
}

// Check represents a health check.
type Health struct {
	Status Status
	Checks []Check
}
