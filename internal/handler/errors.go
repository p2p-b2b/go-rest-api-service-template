package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

var (
	// ErrInvalidID is returned when an invalid ID is provided.
	ErrInvalidID = errors.New("invalid ID")

	// ErrIDRequired is returned when an ID is required.
	ErrIDRequired = errors.New("id is required for this operation")

	// ErrInternalServerError is returned when an internal server error occurs.
	ErrInternalServerError = errors.New("internal server error")
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
