package service

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
type Check struct {
	Name   string                 `json:"name"`
	Kind   string                 `json:"kind"`
	Status Status                 `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// Check represents a health check.
type Health struct {
	Status Status  `json:"status"`
	Checks []Check `json:"checks"`
}
