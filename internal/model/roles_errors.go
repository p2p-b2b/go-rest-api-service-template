package model

import (
	"fmt"

	"github.com/google/uuid"
)

type InvalidRoleNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidRoleNameError) Error() string {
	return fmt.Sprintf("invalid name: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidRoleDescriptionError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidRoleDescriptionError) Error() string {
	return fmt.Sprintf("invalid description: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidRoleIDError struct {
	ID      uuid.UUID
	Message string
}

func (e *InvalidRoleIDError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("invalid role ID: %s", e.ID.String())
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid role ID: %s", e.Message)
	}

	return "invalid role ID"
}

type SystemRoleError struct {
	RoleID string
}

func (e *SystemRoleError) Error() string {
	return fmt.Sprintf("role '%s' is a system role and cannot be modified", e.RoleID)
}

type RoleNotFoundError struct {
	RoleID string
}

func (e *RoleNotFoundError) Error() string {
	return fmt.Sprintf("role not found: %s", e.RoleID)
}

type RoleNameExistsError struct {
	Name string
}

func (e *RoleNameExistsError) Error() string {
	return fmt.Sprintf("role name already exists: %s", e.Name)
}

type RoleIDExistsError struct {
	ID string
}

func (e *RoleIDExistsError) Error() string {
	return fmt.Sprintf("role ID already exists: %s", e.ID)
}

type InvalidRoleUpdateError struct {
	Message string
}

func (e *InvalidRoleUpdateError) Error() string {
	return fmt.Sprintf("invalid role update: %s", e.Message)
}

type InvalidRoleLinkError struct {
	Message string
}

func (e *InvalidRoleLinkError) Error() string {
	return fmt.Sprintf("invalid role link: %s", e.Message)
}

type RoleNameAlreadyExistsError struct {
	Name string
}

func (e *RoleNameAlreadyExistsError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("role name already exists: %s", e.Name)
	}
	return "role name already exists"
}

type RoleIDAlreadyExistsError struct {
	ID string
}

func (e *RoleIDAlreadyExistsError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("role ID already exists: %s", e.ID)
	}
	return "role ID already exists"
}
