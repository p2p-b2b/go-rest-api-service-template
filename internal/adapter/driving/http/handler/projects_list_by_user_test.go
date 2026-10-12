//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driving"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// projectsListStub is the projects port for the two lists, and records which
// one was asked and for whom. Any other method panics: the embedded port is
// nil.
type projectsListStub struct {
	driving.Projects
	input *domain.ListProjectsInput
	asked string
}

func (s *projectsListStub) ListByUserID(_ context.Context, in *domain.ListProjectsInput) (*domain.ListProjectsOutput, error) {
	s.asked, s.input = "ListByUserID", in

	return &domain.ListProjectsOutput{}, nil
}

func (s *projectsListStub) List(_ context.Context, in *domain.ListProjectsInput) (*domain.ListProjectsOutput, error) {
	s.asked, s.input = "List", in

	return &domain.ListProjectsOutput{}, nil
}

func projectsListHandler(t *testing.T, stub *projectsListStub) *ProjectsHandler {
	t.Helper()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	h, err := NewProjectsHandler(ProjectsHandlerConf{Service: stub, OT: ot})
	if err != nil {
		t.Fatalf("NewProjectsHandler: %v", err)
	}

	return h
}

// GET /users/{user_id}/projects names two accounts: whose projects, in the
// path, and who asks, on the verified token. The list is the first one's
// projects that are the second one's too, so each has to reach the port as
// what it is: with the path's user in both places, the list is that user's
// whole list.
func TestListByUserIDPassesThePathsUserAndTheCaller(t *testing.T) {
	t.Parallel()

	stub := &projectsListStub{}
	h := projectsListHandler(t, stub)

	user, caller := uuid.NewV7(), uuid.NewV7()

	req := httptest.NewRequest(http.MethodGet, "/users/"+user.String()+"/projects", nil)
	req.SetPathValue("user_id", user.String())

	rec := httptest.NewRecorder()
	h.listByUserID(rec, withCaller(req, caller))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if stub.asked != "ListByUserID" {
		t.Fatalf("the handler asked the port for %q, want ListByUserID", stub.asked)
	}

	if stub.input.UserID != user {
		t.Errorf("UserID = %s, want the path's user %s", stub.input.UserID, user)
	}

	if stub.input.CallerID != caller {
		t.Errorf("CallerID = %s, want the caller %s", stub.input.CallerID, caller)
	}
}

// A request that reaches the handler with no caller is a route wired without
// its middleware: a 500, and the port is not asked. Never the user's list.
func TestListByUserIDWithNoCallerAsksNothing(t *testing.T) {
	t.Parallel()

	stub := &projectsListStub{}
	h := projectsListHandler(t, stub)

	user := uuid.NewV7()

	req := httptest.NewRequest(http.MethodGet, "/users/"+user.String()+"/projects", nil)
	req.SetPathValue("user_id", user.String())

	rec := httptest.NewRecorder()
	h.listByUserID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500. Body: %s", rec.Code, rec.Body.String())
	}

	if stub.asked != "" {
		t.Errorf("the port was asked for %s with no caller", stub.asked)
	}
}
