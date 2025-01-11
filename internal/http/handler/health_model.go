package handler

// Check represents a health check.
//
// @Description Health check of the service
type Check struct {
	Name   string                 `json:"name" example:"database" format:"string"`
	Kind   string                 `json:"kind" example:"database" format:"string"`
	Status string                 `json:"status" example:"true" format:"boolean"`
	Data   map[string]interface{} `json:"data,omitempty" format:"map"`
}

// Check represents a health check.
//
// @Description Health check of the service
type Health struct {
	Status string  `json:"status" example:"true" format:"boolean"`
	Checks []Check `json:"checks" format:"array"`
}
