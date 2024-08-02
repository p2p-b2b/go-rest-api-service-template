package repository

import "errors"

var (
	ErrInvalidID                    = errors.New("invalid ID")
	ErrInvalidUserFirstName         = errors.New("invalid first name, the first name must be at least 2 characters long")
	ErrInvalidUserLastName          = errors.New("invalid last name, the last name must be at least 2 characters long")
	ErrInvalidUserEmail             = errors.New("invalid email")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)

var (
	ErrInvalidFilter    = errors.New("invalid filter field")
	ErrInvalidSort      = errors.New("invalid sort field")
	ErrInvalidFields    = errors.New("invalid fields field")
	ErrInvalidLimit     = errors.New("invalid limit field")
	ErrInvalidNextToken = errors.New("invalid nextToken field")
	ErrInvalidPrevToken = errors.New("invalid prevToken field")
)
