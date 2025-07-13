package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/resources.go -source=resources.go ResourcesRepository

// ResourcesRepository is the interface for the resources repository methods.
type ResourcesRepository interface {
	SelectByID(ctx context.Context, id uuid.UUID) (*model.Resource, error)
	Select(ctx context.Context, input *model.SelectResourcesInput) (*model.SelectResourcesOutput, error)
}

type ResourcesServiceConf struct {
	Repository    ResourcesRepository
	CacheService  *CacheService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type resourcesServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type ResourcesService struct {
	repository    ResourcesRepository
	cacheService  *CacheService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       resourcesServiceMetrics
}

// NewResourcesService creates a new ResourcesService.
func NewResourcesService(conf ResourcesServiceConf) (*ResourcesService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for ResourcesService"}
	}

	if conf.CacheService == nil {
		return nil, &model.InvalidCacheServiceError{Message: "CacheService is nil, but it is required for ResourcesService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for ResourcesService"}
	}

	service := &ResourcesService{
		repository:   conf.Repository,
		cacheService: conf.CacheService,
		ot:           conf.OT,
	}

	if conf.MetricsPrefix != "" {
		service.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		service.metricsPrefix += "_"
	}

	serviceCalls, err := service.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", service.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the resources service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

// GetByID returns the Resources with the specified ID.
func (ref *ResourcesService) GetByID(ctx context.Context, id uuid.UUID) (*model.Resource, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Resources.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("Resources.id", id.String()),
	)

	if id == uuid.Nil {
		errorType := &model.InvalidResourceIDError{ID: id, Message: "Resource ID cannot be empty"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.GetByID")
	}

	var out *model.Resource
	var err error
	if ref.cacheService.cache == nil {
		slog.Debug("service.Resources.GetByID", "cache", "disabled")

		out, err = ref.repository.SelectByID(ctx, id)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.GetByID")
		}
	} else {
		slog.Debug("service.Resources.GetByID", "cache", "enabled")

		cacheKey := fmt.Sprintf("resource:%s", id.String())
		ctxCache, cancel := context.WithTimeout(ctx, ref.cacheService.queryTimeout)
		defer cancel()

		out, err = FromCacheOrDB(
			ctxCache,
			ref.cacheService.cache,
			cacheKey,
			CacheEncoderTypeGob,
			func() (*model.Resource, error) {
				return ref.repository.SelectByID(ctx, id)
			},
			ref.cacheService.queryTimeout,
			ref.cacheService.entitiesTTL,
		)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.GetByID")
		}
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Resources found", attribute.String("resources.method", out.Action))

	return out, nil
}

// List returns a list of resources.
func (ref *ResourcesService) List(ctx context.Context, input *model.ListResourcesInput) (*model.ListResourcesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Resources.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	var out *model.ListResourcesOutput
	var err error
	if ref.cacheService.cache == nil {
		slog.Debug("service.Resources.List", "cache", "disabled")

		out, err = ref.repository.Select(ctx, input)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.List")
		}
	} else {
		slog.Debug("service.Resources.List", "cache", "enabled")

		cacheKey := fmt.Sprintf("resources:%s", input.UniqueID())
		slog.Debug("service.Resources.List", "cacheKey", cacheKey)

		ctxCache, cancel := context.WithTimeout(ctx, ref.cacheService.queryTimeout)
		defer cancel()

		out, err = FromCacheOrDB(
			ctxCache,
			ref.cacheService.cache,
			cacheKey,
			CacheEncoderTypeGob,
			func() (*model.ListResourcesOutput, error) {
				return ref.repository.Select(ctx, input)
			},
			ref.cacheService.queryTimeout,
			ref.cacheService.entitiesTTL,
		)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.List")
		}
	}

	slog.Debug("service.Resources.List", "resources", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Resources found")

	return out, nil
}

// ListMatches returns a list of policies that match the given action and resource.
func (ref *ResourcesService) ListMatches(ctx context.Context, action, resource string, input *model.ListResourcesInput) (*model.ListResourcesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Resources.ListMatches")
	defer span.End()

	act, err := model.ValidateAction(action)
	if err != nil {
		errType := &model.InvalidActionError{Action: action}
		return nil, o11y.RecordError(ctx, span, errType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.ListMatches")
	}

	res, err := model.ValidateResource(resource)
	if err != nil {
		errType := &model.InvalidResourceError{Resource: resource, Message: "resource  cannot be empty"}
		return nil, o11y.RecordError(ctx, span, errType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Resources.ListMatches")
	}

	var out *model.SelectResourcesOutput
	switch {
	case act == "*" && res == "*":
		out, err = ref.repository.Select(ctx, &model.SelectResourcesInput{
			Filter:    fmt.Sprintf("action = '%s' AND resource = '%s'", act, res),
			Paginator: input.Paginator,
		})
		if err != nil || out == nil || len(out.Items) == 0 {
			return nil, &model.ResourceNotFoundError{
				Message: fmt.Sprintf("does not found exist any resource with action = '%s' and resource as '%s'", action, resource),
			}
		}

	case act == "*" && res != "*" && res != "" && res != "/":

		resourceWithWildcard := convertToSQLRegex(res)

		out, err = ref.repository.Select(ctx, &model.SelectResourcesInput{
			Filter:    fmt.Sprintf("resource ~ '%s'", resourceWithWildcard),
			Paginator: input.Paginator,
		})
		if err != nil || out == nil || len(out.Items) == 0 {
			return nil, &model.ResourceNotFoundError{
				Message: fmt.Sprintf("does not found exist any resource with action = '%s' and resource as '%s'", action, resource),
			}
		}

	case act != "*" && res != "*" && res != "" && res != "/":

		resourceWithWildcard := convertToSQLRegex(res)

		out, err = ref.repository.Select(ctx, &model.SelectResourcesInput{
			Filter:    fmt.Sprintf("action = '%s' AND resource ~ '%s'", act, resourceWithWildcard),
			Paginator: input.Paginator,
		})
		if err != nil || out == nil || len(out.Items) == 0 {
			return nil, &model.ResourceNotFoundError{
				Message: fmt.Sprintf("does not found exist any resource with action = '%s' and resource as '%s'", action, resource),
			}
		}

	case act != "*" && res == "*":
		out, err = ref.repository.Select(ctx, &model.SelectResourcesInput{
			Filter:    fmt.Sprintf("action = '%s' AND resource = '%s'", act, res),
			Paginator: input.Paginator,
		})
		if err != nil || out == nil || len(out.Items) == 0 {
			return nil, &model.ResourceNotFoundError{
				Message: fmt.Sprintf("does not found exist any resource with action = '%s' and resource as '%s'", action, resource),
			}
		}

	default:
		return nil, &model.ResourceNotFoundError{
			Message: fmt.Sprintf("does not found exist any resource with action = '%s' and resource as '%s'", action, resource),
		}
	}

	slog.Debug("service.Resources.ListMatches")
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Resources matched")

	return out, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *ResourcesService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
