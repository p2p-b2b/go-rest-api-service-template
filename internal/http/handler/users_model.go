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
	ErrUserInvalidID            = errors.New("invalid user ID, this must be a valid UUID")
	ErrUserInvalidFirstName     = errors.New("invalid user first name. Must be between" + fmt.Sprintf("%d and %d", UserFirstNameMinLength, UserFirstNameMaxLength) + "characters long")
	ErrUserInvalidLastName      = errors.New("invalid user last name. Must be between" + fmt.Sprintf("%d and %d", UserLastNameMinLength, UserLastNameMaxLength) + "characters long")
	ErrUserInvalidEmail         = errors.New("invalid user email. Must be between" + fmt.Sprintf("%d and %d", UserEmailMinLength, UserEmailMaxLength) + "characters long")
	ErrUserInvalidPassword      = errors.New("invalid user password. Must be at least" + fmt.Sprintf("%d characters long", UserPasswordMinLength) + "characters long")
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

	if len(req.FirstName) < UserFirstNameMinLength || len(req.FirstName) > UserFirstNameMaxLength {
		return ErrUserInvalidFirstName
	}

	if len(req.LastName) < UserLastNameMinLength || len(req.LastName) > UserLastNameMaxLength {
		return ErrUserInvalidLastName
	}

	// minimal email validation
	if len(req.Email) < UserEmailMinLength || len(req.Email) > UserEmailMaxLength {
		return ErrUserInvalidEmail
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(req.Password) < UserPasswordMinLength {
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
		if len(*req.FirstName) < UserFirstNameMinLength || len(*req.FirstName) > UserFirstNameMaxLength {
			return ErrUserInvalidFirstName
		}
	}

	if req.LastName != nil {
		if len(*req.LastName) < UserLastNameMinLength || len(*req.LastName) > UserLastNameMaxLength {
			return ErrUserInvalidLastName
		}
	}

	// minimal email validation
	if req.Email != nil {
		if len(*req.Email) < UserEmailMinLength || len(*req.Email) > UserEmailMaxLength {
			return ErrUserInvalidEmail
		}
	}

	if req.Email != nil {
		if len(*req.Email) >= UserEmailMinLength && len(*req.Email) <= UserEmailMaxLength {
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
