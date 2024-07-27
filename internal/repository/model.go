package repository

import (
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/query"
)

var (
	ErrInvalidID                    = errors.New("invalid ID")
	ErrInvalidFirstName             = errors.New("invalid first name, the first name must be at least 2 characters long")
	ErrInvalidLastName              = errors.New("invalid last name, the last name must be at least 2 characters long")
	ErrInvalidEmail                 = errors.New("invalid email")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)

var (
	// UserFilterFields is a list of valid fields for filtering users.
	UserFilterFields = []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"}

	// UserSortFields is a list of valid fields for sorting users.
	UserSortFields = []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"}

	// UserFields is a list of valid fields for partial responses.
	UserFields = []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"}
)

// User represents a user entity used to model the data stored in the database.
type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	SerialID  int64
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

// InsertUserInput represents the input for the CreateUser method.
type InsertUserInput UserInput

// Validate validates the CreateUserInput.
func (ui *InsertUserInput) Validate() error {
	return (*UserInput)(ui).Validate()
}

// UpdateUserInput represents the input for the UpdateUser method.
type UpdateUserInput UserInput

// Validate validates the UpdateUserInput.
func (ui *UpdateUserInput) Validate() error {
	return (*UserInput)(ui).Validate()
}

// DeleteUserInput represents the input for the DeleteUser method.
type DeleteUserInput struct {
	// ID is the unique identifier of the user.
	ID uuid.UUID `json:"id"`
}

// Validate validates the DeleteUserInput.
func (ui *DeleteUserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrInvalidID
	}

	return nil
}

// ListUserInput represents the common input for the list user method.
type ListUserInput struct {
	Sort      string
	Filter    string
	Fields    []string
	Paginator paginator.Paginator
}

// Validate validates the ListUserInput.
func (ui *ListUserInput) Validate() error {
	if ui.Paginator.Limit < 1 {
		return ErrInvalidLimit
	}

	if ui.Paginator.NextToken != "" {
		if _, err := uuid.Parse(ui.Paginator.NextToken); err != nil {
			return ErrInvalidNextToken
		}
	}

	if ui.Paginator.PrevToken != "" {
		if _, err := uuid.Parse(ui.Paginator.PrevToken); err != nil {
			return ErrInvalidPrevToken
		}
	}

	if ui.Sort != "" && !query.IsValidSort(UserSortFields, ui.Sort) {
		return ErrInvalidSort
	}

	if ui.Filter != "" && !query.IsValidFilter(UserFilterFields, ui.Filter) {
		return ErrInvalidFilter
	}

	for _, field := range ui.Fields {
		if !query.IsValidFields(UserFields, field) {
			return ErrInvalidFields
		}
	}

	return nil
}

// SelectUserInput represents the common input for the select user method.
type SelectUserInput ListUserInput

// ListUserOutput represents the output for the list user method.
type SelectUserOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}