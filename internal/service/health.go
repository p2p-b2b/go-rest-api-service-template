package service

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/health.go -source=health.go HealthRepository

// HealthRepository is the interface for the model repository methods.
type HealthRepository interface {
	DriverName() string
	PingContext(ctx context.Context) error
}

type HealthServiceConf struct {
	Repository    HealthRepository
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type healthServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type HealthService struct {
	repository    HealthRepository
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       healthServiceMetrics
}

// NewHealthService creates a new HealthService.
func NewHealthService(conf HealthServiceConf) (*HealthService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for HealthService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for HealthService"}
	}

	service := &HealthService{
		repository: conf.Repository,
		ot:         conf.OT,
	}
	if conf.MetricsPrefix != "" {
		service.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		service.metricsPrefix += "_"
	}

	serviceCalls, err := service.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", service.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the model service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

// HealthCheck verifies a connection to the repository is still alive.
func (ref *HealthService) HealthCheck(ctx context.Context) (model.Health, error) {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Health.HealthCheck")
	defer span.End()

	// database
	dbStatus := model.StatusUp
	err := ref.repository.PingContext(ctx)
	if err != nil {
		dbStatus = model.StatusDown
	}

	database := model.Check{
		Name:   "database",
		Kind:   ref.repository.DriverName(),
		Status: dbStatus,
	}

	// runtime
	rtStatus := model.StatusUp
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	rt := model.Check{
		Name:   "runtime",
		Kind:   "go",
		Status: rtStatus,
		Data: map[string]any{
			"version":      runtime.Version(),
			"numCPU":       runtime.NumCPU(),
			"numGoroutine": runtime.NumGoroutine(),
			"numCgoCall":   runtime.NumCgoCall(),
			"memStats":     mem,
		},
	}

	// and operator
	allStatus := dbStatus && rtStatus

	health := model.Health{
		Status: allStatus,
		Checks: []model.Check{
			database,
			rt,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "health check successful")

	return health, err
}

// Helper methods for consistent telemetry and error handling

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *HealthService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
