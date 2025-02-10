package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
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
	ErrUserInvalidID            = errors.New("invalid user ID, this must be a valid UUID")
	ErrUserInvalidFirstName     = errors.New("invalid user first name. Must be between" + fmt.Sprintf("%d and %d", ValidUserFirstNameMinLength, ValidUserFirstNameMaxLength) + "characters long")
	ErrUserInvalidLastName      = errors.New("invalid user last name. Must be between" + fmt.Sprintf("%d and %d", ValidUserLastNameMinLength, ValidUserLastNameMaxLength) + "characters long")
	ErrUserInvalidEmail         = errors.New("invalid user email. Must be between" + fmt.Sprintf("%d and %d", ValidUserEmailMinLength, ValidUserEmailMaxLength) + "characters long")
	ErrUserInvalidPassword      = errors.New("invalid user password. Must be at least" + fmt.Sprintf("%d characters long", ValidUserPasswordMinLength) + "characters long")
	ErrUserInvalidService       = errors.New("invalid service")
	ErrUserInvalidOpenTelemetry = errors.New("invalid open telemetry")
)

// User represents a user entity used to model the data stored in the database.
//
// @Description User represents a user entity
type User struct {
	ID        uuid.UUID `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	FirstName string    `json:"first_name,omitempty" example:"John" format:"string"`
	LastName  string    `json:"last_name,omitempty" example:"Doe" format:"string"`
	Email     string    `json:"email,omitempty" example:"my@email.com" format:"email"`
	Disabled  bool      `json:"disabled" example:"false" format:"boolean"`
	CreatedAt time.Time `json:"created_at,omitempty" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt time.Time `json:"updated_at,omitempty" example:"2021-01-01T00:00:00Z" format:"date-time"`
}

// MarshalJSON marshals the user into JSON.
// this is needed to omit zero values from the JSON output.
func (ref User) MarshalJSON() ([]byte, error) {
	type Alias User

	// Define an empty struct to hold omitted fields
	var omitted struct {
		Alias
		ID        string `json:"id,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
		UpdatedAt string `json:"updated_at,omitempty"`
	}

	// Check for zero values and set them in the omitted struct
	if ref.ID == uuid.Nil {
		omitted.ID = ""
	} else {
		omitted.ID = ref.ID.String()
	}

	if ref.CreatedAt.IsZero() {
		omitted.CreatedAt = ""
	} else {
		omitted.CreatedAt = ref.CreatedAt.Format(time.RFC3339)
	}

	if ref.UpdatedAt.IsZero() {
		omitted.UpdatedAt = ""
	} else {
		omitted.UpdatedAt = ref.UpdatedAt.Format(time.RFC3339)
	}

	omitted.Alias = (Alias)(ref)

	return json.Marshal(omitted)
}

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

	if len(req.Password) < ValidUserPasswordMinLength {
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
