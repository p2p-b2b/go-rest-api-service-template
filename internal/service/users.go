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

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/users.go -source=users.go UsersRepository

type UsersRepository interface {
	Insert(ctx context.Context, input *model.InsertUserInput) error
	UpdateByID(ctx context.Context, input *model.UpdateUserInput) error
	DeleteByID(ctx context.Context, input *model.DeleteUserInput) error

	SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	SelectByEmail(ctx context.Context, email string) (*model.User, error)
	SelectByRoleID(ctx context.Context, roleID uuid.UUID, input *model.ListUsersInput) (*model.ListUsersOutput, error)
	Select(ctx context.Context, input *model.SelectUsersInput) (*model.SelectUsersOutput, error)
	SelectAuthz(ctx context.Context, userID uuid.UUID) (map[string]any, error)

	LinkRoles(ctx context.Context, input *model.LinkRolesToUserInput) error
	UnLinkRoles(ctx context.Context, input *model.UnLinkRolesFromUsersInput) error
}

type UsersServiceConf struct {
	Repository    UsersRepository
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type usersServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type UsersService struct {
	repository    UsersRepository
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       usersServiceMetrics
}

func NewUsersService(conf UsersServiceConf) (*UsersService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for UsersService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for UsersService"}
	}

	service := &UsersService{
		repository: conf.Repository,
		ot:         conf.OT,
	}

	if conf.MetricsPrefix != "" {
		service.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		service.metricsPrefix += "_"
	}

	serviceCalls, err := service.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", service.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the user service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

func (ref *UsersService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	if id == uuid.Nil {
		errorType := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByID")
	}

	slog.Debug("service.Users.GetByID", "id", id)
	out, err := ref.repository.SelectByID(ctx, id)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByID")
	}

	slog.Debug("service.Users.GetByID", "email", out.Email)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user found", attribute.String("user.email", out.Email))

	return out, nil
}

func (ref *UsersService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.GetByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	if email == "" {
		errorType := &model.InvalidEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByEmail")
	}

	slog.Debug("service.Users.GetByEmail", "email", email)
	out, err := ref.repository.SelectByEmail(ctx, email)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByEmail")
	}

	slog.Debug("service.Users.GetByEmail", "email", out.Email)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user found", attribute.String("user.email", out.Email))

	return out, nil
}

func (ref *UsersService) Create(ctx context.Context, input *model.CreateUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.Create")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Create")
	}

	span.SetAttributes(attribute.String("user.email", input.Email))

	if input.ID == uuid.Nil {
		var err error
		input.ID, err = uuid.NewV7()
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Create", "failed to generate user ID")
		}
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Create")
	}

	hashPwd, err := HashAndSaltPassword(input.Password)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Create", "failed to hash password")
	}

	rParams := &model.InsertUserInput{
		ID:           input.ID,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PasswordHash: hashPwd,
	}

	if err := ref.repository.Insert(ctx, rParams); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Create")
	}

	slog.Debug("service.Users.Create", "email", input.Email)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user created successfully",
		attribute.String("user.email", input.Email),
		attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersService) UpdateByID(ctx context.Context, input *model.UpdateUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.UpdateByID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UpdateByID")
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UpdateByID")
	}

	rParams := &model.UpdateUserInput{
		ID:        input.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Disabled:  input.Disabled,
	}

	if input.Password != nil && len(*input.Password) < model.ValidUserPasswordMinLength {
		hashPwd, err := HashAndSaltPassword(*input.Password)
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UpdateByID", "failed to hash password")
		}

		rParams.PasswordHash = &hashPwd
	}

	if err := ref.repository.UpdateByID(ctx, rParams); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UpdateByID")
	}

	slog.Debug("service.Users.UpdateByID", "email", input.Email)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user updated successfully",
		attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersService) DeleteByID(ctx context.Context, input *model.DeleteUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.DeleteByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if input.ID == uuid.Nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.DeleteByID")
	}

	rParams := &model.DeleteUserInput{
		ID: input.ID,
	}

	slog.Debug("service.Users.DeleteByID", "qParams", rParams)

	if err := ref.repository.DeleteByID(ctx, rParams); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.DeleteByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user deleted successfully", attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersService) List(ctx context.Context, input *model.ListUsersInput) (*model.ListUsersOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	rParams := &model.SelectUsersInput{
		Sort:      input.Sort,
		Filter:    input.Filter,
		Fields:    input.Fields,
		Paginator: input.Paginator,
	}

	slog.Debug("service.Users.List", "qParams", rParams)

	out, err := ref.repository.Select(ctx, rParams)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.List")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "users listed successfully",
		attribute.Int("count", len(out.Items)))

	return out, nil
}

func (ref *UsersService) LinkRoles(ctx context.Context, input *model.LinkRolesToUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.LinkRoles")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.LinkRoles")
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.LinkRoles")
	}

	if err := ref.repository.LinkRoles(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.LinkRoles")
	}

	slog.Debug("service.Users.LinkRoles", "user.id", input.UserID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "roles linked successfully",
		attribute.String("user.id", input.UserID.String()))

	return nil
}

func (ref *UsersService) UnLinkRoles(ctx context.Context, input *model.UnLinkRolesFromUsersInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.UnLinkRoles")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UnLinkRoles")
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UnLinkRoles")
	}

	if err := ref.repository.UnLinkRoles(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.UnLinkRoles")
	}

	slog.Debug("service.Users.UnLinkRoles", "user.id", input.UserID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "roles unlinked successfully",
		attribute.String("user.id", input.UserID.String()))

	return nil
}

func (ref *UsersService) SelectAuthz(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.SelectAuthz")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	if userID == uuid.Nil {
		errorType := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.SelectAuthz")
	}

	slog.Debug("service.Users.SelectAuthz", "userID", userID)
	out, err := ref.repository.SelectAuthz(ctx, userID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.SelectAuthz")
	}

	slog.Debug("service.Users.SelectAuthz", "userID", userID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user authz found successfully",
		attribute.String("user.id", userID.String()))

	return out, nil
}

// ListByRoleID returns a list of users by role ID
func (ref *UsersService) ListByRoleID(ctx context.Context, roleID uuid.UUID, input *model.ListUsersInput) (*model.ListUsersOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.ListByRoleID")
	defer span.End()

	span.SetAttributes(attribute.String("role.id", roleID.String()))

	if roleID == uuid.Nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.ListByRoleID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.ListByRoleID")
	}

	out, err := ref.repository.SelectByRoleID(ctx, roleID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.ListByRoleID")
	}

	slog.Debug("service.Users.ListByRoleID", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "users found by role ID",
		attribute.String("role.id", roleID.String()),
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *UsersService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
