package service

import "errors"

var (
	ErrRepositoryRequired           = errors.New("repository required")
	ErrOpenTelemetryRequired        = errors.New("OpenTelemetry required")
	ErrInputIsNil                   = errors.New("input is nil")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)
