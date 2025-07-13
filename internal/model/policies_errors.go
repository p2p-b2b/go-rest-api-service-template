package model

import (
	"fmt"

	"github.com/google/uuid"
)

type PolicyNotFoundError struct {
	ID      uuid.UUID
	Message string
}

func (e *PolicyNotFoundError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("policy not found: %s", e.ID.String())
	}
	if e.Message != "" {
		return fmt.Sprintf("policy not found: %s", e.Message)
	}

	return "policy not found"
}

// InvalidPolicyIDError represents an error when the policy ID is invalid.
type InvalidPolicyIDError struct {
	Message string
}

func (e *InvalidPolicyIDError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid policy ID: %s", e.Message)
	}

	return "invalid policy ID"
}

type PolicyIDAlreadyExistsError struct {
	ID      uuid.UUID
	Message string
}

func (e *PolicyIDAlreadyExistsError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("policy ID already exists: %s", e.ID.String())
	}

	if e.Message != "" {
		return fmt.Sprintf("policy ID already exists: %s", e.Message)
	}

	return "policy ID already exists"
}

type InvalidPolicyNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidPolicyNameError) Error() string {
	return fmt.Sprintf("invalid name. Must be between %d and %d characters long", e.MinLength, e.MaxLength)
}

type InvalidPolicyDescriptionError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidPolicyDescriptionError) Error() string {
	return fmt.Sprintf("invalid description. Must be between %d and %d characters long", e.MinLength, e.MaxLength)
}

type InvalidPolicyAllowedActionError struct {
	Message string
}

func (e *InvalidPolicyAllowedActionError) Error() string {
	return fmt.Sprintf("invalid action: %s", e.Message)
}

type InvalidPolicyAllowedResourceError struct {
	Message string
}

func (e *InvalidPolicyAllowedResourceError) Error() string {
	return fmt.Sprintf("invalid resource: %s", e.Message)
}

type PolicyNameAlreadyExistsError struct {
	Name string
}

func (e *PolicyNameAlreadyExistsError) Error() string {
	return fmt.Sprintf("policy name already exists: %s", e.Name)
}

type PolicyIDNotFoundError struct {
	ID uuid.UUID
}

func (e *PolicyIDNotFoundError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("policy ID %s not found", e.ID.String())
	}

	return "policy ID not found"
}

type SystemPolicyError struct {
	PolicyID string
}

func (e *SystemPolicyError) Error() string {
	return fmt.Sprintf("invalid policy ID: %s. System policies cannot be modified", e.PolicyID)
}

type InvalidPolicyLinkRolesError struct {
	Message string
}

func (e *InvalidPolicyLinkRolesError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid policy link roles: %s", e.Message)
	}

	return "invalid policy link roles"
}

type InvalidPolicyCreateError struct {
	Message string
}

func (e *InvalidPolicyCreateError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid policy create: %s", e.Message)
	}

	return "invalid policy create"
}

type InvalidPolicyUpdateError struct {
	Message string
}

func (e *InvalidPolicyUpdateError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid policy update: %s", e.Message)
	}

	return "invalid policy update"
}
