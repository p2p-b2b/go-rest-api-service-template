package service

import (
	"context"
	"runtime"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
	"github.com/p2p-b2b/go-service-template/internal/repository"
)

// this is a mockgen command to generate a mock for UserService
//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/service/user.go -source=user.go UserService

// UserService represents a service for managing users.
type UserService interface {
	// HealthCheck verifies a connection to the repository is still alive.
	HealthCheck(ctx context.Context) (model.Health, error)

	// GetByID returns the user with the specified ID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// Create inserts a new user into the database.
	Create(ctx context.Context, user *model.CreateUserInput) error

	// Update updates the user with the specified ID.
	Update(ctx context.Context, user *model.UpdateUserInput) error

	// Delete deletes the user with the specified ID.
	Delete(ctx context.Context, user *model.DeleteUserInput) error

	// List returns a list of users.
	List(ctx context.Context) ([]*model.ListUserOutput, error)
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

// HealthCheck verifies a connection to the repository is still alive.
func (s *DefaultUserService) HealthCheck(ctx context.Context) (model.Health, error) {
	// database
	dbStatus := model.StatusUp
	err := s.repository.PingContext(ctx)
	if err != nil {
		dbStatus = model.StatusDown
	}

	database := model.Check{
		Name:   "database",
		Status: dbStatus,
	}

	// runtime
	rtStatus := model.StatusUp
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	rt := model.Check{
		Name:   "runtime",
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

// GetByID returns the user with the specified ID.
func (s *DefaultUserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repository.SelectByID(ctx, id)
}

// Create inserts a new user into the database.
func (s *DefaultUserService) Create(ctx context.Context, user *model.CreateUserInput) error {
	return s.repository.Insert(ctx, &model.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       user.Age,
	})
}

// Update updates the user with the specified ID.
func (s *DefaultUserService) Update(ctx context.Context, user *model.UpdateUserInput) error {
	return s.repository.Update(ctx, (*model.User)(user))
}

// Delete deletes the user with the specified ID.
func (s *DefaultUserService) Delete(ctx context.Context, user *model.DeleteUserInput) error {
	return s.repository.Delete(ctx, user.ID)
}

// List returns a list of users.
func (s *DefaultUserService) List(ctx context.Context) ([]*model.ListUserOutput, error) {
	users, err := s.repository.SelectAll(ctx)
	if err != nil {
		return nil, err
	}

	return []*model.ListUserOutput{
		{
			Data:       users,
			TotalCount: len(users),
			Page:       1,
			PageSize:   len(users),
			TotalPages: 1,
		},
	}, nil
}
