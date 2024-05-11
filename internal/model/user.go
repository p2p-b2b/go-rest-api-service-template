package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/paginator"
)

var (
	// ErrSortFieldTooLong is an error that is returned when the sort field is too long.
	ErrSortFieldTooLong = errors.New("sort field is too long")

	// ErrInvalidID is an error that is returned when the ID is not a valid UUID.
	ErrInvalidID = errors.New("invalid ID")
)

// User represents a user.
type User struct {
	// ID is the unique identifier of the user.
	ID uuid.UUID `json:"id"`

	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`

	// Email is the email address of the user.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the time the user was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the input for the CreateUser method.
type CreateUserRequest struct {
	// ID is the unique identifier of the user.
	ID uuid.UUID `json:"id"`

	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`
}

// CreateUserResponse represents the output for the CreateUser method.
type CreateUserResponse User

// UpdateUserRequest represents the input for the UpdateUser method.
type UpdateUserRequest struct {
	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`
}

// DeleteUserInput represents the input for the DeleteUser method.
type DeleteUserInput User

// ListUserRequest represents the input for the ListUser method.
type ListUserRequest struct {
	// Sort is the field to sort by.
	Sort string `json:"sort,omitempty"`

	// Filter is the field to filter by.
	Filter []string `json:"filter,omitempty"`

	// Fields is the fields to return.
	Fields []string `json:"fields,omitempty"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}

func (l *ListUserRequest) Validate() error {
	if len(l.Sort) > 32 {
		return ErrSortFieldTooLong
	}

	if len(l.Fields) > 0 && len(l.Fields) < 15 {
		for _, field := range l.Fields {
			if field != "id" && field != "first_name" && field != "last_name" && field != "email" && field != "created_at" && field != "updated_at" {
				return errors.New("invalid field")
			}
		}
	}

	return nil
}

// ListUserResponse represents a list of users.
type ListUserResponse struct {
	// Items is a list of users.
	Items []*User `json:"data"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}

type SelectAllUserQueryInput struct {
	// Sort is the field to sort by.
	Sort string `json:"sort,omitempty"`

	// Filter is the field to filter by.
	Filter []string `json:"filter,omitempty"`

	// Fields is the fields to return.
	Fields []string `json:"fields,omitempty"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}

type SelectAllUserQueryOutput struct {
	// Items is a list of users.
	Items []*User `json:"data"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}
