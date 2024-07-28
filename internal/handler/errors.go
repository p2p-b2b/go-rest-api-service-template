package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

var (
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidUserID        = errors.New("invalid user ID, this must be a valid UUID")
	ErrInvalidUserFirstName = errors.New("invalid user first name, the length must be at least 2 characters")
	ErrInvalidUserLastName  = errors.New("invalid user last name, the length must be at least 2 characters")
	ErrInvalidUserEmail     = errors.New("invalid user email, the length must be at least 6 characters and must be a valid email address")

	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated, any of these could be empty")

	ErrEncodingPayload = errors.New("error encoding payload")
	ErrDecodingPayload = errors.New("error decoding payload")

	ErrRequiredUUID    = errors.New("required UUID")
	ErrInvalidUUID     = errors.New("invalid UUID")
	ErrUUIDCannotBeNil = errors.New("UUID cannot be nil")

	ErrInvalidFilter    = errors.New("invalid filter field")
	ErrInvalidSort      = errors.New("invalid sort field")
	ErrInvalidFields    = errors.New("invalid fields field")
	ErrInvalidLimit     = errors.New("invalid limit field")
	ErrInvalidNextToken = errors.New("invalid nextToken field")
	ErrInvalidPrevToken = errors.New("invalid prevToken field")
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
