package model

import "time"

// HTTPMessage represents a message to be sent to the client though the HTTP REST API.
//
// @Description HTTPMessage represents a message to be sent to the client though the HTTP REST API.
type HTTPMessage struct {
	Timestamp  time.Time `json:"timestamp" example:"2021-01-01T00:00:00Z" format:"date-time"`
	StatusCode int       `json:"status_code" example:"200" format:"int32"`
	Message    string    `json:"message" example:"Hello, World!" format:"string"`
	Method     string    `json:"method" example:"GET" format:"string"`
	Path       string    `json:"path" example:"/api/v1/hello" format:"string"`
}

func (e *HTTPMessage) String() string {
	return e.Message
}

func (e *HTTPMessage) Error() string {
	return e.Message
}
