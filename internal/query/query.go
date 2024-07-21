package query

import (
	"strconv"
	"strings"
)

var (
	// !=, >=, <= are not supported yet
	// filterComparators = []string{"!=", ">=", "<=", "=", ">", "<"}

	filterComparators = []string{"=", ">", "<"}
	filterOperators   = []string{"AND", "OR"}

	sortOperators = []string{"ASC", "DESC"}
)

// GetFields returns a list of fields for partial response after trimming spaces.
func GetFields(fields string) []string {
	return tokenizeFields(fields)
}

// IsValidFields checks if a list of fields for partial response is valid.
// The fields parameter is a list of valid fields.
// The partial parameter is a string with the fields to validate.
// The function returns true if the fields are valid, false otherwise.
//
// Example:
// IsValidFields(
//
//	[]string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
//	"id, first_name, last_name"
//
// )
func IsValidFields(fields []string, partial string) bool {
	// if fields are empty, then partial is invalid
	if len(fields) == 0 {
		return false
	}

	// if partial is empty, then it is valid
	if partial == "" {
		return true
	}

	// Tokenize the partial string
	tokens := tokenizeFields(partial)

	// check if fields are valid
	for _, token := range tokens {
		if !isValidColumn(token, fields) {
			return false
		}
	}

	return true
}

// tokenizeFields splits a fields string into tokens trimmed by spaces.
func tokenizeFields(fields string) []string {
	var tokens []string
	for _, token := range strings.Split(fields, ",") {
		tokens = append(tokens, strings.TrimSpace(token))
	}

	return tokens
}

// IsValidSort checks if a sort string is valid SQL syntax.
// The columns parameter is a list of valid column names.
// The sort parameter is a string with the sort to validate.
// The function returns true if the sort is valid, false otherwise.
//
// Example:
// IsValidSort(
//
//	[]string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
//	"id ASC, first_name DESC"
//
// )
func IsValidSort(columns []string, sort string) bool {
	// if columns are empty, then sort is invalid
	if len(columns) == 0 {
		return false
	}

	// if sort is empty, then it is valid
	if sort == "" {
		return true
	}

	// Tokenize the sort string
	tokens := tokenizeSort(sort)

	// get the columns in the sort
	cols := getColumnsSort(tokens)

	// columns cannot be zero
	if len(cols) == 0 {
		return false
	}

	// check if cols are valid
	for _, col := range cols {
		if !isValidColumn(col, columns) {
			return false
		}
	}

	// get the operators in the sort
	operators := getOperatorsSort(tokens)

	// each column should have an operator
	if len(cols) != len(operators) {
		return false
	}

	// check if operators are valid
	for _, operator := range operators {
		if !isValidOperatorSort(operator) {
			return false
		}
	}

	return true
}

// IsValidFilter checks if a filter string is valid SQL syntax.
// The columns parameter is a list of valid column names.
// The filter parameter is a string with the filter to validate.
// The function returns true if the filter is valid, false otherwise.
//
// Example:
// IsValidFilter(
//
//	[]string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
//	"id=1 AND first_name='Alice'"
//
// )
func IsValidFilter(columns []string, filter string) bool {
	// if columns are empty, then filter is invalid
	if len(columns) == 0 {
		return false
	}

	// if filter is empty, then it is valid
	if filter == "" {
		return true
	}

	// Tokenize the filter string
	tokens := tokenizeFilter(filter)

	// get the operators in the filter
	operators := getOperatorsFilter(tokens)

	// get the pairs in the filter
	pairs := getPairsFilter(tokens)

	// pairs cannot be zero
	if len(pairs) == 0 {
		return false
	}

	// if pairs are greater than 1, then operators should be equal to pairs - 1
	if len(pairs) > 1 && len(operators) != len(pairs)-1 {
		return false
	}

	// get the columns in the filter
	cols := getColumnsFilter(pairs)

	// columns cannot be zero
	if len(cols) == 0 {
		return false
	}

	// get the values in the filter
	values := getValues(pairs)

	// values cannot be zero
	if len(values) == 0 {
		return false
	}

	// values and columns should be equal
	if len(cols) != len(values) {
		return false
	}

	// get the comparators in the filter
	comparators := getComparators(pairs)

	// comparators cannot be zero
	if len(comparators) == 0 {
		return false
	}

	// comparators and pairs should be equal
	if len(comparators) != len(pairs) {
		return false
	}

	// check if cols are valid
	for _, col := range cols {
		if !isValidColumn(col, columns) {
			return false
		}
	}

	// check if values are valid
	for _, value := range values {
		if !isValue(value) {
			return false
		}
	}

	return true
}

