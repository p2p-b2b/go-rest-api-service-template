package model

import "fmt"

type InvalidUserIDError struct {
	ID      string
	Message string
}

func (e *InvalidUserIDError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid user ID: %s, %s", e.ID, e.Message)
	}

	return fmt.Sprintf("invalid user ID: %s", e.ID)
}

type InvalidUserFirstNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidUserFirstNameError) Error() string {
	return fmt.Sprintf("invalid first name. Must be between %d and %d characters long", e.MinLength, e.MaxLength)
}

type InvalidUserLastNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidUserLastNameError) Error() string {
	return fmt.Sprintf("invalid last name. Must be between %d and %d characters long", e.MinLength, e.MaxLength)
}

type InvalidUserEmailError struct {
	MinLength int
	MaxLength int
	Email     string
}

func (e *InvalidUserEmailError) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("invalid email address: %s. Must be between %d and %d characters long", e.Email, e.MinLength, e.MaxLength)
	}

	return fmt.Sprintf("invalid email address. Must be between %d and %d characters long", e.MinLength, e.MaxLength)
}

type InvalidUserPasswordError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidUserPasswordError) Error() string {
	return fmt.Sprintf("invalid password. Must be between %d and %d characters long", e.MinLength, e.MaxLength)
}

type UserNotFoundError struct {
	ID    string
	Email string
}

func (e *UserNotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("user with ID %s not found", e.ID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user with email %s not found", e.Email)
	}

	return "user not found"
}

type UserAlreadyExistsError struct {
	ID    string
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("user with ID %s already exists", e.ID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user with email %s already exists", e.Email)
	}

	return "user already exists"
}

type UserEmailAlreadyExistsError struct {
	Email string
}

func (e *UserEmailAlreadyExistsError) Error() string {
	return fmt.Sprintf("user with email %s already exists", e.Email)
}
