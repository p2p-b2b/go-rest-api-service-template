package model

import (
	"encoding/json"
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
)

var (
	// ErrSortFieldTooLong is an error that is returned when the sort field is too long.
	ErrSortFieldTooLong = errors.New("sort field is too long")

	// ErrInvalidID is an error that is returned when the ID is not a valid UUID.
	ErrInvalidID = errors.New("invalid ID")

	// ErrInvalidFirstName is an error that is returned when the first name is not valid.
	ErrInvalidFirstName = errors.New("invalid first name, the first name must be at least 2 characters long")

	// ErrInvalidLastName is an error that is returned when the last name is not valid.
	ErrInvalidLastName = errors.New("invalid last name, the last name must be at least 2 characters long")

	// ErrInvalidEmail is an error that is returned when the email is not valid.
	ErrInvalidEmail = errors.New("invalid email")

	// ErrAtLeastOneFieldMustBeUpdated is an error that is returned when at least one field must be updated.
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
	// ID is the unique identifier of the user.
	ID uuid.UUID `json:"id,omitempty"`

	// FirstName is the first name of the user.
	FirstName string `json:"first_name,omitempty"`

	// LastName is the last name of the user.
	LastName string `json:"last_name,omitempty"`

	// Email is the email address of the user.
	Email string `json:"email,omitempty"`

	// Email is the email address of the user.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt is the time the user was last updated.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// SerialID is the serial number of the user used for pagination.
	SerialID int64 `json:"-"`
}

// MarshalJSON marshals the user into JSON.
// this is needed to omit zero values from the JSON output.
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User

	// Define an empty struct to hold omitted fields
	var omitted struct {
		Alias
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
		UpdatedAt string `json:"updated_at,omitempty"`
	}

	// Check for zero values and set them in the omitted struct
	if u.ID == uuid.Nil {
		omitted.ID = ""
	} else {
		omitted.ID = u.ID.String()
	}

	if u.CreatedAt.IsZero() {
		omitted.CreatedAt = ""
	} else {
		omitted.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	}

	if u.UpdatedAt.IsZero() {
		omitted.UpdatedAt = ""
	} else {
		omitted.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	}

	omitted.Alias = (Alias)(u)

	return json.Marshal(omitted)
}

// UserInput represents the common input for the user entity.
type UserInput struct {
	// ID is the unique identifier of the user.
	ID uuid.UUID `json:"id"`

	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`
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
type UpdateUserInput UserInput

// Validate validates the UpdateUserInput.
func (ui *UpdateUserInput) Validate() error {
	return (*UserInput)(ui).Validate()
}

// DeleteUserInput represents the input for the DeleteUser method.
type DeleteUserInput UserInput

// Validate validates the DeleteUserInput.
func (ui *DeleteUserInput) Validate() error {
	return (*UserInput)(ui).Validate()
}

// InsertUserInput represents the input for the InsertUser method.
type InsertUserInput UserInput

// Validate validates the InsertUserInput.
func (ui *InsertUserInput) Validate() error {
	return (*UserInput)(ui).Validate()
}

// CreateUserRequest represents the input for the CreateUser method.
type CreateUserRequest UserInput

// Validate validates the CreateUserRequest.
func (req *CreateUserRequest) Validate() error {
	return (*UserInput)(req).Validate()
}

// UpdateUserRequest represents the input for the UpdateUser method.
type UpdateUserRequest struct {
	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`
}

func (req *UpdateUserRequest) Validate() error {
	if len(req.FirstName) < 2 {
		return ErrInvalidFirstName
	}

	if len(req.LastName) < 2 {
		return ErrInvalidLastName
	}

	// minimal email validation
	if len(req.Email) < 6 {
		return ErrInvalidEmail
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return ErrInvalidEmail
	}

	// at least one field must be updated
	if req.FirstName == "" && req.LastName == "" && req.Email == "" {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	return nil
}

// ListUserInput represents the input for the ListUser method.
type ListUserInput struct {
	// Sort is the field to sort by.
	Sort string `json:"sort,omitempty"`

	// Filter is the field to filter by.
	Filter string `json:"filter,omitempty"`

	// Fields is the fields to return.
	Fields []string `json:"fields,omitempty"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}

// ListUserResponse represents a list of users.
type ListUserResponse struct {
	// Items is a list of users.
	Items []*User `json:"data"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}

// SelectAllUserQueryInput represents the input for the SelectAllUserQuery method.
type SelectAllUserQueryInput ListUserInput

// SelectAllUserQueryOutput represents the output for the SelectAllUserQuery method.
type SelectAllUserQueryOutput ListUserResponse
