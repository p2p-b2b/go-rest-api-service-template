package model

import (
	"time"

	"github.com/google/uuid"
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

// CreateUserInput represents the input for the CreateUser method.
type CreateUserInput struct {
	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// Email is the email address of the user.
	Email string `json:"email"`
}

// UpdateUserInput represents the input for the UpdateUser method.
type UpdateUserInput User

// DeleteUserInput represents the input for the DeleteUser method.
type DeleteUserInput User

// ListUserInput represents the input for the ListUser method.
type ListUserInput struct {
	// User is the user to filter by.
	User *User `json:"user,omitempty"`

	// Sort is the field to sort by.
	Sort string `json:"sort,omitempty"`

	// Order is the order to sort by.
	Order string `json:"order,omitempty"`

	// Filter is the field to filter by.
	Filter string `json:"filter,omitempty"`

	// Fields is the fields to return.
	Fields string `json:"fields,omitempty"`

	// Paginator is the paginator for the list of users.
	Paginator Paginator `json:"paginator,omitempty"`
}

// ListUserOutput represents a list of users.
type ListUserOutput struct {
	// Items is a list of users.
	Items []*User `json:"data"`

	// Paginator is the paginator for the list of users.
	Paginator Paginator `json:"paginator,omitempty"`
}
