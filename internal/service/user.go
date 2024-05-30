package service

import (
	"context"
	"errors"
	"log/slog"
	"runtime"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	opentelemetry "github.com/p2p-b2b/go-rest-api-service-template/internal/opentracing"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	otelMetric "go.opentelemetry.io/otel/metric"
)

// this is a mockgen command to generate a mock for UserService
//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/service/user.go -source=user.go UserService

// Mertics struct for user service
type Metrics struct {
	service_count otelMetric.Int64Counter
}

// UserService represents a service for managing users.
type UserService interface {
	// UserHealthCheck verifies a connection to the repository is still alive.
	UserHealthCheck(ctx context.Context) (model.Health, error)

	// GetUserByID returns the user with the specified ID.
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// GetUserByEmail returns the user with the specified email.
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)

	// CreateUser inserts a new user into the database.
	CreateUser(ctx context.Context, user *model.CreateUserRequest) error

	// UpdateUser updates the user with the specified ID.
	UpdateUser(ctx context.Context, user *model.User) error

	// DeleteUser deletes the user with the specified ID.
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// ListUsers returns a list of users.
	ListUsers(ctx context.Context, params *model.ListUserRequest) (*model.ListUserResponse, error)
}

var (
	ErrGettingUserByID    = errors.New("error getting user by ID")
	ErrGettingUserByEmail = errors.New("error getting user by email")
	ErrInsertingUser      = errors.New("error inserting user")
	ErrIdAlreadyExists    = errors.New("id already exists")
	ErrUpdatingUser       = errors.New("error updating user")
	ErrDeletingUser       = errors.New("error deleting user")
	ErrListingUsers       = errors.New("error listing users")
)

type UserConf struct {
	Repository repository.UserRepository
	Ot         *opentelemetry.Opentelemetry
}
type User struct {
	repository repository.UserRepository
	ot         *opentelemetry.Opentelemetry
	mymetrics  Metrics
}

// NewUserService creates a new UserService.
func NewUserService(conf UserConf) *User {

	http_metric, _ := conf.Ot.GetMeterProdider().Meter("scope").Int64Counter("service.calls", otelMetric.WithDescription("The number of user service"))
	metrics := Metrics{
		service_count: http_metric,
	}

	return &User{
		repository: conf.Repository,
		ot:         conf.Ot,
		mymetrics:  metrics,
	}
}

// UserHealthCheck verifies a connection to the repository is still alive.
func (s *User) UserHealthCheck(ctx context.Context) (model.Health, error) {
	// database
	dbStatus := model.StatusUp
	err := s.repository.PingContext(ctx)
	if err != nil {
		slog.Error("Service HealthCheck", "error", err)
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
	_, span := s.ot.GetTrace().Start(ctx, "User Service: GetUserByID")
	defer span.End()

	user, err := s.repository.SelectByID(ctx, id)
	if err != nil {
		slog.Error("Service GetUserByID", "error", err)
		return nil, ErrGettingUserByID
	}

	s.mymetrics.service_count.Add(ctx, 1)
	return user, nil
}

// GetUserByEmail returns the user with the specified email.
func (s *User) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	_, span := s.ot.GetTrace().Start(ctx, "DB: GetUserByMail")
	defer span.End()

	user, err := s.repository.SelectByEmail(ctx, email)
	if err != nil {
		slog.Error("Service GetUserByEmail", "error", err)
		return nil, ErrGettingUserByEmail
	}

	return user, nil
}

// CreateUser inserts a new user into the database.
func (s *User) CreateUser(ctx context.Context, user *model.CreateUserRequest) error {
	_, span := s.ot.GetTrace().Start(ctx, "DB: CreateUser")
	defer span.End()

	if user == nil {
		return errors.New("user is nil")
	}

	// if user.ID is nil, generate a new UUID
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	newUser := &model.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	if err := s.repository.Insert(ctx, newUser); err != nil {
		pqErr, ok := err.(*pq.Error)

		slog.Error("Service CreateUser", "error", err, "error_code", pqErr.Code)
		if ok {
			if pqErr.Code == "23505" {
				return ErrIdAlreadyExists
			}
		}

		return ErrInsertingUser
	}

	return nil
}

// UpdateUser updates the user with the specified ID.
func (s *User) UpdateUser(ctx context.Context, user *model.User) error {
	_, span := s.ot.GetTrace().Start(ctx, "DB: UpdateUser")
	defer span.End()

	if err := s.repository.Update(ctx, user); err != nil {
		slog.Error("Service UpdateUser", "error", err)
		return ErrUpdatingUser
	}

	return nil
}

// DeleteUser deletes the user with the specified ID.
func (s *User) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, span := s.ot.GetTrace().Start(ctx, "DB: DeleteUser")
	defer span.End()

	if err := s.repository.Delete(ctx, id); err != nil {
		slog.Error("Service DeleteUser", "error", err)
		return ErrDeletingUser
	}

	return nil
}

// ListUsers returns a list of users.
func (s *User) ListUsers(ctx context.Context, lur *model.ListUserRequest) (*model.ListUserResponse, error) {

	_, span := s.ot.GetTrace().Start(ctx, "DB: ListUsers")
	defer span.End()

	qParams := &model.SelectAllUserQueryInput{
		Sort:      lur.Sort,
		Filter:    lur.Filter,
		Fields:    lur.Fields,
		Paginator: lur.Paginator,
	}

	qryOut, err := s.repository.SelectAll(ctx, qParams)
	if err != nil {
		slog.Error("Service ListUsers", "error", err)
		return nil, ErrListingUsers
	}
	if qryOut == nil {
		return nil, nil
	}

	users := qryOut.Items
	if len(users) == 0 {
		slog.Debug("Service List", "message", "no users found")
		return &model.ListUserResponse{
			Items:     users,
			Paginator: paginator.Paginator{},
		}, nil
	}

	return &model.ListUserResponse{
		Items:     users,
		Paginator: qryOut.Paginator,
	}, nil
}
