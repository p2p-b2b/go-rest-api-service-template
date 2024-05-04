package model

// Status is an enumeration of health statuses.
type Status bool

const (
	StatusUp   Status = true
	StatusDown Status = false
)

func (s Status) String() string {
	if s == StatusUp {
		return "UP"
	}
	return "DOWN"
}

// Check represents a health check.
type Check struct {
	Name   string                 `json:"name"`
	Status Status                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// Check represents a health check.
type Health struct {
	Status Status  `json:"status"`
	Checks []Check `json:"checks"`
}
