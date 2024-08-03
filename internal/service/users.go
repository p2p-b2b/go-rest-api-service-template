package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/service/users.go -source=users.go UserRepository

// UserRepository is the interface for the user repository methods.
type UserRepository interface {
	DriverName() string
	Close() error
	PingContext(ctx context.Context) error
	Conn(ctx context.Context) (*sql.Conn, error)
	Insert(ctx context.Context, user *repository.InsertUserInput) error
	Update(ctx context.Context, user *repository.UpdateUserInput) error
	Delete(ctx context.Context, user *repository.DeleteUserInput) error
	SelectByID(ctx context.Context, id uuid.UUID) (*repository.User, error)
	Select(ctx context.Context, params *repository.SelectUserInput) (*repository.SelectUserOutput, error)
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
		attribute.String("component", "service.users"),
		attribute.String("function", "GetUserByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "service.users"),
		attribute.String("function", "GetUserByID"),
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
	}

	span.SetStatus(codes.Ok, "user found")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return user, nil
}

// CreateUser inserts a new user into the database.
func (s *UserService) CreateUser(ctx context.Context, user *CreateUserInput) error {
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

	// validate the user input
	if err := user.Validate(); err != nil {
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
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
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
		return fmt.Errorf("%w: %s", ErrCreatingUser, err)
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
func (s *UserService) UpdateUser(ctx context.Context, user *UpdateUserInput) error {
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

	rParams := &repository.UpdateUserInput{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
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
func (s *UserService) DeleteUser(ctx context.Context, user *DeleteUserInput) error {
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

	rParams := &repository.DeleteUserInput{
		ID: user.ID,
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
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// ListUsers returns a list of users.
func (s *UserService) ListUsers(ctx context.Context, params *ListUserInput) (*ListUserOutput, error) {
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

	rParams := &repository.SelectUserInput{
		Sort:      params.Sort,
		Filter:    params.Filter,
		Fields:    params.Fields,
		Paginator: params.Paginator,
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

	span.SetStatus(codes.Ok, "Users found")
	s.metrics.serviceCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return &ListUserOutput{
		Items:     users,
		Paginator: qryOut.Paginator,
	}, nil
}
