package repository

import (
	"fmt"
	"html/template"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
)

// prettyPrint removes comments and extra spaces from a query.
func prettyPrint(query string, arg ...string) string {
	ws := regexp.MustCompile(`\s+`)

	out := regexp.MustCompile(`--.*\n`).ReplaceAllString(query, "")
	out = strings.ReplaceAll(out, "\n", "")
	out = ws.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)

	if len(arg) > 0 {
		// replace the pattern $1,$2,...., $n by %v
		re := regexp.MustCompile(`\$\d{1,2}`)
		out = re.ReplaceAllString(out, "%v")

		// Convert []string to []any for fmt.Sprintf
		args := make([]any, len(arg))
		for i, a := range arg {
			args[i] = a
		}

		out = fmt.Sprintf(out, args...)
	}

	return out
}

// getFieldValue returns the field value from the given fields slice.
// It checks if the field is empty or if it exists in the fields slice.
// If it exists, it returns the field value.
// If it doesn't exist, it returns an empty string.
// It also checks for the prefix and adds it to the field value if necessary.
// The function is case-insensitive for the "AS" keyword.
// It also checks for the field name in the fields slice.
func getFieldValue(prefix, field string, fields []string) string {
	if field == "" {
		return ""
	}

	for _, f := range fields {
		if strings.HasSuffix(f, " AS "+field) ||
			strings.HasSuffix(f, " as "+field) ||
			strings.HasSuffix(f, " As "+field) ||
			strings.HasSuffix(f, " aS "+field) {
			return f
		} else if strings.Contains(f, field) {
			return prefix + field
		}
	}

	return ""
}

// buildFieldSelection constructs the field selection part of the SQL query based on
// the requested fields, ensuring that ID and SerialID are always included.
func buildFieldSelection(
	sqlFieldsPrefix string,
	fieldsArray []string,
	requestedFields string,
) string {
	var fieldsStr string
	for i, field := range fieldsArray {

		if strings.Contains(field, " AS ") {
			fieldsStr += field + ", "
		} else {
			fieldsStr += sqlFieldsPrefix + field + ", "
		}

		// if it is the last field, remove the last comma
		if i == len(fieldsArray)-1 {
			fieldsStr = strings.TrimSuffix(fieldsStr, ", ")
		}
	}

	// If no specific fields requested, return the full field list
	if requestedFields == "" {
		return fieldsStr
	}

	inputFields := strings.Split(requestedFields, ",")
	fields := make([]string, 0)
	var idIsPresent bool

	for _, field := range inputFields {
		field = strings.TrimSpace(field)

		field := getFieldValue(sqlFieldsPrefix, field, fieldsArray)
		fields = append(fields, field)

		if field == "id" {
			idIsPresent = true
		}
	}

	// ID and SerialID are required for pagination
	if !idIsPresent {
		fields = append(fields, sqlFieldsPrefix+"id")
	}

	fields = append(fields, sqlFieldsPrefix+"serial_id")
	return strings.Join(fields, ", ")
}

// buildPaginationCriteria constructs the SQL criteria for pagination based on token direction.
// It handles both next and previous pagination directions and returns the appropriate SQL WHERE clause
// and sort order to apply to the query.
// The function takes the table alias, token direction, ID, serial ID, and filter query as input parameters.
// It returns the WHERE clause and the internal sort order as a string.
// The function uses the template.HTML type to safely handle HTML content.
// It is important to note that this function does not perform any SQL injection prevention,
// so it should be used with caution and only with trusted input.
func buildPaginationCriteria(
	tableAlias string,
	tokenDirection model.TokenDirection,
	id uuid.UUID,
	serial int64,
	filterQuery string,
) (whereClause template.HTML, internalSort string) {
	// Default sort order (newest to oldest)
	internalSort = fmt.Sprintf("%s.serial_id DESC, %s.id DESC", tableAlias, tableAlias)

	// If filter is empty, start with WHERE, otherwise AND
	filterQueryJoiner := "WHERE"
	if filterQuery != "" {
		filterQueryJoiner = "AND"
	}

	// Handle token directions
	switch tokenDirection {
	case model.TokenDirectionNext:
		// For next token, get records with lower serial ID (older records)
		internalSort = fmt.Sprintf("%s.serial_id DESC, %s.id DESC", tableAlias, tableAlias)
		whereClause = template.HTML(fmt.Sprintf(`
			%s
				%s (%s.serial_id < '%d')
				AND (%s.id < '%s' OR %s.serial_id < '%d')`,
			filterQuery,
			filterQueryJoiner,
			tableAlias,
			serial,
			tableAlias,
			id.String(),
			tableAlias,
			serial,
		))

	case model.TokenDirectionPrev:
		// For prev token, get records with higher serial ID (newer records)
		internalSort = fmt.Sprintf("%s.serial_id ASC, %s.id ASC", tableAlias, tableAlias)
		whereClause = template.HTML(fmt.Sprintf(`
			%s
				%s (%s.serial_id > '%d')
				AND (%s.id > '%s' OR %s.serial_id > '%d')`,
			filterQuery,
			filterQueryJoiner,
			tableAlias,
			serial,
			tableAlias,
			id.String(),
			tableAlias,
			serial,
		))

	default:
		// No pagination token provided, just use the filter
		whereClause = template.HTML(filterQuery)
	}

	return whereClause, internalSort
}

// injectPrefixToFields intelligently injects a prefix to specific fields in a filter query.
// It uses a regular expression that identifies and ignores matches within quoted strings
// to prevent incorrectly modifying string literals.
func injectPrefixToFields(prefix, filter string, allowedFields []string) string {
	if prefix == "" || filter == "" || len(allowedFields) == 0 {
		return filter
	}

	// To prevent ambiguous matches where one allowed field is a substring of another
	// (e.g., "id" and "user_id"), we sort the fields by length in descending order.
	// The regex engine will then try to match longer fields first.
	sort.Slice(allowedFields, func(i, j int) bool {
		return len(allowedFields[i]) > len(allowedFields[j])
	})

	escapedFields := make([]string, len(allowedFields))
	for i, field := range allowedFields {
		escapedFields[i] = regexp.QuoteMeta(field)
	}

	// This pattern has three alternatives, separated by the OR operator `|`:
	// 1. `'[^']*'`: Matches any content within a pair of single quotes.
	// 2. `"[^"]*"`: Matches any content within a pair of double quotes.
	// 3. `\b(field1|field2)\b`: Matches any of the allowed fields as whole words.
	// By placing the quote patterns first, the regex engine prioritizes matching them.
	fieldsPattern := `\b(` + strings.Join(escapedFields, "|") + `)\b`
	pattern := `'[^']*'|"[^"]*"|` + fieldsPattern

	re := regexp.MustCompile(pattern)

	// The callback function inspects each match to decide on the replacement.
	return re.ReplaceAllStringFunc(filter, func(match string) string {
		// Check if the match is a quoted literal (starts with ' or ").
		// If it is, we return it unchanged.
		if strings.HasPrefix(match, "'") || strings.HasPrefix(match, `"`) {
			return match
		}

		// Otherwise, the match must be one of our target fields.
		// We prepend the prefix and return the new string.
		return prefix + match
	})
}
