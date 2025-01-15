package service

import (
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
)

const (
	UsersFirstNameMinLength = 2
	UsersFirstNameMaxLength = 25
	UsersLastNameMinLength  = 2
	UsersLastNameMaxLength  = 25
	UsersEmailMinLength     = 6
	UsersEmailMaxLength     = 50
	UsersPasswordMinLength  = 6
	UsersPasswordMaxLength  = 255
)

var (
	ErrUserInvalidID          = errors.New("invalid user ID. Must be a valid UUID")
	ErrUserInvalidFirstName   = errors.New("invalid first name. Must be between " + fmt.Sprintf("%d and %d", UsersFirstNameMinLength, UsersFirstNameMaxLength) + " characters long")
	ErrUserInvalidLastName    = errors.New("invalid last name. Must be between " + fmt.Sprintf("%d and %d", UsersLastNameMinLength, UsersLastNameMaxLength) + " characters long")
	ErrUserInvalidEmail       = errors.New("invalid email. Must be between " + fmt.Sprintf("%d and %d", UsersEmailMinLength, UsersEmailMaxLength) + " characters long")
	ErrUserInvalidPassword    = errors.New("invalid password. Must be between " + fmt.Sprintf("%d and %d", UsersPasswordMinLength, UsersPasswordMaxLength) + " characters long")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserIDAlreadyExists    = errors.New("user ID already exists")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
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
}

// UserInput represents the common input for the user entity.
type UserInput struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Password  string
	Disabled  bool
}

// Validate validates the user input.
func (ui *UserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(ui.FirstName) < UsersFirstNameMinLength || len(ui.FirstName) > UsersFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(ui.LastName) < UsersLastNameMinLength || len(ui.LastName) > UsersLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	// minimal email validation
	if len(ui.Email) < UsersEmailMinLength || len(ui.Email) > UsersEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(ui.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(ui.Password) < UsersPasswordMinLength {
		return ErrUserInvalidPassword
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
	Password  *string
	Disabled  *bool
}

// Validate validates the UpdateUserInput.
func (ui *UpdateUserInput) Validate() error {
	if reflect.DeepEqual(ui, &UpdateUserInput{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if ui.ID == uuid.Nil {
		return ErrUserInvalidID
	}
	if ui.FirstName != nil && len(*ui.FirstName) < UsersFirstNameMinLength || len(*ui.FirstName) > UsersFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if ui.LastName != nil && len(*ui.LastName) < UsersLastNameMinLength || len(*ui.LastName) > UsersLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	if ui.Email != nil && *ui.Email != "" && len(*ui.Email) < UsersEmailMinLength || len(*ui.Email) > UsersEmailMaxLength {
		return ErrUserInvalidEmail
	}

	if ui.Email != nil && *ui.Email != "" && len(*ui.Email) >= UsersEmailMinLength && len(*ui.Email) <= UsersEmailMaxLength {
		_, err := mail.ParseAddress(*ui.Email)
		if err != nil {
			return ErrUserInvalidEmail
		}
	}

	if ui.Password != nil && len(*ui.Password) < UsersPasswordMinLength {
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

// ListUserInput represents the input for the ListUser method.
type ListUserInput struct {
	Sort      string
	Filter    string
	Fields    []string
	Paginator paginator.Paginator
}

// ListUsersOutput represents the output for the ListUser method.
type ListUsersOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}
