package query

import (
	"regexp"
	"strconv"
	"strings"
)

var sortOperators = []string{"ASC", "DESC"}

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

	// get the operators in the filter
	operators := getOperatorsFilter(filter)

	// get the pairs in the filter
	pairs := getPairsFilter(filter)

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

	// check if cols are valid
	for _, col := range cols {
		if !isValidColumn(col, columns) {
			return false
		}
	}

	// get the values in the filter
	values := getValuesFilter(pairs)

	// values cannot be zero
	if len(values) == 0 {
		return false
	}

	// values and columns should be equal
	if len(cols) != len(values) {
		return false
	}

	// check if values are valid
	for _, value := range values {
		if !isValue(value) {
			return false
		}
	}

	// get the comparators in the filter
	comparators := getComparatorsFilter(pairs)

	// comparators cannot be zero
	if len(comparators) == 0 {
		return false
	}

	// comparators and pairs should be equal
	if len(comparators) != len(pairs) {
		return false
	}

	return true
}

// tokenizeSort splits a sort string into tokens
// separated by commas.
func tokenizeSort(sort string) []string {
	return strings.Split(sort, ",")
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

// isValidOperatorSort checks if a token is a valid operator.
func isValidOperatorSort(value string) bool {
	for _, operator := range sortOperators {
		if value == operator || strings.ToUpper(value) == operator {
			return true
		}
	}
	return false
}

// isValue checks if a token is a valid value.
// Valid values can be single-quoted strings and numbers.
func isValue(value any) bool {
	switch v := value.(type) {
	case string:
		return isQuotedString(v) || isNumber(v)
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

// isNumber checks if a value is a valid number.
func isNumber(value string) bool {
	if _, err := strconv.Atoi(value); err == nil {
		return true
	} else {
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return true
		} else {
			return false
		}
	}
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
func getOperatorsFilter(filter string) []string {
	// https://regex101.com/r/6HPVL2/1
	re := regexp.MustCompile(`\s(AND|OR|and|or)\s`)

	matches := re.FindAllString(filter, -1)
	tokens := make([]string, 0, len(matches))
	for _, match := range matches {
		if match != "" {
			tokens = append(tokens, strings.TrimSpace(match))
		}
	}
	return tokens
}

// getPairs returns the list of column-value pairs in the tokenized filter.
func getPairsFilter(filter string) []string {
	// https://regex101.com/r/3aqJcV/4
	re := regexp.MustCompile(`(\w+\s*(=|!=)\s*('.*?'|".*?"))|(\w+\s{0,}(>=|<=|<|>|=)\s{0,}(\d{1,15}(\.\d{1,15}){0,1})\s{0,}?)`)

	matches := re.FindAllString(filter, -1)
	tokens := make([]string, 0, len(matches))
	for _, match := range matches {
		if match != "" {
			tokens = append(tokens, strings.TrimSpace(match))
		}
	}
	return tokens
}

// getComparatorsFilter returns the list of valid comparators in the pairs values.
func getComparatorsFilter(pairs []string) []string {
	re := regexp.MustCompile(`(\s*(=|!=)\s*)|(\s{0,}(>=|<=|<|>|=)\s{0,})`)

	comparators := make([]string, 0)
	for _, pair := range pairs {
		matches := re.FindAllString(pair, -1)

		for _, match := range matches {
			if match != "" && len(match) <= 2 {
				comparators = append(comparators, strings.TrimSpace(match))
			}
		}
	}

	return comparators
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

// getColumnsFilter returns the list of columns in the pairs values.
func getColumnsFilter(pairs []string) []string {
	re := regexp.MustCompile(`(\s*(=|!=)\s*)|(\s{0,}(>=|<=|<|>|=)\s{0,})`)

	columns := make([]string, 0)
	for _, pair := range pairs {
		p := strings.TrimSpace(pair)
		matches := re.Split(p, 2)

		for i, match := range matches {
			if match != "" {
				// get only the first part of the pair
				if i%2 == 0 {
					columns = append(columns, strings.TrimSpace(match))
				}
			}
		}

	}

	return columns
}

// getValuesFilter returns the list of values in the pairs values.
func getValuesFilter(pairs []string) []string {
	re := regexp.MustCompile(`(\s*(=|!=)\s*)|(\s{0,}(>=|<=|<|>|=)\s{0,})`)

	values := make([]string, 0)
	for _, pair := range pairs {
		p := strings.TrimSpace(pair)
		matches := re.Split(p, 2)

		for i, match := range matches {
			if match != "" {
				// get only the second part of the pair
				if i%2 == 1 {
					values = append(values, strings.TrimSpace(match))
				}
			}
		}

	}

	return values
}
