package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
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
	if err != nil {
		slog.Error("Insert", "error", err)
		return err
	}

	return nil
}

// Update updates the user with the specified ID.
func (s *PGSQLUserRepository) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	var queryFields []string

	if user.FirstName != "" {
		queryFields = append(queryFields, fmt.Sprintf("first_name = '%s'", user.FirstName))
	}

	if user.LastName != "" {
		queryFields = append(queryFields, fmt.Sprintf("last_name = '%s'", user.LastName))
	}

	if user.Email != "" {
		queryFields = append(queryFields, fmt.Sprintf("email = '%s'", user.Email))
	}

	if len(queryFields) == 0 {
		slog.Warn("Update", "error", "no fields to update")
		return nil
	}

	fields := strings.Join(queryFields, ", ")

	query := fmt.Sprintf(`UPDATE users SET %s WHERE id = '%s'`,
		fields,
		user.ID,
	)

	slog.Debug("Update", "query", query)

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		slog.Error("Update", "error", err)
		return err
	}

	return nil
}

// Delete deletes the user with the specified ID.
func (s *PGSQLUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`DELETE FROM users WHERE id = '%s'`, id)

	slog.Debug("Delete", "query", query)

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		slog.Error("Delete", "error", err)
		return err
	}

	return nil
}

// SelectByID returns the user with the specified ID.
func (s *PGSQLUserRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT id, first_name, last_name, email FROM users WHERE id = '%s'`, id)

	slog.Debug("SelectByID", "query", query)

	row := s.db.QueryRowContext(ctx, query)

	var u model.User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email); err != nil {
		return nil, err
	}

	return &u, nil
}

// SelectAll returns a list of users.
func (s *PGSQLUserRepository) SelectAll(ctx context.Context, params *model.SelectAllUserQueryInput) (*model.SelectAllUserQueryOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, s.MaxQueryTimeout)
	defer cancel()

	if params == nil {
		params = &model.SelectAllUserQueryInput{
			Fields: []string{"*"},
			Sort:   "created_at",
			Filter: []string{},
			Paginator: paginator.Paginator{
				Limit: paginator.DefaultLimit,
			},
		}
	}

	fieldsStr := strings.Join(params.Fields, ", ")
	if fieldsStr == "" {
		fieldsStr = "*"
	}

	slog.Debug("SelectAll", "params", params)

	var paginationQuery string

	// if both next and prev tokens are provided, use next token
	if params.Paginator.NextToken != "" && params.Paginator.PrevToken != "" {
		slog.Warn("SelectAll", "error", "both next and prev tokens are provided, going to use next token")

		// clean the prev token
		params.Paginator.PrevToken = ""
	}

	// if next token is provided
	if params.Paginator.NextToken != "" {
		// decode the token
		id, createdAt, err := paginator.DecodeToken(params.Paginator.NextToken)
		if err != nil {
			slog.Error("SelectAll", "error", err)
			return nil, err
		}

		// from newest to oldest
		paginationQuery = fmt.Sprintf(`
                WHERE usrs.created_at < '%s' AND (usrs.id < '%s' OR usrs.created_at < '%s')
                ORDER BY usrs.created_at DESC, usrs.id DESC
                LIMIT %d
            `,
			createdAt.UTC().Format(paginator.DateFormat),
			id.String(),
			createdAt.UTC().Format(paginator.DateFormat),
			params.Paginator.Limit,
		)
	}

	// if prev token is provided
	if params.Paginator.PrevToken != "" {
		// decode the token
		id, createdAt, err := paginator.DecodeToken(params.Paginator.PrevToken)
		if err != nil {
			slog.Error("SelectAll", "error", err)
			return nil, err
		}

		// from newest to oldest
		paginationQuery = fmt.Sprintf(`
                WHERE usrs.created_at > '%s' AND (usrs.id > '%s' OR usrs.created_at > '%s')
                ORDER BY usrs.created_at ASC, usrs.id ASC
                LIMIT %d
                `,
			createdAt.UTC().Format(paginator.DateFormat),
			id.String(),
			createdAt.UTC().Format(paginator.DateFormat),
			params.Paginator.Limit,
		)
	}

	// if no token is provided, first page
	// newest to oldest
	if params.Paginator.NextToken == "" && params.Paginator.PrevToken == "" {
		paginationQuery = fmt.Sprintf(`
            ORDER BY usrs.created_at DESC, id DESC
            LIMIT %d
            `, params.Paginator.Limit)
	}

	query := fmt.Sprintf(`
        WITH usrs AS (
            SELECT %s FROM users usrs %s
        )
        SELECT * FROM usrs ORDER BY created_at DESC, id DESC
    `,
		fieldsStr,
		paginationQuery,
	)

	// helper function to pretty print the query
	prettyPrintQuery := func() string {
		ws := regexp.MustCompile(`\s+`)
		out := ws.ReplaceAllString(query, " ")
		out = strings.ReplaceAll(out, "\n", "")
		out = strings.TrimSpace(out)
		return out
	}
	slog.Debug("SelectAll", "query", prettyPrintQuery())

	// execute the query
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("SelectAll", "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			slog.Error("SelectAll", "error", err)
			return nil, err
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		slog.Error("SelectAll", "error", err)
		return nil, err
	}

	outLen := len(users)

	if outLen == 0 {
		return &model.SelectAllUserQueryOutput{
			Items:     make([]*model.User, 0),
			Paginator: paginator.Paginator{},
		}, nil
	}

	slog.Debug("SelectAll", "next_id", users[outLen-1].ID, "next_created_at", users[outLen-1].CreatedAt)
	slog.Debug("SelectAll", "prev_id", users[0].ID, "prev_created_at", users[0].CreatedAt)

	nextToken := params.Paginator.GenerateToken(users[outLen-1].ID, users[outLen-1].CreatedAt)
	prevToken := params.Paginator.GenerateToken(users[0].ID, users[0].CreatedAt)

	ret := &model.SelectAllUserQueryOutput{
		Items: users,
		Paginator: paginator.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Size:      outLen,
			Limit:     params.Paginator.Limit,
		},
	}

	return ret, nil
}
