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
	UsersFirstNameMinLength = 2
	UsersFirstNameMaxLength = 255
	UsersLastNameMinLength  = 2
	UsersLastNameMaxLength  = 255
	UsersEmailMinLength     = 6
	UsersEmailMaxLength     = 255
	UsersPasswordMinLength  = 6
	UsersPasswordMaxLength  = 255
)

var (
	ErrInvalidUserID        = errors.New("invalid user ID. Must be a valid UUID")
	ErrInvalidUserFirstName = errors.New("invalid first name. Must be between " + fmt.Sprintf("%d and %d", UsersFirstNameMinLength, UsersFirstNameMaxLength) + " characters long")
	ErrInvalidUserLastName  = errors.New("invalid last name. Must be between " + fmt.Sprintf("%d and %d", UsersLastNameMinLength, UsersLastNameMaxLength) + " characters long")
	ErrInvalidUserEmail     = errors.New("invalid email. Must be between " + fmt.Sprintf("%d and %d", UsersEmailMinLength, UsersEmailMaxLength) + " characters long")
	ErrInvalidUserPassword  = errors.New("invalid password. Must be between " + fmt.Sprintf("%d and %d", UsersPasswordMinLength, UsersPasswordMaxLength) + " characters long")
	ErrUserNotFound         = errors.New("user not found")

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

// UserInput represents the common input for the user entity.
type UserInput struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Disabled     bool
}

// Validate validates the user input.
func (ui *UserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrInvalidUserID
	}

	if len(ui.FirstName) < UsersFirstNameMinLength || len(ui.FirstName) > UsersFirstNameMaxLength {
		return ErrInvalidUserFirstName
	}

	if len(ui.LastName) < UsersLastNameMinLength || len(ui.LastName) > UsersLastNameMaxLength {
		return ErrInvalidUserLastName
	}

	if len(ui.Email) < UsersEmailMinLength || len(ui.Email) > UsersEmailMaxLength {
		return ErrInvalidUserEmail
	}

	_, err := mail.ParseAddress(ui.Email)
	if err != nil {
		return ErrInvalidUserEmail
	}

	if len(ui.PasswordHash) < UsersPasswordMinLength {
		return ErrInvalidUserPassword
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
		return ErrInvalidUserID
	}

	if ui.FirstName != nil && *ui.FirstName != "" && len(*ui.FirstName) < UsersFirstNameMinLength || len(*ui.FirstName) > UsersFirstNameMaxLength {
		return ErrInvalidUserFirstName
	}

	if ui.LastName != nil && *ui.LastName != "" && len(*ui.LastName) < UsersLastNameMinLength || len(*ui.LastName) > UsersLastNameMaxLength {
		return ErrInvalidUserLastName
	}

	if ui.Email != nil && *ui.Email != "" && len(*ui.Email) < UsersEmailMinLength || len(*ui.Email) > UsersEmailMaxLength {
		return ErrInvalidUserEmail
	}

	if ui.Email != nil && *ui.Email != "" && len(*ui.Email) < UsersEmailMinLength || len(*ui.Email) > UsersEmailMaxLength {
		return ErrInvalidUserEmail
	}

	if ui.PasswordHash != nil && *ui.PasswordHash != "" && len(*ui.PasswordHash) >= UsersPasswordMinLength && len(*ui.PasswordHash) <= UsersPasswordMaxLength {
		_, err := mail.ParseAddress(*ui.Email)
		if err != nil {
			return ErrInvalidUserEmail
		}
	}

	if ui.PasswordHash != nil && *ui.PasswordHash != "" && len(*ui.PasswordHash) < UsersPasswordMinLength {
		return ErrInvalidUserPassword
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
		return ErrInvalidUserID
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
		if !query.IsValidFields(UserPartialFields, field) {
			return ErrInvalidFields
		}
	}

	return nil
}

// SelectUsersInput represents the common input for the select user method.
type SelectUsersInput ListUserInput

// Validate validates the SelectUsersInput.
func (ui *SelectUsersInput) Validate() error {
	return (*ListUserInput)(ui).Validate()
}

// SelectUsersOutput represents the output for the list user method.
type SelectUsersOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}
