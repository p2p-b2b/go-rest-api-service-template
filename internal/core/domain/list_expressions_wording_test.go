//go:build unit

package domain

import (
	"errors"
	"regexp"
	"slices"
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

// TestEveryListIsJudgedByItsOwnAllowLists: a validator hands
// validateListExpressions its parsers and the allow-lists those parsers
// were built from, and the wording is decided from the lists. Handed
// another list's, or its filter list as its sort list, it would call an
// accepted field refused or a refused one malformed. For every list, a
// field one of its lists holds and another does not is named as not
// sortable, not filterable or not selectable by the one that does not.
func TestEveryListIsJudgedByItsOwnAllowLists(t *testing.T) {
	t.Parallel()

	type lists struct{ sort, filter, partial []string }

	page := Paginator{Limit: 10}

	for name, tc := range map[string]struct {
		lists lists
		input func(sort, filter, fields string) interface{ Validate() error }
	}{
		"IDPTypes": {lists{IDPTypesSortFields, IDPTypesFilterFields, IDPTypesPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectIDPTypesInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"IDPs": {lists{IDPsSortFields, IDPsFilterFields, IDPsPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectIDPsInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"Policies": {lists{PoliciesSortFields, PoliciesFilterFields, PoliciesPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectPoliciesInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"Products": {lists{ProductsSortFields, ProductsFilterFields, ProductsPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectProductsInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"Projects": {lists{ProjectSortFields, ProjectFilterFields, ProjectPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectProjectsInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"RateLimits": {lists{RateLimitsSortFields, RateLimitsFilterFields, RateLimitsPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectRateLimitsInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"ResourcesLimits": {lists{ResourcesLimitsSortFields, ResourcesLimitsFilterFields, ResourcesLimitsPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectResourcesLimitsInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"Resources": {lists{ResourcesSortFields, ResourcesFilterFields, ResourcesPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectResourcesInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"Roles": {lists{RolesSortFields, RolesFilterFields, RolesPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectRolesInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
		"Users": {lists{UsersSortFields, UsersFilterFields, UsersPartialFields}, func(s, f, p string) interface{ Validate() error } {
			return &SelectUsersInput{Sort: s, Filter: f, Fields: p, Paginator: page}
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			all := slices.Concat(tc.lists.sort, tc.lists.filter, tc.lists.partial)
			checked := 0

			for _, field := range all {
				if !listIdentifier.MatchString(field) {
					continue
				}

				for _, c := range []struct {
					in      []string
					build   func() error
					param   string
					refused string
				}{
					{tc.lists.sort, func() error { return tc.input(field+" asc", "", "").Validate() }, FieldSort, "'" + field + "' is not a sortable field"},
					{tc.lists.filter, func() error { return tc.input("", field+" = 'x'", "").Validate() }, FieldFilter, "'" + field + "' is not a filterable field"},
					{tc.lists.partial, func() error { return tc.input("", "", field).Validate() }, FieldFields, "'" + field + "' is not a selectable field"},
				} {
					if slices.Contains(c.in, field) {
						continue
					}

					checked++

					got, ok := listErrorFor(t, c.build(), c.param)
					if !ok || got.Message != c.refused {
						t.Errorf("%s by %q: got %q, want %q", c.param, field, got.Message, c.refused)
					}
				}
			}

			t.Logf("%d fields refused by one list and held by another", checked)
		})
	}
}
