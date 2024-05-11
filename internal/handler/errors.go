package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

var (
	// ErrInvalidID is returned when an invalid ID is provided.
	ErrInvalidID = errors.New("invalid ID")

	// ErrIDRequired is returned when an ID is required.
	ErrIDRequired = errors.New("id is required for this operation")

	// ErrInternalServer is returned when an internal server error occurs.
	ErrInternalServer = errors.New("internal server error")
)

// RESTApiError represents an API error.
type RESTApiError struct {
	StatusCode int `json:"status_code"`
	Message    any `json:"message"`
}

// NewRESTApiError returns a new instance of RESTApiError.
func NewRESTApiError(statusCode int, message error) RESTApiError {
	return RESTApiError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// Error returns the error message.
func (e *RESTApiError) Error() string {
	return fmt.Sprintf("api error: %d", e.StatusCode)
}

// MarshalJSON returns the JSON encoding of the RESTApiError.
func (e *RESTApiError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"status_code": e.StatusCode,
		"message":     e.Message,
	})
}

// UnmarshalJSON parses the JSON-encoded data and stores the result in the RESTApiError.
func (e *RESTApiError) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	e.StatusCode = v["status_code"].(int)
	e.Message = v["message"]

	return nil
}

// InvalidRequestData returns an invalid request data error.
func (e *RESTApiError) InvalidRequestData(errors map[string]string) RESTApiError {
	return RESTApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    errors,
	}
}

// InternalServerError returns an internal server error.
func (e *RESTApiError) InternalServerError() RESTApiError {
	return RESTApiError{
		StatusCode: http.StatusInternalServerError,
		Message:    ErrInternalServer.Error(),
	}
}

// NotFound returns a not found error.
func (e *RESTApiError) NotFound() RESTApiError {
	return RESTApiError{
		StatusCode: http.StatusNotFound,
		Message:    "not found",
	}
}

// RESTApiError returns a bad request error.
func InvalidJSON() RESTApiError {
	return NewRESTApiError(http.StatusBadRequest, errors.New("invalid JSON"))
}

// RESTApiFunc represents a REST API function.
type RESTApiFunc func(w http.ResponseWriter, r *http.Request) error

// Make returns a new HTTP handler function.
func Make(h RESTApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			if apiErr, ok := err.(*RESTApiError); ok {
				w.WriteHeader(apiErr.StatusCode)
				json.NewEncoder(w).Encode(apiErr)
				return
			} else {
				errResp := map[string]any{
					"status_code": http.StatusInternalServerError,
					"message":     err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(errResp)
			}

			slog.Error("HTTP REST API error", "error", err.Error(), "path", r.URL.Path)
		}
	}
}
