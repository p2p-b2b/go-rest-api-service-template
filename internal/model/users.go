package model

import (
	"net/mail"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/qfv"
)

const (
	ValidUserFirstNameMinLength = 2
	ValidUserFirstNameMaxLength = 25
	ValidUserLastNameMinLength  = 2
	ValidUserLastNameMaxLength  = 25
	ValidUserEmailMinLength     = 6
	ValidUserEmailMaxLength     = 50
	ValidUserPasswordMinLength  = 6
	ValidUserPasswordMaxLength  = 100

	UserUserCreatedSuccessfully          = "User created successfully"
	UserUserUpdatedSuccessfully          = "User updated successfully"
	UserUserDeletedSuccessfully          = "User deleted successfully"
	UserRoleLinkedToUserSuccessfully     = "User role linked successfully"
	UserRoleUnlinkedFromUserSuccessfully = "User role unlinked successfully"
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
	Disabled     *bool     `json:"disabled,omitempty" example:"false" format:"boolean"`
	CreatedAt    time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt    time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
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
		return &InvalidUserIDError{ID: ref.ID.String()}
	}

	if len(ref.FirstName) < ValidUserFirstNameMinLength || len(ref.FirstName) > ValidUserFirstNameMaxLength {
		return &InvalidUserFirstNameError{MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength}
	}

	if len(ref.LastName) < ValidUserLastNameMinLength || len(ref.LastName) > ValidUserLastNameMaxLength {
		return &InvalidUserLastNameError{MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength}
	}

	if len(ref.Email) < ValidUserEmailMinLength || len(ref.Email) > ValidUserEmailMaxLength {
		return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: ref.Email}
	}

	_, err := mail.ParseAddress(ref.Email)
	if err != nil {
		return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: ref.Email}
	}

	if ref.PasswordHash != "" {
		if len(ref.PasswordHash) < ValidUserPasswordMinLength || len(ref.PasswordHash) > ValidUserPasswordMaxLength {
			return &InvalidUserPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
	}

	if ref.Password != "" {
		if len(ref.Password) < ValidUserPasswordMinLength || len(ref.Password) > ValidUserPasswordMaxLength {
			return &InvalidUserPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
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
		return &InvalidUserIDError{ID: ref.ID.String()}
	}

	if ref.FirstName != nil {
		if len(*ref.FirstName) < ValidUserFirstNameMinLength || len(*ref.FirstName) > ValidUserFirstNameMaxLength {
			return &InvalidUserFirstNameError{MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength}
		}
	}

	if ref.LastName != nil {
		if len(*ref.LastName) < ValidUserLastNameMinLength || len(*ref.LastName) > ValidUserLastNameMaxLength {
			return &InvalidUserLastNameError{MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength}
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < ValidUserEmailMinLength || len(*ref.Email) > ValidUserEmailMaxLength {
			return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: *ref.Email}
		}
	}

	if ref.Email != nil {
		if len(*ref.Email) < ValidUserEmailMinLength || len(*ref.Email) > ValidUserEmailMaxLength {
			return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: *ref.Email}
		}
	}

	if ref.PasswordHash != nil {
		if len(*ref.PasswordHash) >= ValidUserPasswordMinLength || len(*ref.PasswordHash) <= ValidUserPasswordMaxLength {
			_, err := mail.ParseAddress(*ref.Email)
			if err != nil {
				return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: *ref.Email}
			}
		}
	}

	if ref.PasswordHash != nil {
		if len(*ref.PasswordHash) < ValidUserPasswordMinLength || len(*ref.PasswordHash) > ValidUserPasswordMaxLength {
			return &InvalidUserPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
	}

	if ref.Password != nil {
		if len(*ref.Password) < ValidUserPasswordMinLength || len(*ref.Password) > ValidUserPasswordMaxLength {
			return &InvalidUserPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
	}

	return nil
}

type DeleteUserInput struct {
	ID uuid.UUID
}

func (ref *DeleteUserInput) Validate() error {
	if ref.ID == uuid.Nil {
		return &InvalidUserIDError{ID: ref.ID.String()}
	}

	return nil
}

type SelectUsersInput struct {
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectUsersInput) Validate() error {
	if err := ref.Paginator.Validate(); err != nil {
		return err
	}

	if ref.Sort != "" {
		_, err := qfv.NewSortParser(UserSortFields).Parse(ref.Sort)
		if err != nil {
			return err
		}
	}

	if ref.Filter != "" {
		_, err := qfv.NewFilterParser(UserFilterFields).Parse(ref.Filter)
		if err != nil {
			return err
		}
	}

	if ref.Fields != "" {
		_, err := qfv.NewFieldsParser(UserFilterFields).Parse(ref.Fields)
		if err != nil {
			return err
		}
	}

	return nil
}

type ListUsersInput = SelectUsersInput

type SelectUsersOutput struct {
	Items     []User
	Paginator Paginator
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
		return &InvalidUserIDError{ID: req.ID.String()}
	}

	if len(req.FirstName) < ValidUserFirstNameMinLength || len(req.FirstName) > ValidUserFirstNameMaxLength {
		return &InvalidUserFirstNameError{MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength}
	}

	if len(req.LastName) < ValidUserLastNameMinLength || len(req.LastName) > ValidUserLastNameMaxLength {
		return &InvalidUserLastNameError{MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength}
	}

	// minimal email validation
	if len(req.Email) < ValidUserEmailMinLength || len(req.Email) > ValidUserEmailMaxLength {
		return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: req.Email}
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: req.Email}
	}

	if len(req.Password) < ValidUserPasswordMinLength || len(req.Password) > ValidUserPasswordMaxLength {
		return &InvalidUserPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
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
			return &InvalidUserFirstNameError{MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength}
		}
	}

	if req.LastName != nil {
		if len(*req.LastName) < ValidUserLastNameMinLength || len(*req.LastName) > ValidUserLastNameMaxLength {
			return &InvalidUserLastNameError{MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength}
		}
	}

	// minimal email validation
	if req.Email != nil {
		if len(*req.Email) < ValidUserEmailMinLength || len(*req.Email) > ValidUserEmailMaxLength {
			return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: *req.Email}
		}
	}

	if req.Email != nil {
		if len(*req.Email) >= ValidUserEmailMinLength && len(*req.Email) <= ValidUserEmailMaxLength {
			_, err := mail.ParseAddress(*req.Email)
			if err != nil {
				return &InvalidUserEmailError{MinLength: ValidUserEmailMinLength, MaxLength: ValidUserEmailMaxLength, Email: *req.Email}
			}
		}
	}

	return nil
}

// ListUsersResponse represents a list of users.
//
// @Description ListUsersResponse represents a list of users
type ListUsersResponse struct {
	Items     []User    `json:"items"`
	Paginator Paginator `json:"paginator"`
}
