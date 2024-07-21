package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

var (
	// ErrInvalidUserID is returned when the user ID is not a valid UUID.
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrUserIDRequired is returned when an ID is required.
	ErrUserIDRequired = errors.New("user ID is required")

	// ErrInternalServerError is returned when an internal server error occurs.
	ErrInternalServerError = errors.New("internal server error")

	// ErrEncodingPayload is returned when an error occurs while encoding the payload.
	ErrEncodingPayload = errors.New("error encoding payload")

	// ErrDecodingPayload is returned when an error occurs while decoding the payload.
	ErrDecodingPayload = errors.New("error decoding payload")

	// ErrFirstNameRequired is returned when the first name is required.
	ErrFirstNameRequired = errors.New("first name is required")

	// ErrLastNameRequired is returned when the last name is required.
	ErrLastNameRequired = errors.New("last name is required")

	// ErrEmailRequired is returned when the email is required.
	ErrEmailRequired = errors.New("email is required")

	// ErrAtLeastOneFieldRequired is returned when at least one field is required.
	ErrAtLeastOneFieldRequired = errors.New("at least one field is required")

	// ErrInvalidFilter is returned when the filter is invalid.
	ErrInvalidFilter = errors.New("invalid filter")

	// ErrInvalidSort is returned when the sort is invalid.
	ErrInvalidSort = errors.New("invalid sort")

	// ErrInvalidField is returned when the field is invalid.
	ErrInvalidField = errors.New("invalid field")
)

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// WriteError writes an error log and response to the client with the given status code and message.
// The message is also logged with the request details.
// The response is in JSON format.
// The status code should be one of the http.Status* constants.
// The message should be a human-readable string.
// The request details are logged with the error message.
// The request details include the method, URL, query, user agent, and remote address.
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var err APIError
	err.StatusCode = statusCode
	err.Message = message

	if err := json.NewEncoder(w).Encode(err); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}

	slog.Error(message,
		"status_code", statusCode,
		"method", r.Method,
		"url", r.URL.Path,
		"query", r.URL.RawQuery,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)
}
