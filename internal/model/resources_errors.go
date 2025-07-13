package model

import (
	"fmt"

	"github.com/google/uuid"
)

type ResourceNotFoundError struct {
	ID      string
	Message string
}

func (e *ResourceNotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("resource not found: %s", e.ID)
	}
	if e.Message != "" {
		return e.Message
	}

	return "resource not found"
}

type InvalidResourceIDError struct {
	ID      uuid.UUID
	Message string
}

func (e *InvalidResourceIDError) Error() string {
	if e.ID != uuid.Nil && e.Message != "" {
		return fmt.Sprintf("invalid resource ID: %s, %s", e.ID.String(), e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid resource ID: %s", e.Message)
	}

	return "invalid resource ID"
}

type ResourceIDExistsError struct {
	ID      string
	Message string
}

func (e *ResourceIDExistsError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("resource ID already exists: %s", e.ID)
	}
	if e.Message != "" {
		return fmt.Sprintf("resource ID already exists: %s", e.Message)
	}

	return "resource ID already exists"
}

type InvalidActionError struct {
	Action string
}

func (e *InvalidActionError) Error() string {
	return fmt.Sprintf("invalid action: %s", e.Action)
}

type InvalidResourceError struct {
	Resource string
	Message  string
}

func (e *InvalidResourceError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("invalid resource: %s", e.Resource)
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid resource: %s", e.Message)
	}
	return "invalid resource"
}

type ResourceIDNotFoundError struct {
	ID string
}

func (e *ResourceIDNotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("resource ID not found: %s", e.ID)
	}
	return "resource ID not found"
}
