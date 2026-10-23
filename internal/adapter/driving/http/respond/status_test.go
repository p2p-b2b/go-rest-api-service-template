package respond

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// StatusFor finds a type anywhere in an error's chain, and answers 500 for an
// error it does not know.
func TestStatusFor(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		err  error
		want int
	}{
		"a_type_of_the_table":       {err: &domain.UserNotFoundError{}, want: http.StatusNotFound},
		"the_same_type_wrapped":     {err: fmt.Errorf("reading the account: %w", &domain.UserNotFoundError{}), want: http.StatusNotFound},
		"a_conflict":                {err: &domain.IDPInUseError{}, want: http.StatusConflict},
		"a_dependency_that_is_gone": {err: &domain.IDPUnreachableError{}, want: http.StatusServiceUnavailable},
		"an_error_it_does_not_know": {err: errors.New("the database went away"), want: http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := StatusFor(c.err); got != c.want {
				t.Errorf("StatusFor = %d, want %d", got, c.want)
			}
		})
	}
}

// The table holds each type once, and never at 500: an error it does not
// know is already that.
func TestTheTableHoldsEachTypeOnceAndNoneAt500(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}

	for _, r := range statusTable {
		if seen[r.name] {
			t.Errorf("%s is in the table twice", r.name)
		}

		seen[r.name] = true

		if r.status == http.StatusInternalServerError {
			t.Errorf("%s is in the table at 500, which is what the table answers for what it does not know", r.name)
		}

		if !strings.HasPrefix(r.name, "*domain.") {
			t.Errorf("a row is named %q, want a domain error type", r.name)
		}
	}
}

// A refusal is what the caller got wrong: a status below 500 in the table,
// or a body that could not be decoded, whatever type the decoder gave it.
// Anything else is a failure, a dependency the table answers 503 included.
func TestIsRefusal(t *testing.T) {
	t.Parallel()

	syntaxErr := json.Unmarshal([]byte("{"), &struct{}{})
	if syntaxErr == nil {
		t.Fatal("the decoder accepted a broken body")
	}

	for name, c := range map[string]struct {
		err  error
		want bool
	}{
		"a_missing_row":                    {err: &domain.ProjectNotFoundError{}, want: true},
		"a_body_that_is_not_json":          {err: &DecodeError{Err: syntaxErr}, want: true},
		"a_decoded_body_of_a_known_type":   {err: &DecodeError{Err: &domain.InvalidRequestError{Message: "unknown field"}}, want: true},
		"a_dependency_that_is_gone":        {err: &domain.IDPUnreachableError{}, want: false},
		"an_error_the_table_does_not_know": {err: errors.New("the database went away"), want: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := IsRefusal(c.err); got != c.want {
				t.Errorf("IsRefusal = %v, want %v", got, c.want)
			}
		})
	}
}

// WriteError answers a type of the table with its status and its message,
// and anything else with the fixed 500, which names nothing.
func TestWriteError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), &domain.RoleNotFoundError{Message: "role not found"})

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "role not found") {
		t.Errorf("a missing role is answered %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), errors.New("dial tcp 10.0.0.9:5432: connection refused"))

	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "10.0.0.9") {
		t.Errorf("an error the table does not know is answered %d %s", rec.Code, rec.Body.String())
	}
}
