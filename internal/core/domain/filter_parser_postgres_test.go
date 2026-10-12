package domain

import (
	"errors"
	"testing"
)

// A caller's filter reaches PostgreSQL as the text the caller wrote, so what
// the filter parser accepts is what PostgreSQL is asked to read. This pins,
// in the unit suite, three places where the two disagreed until qfv v1.0.4.
// What the pattern operators take is PostgreSQL's, from its manual
// (9.7. Pattern Matching: `string LIKE pattern`, `text ~ text`).
func TestTheFilterParserReadsWhatPostgresReads(t *testing.T) {
	t.Parallel()

	validate := func(filter string) error {
		return (&SelectUsersInput{Filter: filter, Paginator: Paginator{Limit: 10}}).Validate()
	}

	for _, filter := range []string{
		// A pattern is a string: PostgreSQL has no ~~ that takes a number
		// or a boolean, whatever the column's type is.
		"first_name LIKE 1",
		"first_name NOT ILIKE -1",
		"first_name ~~ 1",
		"first_name SIMILAR TO 1",
		"first_name LIKE TRUE",
		// YES and NO are names to PostgreSQL, and were read as booleans.
		"disabled = YES",
		"disabled = no",
		"disabled IS YES",
	} {
		t.Run("refused/"+filter, func(t *testing.T) {
			t.Parallel()

			err := validate(filter)

			refusal, ok := errors.AsType[*ValidationErrors](err)
			if !ok {
				t.Fatalf("answered %T (%v), want the validation's refusal", err, err)
			}

			onFilter := false

			for _, e := range refusal.Errors {
				onFilter = onFilter || e.Field == "filter"
			}

			if !onFilter {
				t.Errorf("refused, and not on the filter: %v", err)
			}
		})
	}

	for _, filter := range []string{
		// IS is the keyword in lower case too, as every other keyword was.
		"id is null",
		"id is not null",
		"disabled is true",
		"disabled is not false",
		// A number written as a string is a pattern.
		"first_name LIKE '1'",
		"first_name SIMILAR TO '(a|b)%'",
		// The two boolean literals SQL has.
		"disabled = TRUE",
		"disabled = false",
	} {
		t.Run("accepted/"+filter, func(t *testing.T) {
			t.Parallel()

			if err := validate(filter); err != nil {
				t.Errorf("refused: %v", err)
			}
		})
	}
}
