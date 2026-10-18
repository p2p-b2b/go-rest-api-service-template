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

// usersLockStub is the users port for the update, the delete and the unlink
// of roles. It records who it was told asks, and answers the refusal it is
// given. Any other method panics: the embedded port is nil.
type usersLockStub struct {
	driving.Users
	refusal      error
	user, caller uuid.UUID
	asked        int
}

func (s *usersLockStub) UpdateByID(_ context.Context, in *domain.UpdateUserInput) error {
	s.asked++
	s.user, s.caller = in.ID, in.CallerID

	return s.refusal
}

func (s *usersLockStub) DeleteByID(_ context.Context, in *domain.DeleteUserInput) error {
	s.asked++
	s.user, s.caller = in.ID, in.CallerID

	return s.refusal
}

func (s *usersLockStub) UnlinkRoles(_ context.Context, in *domain.UnlinkRolesFromUsersInput) error {
	s.asked++
	s.user, s.caller = in.UserID, in.CallerID

	return s.refusal
}

// rolesUnlinkStub is the roles port for the unlink of users.
type rolesUnlinkStub struct {
	driving.Roles
	refusal error
	users   []uuid.UUID
	caller  uuid.UUID
	asked   int
}

func (s *rolesUnlinkStub) UnlinkUsers(_ context.Context, in *domain.UnlinkUsersFromRoleInput) error {
	s.asked++
	s.caller, s.users = in.CallerID, in.UserIDs

	return s.refusal
}

func handlerTelemetry(t *testing.T) *o11y.OpenTelemetry {
	t.Helper()

	ctx := t.Context()

	return &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}
}

// The four requests that lock an account or take something from it, each
// with the stub that hears it and the handler that serves it.
type lockingRequest struct {
	serve   func(users *usersLockStub, roles *rolesUnlinkStub) http.HandlerFunc
	request func(user, other uuid.UUID) *http.Request
	// heard is what the use-case was told: whom it is about, and who asks.
	heard func(users *usersLockStub, roles *rolesUnlinkStub) (about, caller uuid.UUID, asked int)
}

func lockingRequests(t *testing.T) map[string]lockingRequest {
	t.Helper()

	usersHandler := func(stub *usersLockStub) *UsersHandler {
		h, err := NewUsersHandler(UsersHandlerConf{Service: stub, Recoverer: recovererNeverAsked{}, OT: handlerTelemetry(t)})
		if err != nil {
			t.Fatalf("NewUsersHandler: %v", err)
		}

		return h
	}

	json := func(method, path, body string, values map[string]string) *http.Request {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		for name, value := range values {
			req.SetPathValue(name, value)
		}

		return req
	}

	ofUsers := func(users *usersLockStub, _ *rolesUnlinkStub) (uuid.UUID, uuid.UUID, int) {
		return users.user, users.caller, users.asked
	}

	return map[string]lockingRequest{
		"the_update": {
			serve: func(users *usersLockStub, _ *rolesUnlinkStub) http.HandlerFunc { return usersHandler(users).updateByID },
			request: func(user, _ uuid.UUID) *http.Request {
				return json(http.MethodPut, "/users/"+user.String(), `{"disabled":true}`, map[string]string{"user_id": user.String()})
			},
			heard: ofUsers,
		},
		"the_delete": {
			serve: func(users *usersLockStub, _ *rolesUnlinkStub) http.HandlerFunc { return usersHandler(users).deleteByID },
			request: func(user, _ uuid.UUID) *http.Request {
				return json(http.MethodDelete, "/users/"+user.String(), "", map[string]string{"user_id": user.String()})
			},
			heard: ofUsers,
		},
		"a_users_roles": {
			serve: func(users *usersLockStub, _ *rolesUnlinkStub) http.HandlerFunc {
				return usersHandler(users).unlinkRoles
			},
			request: func(user, other uuid.UUID) *http.Request {
				return json(http.MethodDelete, "/users/"+user.String()+"/roles", `{"role_ids":["`+other.String()+`"]}`, map[string]string{"user_id": user.String()})
			},
			heard: ofUsers,
		},
		"a_roles_users": {
			serve: func(_ *usersLockStub, roles *rolesUnlinkStub) http.HandlerFunc {
				h, err := NewRolesHandler(RolesHandlerConf{Service: roles, OT: handlerTelemetry(t)})
				if err != nil {
					t.Fatalf("NewRolesHandler: %v", err)
				}

				return h.unlinkUsers
			},
			request: func(user, other uuid.UUID) *http.Request {
				return json(http.MethodDelete, "/roles/"+other.String()+"/users", `{"user_ids":["`+user.String()+`"]}`, map[string]string{"role_id": other.String()})
			},
			heard: func(_ *usersLockStub, roles *rolesUnlinkStub) (uuid.UUID, uuid.UUID, int) {
				about := uuid.Nil()
				if len(roles.users) == 1 {
					about = roles.users[0]
				}

				return about, roles.caller, roles.asked
			},
		},
	}
}

// What is done to an administrator account is held to who asks, in the
// use-cases. The caller has to reach them as the caller, beside the account
// the request is about: a use-case that is told nobody asks answers a fault
// for an administrator account and goes on for every other.
func TestWhatLocksAnAccountPassesTheCaller(t *testing.T) {
	t.Parallel()

	for name, c := range lockingRequests(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, caller, other := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()
			users, roles := &usersLockStub{}, &rolesUnlinkStub{}

			rec := httptest.NewRecorder()
			c.serve(users, roles)(rec, withCaller(c.request(user, other), caller))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}

			about, asking, asked := c.heard(users, roles)
			if asked != 1 || about != user || asking != caller {
				t.Errorf("the use-case heard %d requests, about %s as %s; want one, about %s as %s", asked, about, asking, user, caller)
			}
		})
	}
}

// The use-case's refusal is a 403 that says which rule it was, and a request
// with no caller on it is a route wired without its middleware: a 500, with
// the use-case not asked on the three routes that had no caller before.
func TestALockRefusedIsA403(t *testing.T) {
	t.Parallel()

	refusal := &domain.AdministratorProtectedError{Action: "disabled", Message: "the last enabled administrator account"}

	const says = "user cannot be disabled: system/protected resource: the last enabled administrator account"

	for name, c := range lockingRequests(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, other := uuid.NewV7(), uuid.NewV7()
			users, roles := &usersLockStub{refusal: refusal}, &rolesUnlinkStub{refusal: refusal}

			rec := httptest.NewRecorder()
			c.serve(users, roles)(rec, withCaller(c.request(user, other), uuid.NewV7()))

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403. Body: %s", rec.Code, rec.Body.String())
			}

			if !strings.Contains(rec.Body.String(), says) {
				t.Errorf("the answer does not say %q: %s", says, rec.Body.String())
			}

			// And with nobody on the request.
			unasked, unaskedRoles := &usersLockStub{}, &rolesUnlinkStub{}

			rec = httptest.NewRecorder()
			c.serve(unasked, unaskedRoles)(rec, c.request(user, other))

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("with no caller: status %d, want 500. Body: %s", rec.Code, rec.Body.String())
			}

			if _, _, asked := c.heard(unasked, unaskedRoles); asked != 0 {
				t.Errorf("with no caller the use-case was asked %d times", asked)
			}
		})
	}
}
