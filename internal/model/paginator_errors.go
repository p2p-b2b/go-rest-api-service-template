package model

import "fmt"

type InvalidLimitError struct {
	MinLimit int
	MaxLimit int
}

func (e *InvalidLimitError) Error() string {
	return fmt.Sprintf("invalid limit: must be between %d and %d", e.MinLimit, e.MaxLimit)
}

type InvalidTokenError struct {
	Message string
}

func (e *InvalidTokenError) Error() string {
	return fmt.Sprintf("invalid token: %s", e.Message)
}

type InvalidCursorError struct {
	Message string
}

func (e *InvalidCursorError) Error() string {
	return fmt.Sprintf("invalid cursor: %s", e.Message)
}
