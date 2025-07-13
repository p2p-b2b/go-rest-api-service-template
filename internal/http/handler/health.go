package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/health.go -source=health.go HealthService

// HealthService represents the service for the health.
type HealthService interface {
	HealthCheck(ctx context.Context) (model.Health, error)
}

// HealthHandlerConf represents the configuration for the HealthHandler.
type HealthHandlerConf struct {
	Service       HealthService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type healthHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// HealthHandler represents the handler for the health.
type HealthHandler struct {
	service       HealthService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       healthHandlerMetrics
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(conf HealthHandlerConf) (*HealthHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "HealthService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	handler := &HealthHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the health handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *HealthHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /health/status", mdw.ThenFunc(ref.getStatus))
}

// getStatus Get the health of the health service
//
//	@ID				0198042a-f9c5-76be-ba9e-8186a69f48c4
//	@Summary		Check health
//	@Description	Check service health status including database connectivity and system metrics
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	model.Health
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/health/status [get]
func (ref *HealthHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Health.getStatus")
	defer span.End()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := ref.service.HealthCheck(ctxWithTimeout)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Health.getStatus")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Health.getStatus")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "health status checked")
}