// tokenizeSort splits a sort string into tokens
// separated by commas.
func tokenizeSort(sort string) []string {
	return strings.Split(sort, ",")
}

// tokenize splits a filter string into tokens
// separated by spaces.
func tokenizeFilter(filter string) []string {
	// Implement tokenization logic here, e.g., using regexp
	// Split by spaces, handle quotes for values, etc.
	return strings.Split(filter, " ")
}

// isValidColumn checks if a value is a valid column.
func isValidColumn(value string, columns []string) bool {
	for _, column := range columns {
		if value == column {
			return true
		}
	}

	return false
}

// isValidComparator checks if a token is a valid comparator.
func isValidComparator(token string) bool {
	for _, comparator := range filterComparators {
		if token == comparator || strings.ToUpper(token) == comparator {
			return true
		}
	}
	return false
}

// isValidOperatorSort checks if a token is a valid operator.
func isValidOperatorSort(value string) bool {
	for _, operator := range sortOperators {
		if value == operator || strings.ToUpper(value) == operator {
			return true
		}
	}
	return false
}

// isValidOperatorFilter checks if a token is a valid operator.
func isValidOperatorFilter(value string) bool {
	for _, operator := range filterOperators {
		if value == operator || strings.ToUpper(value) == operator {
			return true
		}
	}
	return false
}

// isValue checks if a token is a valid value.
// Valid values can be single-quoted strings and numbers.
func isValue(value any) bool {
	switch value.(type) {
	case string:
		return isQuotedString(value.(string))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

// isQuotedString checks if a value is a valid single-quoted string.
func isQuotedString(value string) bool {
	return value[0] == '\'' && value[len(value)-1] == '\''
}

// getOperatorsSort returns the list of valid operators in the tokenized sort.
func getOperatorsSort(tokens []string) []string {
	operators := make([]string, 0)
	for _, token := range tokens {
		t := strings.TrimSpace(token)

		column := strings.Split(t, " ")
		if len(column) == 2 {
			operators = append(operators, column[1])
		}
	}

	return operators
}

// getOperatorsFilter returns the list of valid operators in the tokenized filter.
func getOperatorsFilter(tokens []string) []string {
	operators := make([]string, 0)
	for _, token := range tokens {
		t := strings.TrimSpace(token)

		if isValidOperatorFilter(t) {
			operators = append(operators, t)
		}
	}
	return operators
}

// getPairs returns the list of column-value pairs in the tokenized filter.
func getPairsFilter(tokens []string) []string {
	pairs := make([]string, 0)
	for _, token := range tokens {
		if !isValidOperatorFilter(token) {
			pairs = append(pairs, token)
		}
	}
	return pairs
}

// getColumnsSort returns the list of valid columns in the sort string.
func getColumnsSort(pairs []string) []string {
	columns := make([]string, 0)
	for _, pair := range pairs {
		p := strings.TrimSpace(pair)
		column := strings.Split(p, " ")
		columns = append(columns, column[0])
	}

	return columns
}

// getColumnsFilter returns the list of valid columns in the pairs values.
func getColumnsFilter(pairs []string) []string {
	columns := make([]string, 0)
	for _, pair := range pairs {
		p := strings.TrimSpace(pair)
		for _, comparator := range filterComparators {
			column := strings.Split(p, comparator)
			if len(column) == 2 {
				columns = append(columns, column[0])
			}
		}
	}

	return columns
}

// getValues returns the list of values in the pairs values.
func getValues(pairs []string) []any {
	values := make([]any, 0)
	for _, pair := range pairs {
		for _, comparator := range filterComparators {
			if strings.Contains(pair, comparator) {
				value := strings.Split(pair, comparator)[1]

				// Check if the value is a quoted string
				if isQuotedString(value) {
					values = append(values, value)
				} else {
					// Try to parse the value as an integer
					if intValue, err := strconv.Atoi(value); err == nil {
						values = append(values, intValue)
					} else {
						// Try to parse the value as a float
						if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
							values = append(values, floatValue)
						} else {
							// Invalid value
							return nil
						}
					}
				}
			}
		}
	}

	return values
}

// getComparators returns the list of valid comparators in the pairs values.
func getComparators(pairs []string) []string {
	comparators := make([]string, 0)
	for _, pair := range pairs {
		for _, comparator := range filterComparators {
			if strings.Contains(pair, comparator) {
				comparators = append(comparators, comparator)
			}
		}
	}

	return comparators
}
