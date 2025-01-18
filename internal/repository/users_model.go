package repository

import (
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/query"
)

const (
	UserFirstNameMinLength = 2
	UserFirstNameMaxLength = 25
	UserLastNameMinLength  = 2
	UserLastNameMaxLength  = 25
	UserEmailMinLength     = 6
	UserEmailMaxLength     = 50
	UserPasswordMinLength  = 6
	UserPasswordMaxLength  = 255
)

var (
	ErrUserInvalidID          = errors.New("invalid user ID. Must be a valid UUID")
	ErrUserInvalidFirstName   = errors.New("invalid first name. Must be between " + fmt.Sprintf("%d and %d", UserFirstNameMinLength, UserFirstNameMaxLength) + " characters long")
	ErrUserInvalidLastName    = errors.New("invalid last name. Must be between " + fmt.Sprintf("%d and %d", UserLastNameMinLength, UserLastNameMaxLength) + " characters long")
	ErrUserInvalidEmail       = errors.New("invalid email. Must be between " + fmt.Sprintf("%d and %d", UserEmailMinLength, UserEmailMaxLength) + " characters long")
	ErrUserInvalidPassword    = errors.New("invalid password. Must be between " + fmt.Sprintf("%d and %d", UserPasswordMinLength, UserPasswordMaxLength) + " characters long")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserIDAlreadyExists    = errors.New("user ID already exists")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
)

var (
	// UserFilterFields is a list of valid fields for filtering users.
	UserFilterFields = []string{"id", "first_name", "last_name", "email", "disabled", "created_at", "updated_at"}

	// UserSortFields is a list of valid fields for sorting users.
	UserSortFields = []string{"id", "first_name", "last_name", "email", "disabled", "created_at", "updated_at"}

	// UserPartialFields is a list of valid fields for partial responses.
	UserPartialFields = []string{"id", "first_name", "last_name", "email", "disabled", "created_at", "updated_at"}
)

// User represents a user entity used to model the data stored in the database.
type User struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Disabled     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	SerialID     int64
}

// InsertUserInput represents the input for the InsertUser method.
type InsertUserInput struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Disabled     bool
}

// Validate validates the InsertUserInput.
func (ui *InsertUserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(ui.FirstName) < UserFirstNameMinLength ||
		len(ui.FirstName) > UserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(ui.LastName) < UserLastNameMinLength ||
		len(ui.LastName) > UserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	if len(ui.Email) < UserEmailMinLength ||
		len(ui.Email) > UserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(ui.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(ui.PasswordHash) < UserPasswordMinLength ||
		len(ui.PasswordHash) > UserPasswordMaxLength {
		return ErrUserInvalidPassword
	}

	return nil
}

// UpdateUserInput represents the input for the UpdateUser method.
type UpdateUserInput struct {
	ID           uuid.UUID
	FirstName    *string
	LastName     *string
	Email        *string
	PasswordHash *string
	Disabled     *bool
	UpdatedAt    *time.Time
}

// Validate validates the UpdateUserInput.
func (ui *UpdateUserInput) Validate() error {
	if reflect.DeepEqual(ui, &UpdateUserInput{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if ui.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if ui.FirstName != nil && *ui.FirstName != "" &&
		len(*ui.FirstName) < UserFirstNameMinLength ||
		len(*ui.FirstName) > UserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if ui.LastName != nil && *ui.LastName != "" &&
		len(*ui.LastName) < UserLastNameMinLength ||
		len(*ui.LastName) > UserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	if ui.Email != nil && *ui.Email != "" &&
		len(*ui.Email) < UserEmailMinLength ||
		len(*ui.Email) > UserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	if ui.Email != nil && *ui.Email != "" &&
		len(*ui.Email) < UserEmailMinLength ||
		len(*ui.Email) > UserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	if ui.PasswordHash != nil && *ui.PasswordHash != "" &&
		len(*ui.PasswordHash) >= UserPasswordMinLength ||
		len(*ui.PasswordHash) <= UserPasswordMaxLength {
		_, err := mail.ParseAddress(*ui.Email)
		if err != nil {
			return ErrUserInvalidEmail
		}
	}

	if ui.PasswordHash != nil && *ui.PasswordHash != "" &&
		len(*ui.PasswordHash) < UserPasswordMinLength ||
		len(*ui.PasswordHash) > UserPasswordMaxLength {
		return ErrUserInvalidPassword
	}

	return nil
}

// DeleteUserInput represents the input for the DeleteUser method.
type DeleteUserInput struct {
	ID uuid.UUID
}

// Validate validates the DeleteUserInput.
func (ui *DeleteUserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	return nil
}

// SelectUsersInput represents the common input for the list user method.
type SelectUsersInput struct {
	Sort      string
	Filter    string
	Fields    []string
	Paginator paginator.Paginator
}

// Validate validates the SelectUsersInput.
func (ui *SelectUsersInput) Validate() error {
	if ui.Paginator.Limit < 1 {
		return ErrInvalidLimit
	}

	if ui.Sort != "" && !query.IsValidSort(UserSortFields, ui.Sort) {
		return ErrInvalidSort
	}

	if ui.Filter != "" && !query.IsValidFilter(UserFilterFields, ui.Filter) {
		return ErrInvalidFilter
	}

	for _, field := range ui.Fields {
		if !query.IsValidFields(UserPartialFields, field) {
			return ErrInvalidFields
		}
	}

	return nil
}

// SelectUsersOutput represents the output for the list user method.
type SelectUsersOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}
