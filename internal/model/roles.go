package model

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

const (
	RoleNameMinLength        = 2
	RoleNameMaxLength        = 100
	RoleDescriptionMinLength = 2
	RoleDescriptionMaxLength = 1000

	RolesRoleCreatedSuccessfully      = "Role created successfully"
	RolesRoleUpdatedSuccessfully      = "Role updated successfully"
	RolesRoleDeletedSuccessfully      = "Role deleted successfully"
	RolesPoliciesLinkedSuccessfully   = "Policies linked successfully"
	RolesPoliciesUnlinkedSuccessfully = "Policies unlinked successfully"
	RolesUsersLinkedSuccessfully      = "Users linked successfully"
	RolesUsersUnlinkedSuccessfully    = "Users unlinked successfully"
)

var (
	// RolesFilterFields is a list of valid fields for filtering models.
	RolesFilterFields = []string{"id", "name", "system", "auto_assign", "created_at", "updated_at"}

	// RolesSortFields is a list of valid fields for sorting models.
	RolesSortFields = []string{"id", "name", "system", "auto_assign", "created_at", "updated_at"}

	// RolesPartialFields is a list of valid fields for partial responses.
	RolesPartialFields = []string{"id", "policy", "name", "description", "system", "auto_assign", "created_at", "updated_at"}
)

// Role represents a user entity used to model the data stored in the database.
//
// @Description Role represents a role.
type Role struct {
	CreatedAt   time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	System      *bool     `json:"system,omitempty,omitzero" example:"false" format:"boolean"`
	AutoAssign  *bool     `json:"auto_assign,omitempty,omitzero" example:"false" format:"boolean"`
	Name        string    `json:"name,omitempty" example:"Role Name" format:"string"`
	Description string    `json:"description,omitempty" example:"This is a role" format:"string"`
	SerialID    int64     `json:"-"`
	ID          uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7ab6-8c97-3e2f8905173a" format:"uuid"`
}

type InsertRoleInput struct {
	Name        string
	Description string
	ID          uuid.UUID
}

