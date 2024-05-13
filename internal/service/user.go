package service

import (
	"context"
	"errors"
	"log/slog"
	"runtime"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
)

// this is a mockgen command to generate a mock for UserService
//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/service/user.go -source=user.go UserService

// UserService represents a service for managing users.
type UserService interface {
	// UserHealthCheck verifies a connection to the repository is still alive.
	UserHealthCheck(ctx context.Context) (model.Health, error)

	// GetUserByID returns the user with the specified ID.
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// CreateUser inserts a new user into the database.
	CreateUser(ctx context.Context, user *model.CreateUserRequest) error

	// UpdateUser updates the user with the specified ID.
	UpdateUser(ctx context.Context, user *model.User) error

	// DeleteUser deletes the user with the specified ID.
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// ListUsers returns a list of users.
	ListUsers(ctx context.Context, params *model.ListUserRequest) (*model.ListUserResponse, error)
}

type DefaultUserServiceConfig struct {
	Repository repository.UserRepository
}

type DefaultUserService struct {
	repository repository.UserRepository
}

// NewUserService creates a new UserService.
func NewDefaultUserService(conf *DefaultUserServiceConfig) *DefaultUserService {
	return &DefaultUserService{
		repository: conf.Repository,
	}
}

// UserHealthCheck verifies a connection to the repository is still alive.
func (s *DefaultUserService) UserHealthCheck(ctx context.Context) (model.Health, error) {
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
func (s *DefaultUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repository.SelectByID(ctx, id)
}

// CreateUser inserts a new user into the database.
func (s *DefaultUserService) CreateUser(ctx context.Context, user *model.CreateUserRequest) error {
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

	return s.repository.Insert(ctx, newUser)
}

// UpdateUser updates the user with the specified ID.
func (s *DefaultUserService) UpdateUser(ctx context.Context, user *model.User) error {
	return s.repository.Update(ctx, user)
}

// DeleteUser deletes the user with the specified ID.
func (s *DefaultUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}

// ListUsers returns a list of users.
func (s *DefaultUserService) ListUsers(ctx context.Context, lur *model.ListUserRequest) (*model.ListUserResponse, error) {
	qParams := &model.SelectAllUserQueryInput{
		Sort:      lur.Sort,
		Filter:    lur.Filter,
		Fields:    lur.Fields,
		Paginator: lur.Paginator,
	}

	qryOut, err := s.repository.SelectAll(ctx, qParams)
	if err != nil {
		slog.Error("Service List", "error", err)
		return nil, err
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
