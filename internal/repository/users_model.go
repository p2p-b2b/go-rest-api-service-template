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

type InsertUserInput struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Disabled     bool
}

func (ref *InsertUserInput) Validate() error {
	if ref.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(ref.FirstName) < UserFirstNameMinLength || len(ref.FirstName) > UserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(ref.LastName) < UserLastNameMinLength || len(ref.LastName) > UserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	if len(ref.Email) < UserEmailMinLength || len(ref.Email) > UserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(ref.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(ref.PasswordHash) < UserPasswordMinLength || len(ref.PasswordHash) > UserPasswordMaxLength {
		return ErrUserInvalidPassword
	}

	return nil
}

type UpdateUserInput struct {
	ID           uuid.UUID
	FirstName    *string
	LastName     *string
	Email        *string
	PasswordHash *string
	Disabled     *bool
	UpdatedAt    *time.Time
}

func (ref *UpdateUserInput) Validate() error {
	if reflect.DeepEqual(ref, &UpdateUserInput{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if ref.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if ref.FirstName != nil {
		if len(*ref.FirstName) < UserFirstNameMinLength || len(*ref.FirstName) > UserFirstNameMaxLength {
			return ErrUserInvalidFirstName
		}
	}

	if ref.LastName != nil {
		if len(*ref.LastName) < UserLastNameMinLength || len(*ref.LastName) > UserLastNameMaxLength {
			return ErrUserInvalidLastName
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < UserEmailMinLength || len(*ref.Email) > UserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < UserEmailMinLength || len(*ref.Email) > UserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if ref.PasswordHash != nil {
		if len(*ref.PasswordHash) >= UserPasswordMinLength || len(*ref.PasswordHash) <= UserPasswordMaxLength {
			_, err := mail.ParseAddress(*ref.Email)
			if err != nil {
				return ErrUserInvalidEmail
			}
		}
	}

	if ref.PasswordHash != nil {
		if len(*ref.PasswordHash) < UserPasswordMinLength || len(*ref.PasswordHash) > UserPasswordMaxLength {
			return ErrUserInvalidPassword
		}
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

type SelectUsersInput struct {
	Sort      string
	Filter    string
	Fields    []string
	Paginator paginator.Paginator
}

func (ref *SelectUsersInput) Validate() error {
	if ref.Paginator.Limit < 1 {
		return ErrInvalidLimit
	}

	if ref.Sort != "" && !query.IsValidSort(UserSortFields, ref.Sort) {
		return ErrInvalidSort
	}

	if ref.Filter != "" && !query.IsValidFilter(UserFilterFields, ref.Filter) {
		return ErrInvalidFilter
	}

	for _, field := range ref.Fields {
		if !query.IsValidFields(UserPartialFields, field) {
			return ErrInvalidFields
		}
	}

	return nil
}

type SelectUsersOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}
