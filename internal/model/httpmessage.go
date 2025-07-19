package model

import "time"

// HTTPMessage represents a message to be sent to the client trough HTTP REST API.
//
//	@Description	HTTPMessage represents a message to be sent to the client trough HTTP REST API.
type HTTPMessage struct {
	Timestamp  time.Time `json:"timestamp" example:"2021-07-01T00:00:00Z" format:"date-time"`
	Message    string    `json:"message" example:"success" format:"string"`
	Method     string    `json:"method" example:"GET" format:"string"`
	Path       string    `json:"path" example:"/api/v1/users" format:"string"`
	StatusCode int       `json:"status_code" example:"200" format:"int32"`
}

// String returns the message as a string.
func (e *HTTPMessage) String() string {
	return e.Message
}

// Error returns the message as an error.
func (e *HTTPMessage) Error() string {
	return e.Message
}
