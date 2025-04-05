package handler

import (
	"errors"
)

var (
	ErrModelServiceRequired  = errors.New("model service required")
	ErrOpenTelemetryRequired = errors.New("open telemetry required")

	ErrInternalServerError          = errors.New("internal server error")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated, any of these could be empty")
	ErrRequiredUUID                 = errors.New("required UUID")
	ErrInvalidUUID                  = errors.New("invalid UUID")
	ErrUUIDCannotBeNil              = errors.New("UUID cannot be nil")

	ErrInvalidLimit     = errors.New("invalid limit field")
	ErrInvalidNextToken = errors.New("invalid nextToken field")
	ErrInvalidPrevToken = errors.New("invalid prevToken field")
)
