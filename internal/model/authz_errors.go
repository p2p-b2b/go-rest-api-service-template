package model

type InvalidAuthzServiceCacheError struct {
	Message string
}

func (e *InvalidAuthzServiceCacheError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "invalid AuthzServiceCache: it is required for ProjectsService"
}
