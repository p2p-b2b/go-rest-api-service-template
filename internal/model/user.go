package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
)

var (
	// ErrSortFieldTooLong is an error that is returned when the sort field is too long.
	ErrSortFieldTooLong = errors.New("sort field is too long")

	// ErrInvalidID is an error that is returned when the ID is not a valid UUID.
	ErrInvalidID = errors.New("invalid ID")

	// ErrInvalidField is an error that is returned when the field is not valid.
	ErrInvalidField = errors.New("invalid field")
)

var (
	// UserFilterFields is a list of valid fields for filtering users.
	UserFilterFields = []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"}

	// UserSortFields is a list of valid fields for sorting users.
	UserSortFields = []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"}
)

// User represents a user entity.
// @Description User information.
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
// @Description Create user request.
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

// UpdateUserRequest represents the input for the UpdateUser method.
// @Description Update user request.
type UpdateUserRequest struct {
	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`
}

// ListUserRequest represents the input for the ListUser method.
// @Description List user request.
type ListUserRequest struct {
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
// @Description Select all users query input.
type SelectAllUserQueryInput struct {
	// Sort is the field to sort by.
	Sort string `json:"sort,omitempty"`

	// Filter is the field to filter by.
	Filter string `json:"filter,omitempty"`

	// Fields is the fields to return.
	Fields []string `json:"fields,omitempty"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}

// SelectAllUserQueryOutput represents the output for the SelectAllUserQuery method.
// @Description Select all users query output.
type SelectAllUserQueryOutput struct {
	// Items is a list of users.
	Items []*User `json:"items"`

	// Paginator is the paginator for the list of users.
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}
