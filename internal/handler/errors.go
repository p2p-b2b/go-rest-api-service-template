package handler

import (
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

	// ErrInvalidRequestBody is returned when an invalid request body is provided.
	ErrInvalidRequestBody = errors.New("invalid request body")
)

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	http.Error(w, message, statusCode)
	slog.Error(message,
		"status_code", statusCode,
		"method", r.Method,
		"url", r.URL.Path,
		"query", r.URL.RawQuery,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)
}
