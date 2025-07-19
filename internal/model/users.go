package model

import (
	"fmt"
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

	UsersUserCreatedSuccessfully          = "User created successfully"
	UsersUserUpdatedSuccessfully          = "User updated successfully"
	UsersUserDeletedSuccessfully          = "User deleted successfully"
	UsersRoleLinkedToUserSuccessfully     = "User role linked successfully"
	UsersRoleUnlinkedFromUserSuccessfully = "User role unlinked successfully"
	UsersUserFound                        = "User found"
)

var (
	// UsersFilterFields is a list of valid fields for filtering users.
	UsersFilterFields = []string{"id", "first_name", "last_name", "email", "disabled", "created_at", "updated_at"}

	// UsersSortFields is a list of valid fields for sorting users.
	UsersSortFields = []string{"id", "first_name", "last_name", "email", "disabled", "created_at", "updated_at"}

	// UsersPartialFields is a list of valid fields for partial responses.
	UsersPartialFields = []string{"id", "first_name", "last_name", "email", "disabled", "created_at", "updated_at"}
)

// User represents a user entity used to model the data stored in the database.
//
//	@Description	User represents a user entity.
type User struct {
	CreatedAt    time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt    time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	Disabled     *bool     `json:"disabled,omitempty" example:"false" format:"boolean"`
	Admin        *bool     `json:"admin,omitempty" example:"false" format:"boolean"`
	FirstName    string    `json:"first_name,omitempty" example:"John" format:"string"`
	LastName     string    `json:"last_name,omitempty" example:"Doe" format:"string"`
	Email        string    `json:"email,omitempty" example:"my@email.com" format:"email"`
	Password     string    `json:"-"`
	PasswordHash string    `json:"-"`
	SerialID     int64     `json:"-"`
	ID           uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7aae-95c6-051c9895119c" format:"uuid"`
}

type InsertUserInput struct {
	FirstName    string
	LastName     string
	Email        string
	Password     string
	PasswordHash string
	ID           uuid.UUID
}

