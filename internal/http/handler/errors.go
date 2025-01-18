package handler

import (
	"errors"
)

var (
	ErrInternalServerError          = errors.New("internal server error")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated, any of these could be empty")
	ErrRequiredUUID                 = errors.New("required UUID")
	ErrInvalidUUID                  = errors.New("invalid UUID")
	ErrUUIDCannotBeNil              = errors.New("UUID cannot be nil")
	ErrInvalidFilter                = errors.New("invalid filter field")
	ErrInvalidSort                  = errors.New("invalid sort field")
	ErrInvalidFields                = errors.New("invalid fields field")
	ErrInvalidLimit                 = errors.New("invalid limit field")
	ErrInvalidNextToken             = errors.New("invalid nextToken field")
	ErrInvalidPrevToken             = errors.New("invalid prevToken field")
)
