package model

import (
	"time"

	"github.com/google/uuid"
)

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

type CreateUserInput User

type UpdateUserInput User

type DeleteUserInput User

type ListUserOutput struct {
	// Data is a list of users.
	Data []*User `json:"data"`

	// TotalCount is the total number of users.
	TotalCount int `json:"total_count"`

	// Page is the current page.
	Page int `json:"page"`

	// PageSize is the number of users per page.
	PageSize int `json:"page_size"`

	TotalPages int `json:"total_pages"`
}
