package service

import "fmt"

var ErrInvalidOpenTelemetry = fmt.Errorf("invalid OpenTelemetry. It must not be nil")

// ------------------------------------------------------------
// Status is an enumeration of health statuses.
type Status bool

// Health statuses enumeration.
const (
	StatusUp   Status = true
	StatusDown Status = false
)

// String returns the string representation of the status.
func (s Status) String() string {
	if s == StatusUp {
		return "UP"
	}
	return "DOWN"
}

// MarshalJSON marshals the status to JSON.
func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// UnmarshalJSON unmarshals the status from JSON.
func (s *Status) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"UP"`:
		*s = StatusUp
	case `"DOWN"`:
		*s = StatusDown
	default:
		*s = StatusDown
	}
	return nil
}

// Check represents a health check.
//
// @Description Health check of the service
type Check struct {
	Name   string                 `json:"name" example:"database" format:"string"`
	Kind   string                 `json:"kind" example:"database" format:"string"`
	Status Status                 `json:"status" example:"true" format:"boolean"`
	Data   map[string]interface{} `json:"data,omitempty" format:"map"`
}

// Check represents a health check.
//
// @Description Health check of the service
type Health struct {
	Status Status  `json:"status" example:"true" format:"boolean"`
	Checks []Check `json:"checks" format:"array"`
}
