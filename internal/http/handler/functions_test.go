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
		{"", uuid.Nil, &InvalidUUIDError{Message: "input is empty"}},
		{"invalid-uuid", uuid.Nil, &InvalidUUIDError{UUID: "invalid-uuid", Message: "invalid UUID length: 12"}},
		{uuid.Nil.String(), uuid.Nil, &InvalidUUIDError{UUID: uuid.Nil.String(), Message: "input is nil"}},
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
	validToken := model.EncodeToken(testID, 10, model.TokenDirectionNext)

	tests := []struct {
		name      string
		nextToken string
		expected  string
		err       error
	}{
		{
			name:      "Empty token",
			nextToken: "",
			expected:  "",
			err:       nil,
		},
		{
			name:      "Invalid token",
			nextToken: "invalid",
			expected:  "",
			err:       &model.InvalidPaginatorTokenError{Message: "invalid token: illegal base64 data at input byte 4"},
		},
		{
			name:      "Valid token",
			nextToken: validToken,
			expected:  validToken,
			err:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseNextTokenQueryParams(test.nextToken)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, test.err, err)
				// assert.Contains(t, err.Error(), test.err.Error())
			}

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParsePrevTokenQueryParams(t *testing.T) {
	testID := uuid.New()
	validPrevToken := model.EncodeToken(testID, 10, model.TokenDirectionPrev)
	validNextToken := model.EncodeToken(testID, 10, model.TokenDirectionNext)

	tests := []struct {
		name      string
		prevToken string
		expected  string
		err       error
	}{
		{
			name:      "Empty token",
			prevToken: "",
			expected:  "",
			err:       nil,
		},
		{
			name:      "Invalid token (bad format)",
			prevToken: "invalid", // Input that causes a base64 decoding error
			expected:  "",
			err:       &model.InvalidPaginatorTokenError{Message: "invalid cursor: invalid token: not base64"},
		},
		{
			name:      "Valid token (encoded as Prev)",
			prevToken: validPrevToken,
			expected:  validPrevToken,
			err:       nil,
		},
		{
			name:      "Invalid token (direction mismatch - next token for prev param)",
			prevToken: validNextToken,
			expected:  "", // On error, the function returns an empty string for the token
			err:       &model.InvalidPaginatorTokenError{Message: "invalid cursor: token direction mismatch: expected prev, got next"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parsePrevTokenQueryParams(test.prevToken)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, test.err, err)
				// Check if the actual error message contains the expected message
				// This provides flexibility if the error messages don't match exactly
				// but the core issue is the same.
				assert.Contains(t, err.Error(), test.err.Error())
			}

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseLimitQueryParams(t *testing.T) {
	tests := []struct {
		limit    string
		expected int
		err      error
	}{
		{"", model.PaginatorDefaultLimit, nil},
		{"invalid", 0, &model.InvalidPaginatorLimitError{MinLimit: model.PaginatorMinLimit, MaxLimit: model.PaginatorMaxLimit}},
		{"0", model.PaginatorDefaultLimit, nil},
		{"5", 5, nil},
		{"-1", 0, &model.InvalidPaginatorLimitError{MinLimit: model.PaginatorMinLimit, MaxLimit: model.PaginatorMaxLimit}},
		{"1000", model.PaginatorMaxLimit, nil},
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
			"nextToken": model.EncodeToken(testID, 10, model.TokenDirectionNext),
			"prevToken": model.EncodeToken(testID, 10, model.TokenDirectionPrev),
			"limit":     "5",
		}

		sortFields := []string{"name", "age"}
		filterFields := []string{"status", "type"}
		fieldsFields := []string{"id", "name"}

		sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(params, fieldsFields, filterFields, sortFields)
		assert.NoError(t, err)
		assert.Equal(t, "name ASC", sort)
		assert.Equal(t, "status='active'", filter)
		assert.Equal(t, "id,name", fields)
		assert.Equal(t, model.EncodeToken(testID, 10, model.TokenDirectionNext), nextToken)
		assert.Equal(t, model.EncodeToken(testID, 10, model.TokenDirectionPrev), prevToken)
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
