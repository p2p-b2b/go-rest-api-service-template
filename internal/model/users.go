package model

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

var (
	ErrInputIsNil                   = errors.New("input is nil")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
	ErrInvalidLimit                 = errors.New("invalid limit field")
	ErrInvalidSort                  = errors.New("invalid sort field")
	ErrInvalidFilter                = errors.New("invalid filter field")
	ErrInvalidFields                = errors.New("invalid fields field")
	ErrInvalidNextToken             = errors.New("invalid nextToken field")
	ErrInvalidPrevToken             = errors.New("invalid prevToken field")
)

const (
	ValidUserFirstNameMinLength = 2
	ValidUserFirstNameMaxLength = 25
	ValidUserLastNameMinLength  = 2
	ValidUserLastNameMaxLength  = 25
	ValidUserEmailMinLength     = 6
	ValidUserEmailMaxLength     = 50
	ValidUserPasswordMinLength  = 6
	ValidUserPasswordMaxLength  = 255
)

var (
	ErrUserInvalidID          = errors.New("invalid user ID. Must be a valid UUID")
	ErrUserInvalidFirstName   = errors.New("invalid first name. Must be between " + fmt.Sprintf("%d and %d", ValidUserFirstNameMinLength, ValidUserFirstNameMaxLength) + " characters long")
	ErrUserInvalidLastName    = errors.New("invalid last name. Must be between " + fmt.Sprintf("%d and %d", ValidUserLastNameMinLength, ValidUserLastNameMaxLength) + " characters long")
	ErrUserInvalidEmail       = errors.New("invalid email. Must be between " + fmt.Sprintf("%d and %d", ValidUserEmailMinLength, ValidUserEmailMaxLength) + " characters long")
	ErrUserInvalidPassword    = errors.New("invalid password. Must be between " + fmt.Sprintf("%d and %d", ValidUserPasswordMinLength, ValidUserPasswordMaxLength) + " characters long")
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
//
// @Description User represents a user entity
type User struct {
	ID           uuid.UUID `json:"id,omitempty,omitzero" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	FirstName    string    `json:"first_name,omitempty" example:"John" format:"string"`
	LastName     string    `json:"last_name,omitempty" example:"Doe" format:"string"`
	Email        string    `json:"email,omitempty" example:"my@email.com" format:"email"`
	Password     string    `json:"-"`
	PasswordHash string    `json:"-"`
	Disabled     bool      `json:"disabled" example:"false" format:"boolean"`
	CreatedAt    time.Time `json:"created_at,omitempty,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt    time.Time `json:"updated_at,omitempty,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	SerialID     int64     `json:"-"`
}

type InsertUserInput struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	Password     string
	PasswordHash string
	Disabled     bool
}

