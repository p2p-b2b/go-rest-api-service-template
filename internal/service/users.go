package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -package=mocks -destination=../../mocks/service/users.go -source=users.go UserRepository

// UserRepository is the interface for the user repository methods.
type UserRepository interface {
	DriverName() string
	Close() error
	PingContext(ctx context.Context) error
	Conn(ctx context.Context) (*sql.Conn, error)
	Insert(ctx context.Context, input *repository.InsertUserInput) error
	Update(ctx context.Context, input *repository.UpdateUserInput) error
	Delete(ctx context.Context, input *repository.DeleteUserInput) error
	SelectUserByID(ctx context.Context, id uuid.UUID) (*repository.User, error)
	SelectUserByEmail(ctx context.Context, email string) (*repository.User, error)
	Select(ctx context.Context, input *repository.SelectUsersInput) (*repository.SelectUsersOutput, error)
}

type UserServiceConf struct {
	Repository    UserRepository
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type userServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type UserService struct {
	repository    UserRepository
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       userServiceMetrics
}

// NewUserService creates a new UserService.
func NewUserService(conf UserServiceConf) (*UserService, error) {
	if conf.Repository == nil {
		return nil, ErrInvalidRepository
	}

	if conf.OT == nil {
		return nil, ErrInvalidOpenTelemetry
	}

	u := &UserService{
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
		slog.Error("service.users.registerMetrics", "error", err)
		return nil, err
	}
	u.metrics.serviceCalls = serviceCalls

	return u, nil
}

// UserHealthCheck verifies a connection to the repository is still alive.
func (s *UserService) UserHealthCheck(ctx context.Context) (Health, error) {
	// database
	dbStatus := StatusUp
	err := s.repository.PingContext(ctx)
	if err != nil {
		slog.Error("service.users.UserHealthCheck", "error", err)
		dbStatus = StatusDown
	}

	database := Check{
		Name:   "database",
		Kind:   s.repository.DriverName(),
		Status: dbStatus,
	}

	// runtime
	rtStatus := StatusUp
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	rt := Check{
		Name:   "runtime",
		Kind:   "go",
		Status: rtStatus,
		Data: map[string]interface{}{
			"version":      runtime.Version(),
			"numCPU":       runtime.NumCPU(),
			"numGoroutine": runtime.NumGoroutine(),
			"numCgoCall":   runtime.NumCgoCall(),
			"memStats":     mem,
		},
	}

	// and operator
	allStatus := dbStatus && rtStatus

	health := Health{
		Status: allStatus,
		Checks: []Check{
			database,
			rt,
		},
	}

	return health, err
}

// GetUserByID returns the user with the specified ID.
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.GetUserByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.GetUserByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.GetUserByID"),
	}

	if id == uuid.Nil {
		slog.Error("service.users.GetUserByID", "error", ErrInvalidUserID)
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrInvalidUserID
	}

	slog.Debug("service.users.GetUserByID", "id", id)
	qryOut, err := s.repository.SelectUserByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.GetUserByID", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	user := &User{
		ID:        qryOut.ID,
		FirstName: qryOut.FirstName,
		LastName:  qryOut.LastName,
		Email:     qryOut.Email,
		Disabled:  qryOut.Disabled,
		CreatedAt: qryOut.CreatedAt,
		UpdatedAt: qryOut.UpdatedAt,
	}

	slog.Debug("service.users.GetUserByID", "email", user.Email)
	span.SetStatus(codes.Ok, "user found")
	span.SetAttributes(attribute.String("user.email", user.Email))
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return user, nil
}

// GetUserByEmail returns the user with the specified email.
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.GetUserByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.GetUserByEmail"),
		attribute.String("user.email", email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.GetUserByEmail"),
	}

	if email == "" {
		slog.Error("service.users.GetUserByEmail", "error", ErrInvalidUserEmail)
		span.SetStatus(codes.Error, ErrInvalidUserEmail.Error())
		span.RecordError(ErrInvalidUserEmail)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrInvalidUserEmail
	}

	slog.Debug("service.users.GetUserByEmail", "email", email)
	qryOut, err := s.repository.SelectUserByEmail(ctx, email)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.GetUserByEmail", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	user := &User{
		ID:        qryOut.ID,
		FirstName: qryOut.FirstName,
		LastName:  qryOut.LastName,
		Email:     qryOut.Email,
		Disabled:  qryOut.Disabled,
		CreatedAt: qryOut.CreatedAt,
		UpdatedAt: qryOut.UpdatedAt,
	}

	slog.Debug("service.users.GetUserByEmail", "email", user.Email)

	span.SetStatus(codes.Ok, "user found")
	span.SetAttributes(attribute.String("user.email", user.Email))
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return user, nil
}

