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

// productsListStub is the products port for the bare list, and records what
// it was asked. Any other method panics: the embedded port is nil.
type productsListStub struct {
	driving.Products
	input *domain.ListProductsInput
}

func (s *productsListStub) List(_ context.Context, in *domain.ListProductsInput) (*domain.ListProductsOutput, error) {
	s.input = in

	return &domain.ListProductsOutput{}, nil
}

// usageListStub is the resources-limits port for its list, and records what
// it was asked.
type usageListStub struct {
	driving.ResourcesLimits
	input *domain.ListResourcesLimitsInput
}

func (s *usageListStub) List(_ context.Context, in *domain.ListResourcesLimitsInput) (*domain.ListResourcesLimitsOutput, error) {
	s.input = in

	return &domain.ListResourcesLimitsOutput{}, nil
}

// GET /products and GET /resources_limits are held to who asks, in the
// repository, so the caller has to reach it: a list handed on with no caller
// is refused there, and one handed on with another account is that
// account's list.
func TestTheBareListsHandTheCallerOn(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	t.Run("products", func(t *testing.T) {
		t.Parallel()

		stub := &productsListStub{}

		h, err := NewProductsHandler(ProductsHandlerConf{Service: stub, OT: ot})
		if err != nil {
			t.Fatalf("NewProductsHandler: %v", err)
		}

		caller := uuid.NewV7()
		rec := httptest.NewRecorder()
		h.list(rec, withCaller(httptest.NewRequest(http.MethodGet, "/products", nil), caller))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}

		if stub.input == nil || stub.input.CallerID != caller {
			t.Errorf("the list was handed on as %+v, want the caller %s", stub.input, caller)
		}
	})

	t.Run("resources_limits", func(t *testing.T) {
		t.Parallel()

		stub := &usageListStub{}

		h, err := NewResourcesLimitsHandler(ResourcesLimitsHandlerConf{Service: stub, OT: ot})
		if err != nil {
			t.Fatalf("NewResourcesLimitsHandler: %v", err)
		}

		caller := uuid.NewV7()
		rec := httptest.NewRecorder()
		h.list(rec, withCaller(httptest.NewRequest(http.MethodGet, "/resources_limits", nil), caller))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}

		if stub.input == nil || stub.input.CallerID != caller {
			t.Errorf("the list was handed on as %+v, want the caller %s", stub.input, caller)
		}
	})
}
