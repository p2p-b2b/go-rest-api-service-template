package handler

import (
	"fmt"
)

var ErrInternalServerError = fmt.Errorf("internal server error")

type InvalidUUIDError struct {
	UUID    string
	Message string
}

func (e *InvalidUUIDError) Error() string {
	if e.UUID != "" && e.Message != "" {
		return fmt.Sprintf("invalid UUID '%s': '%s'", e.UUID, e.Message)
	}

	if e.UUID != "" && e.Message == "" {
		return fmt.Sprintf("invalid UUID: '%s'", e.UUID)
	}

	if e.UUID == "" && e.Message != "" {
		return fmt.Sprintf("invalid UUID: '%s'", e.Message)
	}

	return "invalid UUID"
}

type InvalidServiceError struct {
	Name   string
	Reason string
}

func (e *InvalidServiceError) Error() string {
	if e.Name != "" && e.Reason != "" {
		return fmt.Sprintf("invalid service %s: %s", e.Name, e.Reason)
	}

	if e.Name != "" && e.Reason == "" {
		return fmt.Sprintf("invalid service: %s", e.Name)
	}

	if e.Name == "" && e.Reason != "" {
		return fmt.Sprintf("invalid service: %s", e.Reason)
	}

	return "invalid service"
}

type InvalidOpenTelemetryError struct {
	Name   string
	Reason string
}

func (e *InvalidOpenTelemetryError) Error() string {
	if e.Name != "" && e.Reason != "" {
		return fmt.Sprintf("invalid open telemetry %s: %s", e.Name, e.Reason)
	}

	if e.Name != "" && e.Reason == "" {
		return fmt.Sprintf("invalid open telemetry: %s", e.Name)
	}

	if e.Name == "" && e.Reason != "" {
		return fmt.Sprintf("invalid open telemetry: %s", e.Reason)
	}

	return "invalid open telemetry"
}
