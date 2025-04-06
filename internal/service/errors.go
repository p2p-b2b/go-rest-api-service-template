package service

import "errors"

var (
	ErrRepositoryRequired    = errors.New("repository required")
	ErrOpenTelemetryRequired = errors.New("OpenTelemetry required")
)
