package model

import "fmt"

type InvalidPaginatorLimitError struct {
	MinLimit int
	MaxLimit int
}

func (e *InvalidPaginatorLimitError) Error() string {
	return fmt.Sprintf("invalid limit. must be greater than %d and less than or equal to %d", e.MinLimit, e.MaxLimit)
}

type InvalidPaginatorTokenError struct {
	Message string
}

func (e *InvalidPaginatorTokenError) Error() string {
	return fmt.Sprintf("invalid token: %s", e.Message)
}

type InvalidPaginatorCursorError struct {
	Message string
}

func (e *InvalidPaginatorCursorError) Error() string {
	return fmt.Sprintf("invalid cursor: %s", e.Message)
}