func (ref *InsertRoleInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name
	normalizedName, err := ValidateString(ref.Name, StringValidationOptions{
		MinLength:      RoleNameMinLength,
		MaxLength:      RoleNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "name",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		ref.Name = normalizedName
	}

	// Validate Description
	normalizedDescription, err := ValidateString(ref.Description, StringValidationOptions{
		MinLength:      RoleDescriptionMinLength,
		MaxLength:      RoleDescriptionMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "description",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		ref.Description = normalizedDescription
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type CreateRoleInput = InsertRoleInput

type UpdateRoleInput struct {
	ID          uuid.UUID
	Name        *string
	Description *string
}

func (ref *UpdateRoleInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name (if provided)
	if ref.Name != nil {
		normalizedName, err := ValidateString(*ref.Name, StringValidationOptions{
			MinLength:      RoleNameMinLength,
			MaxLength:      RoleNameMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "name",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*ref.Name = normalizedName
		}
	}

	// Validate Description (if provided)
	if ref.Description != nil {
		normalizedDescription, err := ValidateString(*ref.Description, StringValidationOptions{
			MinLength:      RoleDescriptionMinLength,
			MaxLength:      RoleDescriptionMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "description",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*ref.Description = normalizedDescription
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type DeleteRoleInput struct {
	ID uuid.UUID
}

func (ref *DeleteRoleInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type SelectRolesInput struct {
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectRolesInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate Paginator
	if err := ref.Paginator.Validate(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		} else if ves, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ves.Errors...)
		}
	}

	// Validate Sort
	if ref.Sort != "" {
		if err := ValidateSortExpression(ref.Sort, "sort"); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		}
	}

	// Validate Filter
	if ref.Filter != "" {
		if err := ValidateFilterExpression(ref.Filter, "filter"); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		}
	}

	// Validate Fields
	if ref.Fields != "" {
		if err := ValidateFieldsExpression(ref.Fields, "fields"); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type ListRolesInput = SelectRolesInput

type SelectRolesOutput struct {
	Items     []Role    `json:"items"`
	Paginator Paginator `json:"paginator"`
}

type ListRolesOutput = SelectRolesOutput

type LinkUsersToRoleInput struct {
	UserIDs []uuid.UUID
	RoleID  uuid.UUID
}

func (ref *LinkUsersToRoleInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate RoleID
	if err := ValidateUUID(ref.RoleID, 7, "role_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate UserIDs
	if len(ref.UserIDs) < 1 {
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "user_ids",
			Message: "user_ids must be a list of valid UUIDs",
		})
	} else {
		for i, userID := range ref.UserIDs {
			if err := ValidateUUID(userID, 7, fmt.Sprintf("user_ids[%d]", i)); err != nil {
				if ve, ok := err.(*ValidationError); ok {
					validationErrors.Errors = append(validationErrors.Errors, *ve)
				}
			}
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type UnlinkUsersFromRoleInput = LinkUsersToRoleInput

type LinkPoliciesToRoleInput struct {
	RoleID    uuid.UUID
	PolicyIDs []uuid.UUID
}

func (ref *LinkPoliciesToRoleInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate RoleID
	if err := ValidateUUID(ref.RoleID, 7, "role_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate PolicyIDs
	if len(ref.PolicyIDs) < 1 {
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "policy_ids",
			Message: "policy_ids must be a list of valid UUIDs",
		})
	} else {
		for i, policyID := range ref.PolicyIDs {
			if err := ValidateUUID(policyID, 7, fmt.Sprintf("policy_ids[%d]", i)); err != nil {
				if ve, ok := err.(*ValidationError); ok {
					validationErrors.Errors = append(validationErrors.Errors, *ve)
				}
			}
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type UnlinkPoliciesFromRoleInput = LinkPoliciesToRoleInput

// CreateRoleRequest represents the input for the CreateRole method.
//
// @Description CreateRoleRequest represents the input for the CreateRole method.
type CreateRoleRequest struct {
	Name        string    `json:"name" example:"New role name" format:"string" validate:"required"`
	Description string    `json:"description" example:"This is a role" format:"string" validate:"required"`
	ID          uuid.UUID `json:"id" example:"01980434-b7ff-7aba-a3ef-1b38309c9a1f" format:"uuid" validate:"optional"`
}

// Validate validates the CreateRoleRequest.
func (req *CreateRoleRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(req.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name
	normalizedName, err := ValidateString(req.Name, StringValidationOptions{
		MinLength:      RoleNameMinLength,
		MaxLength:      RoleNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "name",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		req.Name = normalizedName
	}

	// Validate Description
	normalizedDescription, err := ValidateString(req.Description, StringValidationOptions{
		MinLength:      RoleDescriptionMinLength,
		MaxLength:      RoleDescriptionMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "description",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		req.Description = normalizedDescription
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// UpdateRoleRequest represents the input for the UpdateRole method.
//
// @Description UpdateRoleRequest represents the input for the UpdateRole method.
type UpdateRoleRequest struct {
	Name        *string `json:"name" example:"Modified role name" format:"string" validate:"optional"`
	Description *string `json:"description" example:"This is a role" format:"string" validate:"optional"`
}

func (req *UpdateRoleRequest) Validate() error {
	var validationErrors ValidationErrors

	// Check if any field is provided for update
	if reflect.DeepEqual(req, &UpdateRoleRequest{}) {
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "request",
			Message: "at least one field must be provided for update",
		})
		return &validationErrors
	}

	// Validate Name (if provided)
	if req.Name != nil {
		normalizedName, err := ValidateString(*req.Name, StringValidationOptions{
			MinLength:      RoleNameMinLength,
			MaxLength:      RoleNameMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "name",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*req.Name = normalizedName
		}
	}

	// Validate Description (if provided)
	if req.Description != nil {
		normalizedDescription, err := ValidateString(*req.Description, StringValidationOptions{
			MinLength:      RoleDescriptionMinLength,
			MaxLength:      RoleDescriptionMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "description",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*req.Description = normalizedDescription
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// ListRolesResponse represents a list of users.
//
// @Description ListRoleResponse represents a list of users.
type ListRolesResponse = SelectRolesOutput

// LinkUsersToRoleRequest input values for linking users to a role.
//
// @Description LinkUsersToRoleRequest input values for linking users to a role.
type LinkUsersToRoleRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" format:"uuid" validate:"required"`
}

func (req *LinkUsersToRoleRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate UserIDs
	if len(req.UserIDs) < 1 {
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "user_ids",
			Message: "user_ids must be a list of valid UUIDs",
		})
	} else {
		for i, userID := range req.UserIDs {
			if err := ValidateUUID(userID, 7, fmt.Sprintf("user_ids[%d]", i)); err != nil {
				if ve, ok := err.(*ValidationError); ok {
					validationErrors.Errors = append(validationErrors.Errors, *ve)
				}
			}
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// UnlinkUsersFromRoleRequest input values for unlinking users from a role.
//
// @Description UnlinkUsersFromRoleRequest input values for unlinking users from a role.
type UnlinkUsersFromRoleRequest = LinkUsersToRoleRequest

// LinkPoliciesToRoleRequest input values for linking policies to a role.
//
// @Description LinkPoliciesToRoleRequest input values for linking policies to a role.
type LinkPoliciesToRoleRequest struct {
	PolicyIDs []uuid.UUID `json:"policy_ids" format:"uuid" validate:"required"`
}

func (req *LinkPoliciesToRoleRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate PolicyIDs
	if len(req.PolicyIDs) < 1 {
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "policy_ids",
			Message: "policy_ids must be a list of valid UUIDs",
		})
	} else {
		for i, policyID := range req.PolicyIDs {
			if err := ValidateUUID(policyID, 7, fmt.Sprintf("policy_ids[%d]", i)); err != nil {
				if ve, ok := err.(*ValidationError); ok {
					validationErrors.Errors = append(validationErrors.Errors, *ve)
				}
			}
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// UnlinkPoliciesFromRoleRequest input values for unlinking policies from a role.
//
// @Description UnlinkPoliciesFromRoleRequest input values for unlinking policies from a role.
type UnlinkPoliciesFromRoleRequest = LinkPoliciesToRoleRequest
