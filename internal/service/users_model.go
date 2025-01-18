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

// CreateUserInput represents the common input for the user entity.
type CreateUserInput struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Password  string
	Disabled  bool
}

// Validate validates the user input.
func (ui *CreateUserInput) Validate() error {
	if ui.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(ui.FirstName) < UserFirstNameMinLength || len(ui.FirstName) > UserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(ui.LastName) < UserLastNameMinLength || len(ui.LastName) > UserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	// minimal email validation
	if len(ui.Email) < UserEmailMinLength || len(ui.Email) > UserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(ui.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(ui.Password) < UserPasswordMinLength {
		return ErrUserInvalidPassword
	}

	return nil
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
	if ui.FirstName != nil {
		if len(*ui.FirstName) < UserFirstNameMinLength || len(*ui.FirstName) > UserFirstNameMaxLength {
			return ErrUserInvalidFirstName
		}
	}

	if ui.LastName != nil {
		if len(*ui.LastName) < UserLastNameMinLength || len(*ui.LastName) > UserLastNameMaxLength {
			return ErrUserInvalidLastName
		}
	}

	if ui.Email != nil {
		if len(*ui.Email) < UserEmailMinLength || len(*ui.Email) > UserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if ui.Email != nil {
		if len(*ui.Email) >= UserEmailMinLength && len(*ui.Email) <= UserEmailMaxLength {
			_, err := mail.ParseAddress(*ui.Email)
			if err != nil {
				return ErrUserInvalidEmail
			}
		}
	}

	if ui.Password != nil && len(*ui.Password) < UserPasswordMinLength {
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

// ListUsersInput represents the input for the ListUser method.
type ListUsersInput struct {
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
