package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/qfv"
)

const (
	PolicyNameMinLength        = 2
	PolicyNameMaxLength        = 255
	PolicyDescriptionMinLength = 2
	PolicyDescriptionMaxLength = 1024

	PoliciesPolicyCreatedSuccessfully = "Policy created successfully"
	PoliciesPolicyUpdatedSuccessfully = "Policy updated successfully"
	PoliciesPolicyDeletedSuccessfully = "Policy deleted successfully"
	PoliciesRolesLinkedSuccessfully   = "Roles linked successfully"
	PoliciesRolesUnlinkedSuccessfully = "Roles unlinked successfully"

	// https://regex101.com/r/xIOyX2/2
	//                                                  1                      2 this is group of groups and optional 0 opr 7 times
	// this is composed by groups:  ^(/[letters and dash]{1,50} | \*{1} )     ( (/[letters and dash]{1,50}) | (/*{1}) | (/uuid) ){0,7}
	ValidResourceRegex = `^(\/[a-z_]{1,50}|\*{1})((\/[a-z_]{1,50})|(\/\*{1})|(\/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})){0,7}$`

	ValidActionsRegex    = `^(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD|\*)$`
	ValidUUIDOrStarRegex = `[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}|\*{1}`
)

var (
	// PoliciesFilterFields is a list of valid fields for filtering models.
	PoliciesFilterFields = []string{"id", "name", "allowed_action", "allowed_resource", "system", "created_at", "updated_at"}

	// PoliciesSortFields is a list of valid fields for sorting models.
	PoliciesSortFields = []string{"id", "name", "allowed_action", "allowed_resource", "system", "created_at", "updated_at"}

	// PoliciesPartialFields is a list of valid fields for partial responses.
	PoliciesPartialFields = []string{"id", "name", "description", "allowed_action", "allowed_resource", "system", "created_at", "updated_at"}
)

func GetValidActions() string {
	validStr := strings.Trim(ValidActionsRegex, "^()$")
	validStr = strings.ReplaceAll(validStr, "\\", "")
	validStr = strings.ReplaceAll(validStr, "|", ", ")

	return validStr
}

// ValidateAction validates the action string.
func ValidateAction(action string) (validate string, error error) {
	if action == "" {
		return "", fmt.Errorf("action cannot be empty")
	}

	re := regexp.MustCompile(ValidActionsRegex)
	if !re.MatchString(action) {
		return "", fmt.Errorf("invalid action: %s, must be one of %s in Uppercase", action, GetValidActions())
	}

	return action, nil
}

// ValidateResource validates the resource string.
func ValidateResource(resource string) (validate string, error error) {
	if resource == "" {
		return "", fmt.Errorf("resource cannot be empty")
	}

	re := regexp.MustCompile(ValidResourceRegex)
	if !re.MatchString(resource) {
		return "", fmt.Errorf("invalid resource: %s, do not match the required format", resource)
	}

	return resource, nil
}

// Policy represents a user entity used to model the data stored in the database.
//
//	@Description	Policy represents a role.
type Policy struct {
	ID              uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7a93-b5b4-ca4c73283131" format:"uuid"`
	Name            string    `json:"name,omitempty" example:"Policy Name" format:"string"`
	Description     string    `json:"description,omitempty" example:"This is a role" format:"string"`
	System          *bool     `json:"system,omitempty,omitzero" example:"false" format:"boolean"`
	Resource        Resource  `json:"resource,omitzero"`
	AllowedAction   string    `json:"allowed_action,omitempty" example:"GET" format:"string"`
	AllowedResource string    `json:"allowed_resource,omitempty" example:"/projects/39a4707f-536e-433f-8597-6fc0d53a724f/tokens" format:"string"`
	CreatedAt       time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt       time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	SerialID        int64     `json:"-"`
}

type LinkRolesToPolicyInput struct {
	PolicyID uuid.UUID
	RoleIDs  []uuid.UUID
}

