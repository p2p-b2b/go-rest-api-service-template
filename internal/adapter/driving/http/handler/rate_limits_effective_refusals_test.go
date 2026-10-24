//go:build unit

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/respond"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// TestEffectiveRateLimitsRecordsItsRefusals: GET /rate_limits/effective
// refuses a request with no method, with a method that is none, and with no
// endpoint. Each wrote its 400 and returned without recording anything: the
// call was in no count of the handler layer, refused or otherwise, and the
// access line of the 400 had no error_type.
//
// Each must be recorded, as an error the status table answers with the 400
// that was written, so that the status and the outcome agree -- and with the
// body it has always had.
func TestEffectiveRateLimitsRecordsItsRefusals(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		query string
		body  string
		field string
	}{
		{
			name:  "no_method",
			query: "endpoint=/projects/{project_id}/generate",
			body:  "method is required, and must be one of " + domain.GetValidActions(),
			field: "method",
		},
		{
			name:  "a_method_that_is_none",
			query: "method=FETCH&endpoint=/projects/{project_id}/generate",
			body:  "invalid action: FETCH, must be one of " + domain.GetValidActions() + " in Uppercase",
			field: "method",
		},
		{
			name:  "no_endpoint",
			query: "method=GET",
			body:  "endpoint is required, and must be a route template such as /projects/{project_id}/generate",
			field: "endpoint",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// No port: a refused query never reaches it.
			h := &RateLimitsHandler{}
			h.ot, h.metrics = testTelemetry(t)

			req := httptest.NewRequest(http.MethodGet, "/rate_limits/effective?"+tt.query, nil)
			req = req.WithContext(o11y.WithFailure(req.Context()))

			rec := httptest.NewRecorder()
			h.effective(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400. Body: %s", rec.Code, rec.Body.String())
			}

			var answer struct {
				Message string `json:"message"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &answer); err != nil {
				t.Fatalf("the body is not the service's error shape: %v. Body: %s", err, rec.Body.String())
			}

			if answer.Message != tt.body {
				t.Errorf("message = %q, want it unchanged: %q", answer.Message, tt.body)
			}

			recorded, ok := o11y.TakeFailure(req.Context())
			if !ok {
				t.Fatal("nothing was recorded for a 400")
			}

			refusal, isIt := errors.AsType[*domain.InvalidInputError](recorded.Err)
			if !isIt {
				t.Fatalf("recorded %T (%v), want an invalid input", recorded.Err, recorded.Err)
			}

			if refusal.Field != tt.field {
				t.Errorf("the refusal names %q, want %q", refusal.Field, tt.field)
			}

			// The status written and the outcome counted come from the same
			// table: a refusal, and the 400 the handler wrote.
			if !respond.IsRefusal(recorded.Err) {
				t.Error("the 400 is counted as a failure of the service")
			}

			if got := respond.StatusFor(recorded.Err); got != rec.Code {
				t.Errorf("the status table answers the recorded error %d, and the handler wrote %d", got, rec.Code)
			}
		})
	}
}
