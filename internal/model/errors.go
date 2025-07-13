package model

type InvalidInputError struct {
	Message string
}

func (e *InvalidInputError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid input"
}

type InvalidDBConfigurationError struct {
	Message string
}

func (e *InvalidDBConfigurationError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid database configuration"
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

type InvalidRepositoryError struct {
	Message string
}

func (e *InvalidRepositoryError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid repository"
}

type InvalidRegoQueryError struct {
	Message string
}

func (e *InvalidRegoQueryError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid rego query"
}

type InvalidRegoPolicyError struct {
	Message string
}

func (e *InvalidRegoPolicyError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid rego policy"
}

type InvalidCacheServiceError struct {
	Message string
}

func (e *InvalidCacheServiceError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid cache service"
}

type InvalidMailQueueServiceError struct {
	Message string
}

func (e *InvalidMailQueueServiceError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid mail queue service"
}

type InvalidPrivateKeyError struct {
	Message string
}

func (e *InvalidPrivateKeyError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid private key"
}

type InvalidPublicKeyError struct {
	Message string
}

func (e *InvalidPublicKeyError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid public key"
}

type InvalidIssuerError struct {
	Message string
}

func (e *InvalidIssuerError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid issuer"
}

type InvalidAccessTokenDurationError struct {
	Message string
}

func (e *InvalidAccessTokenDurationError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid access token duration"
}

type InvalidRefreshTokenDurationError struct {
	Message string
}

func (e *InvalidRefreshTokenDurationError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid refresh token duration"
}

type InvalidSymmetricKeyError struct {
	Message string
}

func (e *InvalidSymmetricKeyError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid symmetric key"
}

type InvalidServiceError struct {
	Message string
}

func (e *InvalidServiceError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid service"
}

type InvalidRequestError struct {
	Message string
}

func (e *InvalidRequestError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid request"
}

type InternalServerError struct {
	Message string
}

func (e *InternalServerError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "internal server error"
}

type InvalidUUIDError struct {
	UUID    string
	Message string
}

func (e *InvalidUUIDError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.UUID == "" {
		return "invalid UUID: empty string"
	}

	return "invalid UUID: " + e.UUID
}
