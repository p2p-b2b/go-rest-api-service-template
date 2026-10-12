//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driving"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// usersCreateStub is the users port for the create, and records what it was
// asked. Any other method panics: the embedded port is nil.
type usersCreateStub struct {
	driving.Users
	input *domain.CreateUserInput
}

func (s *usersCreateStub) Create(_ context.Context, in *domain.CreateUserInput) error {
	s.input = in

	return nil
}

// An account made through POST /users is the administrator's to vouch for:
// it is stored verified, so the verification link and the recovery's variant
// of it do not apply to it; disabled, as before, for an administrator to
// enable; and marked as disabled by an administrator, so that no link opens
// it.
func TestAnAccountAnAdministratorCreatesIsStoredVerifiedAndDisabled(t *testing.T) {
	t.Parallel()

	stub := &usersCreateStub{}

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	h, err := NewUsersHandler(UsersHandlerConf{Service: stub, Recoverer: recovererNeverAsked{}, OT: ot})
	if err != nil {
		t.Fatalf("NewUsersHandler: %v", err)
	}

	body := `{"first_name":"Jane","last_name":"Doe","email":"jane@example.com","password":"Meadow7Lark!"}`

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	in := stub.input

	if in.Verified == nil || !*in.Verified {
		t.Errorf("Verified = %v, want true: unverified, the public verification link applies to it", in.Verified)
	}

	if in.Disabled == nil || !*in.Disabled {
		t.Errorf("Disabled = %v, want true: no mail is sent for it, an administrator enables it", in.Disabled)
	}

	// And the row says whose doing its `disabled` is: with the mark, no
	// link opens it, whatever becomes of `verified`.
	if !in.DisabledByAdmin {
		t.Error("DisabledByAdmin = false: the account is stored disabled as a registration waiting for its link is")
	}

	if in.LocalAccount == nil || !*in.LocalAccount {
		t.Errorf("LocalAccount = %v, want true", in.LocalAccount)
	}
}
