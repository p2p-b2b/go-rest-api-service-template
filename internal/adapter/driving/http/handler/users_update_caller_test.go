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

// usersUpdateStub is the users port for the update, and records what it was
// asked. It answers the refusal it is given. Any other method panics: the
// embedded port is nil.
type usersUpdateStub struct {
	driving.Users
	input   *domain.UpdateUserInput
	refusal error
}

func (s *usersUpdateStub) UpdateByID(_ context.Context, in *domain.UpdateUserInput) error {
	s.input = in

	return s.refusal
}

// recovererNeverAsked is the password recovery of a handler whose update
// sends no mail.
type recovererNeverAsked struct{ PasswordRecoverer }

func usersHandlerOver(t *testing.T, stub *usersUpdateStub) *UsersHandler {
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

func updateRequest(user uuid.UUID, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPut, "/users/"+user.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("user_id", user.String())

	return req
}

// A change of address is held to who asks, in the use-case. The caller has
// to reach it as the caller, beside the path's user: with the path's user in
// both places the rule would ask the account being changed whether it may
// change itself, and no use-case test could see it.
func TestAUsersUpdatePassesThePathsUserAndTheCaller(t *testing.T) {
	t.Parallel()

	stub := &usersUpdateStub{}
	h := usersHandlerOver(t, stub)

	user, caller := uuid.NewV7(), uuid.NewV7()

	rec := httptest.NewRecorder()
	h.updateByID(rec, withCaller(updateRequest(user, `{"email":"new@example.com"}`), caller))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if stub.input.ID != user {
		t.Errorf("ID = %s, want the path's user %s", stub.input.ID, user)
	}

	if stub.input.CallerID != caller {
		t.Errorf("CallerID = %s, want the caller %s", stub.input.CallerID, caller)
	}

	if stub.input.Email == nil || *stub.input.Email != "new@example.com" {
		t.Errorf("Email = %v, want the body's", stub.input.Email)
	}
}

// The refusal is a 403 in the service's words, and names no address.
func TestAnAddressChangeRefusedIsA403(t *testing.T) {
	t.Parallel()

	stub := &usersUpdateStub{refusal: &domain.AdministratorOnlyError{Entity: "user", Field: domain.FieldEmail}}
	h := usersHandlerOver(t, stub)

	rec := httptest.NewRecorder()
	h.updateByID(rec, withCaller(updateRequest(uuid.NewV7(), `{"email":"new@example.com"}`), uuid.NewV7()))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. Body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "can be changed only by the administrator account") {
		t.Errorf("the body does not say why: %s", rec.Body.String())
	}

	if strings.Contains(rec.Body.String(), "new@example.com") {
		t.Errorf("the address is in the answer: %s", rec.Body.String())
	}
}

// An update that does not know who asks is a wiring mistake, and is not
// handed to the use-case as an update by nobody.
func TestAUsersUpdateWithNoCallerIsNotHandedOn(t *testing.T) {
	t.Parallel()

	stub := &usersUpdateStub{}
	h := usersHandlerOver(t, stub)

	rec := httptest.NewRecorder()
	h.updateByID(rec, updateRequest(uuid.NewV7(), `{"email":"new@example.com"}`))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500. Body: %s", rec.Code, rec.Body.String())
	}

	if stub.input != nil {
		t.Errorf("the update reached the use-case with no caller: %+v", stub.input)
	}
}
