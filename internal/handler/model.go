package handler

import (
	"encoding/json"
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
)

var (
	ErrInvalidID                    = errors.New("invalid ID")
	ErrInvalidFirstName             = errors.New("invalid first name, the first name must be at least 2 characters long")
	ErrInvalidLastName              = errors.New("invalid last name, the last name must be at least 2 characters long")
	ErrInvalidEmail                 = errors.New("invalid email")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)

// User represents a user entity used to model the data stored in the database.
type User struct {
	ID        uuid.UUID `json:"id,omitempty"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// MarshalJSON marshals the user into JSON.
// this is needed to omit zero values from the JSON output.
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User

	// Define an empty struct to hold omitted fields
	var omitted struct {
		Alias
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
		UpdatedAt string `json:"updated_at,omitempty"`
	}

	// Check for zero values and set them in the omitted struct
	if u.ID == uuid.Nil {
		omitted.ID = ""
	} else {
		omitted.ID = u.ID.String()
	}

	if u.CreatedAt.IsZero() {
		omitted.CreatedAt = ""
	} else {
		omitted.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	}

	if u.UpdatedAt.IsZero() {
		omitted.UpdatedAt = ""
	} else {
		omitted.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	}

	omitted.Alias = (Alias)(u)

	return json.Marshal(omitted)
}

// CreateUserRequest represents the input for the CreateUser method.
type CreateUserRequest struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// Validate validates the CreateUserRequest.
func (req *CreateUserRequest) Validate() error {
	if req.ID == uuid.Nil {
		return ErrInvalidID
	}

	if len(req.FirstName) < 2 {
		return ErrInvalidFirstName
	}

	if len(req.LastName) < 2 {
		return ErrInvalidLastName
	}

	// minimal email validation
	if len(req.Email) < 6 {
		return ErrInvalidEmail
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return ErrInvalidEmail
	}

	return nil
}

// UpdateUserRequest represents the input for the UpdateUser method.
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func (req *UpdateUserRequest) Validate() error {
	if len(req.FirstName) < 2 {
		return ErrInvalidFirstName
	}

	if len(req.LastName) < 2 {
		return ErrInvalidLastName
	}

	// minimal email validation
	if len(req.Email) < 6 {
		return ErrInvalidEmail
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return ErrInvalidEmail
	}

	// at least one field must be updated
	if req.FirstName == "" && req.LastName == "" && req.Email == "" {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	return nil
}

// ListUserResponse represents a list of users.
type ListUserResponse struct {
	Items     []*User             `json:"items,omitempty"`
	Paginator paginator.Paginator `json:"paginator,omitempty"`
}
