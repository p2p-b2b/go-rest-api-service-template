package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/policies.go -source=policies.go PoliciesRepository

type ResourcesServiceMethods interface {
	ListMatches(ctx context.Context, action, resource string, input *model.ListResourcesInput) (*model.SelectResourcesOutput, error)
}

// PoliciesRepository is the interface for the policies repository methods.
type PoliciesRepository interface {
	Insert(ctx context.Context, input *model.CreatePolicyInput) error
	UpdateByID(ctx context.Context, input *model.UpdatePolicyInput) error
	DeleteByID(ctx context.Context, input *model.DeletePolicyInput) error

	Select(ctx context.Context, input *model.SelectPoliciesInput) (*model.SelectPoliciesOutput, error)
	SelectByID(ctx context.Context, id uuid.UUID) (*model.Policy, error)
	SelectByRoleID(ctx context.Context, roleID uuid.UUID, input *model.SelectPoliciesInput) (*model.SelectPoliciesOutput, error)

	LinkRoles(ctx context.Context, input *model.LinkRolesToPolicyInput) error
	UnlinkRoles(ctx context.Context, input *model.UnlinkRolesFromPolicyInput) error
}

type PoliciesServiceConf struct {
	Repository       PoliciesRepository
	ResourcesService ResourcesServiceMethods
	CacheService     *CacheService
	OT               *o11y.OpenTelemetry
	MetricsPrefix    string
}

type policiesServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type PoliciesService struct {
	repository       PoliciesRepository
	resourcesService ResourcesServiceMethods
	cacheService     *CacheService
	ot               *o11y.OpenTelemetry
	metricsPrefix    string
	metrics          policiesServiceMetrics
}

// NewPoliciesService creates a new PoliciesService.
func NewPoliciesService(conf PoliciesServiceConf) (*PoliciesService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for PoliciesService"}
	}

	if conf.ResourcesService == nil {
		return nil, &model.InvalidRepositoryError{Message: "ResourcesService is nil, but it is required for PoliciesService"}
	}

	if conf.CacheService == nil {
		return nil, &model.InvalidCacheServiceError{Message: "CacheService is nil, but it is required for PoliciesService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for PoliciesService"}
	}

	service := &PoliciesService{
		repository:       conf.Repository,
		resourcesService: conf.ResourcesService,
		cacheService:     conf.CacheService,
		ot:               conf.OT,
	}

	if conf.MetricsPrefix != "" {
		service.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		service.metricsPrefix += "_"
	}

	serviceCalls, err := service.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", service.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the policies service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

// Create creates a new policy.
func (ref *PoliciesService) Create(ctx context.Context, input *model.CreatePolicyInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.Create")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.Create")
	}

	if input.ID == uuid.Nil {
		var err error
		input.ID, err = uuid.NewV7()
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.Create", "failed to generate policy ID")
		}
	}

	if input.ResourceID == uuid.Nil {
		resources, err := ref.resourcesService.ListMatches(
			ctx,
			input.AllowedAction,
			input.AllowedResource,
			&model.ListResourcesInput{
				Sort: "resource ASC",
				Paginator: model.Paginator{
					Limit: 1,
				},
			},
		)
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.Create")
		}

		if resources == nil || len(resources.Items) == 0 {
			return &model.ResourceNotFoundError{
				Message: fmt.Sprintf("does not found any resource with action = '%s' and resource as '%s'", input.AllowedAction, input.AllowedResource),
			}
		}

		if len(resources.Items) > 1 {
			return &model.InvalidResourceIDError{
				Message: fmt.Sprintf("there are more than one resource with action = '%s' and resource as '%s'", input.AllowedAction, input.AllowedResource),
			}
		}

		input.ResourceID = resources.Items[0].ID
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.Create")
	}

	if err := ref.repository.Insert(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.Create")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Policy created", attribute.String("policy_id", input.ID.String()), attribute.String("policy.name", input.Name))
	return nil
}

// DeleteByID deletes a policy by ID.
func (ref *PoliciesService) DeleteByID(ctx context.Context, input *model.DeletePolicyInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.DeleteByID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.DeleteByID")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.DeleteByID")
	}

	if err := ref.repository.DeleteByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.DeleteByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Policy deleted", attribute.String("policy_id", input.ID.String()))
	return nil
}

// UpdateByID updates a policy by ID.
func (ref *PoliciesService) UpdateByID(ctx context.Context, input *model.UpdatePolicyInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.UpdateByID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.UpdateByID")
	}

	// TODO: the allowed action and resource must be validated by
	// comparing this with the resources table items

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.UpdateByID")
	}

	if err := ref.repository.UpdateByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.UpdateByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Policy updated", attribute.String("policy_id", input.ID.String()))
	return nil
}

// GetByID returns a policy by ID.
func (ref *PoliciesService) GetByID(ctx context.Context, id uuid.UUID) (*model.Policy, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.GetByID")
	defer span.End()

	if id == uuid.Nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.GetByID")
	}

	out, err := ref.repository.SelectByID(ctx, id)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.GetByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Policy retrieved", attribute.String("policy_id", id.String()))
	return out, nil
}

// List returns a list of policies.
func (ref *PoliciesService) List(ctx context.Context, input *model.SelectPoliciesInput) (*model.SelectPoliciesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.List")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.List")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.List")
	}

	policies, err := ref.repository.Select(ctx, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.List")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Policies listed")
	return policies, nil
}

// LinkRoles links roles to a permission.
func (ref *PoliciesService) LinkRoles(ctx context.Context, input *model.LinkRolesToPolicyInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.LinkRoles")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.LinkRoles")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.LinkRoles")
	}

	if err := ref.repository.LinkRoles(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.LinkRoles")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Roles linked to permission", attribute.String("policy_id", input.PolicyID.String()))
	return nil
}

// UnlinkRoles unlinks roles from a permission.
func (ref *PoliciesService) UnlinkRoles(ctx context.Context, input *model.UnlinkRolesFromPolicyInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.UnlinkRoles")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.UnlinkRoles")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.UnlinkRoles")
	}

	if err := ref.repository.UnlinkRoles(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.UnlinkRoles")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Roles unlinked from permission", attribute.String("policy_id", input.PolicyID.String()))
	return nil
}

// ListByRoleID returns a list of policies by role ID.
func (ref *PoliciesService) ListByRoleID(ctx context.Context, roleID uuid.UUID, input *model.ListPoliciesInput) (*model.ListPoliciesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Policies.ListByRoleID")
	defer span.End()

	if roleID == uuid.Nil {
		errorValue := &model.InvalidRoleIDError{Message: "roleID is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.ListByRoleID")
	}

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.ListByRoleID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.ListByRoleID")
	}

	out, err := ref.repository.SelectByRoleID(ctx, roleID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Policies.ListByRoleID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Policies listed by role", attribute.String("role_id", roleID.String()))
	return out, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *PoliciesService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
