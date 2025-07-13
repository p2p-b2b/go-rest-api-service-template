package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTableQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    createTable
		expected string
	}{
		{
			name: "Single column without constraints",
			input: createTable{
				Schema: "public",
				Table:  "users",
				Columns: []createTableColumns{
					{Name: "id", Type: "SERIAL"},
				},
			},
			expected: "CREATE TABLE public.users (id SERIAL);",
		},
		{
			name: "Multiple columns with constraints",
			input: createTable{
				Schema: "public",
				Table:  "users",
				Columns: []createTableColumns{
					{Name: "id", Type: "SERIAL", Constraints: []string{"PRIMARY KEY", "NOT NULL"}},
					{Name: "name", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			expected: "CREATE TABLE public.users (id SERIAL PRIMARY KEY NOT NULL, name VARCHAR(255) NOT NULL);",
		},
		{
			name: "No columns",
			input: createTable{
				Schema:  "public",
				Table:   "empty_table",
				Columns: []createTableColumns{},
			},
			expected: "CREATE TABLE public.empty_table ();",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createTableQuery(tt.input)
			if result != tt.expected {
				t.Errorf("createTableQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		field    string
		fields   []string
		expected string
	}{
		{
			name:     "empty field",
			prefix:   "t.",
			field:    "",
			fields:   []string{"id", "name", "created_at"},
			expected: "",
		},
		{
			name:     "field found with exact match",
			prefix:   "t.",
			field:    "name",
			fields:   []string{"id", "name", "created_at"},
			expected: "t.name",
		},
		{
			name:     "field found with AS alias",
			prefix:   "t.",
			field:    "user_name",
			fields:   []string{"id", "name AS user_name", "created_at"},
			expected: "name AS user_name",
		},
		{
			name:     "field found with partial match",
			prefix:   "t.",
			field:    "id",
			fields:   []string{"user_id", "name", "created_at"},
			expected: "t.id",
		},
		{
			name:     "field not found",
			prefix:   "t.",
			field:    "email",
			fields:   []string{"id", "name", "created_at"},
			expected: "",
		},
		{
			name:     "empty prefix",
			prefix:   "",
			field:    "name",
			fields:   []string{"id", "name", "created_at"},
			expected: "name",
		},
		{
			name:     "multiple fields with AS alias",
			prefix:   "t.",
			field:    "display_name",
			fields:   []string{"id", "first_name AS first", "last_name AS display_name"},
			expected: "last_name AS display_name",
		},
		{
			name:     "non-matching prefix case sensitivity",
			prefix:   "t.",
			field:    "Name",
			fields:   []string{"id", "name", "created_at"},
			expected: "",
		},
		{
			name:     "field found with mixed case",
			prefix:   "t.",
			field:    "UserName",
			fields:   []string{"id", "name AS UserName", "created_at"},
			expected: "name AS UserName",
		},
		{
			name:     "AS in lowercase",
			prefix:   "t.",
			field:    "user_name",
			fields:   []string{"id", "name as user_name", "created_at"},
			expected: "name as user_name",
		},
		{
			name:     "AS in uppercase",
			prefix:   "t.",
			field:    "user_name",
			fields:   []string{"id", "name AS user_name", "created_at"},
			expected: "name AS user_name",
		},
		{
			name:     "AS in mixed case",
			prefix:   "t.",
			field:    "user_name",
			fields:   []string{"id", "name As user_name", "created_at"},
			expected: "name As user_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldValue(tt.prefix, tt.field, tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrettyPrint(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		args     []any
		expected string
	}{
		{
			name: "Simple query without args",
			query: `SELECT *
                  FROM users`,
			args:     nil,
			expected: "SELECT * FROM users",
		},
		{
			name: "Query with SQL comments",
			query: `SELECT *
                  -- This is a comment
                  FROM users`,
			args:     nil,
			expected: "SELECT * FROM users",
		},
		{
			name: "Query with multiple whitespaces and newlines",
			query: `SELECT *
                  FROM   users
                  WHERE  id  =  1`,
			args:     nil,
			expected: "SELECT * FROM users WHERE id = 1",
		},
		{
			name: "Query with string argument",
			query: `SELECT *
                  FROM users
                  WHERE name = $1`,
			args:     []any{"John"},
			expected: "SELECT * FROM users WHERE name = 'John'",
		},
		{
			name: "Query with multiple string arguments",
			query: `SELECT *
                  FROM users
                  WHERE name = $1 AND email = $2`,
			args:     []any{"John", "john@example.com"},
			expected: "SELECT * FROM users WHERE name = 'John' AND email = 'john@example.com'",
		},
		{
			name: "Query with non-string argument",
			query: `SELECT *
                  FROM users
                  WHERE id = $1`,
			args:     []any{123},
			expected: "SELECT * FROM users WHERE id = 123",
		},
		{
			name: "Query with mixed arguments",
			query: `SELECT *
                  FROM users
                  WHERE id = $1 AND name = $2`,
			args:     []any{123, "John"},
			expected: "SELECT * FROM users WHERE id = 123 AND name = 'John'",
		},
		{
			name: "Query with boolean argument",
			query: `SELECT *
                  FROM users
                  WHERE active = $1`,
			args:     []any{true},
			expected: "SELECT * FROM users WHERE active = true",
		},
		{
			name: "Query with UUID argument",
			query: `SELECT *
                  FROM users
                  WHERE id = $1`,
			args:     []any{"550e8400-e29b-41d4-a716-446655440000"},
			expected: "SELECT * FROM users WHERE id = '550e8400-e29b-41d4-a716-446655440000'",
		},
		{
			name: "Complex query with multiple arguments and comments",
			query: `SELECT u.id, u.name, u.email
                  -- Get users
                  FROM users u
                  -- Join with roles
                  JOIN user_roles ur ON u.id = ur.user_id
                  WHERE u.department = $1 AND ur.role = $2
                  -- Filter by active status
                  AND u.active = $3`,
			args:     []any{"Engineering", "Admin", true},
			expected: "SELECT u.id, u.name, u.email FROM users u JOIN user_roles ur ON u.id = ur.user_id WHERE u.department = 'Engineering' AND ur.role = 'Admin' AND u.active = true",
		},
		{
			name: "Query with multi-line comments",
			query: `SELECT *
                  -- This is a comment
                  -- This is another comment
                  FROM users`,
			args:     nil,
			expected: "SELECT * FROM users",
		},
		{
			name: "Query with more than 9 parameters",
			query: `INSERT INTO users (id, name, email, phone, address, city, state, country, zip, active)
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			args:     []any{1, "John", "john@example.com", "1234567890", "123 Main St", "New York", "NY", "USA", "10001", true},
			expected: "INSERT INTO users (id, name, email, phone, address, city, state, country, zip, active) VALUES (1, 'John', 'john@example.com', '1234567890', '123 Main St', 'New York', 'NY', 'USA', '10001', true)",
		},
		{
			name:     "Empty query",
			query:    ``,
			args:     nil,
			expected: "",
		},
		{
			name: "Query with only comments",
			query: `-- This is a comment
                  -- This is another comment`,
			args:     nil,
			expected: "-- This is another comment",
		},
		{
			name: "Query with NULL argument",
			query: `SELECT *
                  FROM users
                  WHERE name = $1`,
			args:     []any{nil},
			expected: "SELECT * FROM users WHERE name = NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prettyPrint(tt.query, tt.args...)
			assert.Equal(t, tt.expected, result, "prettyPrint should transform the query correctly")
		})
	}
}

func TestBuildFieldSelection(t *testing.T) {
	tests := []struct {
		name            string
		sqlFieldsPrefix string
		fieldsArray     []string
		requestedFields string
		expected        string
	}{
		{
			name:            "no requested fields returns all fields",
			sqlFieldsPrefix: "t.",
			fieldsArray:     []string{"id", "name", "created_at"},
			requestedFields: "",
			expected:        "t.id, t.name, t.created_at",
		},
		{
			name:            "specific fields without id adds id and serial_id",
			sqlFieldsPrefix: "t.",
			fieldsArray:     []string{"id", "name", "created_at", "serial_id"},
			requestedFields: "name,created_at",
			expected:        "t.name, t.created_at, t.id, t.serial_id",
		},
		{
			name:            "id field included should not be duplicated",
			sqlFieldsPrefix: "t.",
			fieldsArray:     []string{"id", "name", "created_at", "serial_id"},
			requestedFields: "id,name",
			expected:        "t.id, t.name, t.serial_id",
		},
		{
			name:            "fields with AS aliases",
			sqlFieldsPrefix: "t.",
			fieldsArray:     []string{"id", "name AS user_name", "created_at", "serial_id"},
			requestedFields: "user_name",
			expected:        "name AS user_name, t.id, t.serial_id",
		},
		{
			name:            "id and serial_id both requested",
			sqlFieldsPrefix: "mt.",
			fieldsArray:     []string{"id", "name", "description", "serial_id"},
			requestedFields: "id,name,serial_id",
			expected:        "mt.id, mt.name, mt.serial_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFieldSelection(tt.sqlFieldsPrefix, tt.fieldsArray, tt.requestedFields)
			assert.Equal(t, tt.expected, result, "buildFieldSelection should construct the field selection correctly")
		})
	}
}
