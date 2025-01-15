package service

import "errors"

var (
	ErrInvalidRepository            = errors.New("invalid repository")
	ErrInvalidOpenTelemetry         = errors.New("invalid open telemetry")
	ErrInputIsNil                   = errors.New("input is nil")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)
