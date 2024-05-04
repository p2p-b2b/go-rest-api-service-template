package handler

import "errors"

var (
	// ErrInvalidID is returned when an invalid ID is provided.
	ErrInvalidID = errors.New("invalid ID")

	// ErrIDRequired is returned when an ID is required.
	ErrIDRequired = errors.New("id is required for this operation")

	// ErrInternalServer is returned when an internal server error occurs.
	ErrInternalServer = errors.New("internal server error")
)