func (ref *LinkRolesToPolicyInput) Validate() error {
	if reflect.DeepEqual(ref, &LinkRolesToPolicyInput{}) {
		return &ValidationError{
			Field:   "request",
			Message: "at least one field must be provided",
			Code:    "REQUIRED_FIELD",
		}
	}

	var validationErrors ValidationErrors

	// Validate PolicyID
	if err := ValidateUUID(ref.PolicyID, 7, "policy_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate RoleIDs array
	if len(ref.RoleIDs) == 0 {
		validationErrors.AddError("role_ids", "at least one role ID is required", "REQUIRED")
	}

	for i, roleID := range ref.RoleIDs {
		if err := ValidateUUID(roleID, 7, fmt.Sprintf("role_ids[%d]", i)); err != nil {
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

type UnlinkRolesFromPolicyInput = LinkRolesToPolicyInput

// LinkRolesToPolicyRequest links roles to a policy.
//
//	@Description	Link roles to a policy.
type LinkRolesToPolicyRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" example:"01980434-b7ff-7a96-b0c8-dbabed881cf5" format:"uuid" validate:"required"`
}

func (ref *LinkRolesToPolicyRequest) Validate() error {
	if reflect.DeepEqual(ref, &LinkRolesToPolicyRequest{}) {
		return &ValidationError{
			Field:   "request",
			Message: "at least one field must be provided",
			Code:    "REQUIRED_FIELD",
		}
	}

	var validationErrors ValidationErrors

	// Validate RoleIDs array
	if len(ref.RoleIDs) == 0 {
		validationErrors.AddError("role_ids", "at least one role ID is required", "REQUIRED")
	}

	for i, roleID := range ref.RoleIDs {
		if err := ValidateUUID(roleID, 7, fmt.Sprintf("role_ids[%d]", i)); err != nil {
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

// UnlinkRolesFromPolicyRequest unlinks roles from a policy.
//
//	@Description	Unlink roles from a policy.
type UnlinkRolesFromPolicyRequest = LinkRolesToPolicyRequest

type SelectPoliciesInput struct {
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectPoliciesInput) Validate() error {
	var validationErrors ValidationErrors

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
		if _, err := qfv.NewSortParser(PoliciesSortFields).Parse(ref.Sort); err != nil {
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
		if _, err := qfv.NewFilterParser(PoliciesFilterFields).Parse(ref.Filter); err != nil {
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
		if _, err := qfv.NewFieldsParser(PoliciesPartialFields).Parse(ref.Fields); err != nil {
			validationErrors.AddError("fields", err.Error(), "INVALID_FIELDS")
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type ListPoliciesInput = SelectPoliciesInput

type SelectPoliciesOutput struct {
	Items     []Policy  `json:"items"`
	Paginator Paginator `json:"paginator"`
}

type ListPoliciesOutput = SelectPoliciesOutput

type ListPoliciesResponse = SelectPoliciesOutput

type CreatePolicyInput struct {
	ID              uuid.UUID
	Name            string
	Description     string
	AllowedAction   string
	AllowedResource string
	ResourceID      uuid.UUID
}

func (ref *CreatePolicyInput) Validate() error {
	var errs ValidationErrors

	// Validate policy ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate name
	if _, err := ValidateString(ref.Name, StringValidationOptions{
		MinLength:        PolicyNameMinLength,
		MaxLength:        PolicyNameMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "name",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate description
	if _, err := ValidateString(ref.Description, StringValidationOptions{
		MinLength:        PolicyDescriptionMinLength,
		MaxLength:        PolicyDescriptionMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "description",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate allowed action
	if _, err := ValidateAction(ref.AllowedAction); err != nil {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "allowed_action",
			Message: err.Error(),
			Code:    "INVALID_ACTION",
		})
	}

	// Validate allowed resource
	if _, err := ValidateResource(ref.AllowedResource); err != nil {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "allowed_resource",
			Message: err.Error(),
			Code:    "INVALID_RESOURCE",
		})
	}

	// Validate resource ID
	if err := ValidateUUID(ref.ResourceID, 7, "resource_id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type UpdatePolicyInput struct {
	ID              uuid.UUID
	Name            *string
	Description     *string
	AllowedAction   *string
	AllowedResource *string
}

func (ref *UpdatePolicyInput) Validate() error {
	var errs ValidationErrors

	// Validate policy ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate name if provided
	if ref.Name != nil {
		if _, err := ValidateString(*ref.Name, StringValidationOptions{
			MinLength:        PolicyNameMinLength,
			MaxLength:        PolicyNameMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "name",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate description if provided
	if ref.Description != nil {
		if _, err := ValidateString(*ref.Description, StringValidationOptions{
			MinLength:        PolicyDescriptionMinLength,
			MaxLength:        PolicyDescriptionMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "description",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate allowed action if provided
	if ref.AllowedAction != nil {
		if _, err := ValidateAction(*ref.AllowedAction); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "allowed_action",
				Message: err.Error(),
				Code:    "INVALID_ACTION",
			})
		}
	}

	// Validate allowed resource if provided
	if ref.AllowedResource != nil {
		if _, err := ValidateResource(*ref.AllowedResource); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "allowed_resource",
				Message: err.Error(),
				Code:    "INVALID_RESOURCE",
			})
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type DeletePolicyInput struct {
	ID uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7a9a-9145-07dd540fe352" format:"uuid" validate:"required"`
}

func (ref *DeletePolicyInput) Validate() error {
	var errs ValidationErrors

	// Validate policy ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

// CreatePolicyRequest represents a request to create a policy.
//
//	@Description	Create a policy.
type CreatePolicyRequest struct {
	ID              uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7a9e-b343-668d79691032" format:"uuid" validate:"optional"`
	Name            string    `json:"name" example:"List Policies for project" format:"string" validate:"required"`
	Description     string    `json:"description,omitempty" example:"This allows to list all the policies of a specific project" format:"string" validate:"optional"`
	AllowedAction   string    `json:"allowed_action,omitempty" example:"GET" format:"string" validate:"required"`
	AllowedResource string    `json:"allowed_resource,omitempty" example:"/projects/39a4707f-536e-433f-8597-6fc0d53a724f/tokens" format:"string" validate:"required"`
}

func (ref *CreatePolicyRequest) Validate() error {
	var errs ValidationErrors

	// Check if at least one field is provided
	if reflect.DeepEqual(ref, &CreatePolicyRequest{}) {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "request",
			Message: "at least one field must be updated",
			Code:    "REQUIRED",
		})
	}

	// Validate policy ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate name
	if _, err := ValidateString(ref.Name, StringValidationOptions{
		MinLength:        PolicyNameMinLength,
		MaxLength:        PolicyNameMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "name",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate description
	if _, err := ValidateString(ref.Description, StringValidationOptions{
		MinLength:        PolicyDescriptionMinLength,
		MaxLength:        PolicyDescriptionMaxLength,
		TrimWhitespace:   true,
		AllowEmpty:       false,
		NoControlChars:   true,
		NoHTMLTags:       true,
		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "description",
	}); err != nil {
		errs.Errors = append(errs.Errors, *err.(*ValidationError))
	}

	// Validate allowed action
	if _, err := ValidateAction(ref.AllowedAction); err != nil {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "allowed_action",
			Message: err.Error(),
			Code:    "INVALID_ACTION",
		})
	}

	// Validate allowed resource
	if _, err := ValidateResource(ref.AllowedResource); err != nil {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "allowed_resource",
			Message: err.Error(),
			Code:    "INVALID_RESOURCE",
		})
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}

// UpdatePolicyRequest represents a request to update a policy.
//
//	@Description	Update a policy.
type UpdatePolicyRequest struct {
	Name            *string `json:"name,omitempty" example:"Policy Name" format:"string" validate:"optional"`
	Description     *string `json:"description,omitempty" example:"This is a role" format:"string" validate:"optional"`
	AllowedAction   *string `json:"allowed_action,omitempty" example:"GET" format:"string" validate:"optional"`
	AllowedResource *string `json:"allowed_resource,omitempty" example:"/projects/39a4707f-536e-433f-8597-6fc0d53a724f/tokens" format:"string" validate:"optional"`
}

func (ref *UpdatePolicyRequest) Validate() error {
	var errs ValidationErrors

	// Check if at least one field is provided
	if reflect.DeepEqual(ref, &UpdatePolicyRequest{}) {
		errs.Errors = append(errs.Errors, ValidationError{
			Field:   "request",
			Message: "at least one field must be updated",
			Code:    "REQUIRED",
		})
	}

	// Validate name if provided
	if ref.Name != nil {
		if _, err := ValidateString(*ref.Name, StringValidationOptions{
			MinLength:        PolicyNameMinLength,
			MaxLength:        PolicyNameMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "name",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate description if provided
	if ref.Description != nil {
		if _, err := ValidateString(*ref.Description, StringValidationOptions{
			MinLength:        PolicyDescriptionMinLength,
			MaxLength:        PolicyDescriptionMaxLength,
			TrimWhitespace:   true,
			AllowEmpty:       false,
			NoControlChars:   true,
			NoHTMLTags:       true,
			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "description",
		}); err != nil {
			errs.Errors = append(errs.Errors, *err.(*ValidationError))
		}
	}

	// Validate allowed action if provided
	if ref.AllowedAction != nil {
		if _, err := ValidateAction(*ref.AllowedAction); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "allowed_action",
				Message: err.Error(),
				Code:    "INVALID_ACTION",
			})
		}
	}

	// Validate allowed resource if provided
	if ref.AllowedResource != nil {
		if _, err := ValidateResource(*ref.AllowedResource); err != nil {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   "allowed_resource",
				Message: err.Error(),
				Code:    "INVALID_RESOURCE",
			})
		}
	}

	if errs.HasErrors() {
		return &errs
	}
	return nil
}
