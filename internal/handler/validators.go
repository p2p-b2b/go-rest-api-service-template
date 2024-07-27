package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/query"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
)

var (
	// ErrRequiredUUID is an error that is returned when a UUID is required.
	ErrRequiredUUID = errors.New("required UUID")

	// ErrInvalidID is an error that is returned when the ID is not a valid UUID.
	ErrInvalidUUID = errors.New("invalid UUID")

	// ErrUUIDCannotBeNil is an error that is returned when the UUID is nil.
	ErrUUIDCannotBeNil = errors.New("UUID cannot be nil")
)

// parseUUIDQueryParams parses a string into a UUID.
// If the input is empty, it returns an error.
// If the input is not a valid UUID, it returns an error.
// If the input is a nil UUID, it returns an error.
func parseUUIDQueryParams(input string) (uuid.UUID, error) {
	if input == "" {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrRequiredUUID, input)
	}

	id, err := uuid.Parse(input)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrInvalidUUID, input)
	}

	if id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrUUIDCannotBeNil, input)
	}

	return id, nil
}

var (
	// ErrInvalidFilter is returned when the filter is invalid.
	ErrInvalidFilter = errors.New("invalid filter field")

	// ErrInvalidSort is returned when the sort is invalid.
	ErrInvalidSort = errors.New("invalid sort field")

	// ErrInvalidFields is returned when the field is invalid.
	ErrInvalidFields = errors.New("invalid fields field")

	// ErrInvalidLimit is returned when the limit is invalid.
	ErrInvalidLimit = errors.New("invalid limit field")

	// ErrInvalidNextToken is returned when the nextToken is invalid.
	ErrInvalidNextToken = errors.New("invalid nextToken field")

	// ErrInvalidPrevToken is returned when the prevToken is invalid.
	ErrInvalidPrevToken = errors.New("invalid prevToken field")
)

// parseSortQueryParams parses a string into a sort field.
func parseSortQueryParams(sort string) (string, error) {
	if !query.IsValidSort(repository.UserSortFields, sort) {
		return "", ErrInvalidSort
	}

	return sort, nil
}

// parseFilterQueryParams parses a string into a filter field.
func parseFilterQueryParams(filter string) (string, error) {
	if !query.IsValidFilter(repository.UserFilterFields, filter) {
		return "", ErrInvalidFilter
	}

	return filter, nil
}

// parseFieldsQueryParams parses a string into a list of fields.
func parseFieldsQueryParams(fields string) ([]string, error) {
	if !query.IsValidFields(repository.UserFields, fields) {
		return nil, ErrInvalidFields
	}

	return query.GetFields(fields), nil
}

// parseNextTokenQueryParams parses a string into a nextToken field.
func parseNextTokenQueryParams(nextToken string) (string, error) {
	_, _, err := paginator.DecodeToken(nextToken)
	if err != nil {
		return "", ErrInvalidNextToken
	}

	return nextToken, nil
}

// parsePrevTokenQueryParams parses a string into a prevToken field.
func parsePrevTokenQueryParams(prevToken string) (string, error) {
	_, _, err := paginator.DecodeToken(prevToken)
	if err != nil {
		return "", ErrInvalidNextToken
	}

	return prevToken, nil
}

// parseLimitQueryParams parses a string into a limit field.
func parseLimitQueryParams(limit string) (int, error) {
	var limitInt int
	var err error

	if limit == "" {
		return paginator.DefaultLimit, nil
	}

	// check if this is a valid integer
	if limitInt, err = strconv.Atoi(limit); err != nil {
		return 0, ErrInvalidLimit
	}

	if limitInt == 0 {
		limitInt = paginator.DefaultLimit
	}

	if limitInt < paginator.MinLimit {
		return 0, ErrInvalidLimit
	} else if limitInt > paginator.MaxLimit {
		limitInt = paginator.MaxLimit
	}

	return limitInt, nil
}

// parseListQueryParams parses a list of strings into a list of UUIDs.
func parseListQueryParams(params map[string]any) (sort string, filter string, fields []string, nextToken string, prevToken string, limit int, err error) {
	sort, err = parseSortQueryParams(params["sort"].(string))
	if err != nil {
		return "", "", nil, "", "", 0, err
	}

	filter, err = parseFilterQueryParams(params["filter"].(string))
	if err != nil {
		return "", "", nil, "", "", 0, err
	}

	fields, err = parseFieldsQueryParams(params["fields"].(string))
	if err != nil {
		return "", "", nil, "", "", 0, err
	}

	nextToken, err = parseNextTokenQueryParams(params["nextToken"].(string))
	if err != nil {
		return "", "", nil, "", "", 0, err
	}

	prevToken, err = parsePrevTokenQueryParams(params["prevToken"].(string))
	if err != nil {
		return "", "", nil, "", "", 0, err
	}

	limit, err = parseLimitQueryParams(params["limit"].(string))
	if err != nil {
		return "", "", nil, "", "", 0, err
	}

	return sort, filter, fields, nextToken, prevToken, limit, nil
}
