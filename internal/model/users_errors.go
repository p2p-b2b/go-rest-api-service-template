package model

import (
	"fmt"

	"github.com/google/uuid"
)

type InvalidUserIDError struct {
	ID      uuid.UUID
	Message string
}

func (e *InvalidUserIDError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("invalid user ID: %s", e.ID.String())
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid user ID: %s", e.Message)
	}

	return "invalid user ID"
}

type UserDisabledError struct {
	Username string
}

func (e *UserDisabledError) Error() string {
	return fmt.Sprintf("user '%s' is disabled", e.Username)
}

type InvalidFirstNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidFirstNameError) Error() string {
	return fmt.Sprintf("invalid first name: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidLastNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidLastNameError) Error() string {
	return fmt.Sprintf("invalid last name: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidEmailError struct {
	Email   string
	Message string
}

func (e *InvalidEmailError) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("invalid email '%s'", e.Email)
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid email: %s", e.Message)
	}
	return "invalid email"
}

type InvalidPasswordError struct {
	MinLength int
	MaxLength int
	Message   string
}

func (e *InvalidPasswordError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid password: %s", e.Message)
	}
	return fmt.Sprintf("invalid password: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type UserNotFoundError struct {
	ID    string
	Email string
}

func (e *UserNotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("user not found: %s", e.ID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user not found: %s", e.Email)
	}
	return "user not found"
}

type UserExistsError struct {
	ID    string
	Email string
}

func (e *UserExistsError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("user already exists: %s", e.ID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user already exists: %s", e.Email)
	}
	return "user already exists"
}

type EmailExistsError struct {
	Email string
}

func (e *EmailExistsError) Error() string {
	return fmt.Sprintf("email already exists: %s", e.Email)
}

type InvalidUserUpdateError struct {
	Message string
}

func (e *InvalidUserUpdateError) Error() string {
	return fmt.Sprintf("invalid user update: %s", e.Message)
}

type UserAlreadyExistsError struct {
	ID    string
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("user already exists: %s", e.ID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user already exists: %s", e.Email)
	}
	return "user already exists"
}

type UserEmailAlreadyExistsError struct {
	Email string
}

func (e *UserEmailAlreadyExistsError) Error() string {
	return fmt.Sprintf("email already exists: %s", e.Email)
}
