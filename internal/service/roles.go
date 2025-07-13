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

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/roles.go -source=roles.go RolesRepository

// RolesRepository is the interface for the roles repository methods.
type RolesRepository interface {
	Insert(ctx context.Context, input *model.InsertRoleInput) error
	UpdateByID(ctx context.Context, input *model.UpdateRoleInput) error
	DeleteByID(ctx context.Context, input *model.DeleteRoleInput) error
	SelectByID(ctx context.Context, id uuid.UUID) (*model.Role, error)

	Select(ctx context.Context, input *model.SelectRolesInput) (*model.SelectRolesOutput, error)
	SelectByUserID(ctx context.Context, userID uuid.UUID, input *model.SelectRolesInput) (*model.SelectRolesOutput, error)
	SelectByPolicyID(ctx context.Context, policyID uuid.UUID, input *model.SelectRolesInput) (*model.SelectRolesOutput, error)

	LinkPolicies(ctx context.Context, input *model.LinkPoliciesToRoleInput) error
	UnlinkPolicies(ctx context.Context, input *model.UnlinkPoliciesFromRoleInput) error

	LinkUsers(ctx context.Context, input *model.LinkUsersToRoleInput) error
	UnlinkUsers(ctx context.Context, input *model.UnlinkUsersFromRoleInput) error
}

type RolesServiceConf struct {
	Repository    RolesRepository
	CacheService  *CacheService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type rolesServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type RolesService struct {
	repository    RolesRepository
	cacheService  *CacheService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       rolesServiceMetrics
}

// NewRolesService creates a new RolesService.
func NewRolesService(conf RolesServiceConf) (*RolesService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for RolesService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for RolesService"}
	}

	service := &RolesService{
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
		metric.WithDescription("The number of calls to the roles service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

// GetByID returns the roles with the specified ID.
func (ref *RolesService) GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("roles.id", id.String()))

	if id == uuid.Nil {
		invalidErr := &model.InvalidRoleIDError{Message: "invalid role ID. It is nil"}
		return nil, o11y.RecordError(ctx, span, invalidErr, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.GetByID")
	}

	out, err := ref.repository.SelectByID(ctx, id)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.GetByID")
	}

	slog.Debug("service.Roles.GetByID", "role.id", out.ID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "role found successfully", attribute.String("role.id", out.ID.String()))

	return out, nil
}

// Create inserts a new roles into the database.
func (ref *RolesService) Create(ctx context.Context, input *model.CreateRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.Create")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.Create")
	}

	if input.ID == uuid.Nil {
		var err error
		input.ID, err = uuid.NewV7()
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.Create", "failed to generate role ID")
		}
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.Create")
	}

	if err := ref.repository.Insert(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.Create")
	}

	slog.Debug("service.Roles.Create", "name", input.Name)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "role created successfully",
		attribute.String("role.id", input.ID.String()),
		attribute.String("role.name", input.Name))

	return nil
}

// UpdateByID updates the roles with the specified ID.
func (ref *RolesService) UpdateByID(ctx context.Context, input *model.UpdateRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.UpdateByID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UpdateByID")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UpdateByID")
	}

	if err := ref.repository.UpdateByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UpdateByID")
	}

	slog.Debug("service.Roles.UpdateByID", "name", input.Name)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "role updated successfully",
		attribute.String("role.id", input.ID.String()))

	return nil
}

// DeleteByID deletes the roles with the specified ID.
func (ref *RolesService) DeleteByID(ctx context.Context, input *model.DeleteRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.DeleteByID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.DeleteByID")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.DeleteByID")
	}

	if err := ref.repository.DeleteByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.DeleteByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "role deleted successfully",
		attribute.String("role.id", input.ID.String()))

	return nil
}

