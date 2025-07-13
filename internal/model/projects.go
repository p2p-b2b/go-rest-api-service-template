package model

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/qfv"
)

const (
	ProjectNameMinLength        = 2
	ProjectNameMaxLength        = 70
	ProjectDescriptionMinLength = 2
	ProjectDescriptionMaxLength = 1024

	ProjectsProjectCreatedSuccessfully = "Project created successfully"
	ProjectsProjectUpdatedSuccessfully = "Project updated successfully"
	ProjectsProjectDeletedSuccessfully = "Project deleted successfully"
)

var (
	// ProjectFilterFields is a list of valid fields for filtering models.
	ProjectFilterFields = []string{"id", "name", "disabled", "created_at", "updated_at"}

	// ProjectSortFields is a list of valid fields for sorting models.
	ProjectSortFields = []string{"id", "name", "disabled", "created_at", "updated_at"}

	// ProjectPartialFields is a list of valid fields for partial responses.
	ProjectPartialFields = []string{"id", "name", "description", "disabled", "created_at", "updated_at"}
)

// Project represents a user entity used to model the data stored in the database.
//
// @Description Project represents a project.
type Project struct {
	CreatedAt   time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	Disabled    *bool     `json:"disabled,omitempty" example:"false" format:"boolean"`
	System      *bool     `json:"system,omitempty" example:"false" format:"boolean"`
	Name        string    `json:"name,omitempty" example:"John" format:"string"`
	Description string    `json:"description,omitempty" example:"This is a project" format:"string"`
	SerialID    int64     `json:"-"`
	ID          uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7aa2-bfc2-d862a423985c" format:"uuid"`
}

type InsertProjectInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	Disabled    bool
	System      bool
}

