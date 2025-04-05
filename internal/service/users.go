package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/users.go -source=users.go UsersRepository

// UsersRepository is the interface for the user repository methods.
type UsersRepository interface {
	Insert(ctx context.Context, input *model.InsertUserInput) error
	Update(ctx context.Context, input *model.UpdateUserInput) error
	Delete(ctx context.Context, input *model.DeleteUserInput) error
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
		return nil, ErrRepositoryRequired
	}

	if conf.OT == nil {
		return nil, ErrOpenTelemetryRequired
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
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "service.Users.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.Users.GetByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.Users.GetByID"),
	}

	if id == uuid.Nil {
		slog.Error("service.Users.GetByID", "error", model.ErrUserInvalidID)
		span.SetStatus(codes.Error, model.ErrUserInvalidID.Error())
		span.RecordError(model.ErrUserInvalidID)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, model.ErrUserInvalidID
	}

	slog.Debug("service.Users.GetByID", "user.id", id)
	out, err := ref.repository.SelectByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.GetByID", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}

		return nil, err
	}

	slog.Debug("service.Users.GetByID", "user.email", out.Email)
	span.SetStatus(codes.Ok, "user found")
	span.SetAttributes(attribute.String("user.email", out.Email))
	ref.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return out, nil
}

// GetByEmail returns the user with the specified email.
func (ref *UsersService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "service.Users.GetByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.Users.GetByEmail"),
		attribute.String("user.email", email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.Users.GetByEmail"),
	}

	if email == "" {
		slog.Error("service.Users.GetByEmail", "error", model.ErrUserInvalidEmail)
		span.SetStatus(codes.Error, model.ErrUserInvalidEmail.Error())
		span.RecordError(model.ErrUserInvalidEmail)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, model.ErrUserInvalidEmail
	}

	slog.Debug("service.Users.GetByEmail", "user.email", email)
	out, err := ref.repository.SelectByEmail(ctx, email)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.GetByEmail", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) ||
			errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrUserNotFound
		}

		return nil, err
	}

	slog.Debug("service.Users.GetByEmail", "user.email", out.Email)
	span.SetStatus(codes.Ok, "user found")
	span.SetAttributes(attribute.String("user.email", out.Email))
	ref.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return out, nil
}

// Create inserts a new user into the database.
func (ref *UsersService) Create(ctx context.Context, input *model.CreateUserInput) error {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "service.Users.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.Users.Create"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.Users.Create"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrInputIsNil
	}

	span.SetAttributes(
		attribute.String("user.email", input.Email),
	)

	if input.ID == uuid.Nil {
		input.ID = uuid.New()
	}

	// validate the user input
	if err := input.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.Create", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	hashPwd, err := hashAndSaltPassword(input.Password)
	if err != nil {
		slog.Error("handler.Users.createUser", "error", err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.createUser", "error", err.Error())
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		return err
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
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.Create", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	slog.Debug("service.Users.Create", "user.email", input.Email)
	span.SetStatus(codes.Ok, "User created")
	span.SetAttributes(attribute.String("user.email", input.Email))
	ref.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// Update updates the user with the specified ID.
func (ref *UsersService) Update(ctx context.Context, input *model.UpdateUserInput) error {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "service.Users.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.Users.Update"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.Users.Update"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrInputIsNil
	}

	span.SetAttributes(
		attribute.String("user.id", input.ID.String()),
	)

	if err := input.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.Update", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
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

		hashPwd, err := hashAndSaltPassword(*input.Password)
		if err != nil {
			slog.Error("handler.Users.createUser", "error", err.Error())
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			slog.Error("handler.Users.createUser", "error", err.Error())
			ref.metrics.serviceCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
				),
			)

			return err
		}

		rParams.PasswordHash = &hashPwd
	}

	if err := ref.repository.Update(ctx, rParams); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.Update", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	slog.Debug("service.Users.Update", "user.email", input.Email)
	span.SetStatus(codes.Ok, "User updated")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	ref.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// Delete deletes the user with the specified ID.
func (ref *UsersService) Delete(ctx context.Context, input *model.DeleteUserInput) error {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "service.Users.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.Users.Delete"),
		attribute.String("user.id", input.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.Users.Delete"),
	}

	if input.ID == uuid.Nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrInputIsNil
	}

	rParams := &model.DeleteUserInput{
		ID: input.ID,
	}

	slog.Debug("service.Users.Delete", "qParams", rParams)

	if err := ref.repository.Delete(ctx, rParams); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.Delete", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	span.SetStatus(codes.Ok, "User deleted")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	ref.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// List returns a list of users.
func (ref *UsersService) List(ctx context.Context, input *model.ListUsersInput) (*model.ListUsersOutput, error) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "service.Users.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.Users.List"),
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.Users.List"),
	}

	rParams := &model.SelectUsersInput{
		Sort:      input.Sort,
		Filter:    input.Filter,
		Fields:    input.Fields,
		Paginator: input.Paginator,
	}

	slog.Debug("service.Users.List", "qParams", rParams)

	out, err := ref.repository.Select(ctx, rParams)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.Users.List", "error", err)
		ref.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}

	slog.Debug("service.Users.List", "users.count", len(out.Items))
	span.SetStatus(codes.Ok, "Users found")
	ref.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return &model.ListUsersOutput{
		Items:     out.Items,
		Paginator: out.Paginator,
	}, nil
}
