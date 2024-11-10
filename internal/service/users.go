package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	SelectByID(ctx context.Context, id uuid.UUID) (*repository.User, error)
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
func NewUserService(conf UserServiceConf) *UserService {
	u := &UserService{
		repository: conf.Repository,
		ot:         conf.OT,
	}
	if conf.MetricsPrefix != "" {
		u.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		u.metricsPrefix += "_"
	}

	if err := u.registerMetrics(); err != nil {
		slog.Error("service.users.NewUserService", "error", err)
		panic(err)
	}

	return u
}

// registerMetrics registers the metrics for the user handler.
func (s *UserService) registerMetrics() error {
	serviceCalls, err := s.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", s.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the user service"),
	)
	if err != nil {
		slog.Error("service.users.registerMetrics", "error", err)
		return err
	}
	s.metrics.serviceCalls = serviceCalls

	return nil
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
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("service.users.GetUserByID", "error", ErrInvalidUserID)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrInvalidUserID
	}

	slog.Debug("service.users.GetUserByID", "id", id)
	qryOut, err := s.repository.SelectByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.GetUserByID", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, ErrGettingUserByID
	}

	if qryOut == nil {
		span.SetStatus(codes.Error, ErrGettingUserByID.Error())
		span.RecordError(ErrGettingUserByID)
		slog.Error("service.users.GetUserByID", "error", ErrGettingUserByID)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrGettingUserByID
	}

	user := &User{
		ID:        qryOut.ID,
		FirstName: qryOut.FirstName,
		LastName:  qryOut.LastName,
		Email:     qryOut.Email,
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

// CreateUser inserts a new user into the database.
func (s *UserService) CreateUser(ctx context.Context, input *CreateUserInput) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users.CreateUser"),
		attribute.String("user.email", input.Email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.CreateUser"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrCreatingUser.Error())
		span.RecordError(ErrCreatingUser)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrCreatingUser
	}

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

	rParams := &repository.InsertUserInput{
		ID:        input.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
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

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if strings.Contains(pgErr.Message, "_pkey") {
					return ErrUserIDAlreadyExists
				}

				if strings.Contains(pgErr.Message, "_email") {
					return ErrUserEmailAlreadyExists
				}

				return ErrCreatingUser
			}
		}

		return ErrCreatingUser
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
		attribute.String("user.id", input.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users.UpdateUser"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrUpdatingUser.Error())
		span.RecordError(ErrUpdatingUser)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrUpdatingUser
	}

	rParams := &repository.UpdateUserInput{
		ID:        input.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
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

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if strings.Contains(pgErr.Message, "_email") {
					return ErrUserEmailAlreadyExists
				}
			}
		}

		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}

		return ErrUpdatingUser
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
		span.SetStatus(codes.Error, ErrDeletingUser.Error())
		span.RecordError(ErrDeletingUser)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrDeletingUser
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
		return ErrDeletingUser
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

	if qryOut == nil {
		span.SetStatus(codes.Error, ErrListingUsers.Error())
		span.RecordError(ErrListingUsers)
		slog.Error("service.users.ListUsers", "error", ErrListingUsers)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, nil
	}

	users := make([]*User, len(qryOut.Items))
	for i, u := range qryOut.Items {
		users[i] = &User{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
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
