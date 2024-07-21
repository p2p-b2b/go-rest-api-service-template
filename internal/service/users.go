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
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/service/users.go -source=users.go UserRepository

// UserRepository represents a repository for managing users.
type UserRepository interface {
	// DriverName returns the name of the driver.
	DriverName() string

	// Close closes the repository, releasing any open resources.
	Close() error

	// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
	PingContext(ctx context.Context) error

	// Conn returns the connection to the repository.
	Conn(ctx context.Context) (*sql.Conn, error)

	// Insert a new user into the database.
	Insert(ctx context.Context, user *model.User) error

	// Update updates the user.
	Update(ctx context.Context, user *model.User) error

	// Delete deletes the user.
	Delete(ctx context.Context, user *model.User) error

	// SelectByID returns the user with the specified ID.
	SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// SelectByEmail returns the user with the specified email.
	SelectByEmail(ctx context.Context, email string) (*model.User, error)

	// SelectAll returns a list of users.
	SelectAll(ctx context.Context, params *model.SelectAllUserQueryInput) (*model.SelectAllUserQueryOutput, error)
}

var (
	ErrGettingUserByID     = errors.New("error getting user by ID")
	ErrGettingUserByEmail  = errors.New("error getting user by email")
	ErrCreatingUser        = errors.New("error creating user")
	ErrUserIDAlreadyExists = errors.New("id already exists")
	ErrUpdatingUser        = errors.New("error updating user")
	ErrDeletingUser        = errors.New("error deleting user")
	ErrListingUsers        = errors.New("error listing users")
)

type UserServiceConf struct {
	Repository    UserRepository
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type userServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type User struct {
	repository    UserRepository
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       userServiceMetrics
}

// NewUserService creates a new UserService.
func NewUserService(conf UserServiceConf) *User {
	u := &User{
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
func (s *User) registerMetrics() error {
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
func (s *User) UserHealthCheck(ctx context.Context) (model.Health, error) {
	// database
	dbStatus := model.StatusUp
	err := s.repository.PingContext(ctx)
	if err != nil {
		slog.Error("service.users.UserHealthCheck", "error", err)
		dbStatus = model.StatusDown
	}

	database := model.Check{
		Name:   "database",
		Kind:   s.repository.DriverName(),
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

	health := model.Health{
		Status: allStatus,
		Checks: []model.Check{
			database,
			rt,
		},
	}

	return health, err
}

// GetUserByID returns the user with the specified ID.
func (s *User) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.GetUserByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users"),
		attribute.String("function", "GetUserByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "GetUserByID"),
	}

	user, err := s.repository.SelectByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.GetUserByID", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, ErrGettingUserByID
	}

	span.SetStatus(codes.Ok, "user found")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return user, nil
}

// GetUserByEmail returns the user with the specified email.
func (s *User) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.GetUserByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users"),
		attribute.String("function", "GetUserByEmail"),
		attribute.String("user.email", email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "GetUserByEmail"),
	}

	user, err := s.repository.SelectByEmail(ctx, email)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.GetUserByEmail", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, ErrGettingUserByEmail
	}

	span.SetStatus(codes.Ok, "User found")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return user, nil
}

// CreateUser inserts a new user into the database.
func (s *User) CreateUser(ctx context.Context, user *model.User) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users"),
		attribute.String("function", "CreateUser"),
		attribute.String("user.email", user.Email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "CreateUser"),
	}

	if user == nil {
		span.SetStatus(codes.Error, ErrCreatingUser.Error())
		span.RecordError(ErrCreatingUser)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrCreatingUser
	}

	// if user.ID is nil, generate a new UUID
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if err := s.repository.Insert(ctx, user); err != nil {
		pgxErr, ok := err.(*pgconn.PgError)

		slog.Error("service.users.CreateUser", "error", err, "error_code", pgxErr.Code)
		if ok {
			if pgxErr.Code == "23505" {
				span.SetStatus(codes.Error, "ID already exists")
				span.RecordError(ErrUserIDAlreadyExists)
				s.metrics.serviceCalls.Add(ctx, 1,
					metric.WithAttributes(
						append(metricCommonAttributes, attribute.String("successful", "false"))...,
					),
				)
				return ErrUserIDAlreadyExists
			}
		}

		return ErrCreatingUser
	}

	span.SetStatus(codes.Ok, "User created")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)
	return nil
}

// UpdateUser updates the user with the specified ID.
func (s *User) UpdateUser(ctx context.Context, user *model.User) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.UpdateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users"),
		attribute.String("function", "UpdateUser"),
		attribute.String("user.id", user.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "UpdateUser"),
	}

	if user == nil {
		span.SetStatus(codes.Error, ErrUpdatingUser.Error())
		span.RecordError(ErrUpdatingUser)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrUpdatingUser
	}

	if err := s.repository.Update(ctx, user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("service.users.UpdateUser", "error", err)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUpdatingUser
	}

	span.SetStatus(codes.Ok, "User updated")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// DeleteUser deletes the user with the specified ID.
func (s *User) DeleteUser(ctx context.Context, user *model.User) error {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.DeleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users"),
		attribute.String("function", "DeleteUser"),
		attribute.String("user.id", user.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "DeleteUser"),
	}

	if user.ID == uuid.Nil {
		span.SetStatus(codes.Error, ErrDeletingUser.Error())
		span.RecordError(ErrDeletingUser)
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return ErrDeletingUser
	}

	if err := s.repository.Delete(ctx, user); err != nil {
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
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// ListUsers returns a list of users.
func (s *User) ListUsers(ctx context.Context, params *model.ListUserRequest) (*model.ListUserResponse, error) {
	ctx, span := s.ot.Traces.Tracer.Start(ctx, "service.users.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "service.users"),
		attribute.String("function", "ListUsers"),
		attribute.String("sort", params.Sort),
		attribute.StringSlice("fields", params.Fields),
		attribute.String("filter", params.Filter),
		attribute.Int("limit", params.Paginator.Limit),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "ListUsers"),
	}

	qParams := &model.SelectAllUserQueryInput{
		Sort:      params.Sort,
		Filter:    params.Filter,
		Fields:    params.Fields,
		Paginator: params.Paginator,
	}

	slog.Debug("service.users.ListUsers", "qParams", qParams)

	qryOut, err := s.repository.SelectAll(ctx, qParams)
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

	users := qryOut.Items
	if len(users) == 0 {
		slog.Debug("service.users.ListUsers", "message", "no users found")
		span.SetStatus(codes.Error, "no users found")
		s.metrics.serviceCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "true"))...,
			),
		)
		return &model.ListUserResponse{
			Items:     users,
			Paginator: paginator.Paginator{},
		}, nil
	}

	span.SetStatus(codes.Ok, "Users found")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return &model.ListUserResponse{
		Items:     users,
		Paginator: qryOut.Paginator,
	}, nil
}