func (ref *InsertUserInput) Validate() error {
	var errs ValidationErrors

	// Validate user ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate first name
	if _, err := ValidateString(ref.FirstName, StringValidationOptions{
		MinLength:        ValidUserFirstNameMinLength,
		MaxLength:        ValidUserFirstNameMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "first_name",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate last name
	if _, err := ValidateString(ref.LastName, StringValidationOptions{
		MinLength:        ValidUserLastNameMinLength,
		MaxLength:        ValidUserLastNameMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "last_name",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate email
	if _, err := ValidateEmail(ref.Email, "email"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate password hash if provided
	if ref.PasswordHash != "" {
		if _, err := ValidateString(ref.PasswordHash, StringValidationOptions{
			MinLength:      ValidUserPasswordMinLength,
			MaxLength:      ValidUserPasswordMaxLength,
			TrimWhitespace: false, // Don't trim password hashes
			AllowEmpty:     false,
			NoControlChars: true,
			NoNullBytes:    true,
			FieldName:      "password_hash",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate password if provided
	if ref.Password != "" {
		if err := ValidatePassword(ref.Password, "password"); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type CreateUserInput = InsertUserInput

type UpdateUserInput struct {
	FirstName    *string
	LastName     *string
	Email        *string
	Password     *string
	PasswordHash *string
	Disabled     *bool
	ID           uuid.UUID
}

func (ref *UpdateUserInput) Validate() error {
	var errs ValidationErrors

	// Check if at least one field is provided for update
	if reflect.DeepEqual(ref, &UpdateUserInput{}) {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "request",
			Message: "at least one field must be updated",
			Code:    "REQUIRED",
		})
	}

	// Validate user ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate first name if provided
	if ref.FirstName != nil {
		if _, err := ValidateString(*ref.FirstName, StringValidationOptions{
			MinLength:        ValidUserFirstNameMinLength,
			MaxLength:        ValidUserFirstNameMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "first_name",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate last name if provided
	if ref.LastName != nil {
		if _, err := ValidateString(*ref.LastName, StringValidationOptions{
			MinLength:        ValidUserLastNameMinLength,
			MaxLength:        ValidUserLastNameMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "last_name",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate email if provided
	if ref.Email != nil {
		if _, err := ValidateEmail(*ref.Email, "email"); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate password hash if provided
	if ref.PasswordHash != nil && *ref.PasswordHash != "" {
		if _, err := ValidateString(*ref.PasswordHash, StringValidationOptions{
			MinLength:      ValidUserPasswordMinLength,
			MaxLength:      ValidUserPasswordMaxLength,
			TrimWhitespace: false, // Don't trim password hashes
			AllowEmpty:     false,
			NoControlChars: true,
			NoNullBytes:    true,
			FieldName:      "password_hash",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate password if provided
	if ref.Password != nil && *ref.Password != "" {
		if err := ValidatePassword(*ref.Password, "password"); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type DeleteUserInput struct {
	ID uuid.UUID
}

func (ref *DeleteUserInput) Validate() error {
	// Validate UUID
	if err := ValidateUUID(ref.ID, 0, "id"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidUserIDError{Message: valErr.Message}
		}
		return &InvalidUserIDError{Message: "user ID cannot be empty or nil"}
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
	var errs ValidationErrors

	// Validate paginator
	if err := ref.Paginator.Validate(); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			errs.Errors = append(errs.Errors, *validationErr)
		} else if validationErrs, ok := err.(*ValidationErrors); ok {
			errs.Errors = append(errs.Errors, validationErrs.Errors...)
		} else {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "paginator",
				Message: err.Error(),
				Code:    "INVALID",
			})
		}
	}

	// Validate sort expression
	if err := ValidateSortExpression(ref.Sort, "sort"); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			errs.Errors = append(errs.Errors, *validationErr)
		}
	}

	// Validate filter expression
	if err := ValidateFilterExpression(ref.Filter, "filter"); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			errs.Errors = append(errs.Errors, *validationErr)
		}
	}

	// Validate fields expression
	if err := ValidateFieldsExpression(ref.Fields, "fields"); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			errs.Errors = append(errs.Errors, *validationErr)
		}
	}

	// Additional validation using existing parsers
	if ref.Sort != "" {
		if _, err := qfv.NewSortParser(UsersSortFields).Parse(ref.Sort); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "sort",
				Message: fmt.Sprintf("invalid sort expression: %v", err),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	if ref.Filter != "" {
		if _, err := qfv.NewFilterParser(UsersFilterFields).Parse(ref.Filter); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "filter",
				Message: fmt.Sprintf("invalid filter expression: %v", err),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	if ref.Fields != "" {
		if _, err := qfv.NewFieldsParser(UsersPartialFields).Parse(ref.Fields); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "fields",
				Message: fmt.Sprintf("invalid fields expression: %v", err),
				Code:    "INVALID_FORMAT",
			})
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type ListUsersInput = SelectUsersInput

type SelectUsersOutput struct {
	Items     []User    `json:"items"`
	Paginator Paginator `json:"paginator"`
}

type ListUsersOutput = SelectUsersOutput

type LinkRolesToUserInput struct {
	UserID  uuid.UUID
	RoleIDs []uuid.UUID
}

func (ref *LinkRolesToUserInput) Validate() error {
	var errs ValidationErrors

	// Validate user ID
	if err := ValidateUUID(ref.UserID, 7, "user_id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate role IDs array
	if len(ref.RoleIDs) == 0 {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "role_ids",
			Message: "at least one role ID is required",
			Code:    "REQUIRED",
		})
	}

	for i, roleID := range ref.RoleIDs {
		if err := ValidateUUID(roleID, 7, fmt.Sprintf("role_ids[%d]", i)); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type UnLinkRolesFromUsersInput = LinkRolesToUserInput

// CreateUserRequest represents the input for the CreateUser method.
//
//	@Description	Create user request.
type CreateUserRequest struct {
	FirstName string    `json:"first_name" example:"John" format:"string" validate:"required"`
	LastName  string    `json:"last_name" example:"Doe" format:"string" validate:"required"`
	Email     string    `json:"email" example:"my@email.com" format:"email" validate:"required"`
	Password  string    `json:"password" example:"ThisIs4Passw0rd" format:"string" validate:"required"`
	ID        uuid.UUID `json:"id" example:"01980434-b7ff-7ab2-b903-524ba1d47616" format:"uuid" validate:"optional"`
}

// Validate validates the CreateUserRequest.
func (req *CreateUserRequest) Validate() error {
	var errs ValidationErrors

	// Validate user ID
	if err := ValidateUUID(req.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate first name
	if _, err := ValidateString(req.FirstName, StringValidationOptions{
		MinLength:        ValidUserFirstNameMinLength,
		MaxLength:        ValidUserFirstNameMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "first_name",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate last name
	if _, err := ValidateString(req.LastName, StringValidationOptions{
		MinLength:        ValidUserLastNameMinLength,
		MaxLength:        ValidUserLastNameMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "last_name",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate email
	if _, err := ValidateEmail(req.Email, "email"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate password
	if err := ValidatePassword(req.Password, "password"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

// UpdateUserRequest represents the input for the UpdateUser method.
//
//	@Description	Update user request.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name" example:"John" format:"string" validate:"optional"`
	LastName  *string `json:"last_name" example:"Doe" format:"string" validate:"optional"`
	Email     *string `json:"email" example:"my@email.com" format:"email" validate:"optional"`
	Password  *string `json:"password" example:"ThisIs4Passw0rd" format:"string" validate:"optional"`
	Disabled  *bool   `json:"disabled" example:"false" format:"boolean" validate:"optional"`
}

func (req *UpdateUserRequest) Validate() error {
	var errs ValidationErrors

	// Check if at least one field is provided for update
	if reflect.DeepEqual(req, &UpdateUserRequest{}) {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "request",
			Message: "at least one field must be updated",
			Code:    "REQUIRED",
		})
	}

	// Validate first name if provided
	if req.FirstName != nil {
		if _, err := ValidateString(*req.FirstName, StringValidationOptions{
			MinLength:        ValidUserFirstNameMinLength,
			MaxLength:        ValidUserFirstNameMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "first_name",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate last name if provided
	if req.LastName != nil {
		if _, err := ValidateString(*req.LastName, StringValidationOptions{
			MinLength:        ValidUserLastNameMinLength,
			MaxLength:        ValidUserLastNameMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "last_name",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate email if provided
	if req.Email != nil {
		if _, err := ValidateEmail(*req.Email, "email"); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate password if provided
	if req.Password != nil {
		if err := ValidatePassword(*req.Password, "password"); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

// LinkRolesToUserRequest represents the input for the LinkRoles method.
//
//	@Description	Link roles request.
type LinkRolesToUserRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" format:"uuid" validate:"required"`
}

func (req *LinkRolesToUserRequest) Validate() error {
	var errs ValidationErrors

	// Check if at least one field is provided
	if reflect.DeepEqual(req, &LinkRolesToUserRequest{}) {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "request",
			Message: "at least one role ID must be provided",
			Code:    "REQUIRED",
		})
	}

	// Validate role IDs array
	if len(req.RoleIDs) == 0 {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "role_ids",
			Message: "at least one role ID is required",
			Code:    "REQUIRED",
		})
	}

	for i, roleID := range req.RoleIDs {
		if err := ValidateUUID(roleID, 7, fmt.Sprintf("role_ids[%d]", i)); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

// UnlinkRolesFromUserRequest represents the input for the UnLinkRoles method.
//
//	@Description	Unlink roles request.
type UnlinkRolesFromUserRequest = LinkRolesToUserRequest

// ListUsersResponse represents a list of users.
//
//	@Description	List of users.
type ListUsersResponse = SelectUsersOutput
