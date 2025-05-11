package model

import "errors"

var ErrAtLeastOneFieldMustBeUpdated = errors.New("at least one field must be updated")

type InputIsInvalidError struct {
	Message string
}

func (e *InputIsInvalidError) Error() string {
	return e.Message
}

type InvalidDBConfigurationError struct {
	Message string
}

func (e *InvalidDBConfigurationError) Error() string {
	return e.Message
}

type InvalidDBMaxPingTimeoutError struct {
	Message string
}

func (e *InvalidDBMaxPingTimeoutError) Error() string {
	return e.Message
}

type InvalidDBMaxQueryTimeoutError struct {
	Message string
}

func (e *InvalidDBMaxQueryTimeoutError) Error() string {
	return e.Message
}

type InvalidRepositoryConfigurationError struct {
	Message string
}

func (e *InvalidRepositoryConfigurationError) Error() string {
	return e.Message
}

type InvalidOTConfigurationError struct {
	Message string
}

func (e *InvalidOTConfigurationError) Error() string {
	return e.Message
}

type InvalidByteSequenceError struct {
	Message string
}

func (e *InvalidByteSequenceError) Error() string {
	return e.Message
}

type InvalidMessageFormatError struct {
	Message string
}

func (e *InvalidMessageFormatError) Error() string {
	return e.Message
}
