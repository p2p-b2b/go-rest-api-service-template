package repositorypg

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// filterRefusal maps the SQLSTATEs a filter can raise, on a list that had
// one, to the caller's 400 in our words, and leaves everything else alone.
func TestAFilterPostgresRefusesIsTheCallersError(t *testing.T) {
	t.Parallel()

	const clause = "WHERE (usrs.id = 'nope')"
	const statement = "SELECT usrs.id FROM users AS usrs WHERE (usrs.id = 'nope') ORDER BY usrs.id"

	refused := func(t *testing.T, err error, want string) {
		t.Helper()

		got, ok := errors.AsType[*domain.InvalidInputError](err)
		if !ok || got.Field != domain.FieldFilter || got.Message != want {
			t.Fatalf("got %T (%v), want the filter refused with %q", err, err, want)
		}

		if _, wraps := errors.AsType[*pgconn.PgError](err); wraps {
			t.Error("the refusal wraps the driver's error, whose text is Postgres's")
		}
	}

	for code, want := range filterRefusals {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			refused(t, filterRefusal(clause, statement, &pgconn.PgError{Code: code, Message: `invalid input syntax for type uuid: "nope"`}), want)
		})
	}

	t.Run("without_a_filter_it_is_ours", func(t *testing.T) {
		t.Parallel()

		pgErr := &pgconn.PgError{Code: "22P02"}
		if err := filterRefusal("", statement, pgErr); !errors.Is(err, pgErr) {
			t.Errorf("a refusal on a list with no filter became %v", err)
		}
	})

	t.Run("a_code_a_filter_does_not_raise_is_ours", func(t *testing.T) {
		t.Parallel()

		pgErr := &pgconn.PgError{Code: "2201W"} // a negative LIMIT
		if err := filterRefusal(clause, statement, pgErr); !errors.Is(err, pgErr) {
			t.Errorf("%v became the caller's", err)
		}
	})

	t.Run("not_a_postgres_error", func(t *testing.T) {
		t.Parallel()

		other := errors.New("the connection is closed")
		if err := filterRefusal(clause, statement, other); !errors.Is(err, other) {
			t.Errorf("became %v", err)
		}
	})

	// A syntax error is the filter's only when Postgres stopped inside it.
	inside := int32(strings.Index(statement, "'nope'") + 1)
	outside := int32(strings.Index(statement, "ORDER") + 1)

	t.Run("a_syntax_error_inside_the_filter", func(t *testing.T) {
		t.Parallel()

		refused(t, filterRefusal(clause, statement, &pgconn.PgError{Code: pgSyntaxError, Position: inside}), filterExpressionRefused)
	})

	t.Run("a_syntax_error_outside_the_filter_is_ours", func(t *testing.T) {
		t.Parallel()

		pgErr := &pgconn.PgError{Code: pgSyntaxError, Position: outside}
		if err := filterRefusal(clause, statement, pgErr); !errors.Is(err, pgErr) {
			t.Errorf("a syntax error at %d, after the filter, became %v", outside, err)
		}
	})

	t.Run("a_syntax_error_with_no_position_is_ours", func(t *testing.T) {
		t.Parallel()

		pgErr := &pgconn.PgError{Code: pgSyntaxError}
		if err := filterRefusal(clause, statement, pgErr); !errors.Is(err, pgErr) {
			t.Errorf("became %v", err)
		}
	})
}

// repositorySources returns the package's Go files that are not tests.
func repositorySources(t *testing.T) map[string]string {
	t.Helper()

	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	sources := map[string]string{}

	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		sources[path] = string(body)
	}

	return sources
}

// TestNoDriverTextInABody: the four errors every repository answers as a 400
// carry our sentence, never pgErr.Message, which named the statement's
// aliases and echoed the caller's input ("column usrs.nosuch does not
// exist"), and which a Postgres or driver upgrade rewrites.
func TestNoDriverTextInABody(t *testing.T) {
	t.Parallel()

	driverText := regexp.MustCompile(`&domain\.(InvalidByteSequenceError|InvalidMessageFormatError|UndefinedColumnError|DatatypeMismatchError)\{[^}]*pgErr\.Message`)

	for path, body := range repositorySources(t) {
		for _, found := range driverText.FindAllString(body, -1) {
			t.Errorf("%s: %s: a 400 that carries Postgres's text", path, found)
		}
	}
}

// TestEveryFilteredListMapsPostgresRefusals: a list that writes the caller's
// filter into its statement hands every error of its query and its rows to
// filterRefusal, or it answers `id = 'nope'` with a 500.
func TestEveryFilteredListMapsPostgresRefusals(t *testing.T) {
	t.Parallel()

	writesAFilter := regexp.MustCompile(`filterQuery = fmt\.Sprintf\("(WHERE|AND) \(%s\)"`)
	lists := 0

	for path, body := range repositorySources(t) {
		for _, fn := range strings.Split(body, "\nfunc ")[1:] {
			if !writesAFilter.MatchString(fn) {
				continue
			}

			lists++
			name := fn[:strings.Index(fn, "(ctx")]
			queries := strings.Count(fn, "ref.db.Query(")
			mapped := strings.Count(fn, "filterRefusal(filterQuery, ")
			forwarded := strings.Contains(fn, "filterQuery, query)")

			// Each query's error, and its rows' error, or the rows handed on
			// with the filter to the function that reads them.
			if mapped < 2*queries && !(forwarded && mapped >= queries) {
				t.Errorf("%s %s: %d queries, %d errors handed to filterRefusal", path, name, queries, mapped)
			}
		}
	}

	if lists < 17 {
		t.Errorf("%d lists write a filter, want at least 17: the guard is reading nothing", lists)
	}
}
