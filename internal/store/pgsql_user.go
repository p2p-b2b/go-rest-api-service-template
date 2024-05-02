package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
)

type PGSQLUserStoreConfig struct {
	DB              *sql.DB
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
}

// this implement repository.UserRepository
// PGSQLUserStore is a PostgreSQL store.
type PGSQLUserStore struct {
	// DB is the PostgreSQL database.
	DB *sql.DB

	// MaxQueryTimeout is the maximum time a query can take.
	MaxPingTimeout time.Duration

	// MaxQueryTimeout is the maximum time a query can take.
	MaxQueryTimeout time.Duration
}

// NewPGSQLUserStore creates a new PGSQLUserStore.
func NewPGSQLUserStore(conf PGSQLUserStoreConfig) *PGSQLUserStore {
	return &PGSQLUserStore{
		DB:              conf.DB,
		MaxPingTimeout:  conf.MaxPingTimeout,
		MaxQueryTimeout: conf.MaxQueryTimeout,
	}
}

// Close closes the repository, releasing any open resources.
func (s *PGSQLUserStore) Close() error {
	return s.DB.Close()
}

// Ping verifies a connection to the repository is still alive, establishing a connection if necessary.
func (s *PGSQLUserStore) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxPingTimeout)
	defer cancel()

	return s.DB.PingContext(ctx)
}

// GetByID returns the user with the specified ID.
func (s *PGSQLUserStore) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `SELECT id, first_name, last_name, age FROM users WHERE id = $1`

	row := s.DB.QueryRowContext(ctx, query, id)

	var u model.User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Age); err != nil {
		return nil, err
	}

	return &u, nil
}

// Create inserts a new user into the database.
func (s *PGSQLUserStore) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `INSERT INTO users (id, first_name, last_name, age) VALUES ($1, $2, $3, $4)`

	_, err := s.DB.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.Age)
	return err
}

// Update updates the user with the specified ID.
func (s *PGSQLUserStore) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `UPDATE users SET first_name = $1, last_name = $2, age = $3 WHERE id = $4`

	_, err := s.DB.ExecContext(ctx, query, user.FirstName, user.LastName, user.Age, user.ID)
	return err
}

// Delete deletes the user with the specified ID.
func (s *PGSQLUserStore) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `DELETE FROM users WHERE id = $1`

	_, err := s.DB.ExecContext(ctx, query, id)
	return err
}

// List returns a list of users.
func (s *PGSQLUserStore) List(ctx context.Context) ([]*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `SELECT id, first_name, last_name, age FROM users`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}
