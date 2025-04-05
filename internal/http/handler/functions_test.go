package handler

import (
	"testing"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/qfv"
	"github.com/stretchr/testify/assert"
)

func TestParseUUIDQueryParams(t *testing.T) {
	testID := uuid.New()

	tests := []struct {
		input    string
		expected uuid.UUID
		err      error
	}{
		{"", uuid.Nil, ErrRequiredUUID},
		{"invalid-uuid", uuid.Nil, ErrInvalidUUID},
		{uuid.Nil.String(), uuid.Nil, ErrUUIDCannotBeNil},
		{testID.String(), testID, nil},
	}

	for _, test := range tests {
		result, err := parseUUIDQueryParams(test.input)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParseSortQueryParams(t *testing.T) {
	allowedFields := []string{"name", "age"}

	tests := []struct {
		sort     string
		expected string
		err      error
	}{
		{"name ASC", "name ASC", nil},
		{"age ASC, name DESC", "age ASC, name DESC", nil},
		{"invalid", "", &qfv.QFVSortError{Field: "invalid", Message: "field not allowed for sorting"}},
	}

	for _, test := range tests {
		result, err := parseSortQueryParams(test.sort, allowedFields)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParseFilterQueryParams(t *testing.T) {
	allowedFields := []string{"status", "type"}

	tests := []struct {
		filter   string
		expected string
		err      error
	}{
		{"status='active'", "status='active'", nil},
		{"status='active' AND type=1", "status='active' AND type=1", nil},
		{"invalid", "", &qfv.QFVFilterError{Field: "", Message: "parsing errors: [error on field 'invalid': field not allowed error on field 'invalid': unexpected token after field]"}},
	}

	for _, test := range tests {
		result, err := parseFilterQueryParams(test.filter, allowedFields)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParseFieldsQueryParams(t *testing.T) {
	allowedFields := []string{"id", "name"}

	tests := []struct {
		fields   string
		expected string
		err      error
	}{
		{"id,name", "id,name", nil},
		{"invalid", "", &qfv.QFVFieldsError{Field: "invalid", Message: "unknown field"}},
	}

	for _, test := range tests {
		result, err := parseFieldsQueryParams(test.fields, allowedFields)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParseNextTokenQueryParams(t *testing.T) {
	testID := uuid.New()

	tests := []struct {
		nextToken string
		expected  string
		err       error
	}{
		{"", "", nil},
		{"invalid", "", ErrInvalidNextToken},
		{model.EncodeToken(testID, 10), model.EncodeToken(testID, 10), nil},
	}

	for _, test := range tests {
		result, err := parseNextTokenQueryParams(test.nextToken)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParsePrevTokenQueryParams(t *testing.T) {
	testID := uuid.New()

	tests := []struct {
		prevToken string
		expected  string
		err       error
	}{
		{"", "", nil},
		{"invalid", "", ErrInvalidPrevToken},
		{model.EncodeToken(testID, 10), model.EncodeToken(testID, 10), nil},
	}

	for _, test := range tests {
		result, err := parsePrevTokenQueryParams(test.prevToken)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParseLimitQueryParams(t *testing.T) {
	tests := []struct {
		limit    string
		expected int
		err      error
	}{
		{"", model.DefaultLimit, nil},
		{"invalid", 0, ErrInvalidLimit},
		{"0", model.DefaultLimit, nil},
		{"5", 5, nil},
		{"-1", 0, ErrInvalidLimit},
		{"1000", model.MaxLimit, nil},
	}

	for _, test := range tests {
		result, err := parseLimitQueryParams(test.limit)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestParseListQueryParams(t *testing.T) {
	t.Run("TestParseListQueryParams", func(t *testing.T) {
		testID := uuid.New()

		params := map[string]any{
			"sort":      "name ASC",
			"filter":    "status='active'",
			"fields":    "id, name",
			"nextToken": model.EncodeToken(testID, 10),
			"prevToken": model.EncodeToken(testID, 10),
			"limit":     "5",
		}

		sortFields := []string{"name", "age"}
		filterFields := []string{"status", "type"}
		fieldsFields := []string{"id", "name"}

		sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(params, fieldsFields, filterFields, sortFields)
		assert.NoError(t, err)
		assert.Equal(t, "name ASC", sort)
		assert.Equal(t, "status='active'", filter)
		assert.Equal(t, "id, name", fields)
		assert.Equal(t, model.EncodeToken(testID, 10), nextToken)
		assert.Equal(t, model.EncodeToken(testID, 10), prevToken)
		assert.Equal(t, 5, limit)
	})

	t.Run("TestParseListQueryParamsWithInvalidSort", func(t *testing.T) {
		params := map[string]any{
			"sort":      "invalid",
			"filter":    "status='active'",
			"fields":    "id, name",
			"nextToken": "",
			"prevToken": "",
			"limit":     "5",
		}

		sortFields := []string{"name", "age"}
		filterFields := []string{"status", "type"}
		fieldsFields := []string{"id", "name"}
		_, _, _, _, _, _, err := parseListQueryParams(params, fieldsFields, filterFields, sortFields)
		assert.Error(t, err)
		assert.Equal(t, &qfv.QFVSortError{Field: "invalid", Message: "field not allowed for sorting"}, err)
	})
}
