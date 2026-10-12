//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driving"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// usersLinksStub is the users port for the two changes of a user's project
// links, and records what it was asked. It answers the refusal it is given.
// Any other method panics: the embedded port is nil.
type usersLinksStub struct {
	driving.Users
	input   *domain.LinkProjectsToUserInput
	refusal error
}

func (s *usersLinksStub) LinkProjects(_ context.Context, in *domain.LinkProjectsToUserInput) error {
	s.input = in

	return s.refusal
}

func (s *usersLinksStub) UnlinkProjects(_ context.Context, in *domain.UnlinkProjectsFromUserInput) error {
	s.input = in

	return s.refusal
}

func usersLinksHandler(t *testing.T, stub *usersLinksStub) *UsersHandler {
	t.Helper()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	h, err := NewUsersHandler(UsersHandlerConf{Service: stub, Recoverer: recovererNeverAsked{}, OT: ot})
	if err != nil {
		t.Fatalf("NewUsersHandler: %v", err)
	}

	return h
}

// POST and DELETE /users/{user_id}/projects take their projects from the
// body, and the ids are held to who asks, in the repository. The caller has
// to reach it as the caller, beside the path's user: with the path's user in
// both places the projects would be held to the user being changed.
func TestAUsersProjectLinksPassThePathsUserAndTheCaller(t *testing.T) {
	t.Parallel()

	for name, method := range map[string]string{"link": http.MethodPost, "unlink": http.MethodDelete} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stub := &usersLinksStub{}
			h := usersLinksHandler(t, stub)

			user, caller, project := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

			req := httptest.NewRequest(method, "/users/"+user.String()+"/projects", strings.NewReader(`{"project_ids":["`+project.String()+`"]}`))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("user_id", user.String())

			rec := httptest.NewRecorder()

			if method == http.MethodPost {
				h.linkProjects(rec, withCaller(req, caller))
			} else {
				h.unlinkProjects(rec, withCaller(req, caller))
			}

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}

			if stub.input.UserID != user {
				t.Errorf("UserID = %s, want the path's user %s", stub.input.UserID, user)
			}

			if stub.input.CallerID != caller {
				t.Errorf("CallerID = %s, want the caller %s", stub.input.CallerID, caller)
			}

			if len(stub.input.ProjectIDs) != 1 || stub.input.ProjectIDs[0] != project {
				t.Errorf("ProjectIDs = %v, want the body's", stub.input.ProjectIDs)
			}
		})
	}
}

// A project that is not the caller's is a 400 that names the id the caller
// sent, for a link and for an unlink. And a change that does not know who
// asks is not handed on.
func TestAProjectThatIsNotTheCallersIsA400(t *testing.T) {
	t.Parallel()

	foreign := uuid.NewV7()

	for name, method := range map[string]string{"link": http.MethodPost, "unlink": http.MethodDelete} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stub := &usersLinksStub{refusal: &domain.ProjectNotOfCallerError{ProjectID: foreign}}
			h := usersLinksHandler(t, stub)

			user := uuid.NewV7()
			body := `{"project_ids":["` + foreign.String() + `"]}`

			req := httptest.NewRequest(method, "/users/"+user.String()+"/projects", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("user_id", user.String())

			rec := httptest.NewRecorder()

			serve := h.linkProjects
			if method == http.MethodDelete {
				serve = h.unlinkProjects
			}

			serve(rec, withCaller(req, uuid.NewV7()))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400. Body: %s", rec.Code, rec.Body.String())
			}

			if !strings.Contains(rec.Body.String(), foreign.String()+" is not one of your projects") {
				t.Errorf("the body does not say which project: %s", rec.Body.String())
			}

			// With no caller on the request: a wiring mistake, a 500, and
			// nothing handed to the use-case.
			unasked := &usersLinksStub{}
			noCaller := httptest.NewRequest(method, "/users/"+user.String()+"/projects", strings.NewReader(body))
			noCaller.Header.Set("Content-Type", "application/json")
			noCaller.SetPathValue("user_id", user.String())

			rec = httptest.NewRecorder()

			serveUnasked := usersLinksHandler(t, unasked).linkProjects
			if method == http.MethodDelete {
				serveUnasked = usersLinksHandler(t, unasked).unlinkProjects
			}

			serveUnasked(rec, noCaller)

			if rec.Code != http.StatusInternalServerError || unasked.input != nil {
				t.Errorf("with no caller: status %d, handed on %+v", rec.Code, unasked.input)
			}
		})
	}
}
