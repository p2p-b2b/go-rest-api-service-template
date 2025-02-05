package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"go.opentelemetry.io/otel/metric"
)

var (
	ErrInvalidService       = errors.New("invalid service")
	ErrInvalidOpenTelemetry = errors.New("invalid open telemetry")
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -package=mocks -destination=../../../mocks/handler/health.go -source=health.go HealthService

// HealthService represents the service for the health.
type HealthService interface {
	HealthCheck(ctx context.Context) (service.Health, error)
}

// HealthHandler represents the handler for the health.
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
		slog.Error("service is required")
		return nil, ErrInvalidService
	}

	if conf.OT == nil {
		slog.Error("open telemetry is required")
		return nil, ErrInvalidOpenTelemetry
	}

	uh := &HealthHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		uh.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		uh.metricsPrefix += "_"
	}

	handlerCalls, err := uh.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", uh.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the health handler"),
	)
	if err != nil {
		slog.Error("handler.Health.registerMetrics", "error", err)
		return nil, err
	}
	uh.metrics.handlerCalls = handlerCalls

	return uh, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *HealthHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /health/status", mdw.ThenFunc(ref.getStatus))
}

// getStatus Get the health of the health service
//
//	@ID				0986a6ff-aa83-4b06-9a16-7e338eaa50d1
//	@Summary		Check health status
//	@Description	Check health status of the service pinging the database and go metrics
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	Health
//	@Failure		500	{object}	respond.HTTPMessage
//	@Router			/health/status [get]
func (ref *HealthHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	sHealth, err := ref.service.HealthCheck(ctx)
	if err != nil {
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	health := &Health{
		Status: sHealth.Status.String(),
		Checks: make([]Check, len(sHealth.Checks)),
	}

	for i, sCheck := range sHealth.Checks {
		health.Checks[i] = Check{
			Name:   sCheck.Name,
			Kind:   sCheck.Kind,
			Status: sCheck.Status.String(),
			Data:   sCheck.Data,
		}
	}

	if err := respond.WriteJSONData(w, http.StatusOK, health); err != nil {
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}
}
