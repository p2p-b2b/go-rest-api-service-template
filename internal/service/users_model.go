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
	ValidValidUserFirstNameMinLength = 2
	ValidValidUserFirstNameMaxLength = 25
	ValidValidUserLastNameMinLength  = 2
	ValidUserLastNameMaxLength       = 25
	ValidUserEmailMinLength          = 6
	ValidUserEmailMaxLength          = 50
	ValidUserPasswordMinLength       = 6
	ValidUserPasswordMaxLength       = 255
)

var (
	ErrUserInvalidID          = errors.New("invalid user ID. Must be a valid UUID")
	ErrUserInvalidFirstName   = errors.New("invalid first name. Must be between " + fmt.Sprintf("%d and %d", ValidValidUserFirstNameMinLength, ValidValidUserFirstNameMaxLength) + " characters long")
	ErrUserInvalidLastName    = errors.New("invalid last name. Must be between " + fmt.Sprintf("%d and %d", ValidValidUserLastNameMinLength, ValidUserLastNameMaxLength) + " characters long")
	ErrUserInvalidEmail       = errors.New("invalid email. Must be between " + fmt.Sprintf("%d and %d", ValidUserEmailMinLength, ValidUserEmailMaxLength) + " characters long")
	ErrUserInvalidPassword    = errors.New("invalid password. Must be between " + fmt.Sprintf("%d and %d", ValidUserPasswordMinLength, ValidUserPasswordMaxLength) + " characters long")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserIDAlreadyExists    = errors.New("user ID already exists")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
)

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

type CreateUserInput struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Password  string
	Disabled  bool
}

func (ref *CreateUserInput) Validate() error {
	if ref.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(ref.FirstName) < ValidValidUserFirstNameMinLength || len(ref.FirstName) > ValidValidUserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(ref.LastName) < ValidValidUserLastNameMinLength || len(ref.LastName) > ValidUserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	// minimal email validation
	if len(ref.Email) < ValidUserEmailMinLength || len(ref.Email) > ValidUserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(ref.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(ref.Password) < ValidUserPasswordMinLength {
		return ErrUserInvalidPassword
	}

	return nil
}

type UpdateUserInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
	Disabled  *bool
}

func (ref *UpdateUserInput) Validate() error {
	if reflect.DeepEqual(ref, &UpdateUserInput{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if ref.ID == uuid.Nil {
		return ErrUserInvalidID
	}
	if ref.FirstName != nil {
		if len(*ref.FirstName) < ValidValidUserFirstNameMinLength || len(*ref.FirstName) > ValidValidUserFirstNameMaxLength {
			return ErrUserInvalidFirstName
		}
	}

	if ref.LastName != nil {
		if len(*ref.LastName) < ValidValidUserLastNameMinLength || len(*ref.LastName) > ValidUserLastNameMaxLength {
			return ErrUserInvalidLastName
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < ValidUserEmailMinLength || len(*ref.Email) > ValidUserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) >= ValidUserEmailMinLength && len(*ref.Email) <= ValidUserEmailMaxLength {
			_, err := mail.ParseAddress(*ref.Email)
			if err != nil {
				return ErrUserInvalidEmail
			}
		}
	}

	if ref.Password != nil && len(*ref.Password) < ValidUserPasswordMinLength {
		return ErrUserInvalidPassword
	}

	return nil
}

type DeleteUserInput struct {
	ID uuid.UUID
}

func (ref *DeleteUserInput) Validate() error {
	if ref.ID == uuid.Nil {
		return ErrUserInvalidID
	}
	return nil
}

type ListUsersInput struct {
	Sort      string
	Filter    string
	Fields    []string
	Paginator paginator.Paginator
}

type ListUsersOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}
