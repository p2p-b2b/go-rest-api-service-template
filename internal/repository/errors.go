package repository

import "errors"

var (
	ErrDBInvalidConfiguration       = errors.New("invalid database configuration. It is nil")
	ErrDBInvalidMaxPingTimeout      = errors.New("invalid max ping timeout. It must be greater than 10 millisecond")
	ErrDBInvalidMaxQueryTimeout     = errors.New("invalid max query timeout. It must be greater than 10 millisecond")
	ErrOTInvalidConfiguration       = errors.New("invalid OpenTelemetry configuration. It is nil")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")

	ErrInputIsNil       = errors.New("input is nil")
	ErrInvalidFilter    = errors.New("invalid filter field")
	ErrInvalidSort      = errors.New("invalid sort field")
	ErrInvalidFields    = errors.New("invalid fields field")
	ErrInvalidLimit     = errors.New("invalid limit field")
	ErrInvalidNextToken = errors.New("invalid nextToken field")
	ErrInvalidPrevToken = errors.New("invalid prevToken field")
)
