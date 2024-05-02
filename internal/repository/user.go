package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
)

// this is a mockgen command to generate a mock for UserRepository
//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/user_repository.go -source=user.go UserRepository

// UserRepository represents a repository for managing users.
type UserRepository interface {
	// Close closes the repository, releasing any open resources.
	Close() error

	// Ping verifies a connection to the repository is still alive, establishing a connection if necessary.
	Ping(ctx context.Context) error

	// GetByID returns the user with the specified ID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// Create inserts a new user into the database.
	Create(ctx context.Context, user *model.User) error

	// Update updates the user with the specified ID.
	Update(ctx context.Context, user *model.User) error

	// Delete deletes the user with the specified ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// List returns a list of users.
	List(ctx context.Context) ([]*model.User, error)
}
