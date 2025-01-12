package service

import "errors"

var (
	ErrInvalidRepository            = errors.New("invalid repository")
	ErrInvalidOpenTelemetry         = errors.New("invalid open telemetry")
	ErrGettingUserByID              = errors.New("error getting user by ID")
	ErrGettingUserByEmail           = errors.New("error getting user by email")
	ErrCreatingUser                 = errors.New("error creating user")
	ErrUpdatingUser                 = errors.New("error updating user")
	ErrDeletingUser                 = errors.New("error deleting user")
	ErrListingUsers                 = errors.New("error listing users")
	ErrInvalidID                    = errors.New("invalid ID")
	ErrInvalidFirstName             = errors.New("invalid first name, the first name must be at least 2 characters long")
	ErrInvalidLastName              = errors.New("invalid last name, the last name must be at least 2 characters long")
	ErrInvalidEmail                 = errors.New("invalid email")
	ErrInvalidFilter                = errors.New("invalid filter field")
	ErrInvalidSort                  = errors.New("invalid sort field")
	ErrInvalidFields                = errors.New("invalid fields field")
	ErrInvalidLimit                 = errors.New("invalid limit field")
	ErrInvalidNextToken             = errors.New("invalid nextToken field")
	ErrInvalidPrevToken             = errors.New("invalid prevToken field")
	ErrIdAlreadyExists              = errors.New("id already exists")
	ErrInputIsNil                   = errors.New("input is nil")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)
