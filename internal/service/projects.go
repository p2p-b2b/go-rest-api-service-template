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

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/projects.go -source=projects.go ProjectsRepository

// ProjectsRepository is the interface for the projects repository methods.
type ProjectsRepository interface {
	Insert(ctx context.Context, input *model.InsertProjectInput) error
	UpdateByID(ctx context.Context, input *model.UpdateProjectInput) error
	DeleteByID(ctx context.Context, input *model.DeleteProjectInput) error
	SelectByID(ctx context.Context, id, userID uuid.UUID) (*model.Project, error)
	Select(ctx context.Context, input *model.SelectProjectsInput) (*model.SelectProjectsOutput, error)
}

type ProjectsServiceConf struct {
	Repository    ProjectsRepository
	CacheService  *CacheService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type projectServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type ProjectsService struct {
	repository    ProjectsRepository
	cacheService  *CacheService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       projectServiceMetrics
}

// NewProjectsService creates a new ProjectsService.
func NewProjectsService(conf ProjectsServiceConf) (*ProjectsService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for ProjectsService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for ProjectsService"}
	}

	service := &ProjectsService{
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
		metric.WithDescription("The number of calls to the projects service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

// GetByID returns the projects with the specified ID.
func (ref *ProjectsService) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Project, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Projects.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("projects.id", id.String()))

	if id == uuid.Nil {
		errorType := &model.InvalidProjectIDError{Message: "project ID is nil"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.GetByID")
	}

	out, err := ref.repository.SelectByID(ctx, id, userID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.GetByID")
	}

	slog.Debug("service.Projects.GetByID", "project.id", out.ID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "project found successfully", attribute.String("project.id", out.ID.String()))

	return out, nil
}

// Create inserts a new projects into the database.
func (ref *ProjectsService) Create(ctx context.Context, input *model.CreateProjectInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Projects.Create")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.Create")
	}

	span.SetAttributes(attribute.String("projects.name", input.Name))

	if input.ID == uuid.Nil {
		input.ID = uuid.Must(uuid.NewV7())
	}

	// validate the projects input
	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.Create")
	}

	if err := ref.repository.Insert(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.Create")
	}

	// remove cache key for authz
	if ref.cacheService != nil {
		slog.Debug("service.Projects.Create", "what", "removing cache", "id", fmt.Sprintf("authz:%s", input.UserID.String()))
		ref.cacheService.Remove(ctx, fmt.Sprintf("authz:%s", input.UserID.String()))
	}

	slog.Debug("service.Projects.Create", "name", input.Name)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "project created successfully",
		attribute.String("project.id", input.ID.String()),
		attribute.String("project.name", input.Name))

	return nil
}

// UpdateByID updates the projects with the specified ID.
func (ref *ProjectsService) UpdateByID(ctx context.Context, input *model.UpdateProjectInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Projects.UpdateByID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.UpdateByID")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.UpdateByID")
	}

	span.SetAttributes(attribute.String("projects.id", input.ID.String()))

	if err := ref.repository.UpdateByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.UpdateByID")
	}

	slog.Debug("service.Projects.UpdateByID", "name", input.Name)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "project updated successfully",
		attribute.String("project.id", input.ID.String()))

	return nil
}

// DeleteByID deletes the projects with the specified ID.
func (ref *ProjectsService) DeleteByID(ctx context.Context, input *model.DeleteProjectInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Projects.DeleteByID")
	defer span.End()

	span.SetAttributes(attribute.String("projects.id", input.ID.String()))

	if input.ID == uuid.Nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.DeleteByID")
	}

	if err := ref.repository.DeleteByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.DeleteByID")
	}

	// TODO: remove cache key for authz, userIDs are needed
	// remove cache key for authz
	// if ref.cacheService != nil {
	//  slog.Debug("service.Projects.DeleteByID", "what", "removing cache", "id", fmt.Sprintf("authz:%s", input.UserID.String()))
	//   ref.cacheService.Remove(ctx, fmt.Sprintf("authz:%s", input.UserID.String()))
	// }

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "project deleted successfully",
		attribute.String("project.id", input.ID.String()))

	return nil
}

// List returns a list of models.
func (ref *ProjectsService) List(ctx context.Context, input *model.ListProjectsInput) (*model.ListProjectsOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Projects.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	out, err := ref.repository.Select(ctx, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Projects.List")
	}

	slog.Debug("service.Projects.List", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "projects listed successfully",
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *ProjectsService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