func (ref *InsertProjectInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate UserID
	if err := ValidateUUID(ref.UserID, 7, "user_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name
	normalizedName, err := ValidateString(ref.Name, StringValidationOptions{
		MinLength:      ProjectNameMinLength,
		MaxLength:      ProjectNameMaxLength,
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
		MinLength:      ProjectDescriptionMinLength,
		MaxLength:      ProjectDescriptionMaxLength,
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

type CreateProjectInput = InsertProjectInput

type UpdateProjectInput struct {
	Name        *string
	Description *string
	Disabled    *bool
	ID          uuid.UUID
	UserID      uuid.UUID
}

func (ref *UpdateProjectInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if err := ValidateUUID(ref.UserID, 7, "user_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if ref.Name != nil {
		normalizedName, err := ValidateString(*ref.Name, StringValidationOptions{
			MinLength:      ProjectNameMinLength,
			MaxLength:      ProjectNameMaxLength,
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

	if ref.Description != nil {
		normalizedDescription, err := ValidateString(*ref.Description, StringValidationOptions{
			MinLength:      ProjectDescriptionMinLength,
			MaxLength:      ProjectDescriptionMaxLength,
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

type DeleteProjectInput struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func (ref *DeleteProjectInput) Validate() error {
	var validationErrors ValidationErrors
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if err := ValidateUUID(ref.UserID, 7, "user_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type SelectProjectsInput struct {
	UserID    uuid.UUID
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectProjectsInput) Validate() error {
	var validationErrors ValidationErrors

	if err := ValidateUUID(ref.UserID, 7, "user_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate paginator
	if err := ref.Paginator.Validate(); err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ve.Errors...)
		} else {
			validationErrors.AddError("paginator", err.Error(), "INVALID_PAGINATION")
		}
	}

	// Validate sort expression
	if err := ValidateSortExpression(ref.Sort, "sort"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Additional sort validation using existing parser
	if ref.Sort != "" {
		if _, err := qfv.NewSortParser(ProjectSortFields).Parse(ref.Sort); err != nil {
			validationErrors.AddError("sort", err.Error(), "INVALID_SORT")
		}
	}

	// Validate filter expression
	if err := ValidateFilterExpression(ref.Filter, "filter"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Additional filter validation using existing parser
	if ref.Filter != "" {
		if _, err := qfv.NewFilterParser(ProjectFilterFields).Parse(ref.Filter); err != nil {
			validationErrors.AddError("filter", err.Error(), "INVALID_FILTER")
		}
	}

	// Validate fields expression
	if err := ValidateFieldsExpression(ref.Fields, "fields"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Additional fields validation using existing parser
	if ref.Fields != "" {
		if _, err := qfv.NewFieldsParser(ProjectPartialFields).Parse(ref.Fields); err != nil {
			validationErrors.AddError("fields", err.Error(), "INVALID_FIELDS")
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type ListProjectsInput = SelectProjectsInput

type SelectProjectsOutput struct {
	Items     []Project `json:"items"`
	Paginator Paginator `json:"paginator"`
}

type ListProjectsOutput = SelectProjectsOutput

// CreateProjectRequest represents the inputs necessary to create a new project.
//
// @Description CreateProjectRequest represents the inputs necessary to create a new project.
type CreateProjectRequest struct {
	Name        string    `json:"name" example:"New project name" format:"string" validate:"required"`
	Description string    `json:"description" example:"This is a new project" format:"string" validate:"required"`
	ID          uuid.UUID `json:"id" example:"01980434-b7ff-7aa6-a131-a7c3590a1ce1" format:"uuid" validate:"optional"`
}

func (req *CreateProjectRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(req.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name
	normalizedName, err := ValidateString(req.Name, StringValidationOptions{
		MinLength:      ProjectNameMinLength,
		MaxLength:      ProjectNameMaxLength,
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
		MinLength:      ProjectDescriptionMinLength,
		MaxLength:      ProjectDescriptionMaxLength,
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

// UpdateProjectRequest represents the inputs necessary to update a project.
//
// @Description UpdateProjectRequest represents the inputs necessary to update a project.
type UpdateProjectRequest struct {
	Name        *string `json:"name" example:"New project name" format:"string"`
	Description *string `json:"description" example:"This is a new project data" format:"string"`
	Disabled    *bool   `json:"disabled" example:"false" format:"boolean"`
}

func (req *UpdateProjectRequest) Validate() error {
	if reflect.DeepEqual(req, &UpdateProjectRequest{}) {
		return &ValidationError{
			Field:   "request",
			Message: "at least one field must be provided for update",
			Code:    "REQUIRED_FIELD",
		}
	}

	var validationErrors ValidationErrors

	if req.Name != nil {
		normalizedName, err := ValidateString(*req.Name, StringValidationOptions{
			MinLength:      ProjectNameMinLength,
			MaxLength:      ProjectNameMaxLength,
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

	if req.Description != nil {
		normalizedDescription, err := ValidateString(*req.Description, StringValidationOptions{
			MinLength:      ProjectDescriptionMinLength,
			MaxLength:      ProjectDescriptionMaxLength,
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

// ListProjectsResponse represents a list of users.
//
// @Description ListProjectsResponse represents a list of users.
type ListProjectsResponse = SelectProjectsOutput

// LinkUsersToProjectRequest represents the input for the LinkUserToProject method.
//
// @Description LinkUsersToProjectRequest represents the input for the LinkUserToProject method.
type LinkUsersToProjectRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" format:"uuid"`
}

func (req *LinkUsersToProjectRequest) Validate() error {
	var validationErrors ValidationErrors

	if len(req.UserIDs) < 1 {
		validationErrors.AddError("user_ids", "at least one user ID is required", "REQUIRED")
		return &validationErrors
	}

	for i, userID := range req.UserIDs {
		if err := ValidateUUID(userID, 7, fmt.Sprintf("user_ids[%d]", i)); err != nil {
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

// UnlinkUsersFromProjectRequest represents the input for the UnlinkUserFromProject method.
//
// @Description UnlinkUsersFromProjectRequest represents the input for the UnlinkUserFromProject method.
type UnlinkUsersFromProjectRequest = LinkUsersToProjectRequest
