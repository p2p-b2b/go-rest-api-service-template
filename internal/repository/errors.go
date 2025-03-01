package repository

import "errors"

var (
	ErrDBInvalidConfiguration   = errors.New("invalid database configuration. It is nil")
	ErrDBInvalidMaxPingTimeout  = errors.New("invalid max ping timeout. It must be greater than 10 millisecond")
	ErrDBInvalidMaxQueryTimeout = errors.New("invalid max query timeout. It must be greater than 10 millisecond")
	ErrOTInvalidConfiguration   = errors.New("invalid OpenTelemetry configuration. It is nil")
)
