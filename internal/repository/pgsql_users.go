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

type PGSQLUserRepositoryConfig struct {
	DB              *sql.DB
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type pgsqlUserRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// this implement repository.UserRepository
// PGSQLUserRepository is a PostgreSQL store.
type PGSQLUserRepository struct {
	db              *sql.DB
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         pgsqlUserRepositoryMetrics
}

// NewPGSQLUserRepository creates a new PGSQLUserRepository.
func NewPGSQLUserRepository(conf PGSQLUserRepositoryConfig) *PGSQLUserRepository {
	r := &PGSQLUserRepository{
		db:              conf.DB,
		maxPingTimeout:  conf.MaxPingTimeout,
		maxQueryTimeout: conf.MaxQueryTimeout,
		ot:              conf.OT,
	}
	if conf.MetricsPrefix != "" {
		r.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		r.metricsPrefix += "_"
	}

	if err := r.registerMetrics(); err != nil {
		slog.Error("failed to register metrics", "error", err)
		panic(err)
	}

	return r
}

// registerMetrics registers the metrics for the user handler.
func (r *PGSQLUserRepository) registerMetrics() error {
	repositoryCalls, err := r.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", r.metricsPrefix, "repositories_calls_total"),
		metric.WithDescription("The number of calls to the user repository"),
	)
	if err != nil {
		slog.Error("repository.users.registerMetrics", "error", err)
		return err
	}
	r.metrics.repositoryCalls = repositoryCalls

	return nil
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

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.PingContext")
	defer span.End()

	return r.db.PingContext(ctx)
}

// Insert a new user into the database.
func (r *PGSQLUserRepository) Insert(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Insert")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Insert"),
		attribute.String("user.id", user.ID.String()),
	)

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
		slog.Error("repository.users.Insert", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("driver", r.DriverName()),
				attribute.String("component", "repository.user"),
				attribute.String("function", "Insert"),
				attribute.String("successful", "false"),
			),
		)
		return err
	}

	span.SetStatus(codes.Ok, "user inserted successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("driver", r.DriverName()),
			attribute.String("component", "repository.user"),
			attribute.String("function", "Insert"),
			attribute.String("successful", "true"),
		),
	)
	return nil
}

// Update updates the user with the specified ID.
func (r *PGSQLUserRepository) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Update"),
		attribute.String("user.id", user.ID.String()),
	)

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
		slog.Error("repository.users.Update", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("driver", r.DriverName()),
				attribute.String("component", "repository.user"),
				attribute.String("function", "Update"),
				attribute.String("successful", "false"),
			),
		)
		return err
	}

	span.SetStatus(codes.Ok, "user updated successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("component", "repository.user"),
			attribute.String("function", "Update"),
			attribute.String("successful", "true"),
		),
	)

	return nil
}

// Delete deletes the user with the specified ID.
func (r *PGSQLUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Delete"),
		attribute.String("user.id", id.String()),
	)

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
		slog.Error("repository.users.Delete", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("driver", r.DriverName()),
				attribute.String("component", "repository.user"),
				attribute.String("function", "Delete"),
				attribute.String("successful", "false"),
			),
		)
		return err
	}

	span.SetStatus(codes.Ok, "user deleted successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("driver", r.DriverName()),
			attribute.String("component", "repository.user"),
			attribute.String("function", "Delete"),
			attribute.String("successful", "true"),
		),
	)

	return nil
}

// SelectByID returns the user with the specified ID.
func (r *PGSQLUserRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.SelectByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "SelectByID"),
		attribute.String("user.id", id.String()),
	)

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
		slog.Error("repository.users.SelectByID", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("driver", r.DriverName()),
				attribute.String("component", "repository.user"),
				attribute.String("function", "SelectByID"),
				attribute.String("successful", "false"),
			),
		)
		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("driver", r.DriverName()),
			attribute.String("component", "repository.user"),
			attribute.String("function", "SelectByID"),
			attribute.String("successful", "true"),
		),
	)
	return &u, nil
}

// SelectByEmail returns the user with the specified email.
func (r *PGSQLUserRepository) SelectByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.SelectByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "SelectByEmail"),
		attribute.String("user.email", email),
	)

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
		slog.Error("repository.users.SelectByEmail", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("driver", r.DriverName()),
				attribute.String("component", "repository.user"),
				attribute.String("function", "SelectByEmail"),
				attribute.String("successful", "true"),
			),
		)
		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("driver", r.DriverName()),
			attribute.String("component", "repository.user"),
			attribute.String("function", "SelectByEmail"),
			attribute.String("successful", "true"),
		),
	)

	return &u, nil
}

// SelectAll returns a list of users.
func (r *PGSQLUserRepository) SelectAll(ctx context.Context, params *model.SelectAllUserQueryInput) (*model.SelectAllUserQueryOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.SelectAll")
	defer span.End()

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

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "SelectByEmail"),
		attribute.String("user.sort", params.Sort),
		attribute.Int("user.limit", params.Paginator.Limit),
		attribute.String("user.fields", strings.Join(params.Fields, ",")),
		attribute.String("user.filter", strings.Join(params.Filter, ",")),
	)

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
			slog.Error("repository.users.SelectAll", "error", err)
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
			slog.Error("repository.users.SelectAll", "error", err)
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("driver", r.DriverName()),
					attribute.String("component", "repository.user"),
					attribute.String("function", "SelectByEmail"),
					attribute.String("successful", "false"),
				),
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
	slog.Debug("repository.users.SelectAll", "query", prettyPrint(query))

	// execute the query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("repository.users.SelectAll", "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			span.SetStatus(codes.Error, "scan failed")
			span.RecordError(err)
			slog.Error("repository.users.SelectAll", "error", err)
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("driver", r.DriverName()),
					attribute.String("component", "repository.user"),
					attribute.String("function", "SelectAll"),
					attribute.String("successful", "false"),
				),
			)
			return nil, err
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		span.SetStatus(codes.Error, "rows failed")
		span.RecordError(err)
		slog.Error("repository.users.SelectAll", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("driver", r.DriverName()),
				attribute.String("component", "repository.user"),
				attribute.String("function", "SelectAll"),
				attribute.String("successful", "false"),
			),
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

	slog.Debug("repository.users.SelectAll", "next_id", users[outLen-1].ID, "next_created_at", users[outLen-1].CreatedAt)
	slog.Debug("repository.users.SelectAll", "prev_id", users[0].ID, "prev_created_at", users[0].CreatedAt)

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

	span.SetStatus(codes.Ok, "users selected successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("driver", r.DriverName()),
			attribute.String("component", "repository.user"),
			attribute.String("function", "SelectAll"),
			attribute.String("successful", "true"),
		),
	)

	return ret, nil
}