// CreateUser inserts a new user into the database.
func (s *UserService) CreateUser(ctx context.Context, input *CreateUserInput) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.CreateUser"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.CreateUser"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		s.metrics.serviceCalls.Add(ctx, 1,
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
		slog.Error("service.users.CreateUser", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	hashPwd, err := hashAndSaltPassword(input.Password)
	if err != nil {
		slog.Error("handler.users.createUser", "error", err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.createUser", "error", err.Error())
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		return err
	}

	rParams := &repository.InsertUserInput{
		ID:           input.ID,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		Disabled:     input.Disabled,
		PasswordHash: hashPwd,
	}

	if err := s.repository.Insert(ctx, rParams); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.CreateUser", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, repository.ErrUserEmailAlreadyExists) {
			return ErrUserEmailAlreadyExists
		}

		if errors.Is(err, repository.ErrUserIDAlreadyExists) {
			return ErrUserIDAlreadyExists
		}

		return err
	}

	slog.Debug("service.users.CreateUser", "email", input.Email)
	span.SetStatus(codes.Ok, "User created")
	span.SetAttributes(attribute.String("user.email", input.Email))
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)
	return nil
}

// UpdateUser updates the user with the specified ID.
func (s *UserService) UpdateUser(ctx context.Context, input *UpdateUserInput) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.UpdateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.UpdateUser"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.UpdateUser"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		s.metrics.serviceCalls.Add(ctx, 1,
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
		slog.Error("service.users.UpdateUser", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	rParams := &repository.UpdateUserInput{
		ID:        input.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Disabled:  input.Disabled,
	}

	// update the password if it is provided
	if input.Password != nil && len(*input.Password) < UsersPasswordMinLength {

		hashPwd, err := hashAndSaltPassword(*input.Password)
		if err != nil {
			slog.Error("handler.users.createUser", "error", err.Error())
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			slog.Error("handler.users.createUser", "error", err.Error())
			s.metrics.serviceCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
				),
			)

			return err
		}

		rParams.PasswordHash = &hashPwd
	}

	if err := s.repository.Update(ctx, rParams); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.UpdateUser", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}

		if errors.Is(err, repository.ErrUserEmailAlreadyExists) {
			return ErrUserEmailAlreadyExists
		}

		return err
	}

	slog.Debug("service.users.UpdateUser", "email", input.Email)
	span.SetStatus(codes.Ok, "User updated")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// DeleteUser deletes the user with the specified ID.
func (s *UserService) DeleteUser(ctx context.Context, input *DeleteUserInput) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.DeleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.DeleteUser"),
		attribute.String("user.id", input.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.DeleteUser"),
	}

	if input.ID == uuid.Nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrInputIsNil
	}

	rParams := &repository.DeleteUserInput{
		ID: input.ID,
	}

	slog.Debug("service.users.DeleteUser", "qParams", rParams)

	if err := s.repository.Delete(ctx, rParams); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.DeleteUser", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	span.SetStatus(codes.Ok, "User deleted")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// ListUsers returns a list of users.
func (s *UserService) ListUsers(ctx context.Context, input *ListUserInput) (*ListUsersOutput, error) {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.ListUsers"),
		attribute.String("sort", input.Sort),
		attribute.StringSlice("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.ListUsers"),
	}

	rParams := &repository.SelectUsersInput{
		Sort:      input.Sort,
		Filter:    input.Filter,
		Fields:    input.Fields,
		Paginator: input.Paginator,
	}

	slog.Debug("service.users.ListUsers", "qParams", rParams)

	qryOut, err := s.repository.Select(ctx, rParams)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.ListUsers", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}

	users := make([]*User, len(qryOut.Items))
	for i, u := range qryOut.Items {
		users[i] = &User{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Disabled:  u.Disabled,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	slog.Debug("service.users.ListUsers", "users", len(users))
	span.SetStatus(codes.Ok, "Users found")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return &ListUsersOutput{
		Items:     users,
		Paginator: qryOut.Paginator,
	}, nil
}
