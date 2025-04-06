package model

import "errors"

var (
	ErrInputIsNil                   = errors.New("input is nil")
	ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")
)
