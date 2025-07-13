package model

import "fmt"

type InvalidJWTError struct {
	Value   string
	Message string
}

func (e *InvalidJWTError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("invalid JWT '%s'", e.Value)
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid JWT: %s", e.Message)
	}
	return "invalid JWT"
}

type InvalidSenderError struct {
	Message string
}

func (e *InvalidSenderError) Error() string {
	return fmt.Sprintf("invalid sender: %s", e.Message)
}

type InvalidRefreshTokenError struct {
	Message string
}

func (e *InvalidRefreshTokenError) Error() string {
	return fmt.Sprintf("invalid refresh token: %s", e.Message)
}

type UserAlreadyVerifiedError struct {
	Email string
}

func (e *UserAlreadyVerifiedError) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("user '%s' is already verified", e.Email)
	}
	return "user is already verified"
}

type InvalidVerificationEndpointError struct {
	Endpoint string
	Message  string
}

func (e *InvalidVerificationEndpointError) Error() string {
	if e.Endpoint != "" {
		return fmt.Sprintf("invalid verification endpoint '%s'", e.Endpoint)
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid verification endpoint: %s", e.Message)
	}

	return "invalid verification endpoint"
}
