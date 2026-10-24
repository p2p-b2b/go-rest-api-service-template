package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

type fakeMembership struct {
	answer domain.ProjectMembership
	err    error
	calls  int
}

func (f *fakeMembership) Membership(_ context.Context, _, _ uuid.UUID) (domain.ProjectMembership, error) {
	f.calls++
	return f.answer, f.err
}

func TestRequireProjectMembership(t *testing.T) {
	t.Parallel()

	project := uuid.NewV7()
	user := uuid.NewV7()

	// A mux, so r.PathValue is populated the way it is in the real chain.
	serve := func(checker *fakeMembership, pattern, target string, withClaims bool) *httptest.ResponseRecorder {
		mux := http.NewServeMux()
		mux.Handle(pattern, RequireProjectMembership(checker)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))

		req := httptest.NewRequest(http.MethodGet, target, nil)
		if withClaims {
			req = req.WithContext(context.WithValue(req.Context(), JwtClaims, map[string]any{"sub": user.String()}))
		}

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	t.Run("a_member_passes", func(t *testing.T) {
		t.Parallel()
		f := &fakeMembership{answer: domain.ProjectMembershipMember}
		rec := serve(f, "GET /projects/{project_id}/embedding_configs", "/projects/"+project.String()+"/embedding_configs", true)
		if rec.Code != http.StatusOK || f.calls != 1 {
			t.Fatalf("status %d, calls %d", rec.Code, f.calls)
		}
	})

	t.Run("an_administrator_passes", func(t *testing.T) {
		t.Parallel()
		f := &fakeMembership{answer: domain.ProjectMembershipAdmin}
		rec := serve(f, "GET /projects/{project_id}/embedding_configs", "/projects/"+project.String()+"/embedding_configs", true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
	})

	t.Run("a_non_member_is_refused_as_not_found", func(t *testing.T) {
		t.Parallel()
		// 404, not 403: the same answer a missing project gets, so the refusal
		// does not confirm the project exists.
		f := &fakeMembership{answer: domain.ProjectMembershipNone}
		rec := serve(f, "GET /projects/{project_id}/embedding_configs", "/projects/"+project.String()+"/embedding_configs", true)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status %d, want 404: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("the_refusal_is_the_access_lines_not_a_line_of_its_own", func(t *testing.T) {
		t.Parallel()
		// A refusal is no error of the service: it is handed to the holder,
		// for the access line to carry at INFO, and it was a WARN line
		// besides. The body is the one a missing project gets.
		mux := http.NewServeMux()
		mux.Handle("GET /projects/{project_id}/products", RequireProjectMembership(&fakeMembership{answer: domain.ProjectMembershipNone})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))

		ctx := context.WithValue(o11y.WithFailure(t.Context()), JwtClaims, map[string]any{"sub": user.String()})
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/projects/"+project.String()+"/products", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status %d, want 404", rec.Code)
		}

		held, ok := o11y.FailureFrom(ctx)
		if !ok {
			t.Fatal("the refusal was not handed to the access line")
		}

		if !strings.Contains(held.Func, "middleware.RequireProjectMembership") {
			t.Errorf("the refusal is said to arise in %s, want the middleware", held.Func)
		}

		notFound, isNotFound := errors.AsType[*domain.ProjectNotFoundError](held.Err)
		if !isNotFound || notFound.ID != project {
			t.Fatalf("held %v, want the project's ProjectNotFoundError", held.Err)
		}

		want := (&domain.ProjectNotFoundError{ID: project}).Error()
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("body %s, want %q", rec.Body.String(), want)
		}
	})

	t.Run("a_store_fault_fails_closed", func(t *testing.T) {
		t.Parallel()
		f := &fakeMembership{err: errors.New("db away")}
		rec := serve(f, "GET /projects/{project_id}/embedding_configs", "/projects/"+project.String()+"/embedding_configs", true)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status %d, want 500", rec.Code)
		}
	})

	t.Run("the_nil_project_is_a_400", func(t *testing.T) {
		t.Parallel()
		// It parses, and the use-case refuses it as an invalid input: the
		// status table's 400, not the store fault's 500.
		f := &fakeMembership{err: &domain.InvalidInputError{Field: "project_id", Value: uuid.Nil().String(), Message: "project ID is nil"}}
		rec := serve(f, "GET /projects/{project_id}/products", "/projects/"+uuid.Nil().String()+"/products", true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status %d, want 400: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("a_route_without_a_project_is_not_asked", func(t *testing.T) {
		t.Parallel()
		f := &fakeMembership{answer: domain.ProjectMembershipNone}
		rec := serve(f, "GET /roles", "/roles", true)
		if rec.Code != http.StatusOK || f.calls != 0 {
			t.Fatalf("status %d, calls %d; a route with no project_id must pass untouched", rec.Code, f.calls)
		}
	})

	t.Run("no_claims_is_401", func(t *testing.T) {
		t.Parallel()
		f := &fakeMembership{answer: domain.ProjectMembershipMember}
		rec := serve(f, "GET /projects/{project_id}/embedding_configs", "/projects/"+project.String()+"/embedding_configs", false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status %d, want 401", rec.Code)
		}
	})
}
