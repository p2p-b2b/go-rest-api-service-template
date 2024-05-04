package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserInput User

type UpdateUserInput User

type DeleteUserInput User

type ListUserOutput struct {
	Data []*User `json:"data"`

	// pagination
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}