func (ref *InsertUserInput) Validate() error {
	if ref.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(ref.FirstName) < ValidUserFirstNameMinLength || len(ref.FirstName) > ValidUserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(ref.LastName) < ValidUserLastNameMinLength || len(ref.LastName) > ValidUserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	if len(ref.Email) < ValidUserEmailMinLength || len(ref.Email) > ValidUserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(ref.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if ref.PasswordHash != "" {
		if len(ref.PasswordHash) < ValidUserPasswordMinLength || len(ref.PasswordHash) > ValidUserPasswordMaxLength {
			return ErrUserInvalidPassword
		}
	}

	if ref.Password != "" {
		if len(ref.Password) < ValidUserPasswordMinLength || len(ref.Password) > ValidUserPasswordMaxLength {
			return ErrUserInvalidPassword
		}
	}

	return nil
}

type CreateUserInput = InsertUserInput

type UpdateUserInput struct {
	ID           uuid.UUID
	FirstName    *string
	LastName     *string
	Email        *string
	Password     *string
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
		if len(*ref.FirstName) < ValidUserFirstNameMinLength || len(*ref.FirstName) > ValidUserFirstNameMaxLength {
			return ErrUserInvalidFirstName
		}
	}

	if ref.LastName != nil {
		if len(*ref.LastName) < ValidUserLastNameMinLength || len(*ref.LastName) > ValidUserLastNameMaxLength {
			return ErrUserInvalidLastName
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < ValidUserEmailMinLength || len(*ref.Email) > ValidUserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < ValidUserEmailMinLength || len(*ref.Email) > ValidUserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if ref.PasswordHash != nil {
		if len(*ref.PasswordHash) >= ValidUserPasswordMinLength || len(*ref.PasswordHash) <= ValidUserPasswordMaxLength {
			_, err := mail.ParseAddress(*ref.Email)
			if err != nil {
				return ErrUserInvalidEmail
			}
		}
	}

	if ref.PasswordHash != nil {
		if len(*ref.PasswordHash) < ValidUserPasswordMinLength || len(*ref.PasswordHash) > ValidUserPasswordMaxLength {
			return ErrUserInvalidPassword
		}
	}

	if ref.Password != nil {
		if len(*ref.Password) < ValidUserPasswordMinLength || len(*ref.Password) > ValidUserPasswordMaxLength {
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

type ListUsersInput = SelectUsersInput

type SelectUsersOutput struct {
	Items     []*User
	Paginator paginator.Paginator
}

type ListUsersOutput = SelectUsersOutput

// CreateUserRequest represents the input for the CreateUser method.
//
// @Description CreateUserRequest represents the input for the CreateUser method
type CreateUserRequest struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	FirstName string    `json:"first_name" example:"John" format:"string"`
	LastName  string    `json:"last_name" example:"Doe" format:"string"`
	Email     string    `json:"email" example:"my@email.com" format:"email"`
	Password  string    `json:"password" example:"ThisIs4Passw0rd" format:"string"`
}

// Validate validates the CreateUserRequest.
func (req *CreateUserRequest) Validate() error {
	if req.ID == uuid.Nil {
		return ErrUserInvalidID
	}

	if len(req.FirstName) < ValidUserFirstNameMinLength || len(req.FirstName) > ValidUserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(req.LastName) < ValidUserLastNameMinLength || len(req.LastName) > ValidUserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	// minimal email validation
	if len(req.Email) < ValidUserEmailMinLength || len(req.Email) > ValidUserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(req.Password) < ValidUserPasswordMinLength || len(req.Password) > ValidUserPasswordMaxLength {
		return ErrUserInvalidPassword
	}

	return nil
}

// UpdateUserRequest represents the input for the UpdateUser method.
//
// @Description UpdateUserRequest represents the input for the UpdateUser method
type UpdateUserRequest struct {
	FirstName *string `json:"first_name" example:"John" format:"string"`
	LastName  *string `json:"last_name" example:"Doe" format:"string"`
	Email     *string `json:"email" example:"my@email.com" format:"email"`
	Password  *string `json:"password" example:"ThisIs4Passw0rd" format:"string"`
	Disabled  *bool   `json:"disabled" example:"false" format:"boolean"`
}

func (req *UpdateUserRequest) Validate() error {
	if reflect.DeepEqual(req, &UpdateUserRequest{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if req.FirstName != nil {
		if len(*req.FirstName) < ValidUserFirstNameMinLength || len(*req.FirstName) > ValidUserFirstNameMaxLength {
			return ErrUserInvalidFirstName
		}
	}

	if req.LastName != nil {
		if len(*req.LastName) < ValidUserLastNameMinLength || len(*req.LastName) > ValidUserLastNameMaxLength {
			return ErrUserInvalidLastName
		}
	}

	// minimal email validation
	if req.Email != nil {
		if len(*req.Email) < ValidUserEmailMinLength || len(*req.Email) > ValidUserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if req.Email != nil {
		if len(*req.Email) >= ValidUserEmailMinLength && len(*req.Email) <= ValidUserEmailMaxLength {
			_, err := mail.ParseAddress(*req.Email)
			if err != nil {
				return ErrUserInvalidEmail
			}
		}
	}

	return nil
}

// ListUsersResponse represents a list of users.
//
// @Description ListUsersResponse represents a list of users
type ListUsersResponse struct {
	Items     []*User             `json:"items"`
	Paginator paginator.Paginator `json:"paginator"`
}
