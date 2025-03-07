package model

import "time"

// HTTPMessage represents a message to be sent to the client though the HTTP REST API.
//
// @Description HTTPMessage represents a message to be sent to the client though the HTTP REST API.
type HTTPMessage struct {
	Timestamp  time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code"`
	Message    string    `json:"message"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
}

func (e *HTTPMessage) String() string {
	return e.Message
}

func (e *HTTPMessage) Error() string {
	return e.Message
}
