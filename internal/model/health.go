package model

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
type Check struct {
	// Name is the name of the check.
	Name string `json:"name"`

	// Kind is the kind of check.
	Kind string `json:"kind,omitempty"`

	// Status is the status of the check.
	Status Status `json:"status"`

	// Data is an optional field that can be used to provide additional information about the check.
	Data map[string]interface{} `json:"data,omitempty"`
}

// Check represents a health check.
type Health struct {
	// Status is the status of the health check.
	Status Status `json:"status"`

	// Checks is a list of health checks.
	Checks []Check `json:"checks"`
}