// List returns a list of models.
func (ref *RolesService) List(ctx context.Context, input *model.ListRolesInput) (*model.ListRolesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.List")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.List")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.List")
	}

	out, err := ref.repository.Select(ctx, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.List")
	}

	slog.Debug("service.Roles.List", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "roles listed successfully",
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// LinkUsers links users to a user.
func (ref *RolesService) LinkUsers(ctx context.Context, input *model.LinkUsersToRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.LinkUsers")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.LinkUsers")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.LinkUsers")
	}

	span.SetAttributes(attribute.String("roles.id", input.RoleID.String()))

	if err := ref.repository.LinkUsers(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.LinkUsers")
	}

	// remove cache key for authz
	if ref.cacheService != nil {
		for _, userID := range input.UserIDs {
			slog.Debug("service.Roles.LinkUsers", "what", "removing cache", "id", fmt.Sprintf("authz:%s", userID.String()))
			ref.cacheService.Remove(ctx, fmt.Sprintf("authz:%s", userID.String()))
		}
	}

	slog.Debug("service.Roles.LinkUsers", "roles.id", input.RoleID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "users linked to role successfully",
		attribute.String("role.id", input.RoleID.String()))

	return nil
}

// UnlinkUsers unlinks users from a role.
func (ref *RolesService) UnlinkUsers(ctx context.Context, input *model.UnlinkUsersFromRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.UnlinkUsers")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UnlinkUsers")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UnlinkUsers")
	}

	span.SetAttributes(attribute.String("roles.id", input.RoleID.String()))

	if err := ref.repository.UnlinkUsers(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UnlinkUsers")
	}

	if ref.cacheService != nil {
		for _, userID := range input.UserIDs {
			slog.Debug("service.Roles.UnlinkUsers", "what", "removing cache", "id", fmt.Sprintf("authz:%s", userID.String()))
			ref.cacheService.Remove(ctx, fmt.Sprintf("authz:%s", userID.String()))
		}
	}

	slog.Debug("service.Roles.UnlinkUsers", "roles.id", input.RoleID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "users unlinked from role successfully",
		attribute.String("role.id", input.RoleID.String()))

	return nil
}

// LinkPolicies links permission to a role.
func (ref *RolesService) LinkPolicies(ctx context.Context, input *model.LinkPoliciesToRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.LinkPolicies")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.LinkPolicies")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.LinkPolicies")
	}

	span.SetAttributes(attribute.String("roles.id", input.RoleID.String()))

	if err := ref.repository.LinkPolicies(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.LinkPolicies")
	}

	slog.Debug("service.Roles.LinkPolicies", "roles.id", input.RoleID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "policies linked to role successfully",
		attribute.String("role.id", input.RoleID.String()))

	return nil
}

// UnlinkPolicies unlinks permission from a role.
func (ref *RolesService) UnlinkPolicies(ctx context.Context, input *model.UnlinkPoliciesFromRoleInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.UnlinkPolicies")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UnlinkPolicies")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UnlinkPolicies")
	}

	span.SetAttributes(attribute.String("roles.id", input.RoleID.String()))

	if err := ref.repository.UnlinkPolicies(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.UnlinkPolicies")
	}

	slog.Debug("service.Roles.UnlinkPolicies", "roles.id", input.RoleID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "policies unlinked from role successfully",
		attribute.String("role.id", input.RoleID.String()))

	return nil
}

// ListByUserID returns a list of roles for a user.
func (ref *RolesService) ListByUserID(ctx context.Context, userID uuid.UUID, input *model.ListRolesInput) (*model.ListRolesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.ListByUserID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.ListByUserID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.ListByUserID")
	}

	out, err := ref.repository.SelectByUserID(ctx, userID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.ListByUserID")
	}

	slog.Debug("service.Roles.ListByUserID", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "roles found by user ID",
		attribute.String("user.id", userID.String()),
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// ListByPolicyID returns a list of roles for a policy.
func (ref *RolesService) ListByPolicyID(ctx context.Context, policyID uuid.UUID, input *model.ListRolesInput) (*model.ListRolesOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Roles.ListByPolicyID")
	defer span.End()

	span.SetAttributes(attribute.String("policy.id", policyID.String()))

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.ListByPolicyID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.ListByPolicyID")
	}

	out, err := ref.repository.SelectByPolicyID(ctx, policyID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Roles.ListByPolicyID")
	}

	slog.Debug("service.Roles.ListByPolicyID", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "roles found by policy ID",
		attribute.String("policy.id", policyID.String()),
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *RolesService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
