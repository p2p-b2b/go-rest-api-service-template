//go:build unit

package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// A first sign-in stores its account and its identity in one transaction,
// and what refuses that transaction reaches the callback: the identity
// linked meanwhile (409), the provider deleted meanwhile (the identity's
// foreign key, 404), the users limit reached (409, as POST /users answers
// it). The last two answered 500.
func TestAFirstSignInsRefusalsAreAnsweredAsDocumented(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	metrics, err := o11y.NewLayerMetrics(ot.Metrics.Meter, "")
	if err != nil {
		t.Fatalf("NewLayerMetrics: %v", err)
	}

	h := &AuthnIDPsHandler{metrics: metrics}

	for name, c := range map[string]struct {
		err    error
		status int
	}{
		"the_identity_was_linked_meanwhile": {err: &domain.UserIdentityAlreadyLinkedError{}, status: http.StatusConflict},
		"the_provider_was_deleted_meanwhile": {
			err:    &domain.UserIdentityNotFoundError{Message: "the account or the provider does not exist"},
			status: http.StatusNotFound,
		},
		"the_users_limit_is_reached": {
			err:    &domain.ResourcesLimitsHardLimitReachedError{},
			status: http.StatusConflict,
		},
		"the_address_was_taken_meanwhile": {err: &domain.IDPIdentityNotLinkedError{}, status: http.StatusUnauthorized},
		"a_fault":                         {err: errors.New("a fault"), status: http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, span := ot.Traces.Tracer.Start(ctx, "test")
			defer span.End()

			rec := httptest.NewRecorder()
			h.writeCallbackError(rec, httptest.NewRequest(http.MethodGet, "/auth/idp/x/callback", nil), span, time.Now(), nil, c.err)

			if rec.Code != c.status {
				t.Errorf("status = %d, want %d. Body: %s", rec.Code, c.status, rec.Body.String())
			}
		})
	}
}
