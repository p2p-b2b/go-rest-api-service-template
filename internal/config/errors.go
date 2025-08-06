package config

import "fmt"

// InvalidConfigurationError represents an error for invalid configuration
// It is used to provide detailed information about the configuration error
// including the field, value, and a message describing the error.
type InvalidConfigurationError struct {
	Field   string
	Value   any
	Message string
}

func (e *InvalidConfigurationError) Error() string {
	if e.Value != nil && e.Value != "" {
		return fmt.Sprintf("%s (field: '%s', value: '%v')", e.Message, e.Field, e.Value)
	}

	if e.Field == "" {
		return e.Message
	}

	return fmt.Sprintf("%s (field: %s)", e.Message, e.Field)
}
