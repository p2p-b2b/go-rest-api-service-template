package service

import (
	"context"
	"database/sql"
	"errors"
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

// UsersRepository is the interface for the user repository methods.
type UsersRepository interface {
	Insert(ctx context.Context, input *model.InsertUserInput) error
	UpdateByID(ctx context.Context, input *model.UpdateUserInput) error
	DeleteByID(ctx context.Context, input *model.DeleteUserInput) error
	SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	SelectByEmail(ctx context.Context, email string) (*model.User, error)
	Select(ctx context.Context, input *model.SelectUsersInput) (*model.SelectUsersOutput, error)
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

// NewUsersService creates a new UsersService.
func NewUsersService(conf UsersServiceConf) (*UsersService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryConfigurationError{Message: "invalid repository configuration. It is nil"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "invalid OpenTelemetry configuration. It is nil"}
	}

	u := &UsersService{
		repository: conf.Repository,
		ot:         conf.OT,
	}
	if conf.MetricsPrefix != "" {
		u.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		u.metricsPrefix += "_"
	}

	serviceCalls, err := u.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", u.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the user service"),
	)
	if err != nil {
		slog.Error("service.Users.NewUsersService", "error", err)
		return nil, err
	}
	u.metrics.serviceCalls = serviceCalls

	return u, nil
}

// GetByID returns the user with the specified ID.
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

// GetByEmail returns the user with the specified email.
func (ref *UsersService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.GetByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	if email == "" {
		errorType := &model.InvalidUserEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByEmail")
	}

	slog.Debug("service.Users.GetByEmail", "email", email)
	out, err := ref.repository.SelectByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, o11y.RecordError(ctx, span, &model.UserNotFoundError{Email: email}, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByEmail")
		}
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.GetByEmail")
	}

	slog.Debug("service.Users.GetByEmail", "email", out.Email)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user found", attribute.String("user.email", out.Email))

	return out, nil
}

// Create inserts a new user into the database.
func (ref *UsersService) Create(ctx context.Context, input *model.CreateUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.Create")
	defer span.End()

	if input == nil {
		errorValue := &model.InputIsInvalidError{Message: "input cannot be nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Create")
	}

	span.SetAttributes(attribute.String("user.email", input.Email))

	if input.ID == uuid.Nil {
		input.ID = uuid.New()
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
		Disabled:     input.Disabled,
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

// Update updates the user with the specified ID.
func (ref *UsersService) UpdateByID(ctx context.Context, input *model.UpdateUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.Update")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Update")
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Update")
	}

	rParams := &model.UpdateUserInput{
		ID:        input.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Disabled:  input.Disabled,
	}

	// update the password if it is provided
	if input.Password != nil && len(*input.Password) < model.ValidUserPasswordMinLength {

		hashPwd, err := HashAndSaltPassword(*input.Password)
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Update", "failed to hash password")
		}

		rParams.PasswordHash = &hashPwd
	}

	if err := ref.repository.UpdateByID(ctx, rParams); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.Update")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "user updated successfully", attribute.String("user.id", input.ID.String()))

	return nil
}

// Delete deletes the user with the specified ID.
func (ref *UsersService) DeleteByID(ctx context.Context, input *model.DeleteUserInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Users.DeleteByID")
	defer span.End()

	if input.ID == uuid.Nil {
		errorValue := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Users.DeleteByID")
	}
	span.SetAttributes(attribute.String("user.id", input.ID.String()))

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

// List returns a list of users.
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
