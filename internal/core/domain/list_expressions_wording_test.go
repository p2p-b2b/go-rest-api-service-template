//go:build unit

package domain

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

func listErrorFor(t *testing.T, err error, field string) (ValidationError, bool) {
	t.Helper()

	verrs, ok := errors.AsType[*ValidationErrors](err)
	if !ok {
		return ValidationError{}, false
	}

	for _, e := range verrs.Errors {
		if e.Field == field {
			return e, true
		}
	}

	return ValidationError{}, false
}

// TestListExpressionsAreRefusedInOurWords: a refused sort, filter or fields
// expression is answered in the service's wording under one code, never the
// qfv library's error text, which a qfv release could reword.
func TestListExpressionsAreRefusedInOurWords(t *testing.T) {
	t.Parallel()

	products := func(sort, filter, fields string) error {
		return (&SelectProductsInput{Sort: sort, Filter: filter, Fields: fields, Paginator: Paginator{Limit: 10}}).Validate()
	}

	cases := []struct {
		name, sort, filter, fields, field, want string
	}{
		{name: "sort_unknown_field", sort: "nope asc", field: FieldSort, want: "'nope' is not a sortable field"},
		{name: "sort_missing_direction", sort: "name", field: FieldSort, want: "must be a comma-separated list of 'field asc' or 'field desc'"},
		{name: "sort_bad_direction", sort: "name sideways", field: FieldSort, want: "must be a comma-separated list of 'field asc' or 'field desc'"},
		{name: "filter_unknown_field", filter: "nope = 'x'", field: FieldFilter, want: "'nope' is not a filterable field"},
		{name: "filter_syntax", filter: "name = ", field: FieldFilter, want: "is not a valid filter expression"},
		{name: "filter_unknown_field_in_a_compound", filter: "name = 'a' AND nope = 'b'", field: FieldFilter, want: "'nope' is not a filterable field"},
		{name: "fields_unknown_field", fields: "name,nope", field: FieldFields, want: "'nope' is not a selectable field"},
		{name: "fields_empty_member", fields: "name,,id", field: FieldFields, want: "must be a comma-separated list of field names"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := listErrorFor(t, products(tc.sort, tc.filter, tc.fields), tc.field)
			if !ok {
				t.Fatalf("no %s error", tc.field)
			}

			if got.Message != tc.want || got.Code != ListExpressionCode {
				t.Fatalf("got %q (%s), want %q (%s)", got.Message, got.Code, tc.want, ListExpressionCode)
			}
		})
	}
}

// TestAListExpressionEchoesOnlyAnIdentifier: whatever the client sends, the
// message is either the generic one or names a short plain identifier;
// nothing else from the input is repeated back.
func TestAListExpressionEchoesOnlyAnIdentifier(t *testing.T) {
	t.Parallel()

	named := regexp.MustCompile(`^'[A-Za-z_][A-Za-z0-9_.]{0,63}' is not a filterable field$`)

	for _, filter := range []string{"<script> == 1", "na\x00me == 1", "\x00 == 1", strings.Repeat("a", 200) + " == 1", "'); DROP TABLE x; -- == 1"} {
		err := (&SelectProductsInput{Filter: filter, Paginator: Paginator{Limit: 10}}).Validate()

		got, ok := listErrorFor(t, err, FieldFilter)
		if !ok {
			t.Fatalf("%q: no filter error", filter)
		}

		if got.Message != "is not a valid filter expression" && !named.MatchString(got.Message) {
			t.Errorf("%q: message %q repeats more than an identifier", filter, got.Message)
		}
	}
}
