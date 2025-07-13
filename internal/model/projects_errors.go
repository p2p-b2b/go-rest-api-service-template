package model

import (
	"fmt"

	"github.com/google/uuid"
)

type ProjectNotFoundError struct {
	ID      uuid.UUID
	Message string
}

func (e *ProjectNotFoundError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("project not found: %s", e.ID.String())
	}
	if e.Message != "" {
		return fmt.Sprintf("project not found: %s", e.Message)
	}

	return "project not found"
}

type SystemProjectError struct {
	ID      uuid.UUID
	Name    string
	Message string
}

func (e *SystemProjectError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("project '%s' is a system project", e.ID.String())
	}
	if e.Name != "" {
		return fmt.Sprintf("project '%s' is a system project", e.Name)
	}
	if e.Message != "" {
		return fmt.Sprintf("project is a system project: %s", e.Message)
	}

	return "project is a system project"
}

type InvalidProjectDescriptionError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidProjectDescriptionError) Error() string {
	return fmt.Sprintf("invalid description: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidProjectNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidProjectNameError) Error() string {
	return fmt.Sprintf("invalid name: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidProjectIDError struct {
	ID      uuid.UUID
	Message string
}

func (e *InvalidProjectIDError) Error() string {
	if e.ID != uuid.Nil && e.Message != "" {
		return fmt.Sprintf("invalid project ID: %s, %s", e.ID.String(), e.Message)
	}
	if e.ID != uuid.Nil && e.Message == "" {
		return fmt.Sprintf("invalid project ID: %s", e.ID.String())
	}
	if e.ID == uuid.Nil && e.Message != "" {
		return fmt.Sprintf("invalid project ID: %s", e.Message)
	}

	return "invalid project ID"
}

type ProjectNameExistsError struct {
	Name string
}

func (e *ProjectNameExistsError) Error() string {
	return fmt.Sprintf("project name '%s' already exists", e.Name)
}

type ProjectIDExistsError struct {
	ID string
}

func (e *ProjectIDExistsError) Error() string {
	return fmt.Sprintf("project ID '%s' already exists", e.ID)
}

type InvalidProjectUpdateError struct {
	Message string
}

func (e *InvalidProjectUpdateError) Error() string {
	return fmt.Sprintf("invalid project update: %s", e.Message)
}

type ProjectIDNotFoundError struct {
	ID uuid.UUID
}

func (e *ProjectIDNotFoundError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("project not found: %s", e.ID.String())
	}
	return "project not found"
}

type ProjectIDAlreadyExistsError struct {
	ID uuid.UUID
}

func (e *ProjectIDAlreadyExistsError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("project ID already exists: %s", e.ID.String())
	}
	return "project ID already exists"
}

type ProjectNameAlreadyExistsError struct {
	Name string
}

func (e *ProjectNameAlreadyExistsError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("project name already exists: %s", e.Name)
	}
	return "project name already exists"
}
