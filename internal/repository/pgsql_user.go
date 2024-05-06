package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
)

type PGSQLUserRepositoryConfig struct {
	DB              *sql.DB
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
}

// this implement repository.UserRepository
// PGSQLUserRepository is a PostgreSQL store.
type PGSQLUserRepository struct {
	// DB is the PostgreSQL database.
	db *sql.DB

	// MaxQueryTimeout is the maximum time a query can take.
	MaxPingTimeout time.Duration

	// MaxQueryTimeout is the maximum time a query can take.
	MaxQueryTimeout time.Duration
}

// NewPGSQLUserRepository creates a new PGSQLUserRepository.
func NewPGSQLUserRepository(conf PGSQLUserRepositoryConfig) *PGSQLUserRepository {
	return &PGSQLUserRepository{
		db:              conf.DB,
		MaxPingTimeout:  conf.MaxPingTimeout,
		MaxQueryTimeout: conf.MaxQueryTimeout,
	}
}

// DriverName returns the name of the driver.
func (s *PGSQLUserRepository) DriverName() string {
	return sql.Drivers()[0]
}

// Conn returns the connection to the repository.
func (s *PGSQLUserRepository) Conn(ctx context.Context) (*sql.Conn, error) {
	return s.db.Conn(ctx)
}

// Close closes the repository, releasing any open resources.
func (s *PGSQLUserRepository) Close() error {
	return s.db.Close()
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (s *PGSQLUserRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxPingTimeout)
	defer cancel()

	return s.db.PingContext(ctx)
}

// Insert a new user into the database.
func (s *PGSQLUserRepository) Insert(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = "INSERT INTO users (id, first_name, last_name, email) VALUES ($1, $2, $3, $4)"

	_, err := s.db.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.Email)

	return err
}

// Update updates the user with the specified ID.
func (s *PGSQLUserRepository) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `UPDATE users SET first_name = $1, last_name = $2, email = $3 WHERE id = $4`

	_, err := s.db.ExecContext(ctx, query, user.FirstName, user.LastName, user.Email, user.ID)
	return err
}

// Delete deletes the user with the specified ID.
func (s *PGSQLUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `DELETE FROM users WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// SelectByID returns the user with the specified ID.
func (s *PGSQLUserRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `SELECT id, first_name, last_name, email FROM users WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, id)

	var u model.User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email); err != nil {
		return nil, err
	}

	return &u, nil
}

// SelectAll returns a list of users.
func (s *PGSQLUserRepository) SelectAll(ctx context.Context, params *model.ListUserInput) (*model.ListUserOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	const query = `SELECT id, first_name, last_name, email FROM users ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return &model.ListUserOutput{
		Items:    users,
		Next:     "",
		Previous: "",
		Total:    len(users),
	}, nil
}
