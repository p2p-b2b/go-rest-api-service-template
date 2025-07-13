package model

// Status represents the health status of a service.
//
// @Description Health status of a service.
type Status bool

// String returns the string representation of the status.
func (val Status) String() string {
	if val == StatusUp {
		return "UP"
	}
	return "DOWN"
}

// Health statuses enumeration.
//
// @Description Health statuses enumeration.
const (
	StatusUp   Status = true
	StatusDown Status = false
)

// Check represents a health check.
//
// @Description Health check of the service.
type Check struct {
	Data   map[string]any `json:"data"`
	Name   string         `json:"name" example:"database" format:"string"`
	Kind   string         `json:"kind" example:"database" format:"string"`
	Status Status         `json:"status" example:"True" format:"string"`
}

// Health represents a health check.
//
// @Description Health check of the service.
type Health struct {
	Checks []Check `json:"checks"`
	Status Status  `json:"status" example:"True" format:"string"`
}
