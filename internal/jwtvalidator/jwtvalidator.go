// Package jwtvalidator provides an interface for validating JWT tokens.
package jwtvalidator

import (
	"context"
)

// Validator is an interface for validating JWT tokens.
type Validator interface {
	// Validate validates a JWT token and returns the claims if the token is valid.
	Validate(ctx context.Context, token string) (claims map[string]any, err error)

	// GetClientID returns the client ID of the Validator.
	GetClientID() string
}
