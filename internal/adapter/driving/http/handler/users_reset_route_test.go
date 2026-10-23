//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// recovererAnswering is the recovery port answering with err.
type recovererAnswering struct{ err error }

func (r recovererAnswering) RecoverPasswordOf(context.Context, uuid.UUID) error { return r.err }

// POST /users/{user_id}/password/reset answers what its annotations say for
// each refusal of the recovery: the account's absence is a 404, a stored
// address the recovery's rule refuses is a 400 (it was a 500: the rule's
// error is the singular ValidationError, which the route did not match),
// and a spent budget is a 429 that says when to come back, as the public
// route's does (it said nothing).
func TestAnAdministratorsResetAnswersEachRefusalAsDocumented(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err        error
		name       string
		retryAfter string
		status     int
	}{
		{name: "the_account_is_not_there", err: &domain.UserNotFoundError{ID: uuid.NewV7()}, status: http.StatusNotFound},
		{name: "the_stored_address_is_refused", err: &domain.ValidationError{Field: domain.FieldEmail, Message: "not an address"}, status: http.StatusBadRequest},
		{name: "the_budget_is_spent", err: &domain.TooManyRecoveryRequestsError{RetryAfter: 90 * time.Second}, status: http.StatusTooManyRequests, retryAfter: "90"},
		{name: "anything_else", err: context.DeadlineExceeded, status: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ot := &o11y.OpenTelemetry{
				Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
				Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
			}

			h, err := NewUsersHandler(UsersHandlerConf{Service: &usersCreateStub{}, Recoverer: recovererAnswering{err: tc.err}, OT: ot})
			if err != nil {
				t.Fatalf("NewUsersHandler: %v", err)
			}

			id := uuid.NewV7()
			req := httptest.NewRequest(http.MethodPost, "/users/"+id.String()+"/password/reset", nil)
			req.SetPathValue("user_id", id.String())

			rec := httptest.NewRecorder()
			h.resetPassword(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body.String())
			}

			if got := rec.Header().Get("Retry-After"); got != tc.retryAfter {
				t.Errorf("Retry-After = %q, want %q", got, tc.retryAfter)
			}
		})
	}
}
