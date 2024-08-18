package handler

import (
	"encoding/json"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
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
		return ErrInvalidUserID
	}

	if len(req.FirstName) < 2 {
		return ErrInvalidUserFirstName
	}

	if len(req.LastName) < 2 {
		return ErrInvalidUserLastName
	}

	// minimal email validation
	if len(req.Email) < 6 {
		return ErrInvalidUserEmail
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return ErrInvalidUserEmail
	}

	return nil
}

// UpdateUserRequest represents the input for the UpdateUser method.
type UpdateUserRequest struct {
	FirstName *string
	LastName  *string
	Email     *string
}

func (req *UpdateUserRequest) Validate() error {
	// check if req is equal to the empty struct
	if *req == (UpdateUserRequest{}) {
		return ErrAtLeastOneFieldMustBeUpdated
	}

	if req.FirstName != nil && *req.FirstName != "" && len(*req.FirstName) < 2 {
		return ErrInvalidUserFirstName
	}

	if req.LastName != nil && *req.LastName != "" && len(*req.LastName) < 2 {
		return ErrInvalidUserLastName
	}

	// minimal email validation
	if req.Email != nil && *req.Email != "" {
		if len(*req.Email) < 6 {
			return ErrInvalidUserEmail
		}

		_, err := mail.ParseAddress(*req.Email)
		if err != nil {
			return ErrInvalidUserEmail
		}
	}

	return nil
}

// ListUserResponse represents a list of users.
type ListUserResponse struct {
	Items     []*User             `json:"items"`
	Paginator paginator.Paginator `json:"paginator"`
}
