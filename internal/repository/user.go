package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
)

// this is a mockgen command to generate a mock for UserRepository
//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/repository/user.go -source=user.go UserRepository

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

	// Update updates the user with the specified ID.
	Update(ctx context.Context, user *model.User) error

	// Delete deletes the user with the specified ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// SelectByID returns the user with the specified ID.
	SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// SelectAll returns a list of users.
	SelectAll(ctx context.Context, params *model.ListUserInput) (*model.ListUserOutput, error)
}
