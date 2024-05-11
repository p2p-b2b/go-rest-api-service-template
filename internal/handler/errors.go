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

type RESTApiError struct {
	StatusCode int `json:"status_code"`
	Message    any `json:"message"`
}

func NewRESTApiError(statusCode int, message error) RESTApiError {
	return RESTApiError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *RESTApiError) Error() string {
	return fmt.Sprintf("api error: %d", e.StatusCode)
}

func (e *RESTApiError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"status_code": e.StatusCode,
		"message":     e.Message,
	})
}

func (e *RESTApiError) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	e.StatusCode = v["status_code"].(int)
	e.Message = v["message"]

	return nil
}

func (e *RESTApiError) InvalidRequestData(errors map[string]string) RESTApiError {
	return RESTApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    errors,
	}
}

func (e *RESTApiError) InternalServerError() RESTApiError {
	return RESTApiError{
		StatusCode: http.StatusInternalServerError,
		Message:    ErrInternalServer.Error(),
	}
}

func (e *RESTApiError) NotFound() RESTApiError {
	return RESTApiError{
		StatusCode: http.StatusNotFound,
		Message:    "not found",
	}
}

func InvalidJSON() RESTApiError {
	return NewRESTApiError(http.StatusBadRequest, errors.New("invalid JSON"))
}

type RESTApiFunc func(w http.ResponseWriter, r *http.Request) error

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
