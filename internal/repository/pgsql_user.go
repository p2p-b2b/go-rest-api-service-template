package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

var repositoryCalls metric.Int64Counter

type PGSQLUserRepositoryConfig struct {
	DB              *sql.DB
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
}

// this implement repository.UserRepository
// PGSQLUserRepository is a PostgreSQL store.
type PGSQLUserRepository struct {
	// DB is the PostgreSQL database.
	db *sql.DB

	// MaxQueryTimeout is the maximum time a query can take.
	maxPingTimeout time.Duration

	// MaxQueryTimeout is the maximum time a query can take.
	maxQueryTimeout time.Duration

	// Tracer for openTelemetry
	ot *o11y.OpenTelemetry
}

// NewPGSQLUserRepository creates a new PGSQLUserRepository.
func NewPGSQLUserRepository(conf PGSQLUserRepositoryConfig) *PGSQLUserRepository {
	return &PGSQLUserRepository{
		db:              conf.DB,
		maxPingTimeout:  conf.MaxPingTimeout,
		maxQueryTimeout: conf.MaxQueryTimeout,
		ot:              conf.OT,
	}
}

// RegisterMetrics registers the metrics for the user handler.
func (s *PGSQLUserRepository) RegisterMetrics() error {
	var err error
	repositoryCalls, err = s.ot.Metrics.Meter.Int64Counter(
		"repository_calls",
		metric.WithDescription("The number of calls to the user repository"),
	)

	return err
}

// DriverName returns the name of the driver.
func (r *PGSQLUserRepository) DriverName() string {
	return sql.Drivers()[0]
}

// Conn returns the connection to the repository.
func (r *PGSQLUserRepository) Conn(ctx context.Context) (*sql.Conn, error) {
	return r.db.Conn(ctx)
}

// Close closes the repository, releasing any open resources.
func (r *PGSQLUserRepository) Close() error {
	return r.db.Close()
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (r *PGSQLUserRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxPingTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "User Repository: PingContext")
	defer span.End()

	return r.db.PingContext(ctx)
}

// Insert a new user into the database.
func (r *PGSQLUserRepository) Insert(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "User Repository: Insert")
	defer span.End()

	query := fmt.Sprintf(`
        INSERT INTO users (id, first_name, last_name, email)
        VALUES ('%s', '%s', '%s', '%s')`,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
	)

	slog.Debug("Insert", "query", prettyPrint(query))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("Insert", "error", err)
		repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("method", "Insert")),
			metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
		)
		return err
	}

	repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", "Insert")),
		metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
	)
	return nil
}

// Update updates the user with the specified ID.
func (r *PGSQLUserRepository) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "user Repository: Update")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", user.ID.String()))

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

	query := fmt.Sprintf(`
        UPDATE users SET %s
        WHERE id = '%s'`,
		fields,
		user.ID,
	)

	slog.Debug("Update", "query", prettyPrint(query))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("Update", "error", err)
		repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("method", "Update")),
			metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
		)
		return err
	}

	repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", "Update")),
		metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
	)
	return nil
}

// Delete deletes the user with the specified ID.
func (r *PGSQLUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "User Repository: Delete")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	query := fmt.Sprintf(`
        DELETE FROM users
        WHERE id = '%s'`,
		id,
	)

	slog.Debug("Delete", "query", query)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("Delete", "error", err)
		repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("method", "Delete")),
			metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
		)
		return err
	}

	repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", "Delete")),
		metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
	)
	return nil
}

// SelectByID returns the user with the specified ID.
func (r *PGSQLUserRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "User Repository: SelectByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	query := fmt.Sprintf(`
        SELECT id, first_name, last_name, email
        FROM users
        WHERE id = '%s'`,
		id,
	)

	slog.Debug("SelectByID", "query", prettyPrint(query))

	row := r.db.QueryRowContext(ctx, query)

	var u model.User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email); err != nil {
		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		slog.Error("SelectByID", "error", err)
		repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("method", "SelectByID")),
			metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
		)
		return nil, err
	}

	repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", "SelectByID")),
		metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
	)
	return &u, nil
}

// SelectByEmail returns the user with the specified email.
func (r *PGSQLUserRepository) SelectByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "User Repository: SelectByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	query := fmt.Sprintf(`
        SELECT id, first_name, last_name, email
        FROM users
        WHERE email = '%s'`,
		email,
	)

	slog.Debug("SelectByEmail", "query", prettyPrint(query))

	row := r.db.QueryRowContext(ctx, query)

	var u model.User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email); err != nil {
		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		slog.Error("SelectByEmail", "error", err)
		repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("method", "SelectByEmail")),
			metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
		)
		return nil, err
	}

	repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", "SelectByEmail")),
		metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
	)
	return &u, nil
}

// SelectAll returns a list of users.
func (r *PGSQLUserRepository) SelectAll(ctx context.Context, params *model.SelectAllUserQueryInput) (*model.SelectAllUserQueryOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "User Repository: SelectAll")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.fields", strings.Join(params.Fields, ",")),
		attribute.String("user.sort", params.Sort),
		attribute.String("user.filter", strings.Join(params.Filter, ",")),
		attribute.Int("user.limit", params.Paginator.Limit),
		attribute.String("user.next_token", params.Paginator.NextToken),
		attribute.String("user.prev_token", params.Paginator.PrevToken),
	)

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
		slog.Warn("SelectAll",
			"message",
			"both next and prev tokens are provided, going to use next token")

		// clean the prev token
		params.Paginator.PrevToken = ""
	}

	// if next token is provided
	if params.Paginator.NextToken != "" {
		// decode the token
		id, createdAt, err := paginator.DecodeToken(params.Paginator.NextToken)
		if err != nil {
			span.SetStatus(codes.Error, "invalid token")
			span.RecordError(err)
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
			span.SetStatus(codes.Error, "invalid token")
			span.RecordError(err)
			slog.Error("SelectAll", "error", err)
			repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(attribute.String("method", "SelectByEmail")),
				metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
			)
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
	slog.Debug("SelectAll", "query", prettyPrint(query))

	// execute the query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("SelectAll", "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			span.SetStatus(codes.Error, "scan failed")
			span.RecordError(err)
			slog.Error("SelectAll", "error", err)
			repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(attribute.String("method", "SelectAll")),
				metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
			)
			return nil, err
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		span.SetStatus(codes.Error, "rows failed")
		span.RecordError(err)
		slog.Error("SelectAll", "error", err)
		repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("method", "SelectAll")),
			metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", false))),
		)
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

	repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", "SelectAll")),
		metric.WithAttributes(attribute.String("successful", fmt.Sprintf("%t", true))),
	)

	return ret, nil
}
