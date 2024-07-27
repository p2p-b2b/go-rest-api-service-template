package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

var (
	// ErrUserIsNil is an error that is returned when the user is nil.
	ErrUserIsNil = fmt.Errorf("user is nil")

	// ErrUserIDIsNil is an error that is returned when the user ID is nil.
	ErrUserIDIsNil = fmt.Errorf("user ID is nil")

	// ErrFunctionParameterIsNil is an error that is returned when a function parameter is nil.
	ErrFunctionParameterIsNil = fmt.Errorf("function parameter is nil")
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
func (r *PGSQLUserRepository) Insert(ctx context.Context, user *model.InsertUserInput) error {
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

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Insert"),
	}

	if user == nil {
		span.SetStatus(codes.Error, ErrUserIsNil.Error())
		span.RecordError(ErrUserIsNil)
		slog.Error("repository.users.Insert", "error", ErrUserIsNil.Error())
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserIsNil
	}

	if err := user.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.users.Insert", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return err
	}

	query := fmt.Sprintf(`
        INSERT INTO users (id, first_name, last_name, email)
        VALUES ('%s', '%s', '%s', '%s')`,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
	)

	slog.Debug("repository.users.Insert", "query", prettyPrint(query))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.SetStatus(codes.Error, pgErr.Error())
			span.RecordError(pgErr)
			slog.Error("repository.users.Insert", "error", pgErr.Error())
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)

			// return the pgErr as an error without the (SQLSTATE XXXXX) suffix
			// remove the SQLSTATE XXXXX suffix
			// https://github.com/jackc/pgconn/blob/1860f4e57204614f40d05a5c76a43e8d80fde9da/errors.go#L52
			errMessage := strings.Split(pgErr.Message, "(SQLSTATE")[0]
			return fmt.Errorf("%s", errMessage)
		}

		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("repository.users.Insert", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return err
	}

	span.SetStatus(codes.Ok, "user inserted successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)
	return nil
}

// Update updates the user with the specified ID.
func (r *PGSQLUserRepository) Update(ctx context.Context, user *model.UpdateUserInput) error {
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

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Update"),
	}

	if user == nil {
		span.SetStatus(codes.Error, ErrUserIsNil.Error())
		span.RecordError(ErrUserIsNil)
		slog.Error("repository.users.Update", "error", "user is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserIsNil
	}

	if user.ID == uuid.Nil {
		span.SetStatus(codes.Error, ErrUserIDIsNil.Error())
		span.RecordError(ErrUserIDIsNil)
		slog.Error("repository.users.Update", "error", "user id is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserIDIsNil
	}

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

	slog.Debug("repository.users.Update", "query", prettyPrint(query))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.SetStatus(codes.Error, pgErr.Error())
			span.RecordError(pgErr)
			slog.Error("repository.users.Update", "error", pgErr.Error())
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)

			// return the pgErr as an error without the (SQLSTATE XXXXX) suffix
			// remove the SQLSTATE XXXXX suffix
			// https://github.com/jackc/pgconn/blob/1860f4e57204614f40d05a5c76a43e8d80fde9da/errors.go#L52
			errMessage := strings.Split(pgErr.Message, "(SQLSTATE")[0]
			return fmt.Errorf("%s", errMessage)
		}

		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("repository.users.Update", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return err
	}

	span.SetStatus(codes.Ok, "user updated successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// Delete deletes the user with the specified ID.
func (r *PGSQLUserRepository) Delete(ctx context.Context, user *model.DeleteUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Delete"),
		attribute.String("user.id", user.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "Delete"),
	}

	if user == nil {
		span.SetStatus(codes.Error, ErrUserIsNil.Error())
		span.RecordError(ErrUserIsNil)
		slog.Error("repository.users.Delete", "error", "user is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserIsNil
	}

	if user.ID == uuid.Nil {
		span.SetStatus(codes.Error, ErrUserIDIsNil.Error())
		span.RecordError(ErrUserIDIsNil)
		slog.Error("repository.users.Delete", "error", "user id is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserIDIsNil
	}

	query := fmt.Sprintf(`
        DELETE FROM users
        WHERE id = '%s'`,
		user.ID,
	)

	slog.Debug("repository.users.Delete", "query", prettyPrint(query))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.SetStatus(codes.Error, pgErr.Error())
			span.RecordError(pgErr)
			slog.Error("repository.users.Delete", "error", pgErr.Error())
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)

			// return the pgErr as an error without the (SQLSTATE XXXXX) suffix
			// remove the SQLSTATE XXXXX suffix
			// https://github.com/jackc/pgconn/blob/1860f4e57204614f40d05a5c76a43e8d80fde9da/errors.go#L52
			errMessage := strings.Split(pgErr.Message, "(SQLSTATE")[0]
			return fmt.Errorf("%s", errMessage)
		}

		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("repository.users.Delete", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return err
	}

	span.SetStatus(codes.Ok, "user deleted successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
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

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "SelectByID"),
	}

	if id == uuid.Nil {
		span.SetStatus(codes.Error, ErrUserIDIsNil.Error())
		span.RecordError(ErrUserIDIsNil)
		slog.Error("repository.users.SelectByID", "error", "id is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrUserIDIsNil
	}

	query := fmt.Sprintf(`
        SELECT id, first_name, last_name, email, created_at, updated_at
        FROM users
        WHERE id = '%s'`,
		id,
	)

	slog.Debug("repository.users.SelectByID", "query", prettyPrint(query))

	row := r.db.QueryRowContext(ctx, query)

	var u model.User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.SetStatus(codes.Error, pgErr.Error())
			span.RecordError(pgErr)
			slog.Error("repository.users.SelectByID", "error", pgErr.Error())
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)

			// return the pgErr as an error without the (SQLSTATE XXXXX) suffix
			// remove the SQLSTATE XXXXX suffix
			// https://github.com/jackc/pgconn/blob/1860f4e57204614f40d05a5c76a43e8d80fde9da/errors.go#L52
			errMessage := strings.Split(pgErr.Message, "(SQLSTATE")[0]
			return nil, fmt.Errorf("%s", errMessage)
		}

		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		slog.Error("repository.users.SelectByID", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)
	return &u, nil
}

// SelectAll selects all users from the repository.
func (r *PGSQLUserRepository) SelectAll(ctx context.Context, params *model.SelectAllUserQueryInput) (*model.SelectAllUserQueryOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.SelectAll")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("method", "SelectByEmail"),
		attribute.String("user.sort", params.Sort),
		attribute.String("user.filter", params.Filter),
		attribute.Int("user.limit", params.Paginator.Limit),
		attribute.String("user.fields", strings.Join(params.Fields, ",")),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user"),
		attribute.String("function", "SelectAll"),
	}

	if params == nil {
		span.SetStatus(codes.Error, ErrFunctionParameterIsNil.Error())
		span.RecordError(ErrFunctionParameterIsNil)
		slog.Error("repository.users.SelectAll", "error", "params is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, ErrFunctionParameterIsNil
	}

	// if no fields are provided, select all fields
	fieldsStr := "usrs.*"
	if params.Fields[0] != "" {
		fields := make([]string, 0)
		for _, field := range params.Fields {
			fields = append(fields, "usrs."+field)
		}

		fields = append(fields, "usrs.serial_id")
		fieldsStr = strings.Join(fields, ", ")
	}

	var filterQuery string
	if params.Filter != "" {
		filterQuery = fmt.Sprintf("AND (%s)", params.Filter)
	}

	var paginationQuery string
	// if both next and prev tokens are provided, use next token
	if params.Paginator.NextToken != "" && params.Paginator.PrevToken != "" {
		slog.Warn("repository.user.SelectAll",
			"message",
			"both next and prev tokens are provided, going to use next token")

		// clean the prev token
		params.Paginator.PrevToken = ""
	}

	// if next token is provided
	if params.Paginator.NextToken != "" {
		// decode the token
		id, serial, err := paginator.DecodeToken(params.Paginator.NextToken)
		if err != nil {
			slog.Error("repository.user.SelectAll", "error", err)
			span.SetStatus(codes.Error, "invalid token")
			span.RecordError(err)
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)
			return nil, err
		}
		// from newest to oldest
		paginationQuery = fmt.Sprintf(`
            WHERE usrs.serial_id < '%d'
                AND (usrs.id < '%s' OR usrs.serial_id < '%d')
                %s
            ORDER BY usrs.serial_id DESC, usrs.id DESC
            LIMIT %d
        `,
			serial,
			id.String(),
			serial,
			filterQuery,
			params.Paginator.Limit,
		)
	}

	// if prev token is provided
	if params.Paginator.PrevToken != "" {
		// decode the token
		id, serial, err := paginator.DecodeToken(params.Paginator.PrevToken)
		if err != nil {
			slog.Error("repository.user.SelectAll", "error", err)
			span.SetStatus(codes.Error, "invalid token")
			span.RecordError(err)
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)
			return nil, err
		}

		// from newest to oldest
		paginationQuery = fmt.Sprintf(`
                WHERE usrs.serial_id > '%d'
                    AND (usrs.id > '%s' OR usrs.serial_id > '%d')
                    %s
                ORDER BY usrs.serial_id ASC, usrs.id ASC
                LIMIT %d`,
			serial,
			id.String(),
			serial,
			filterQuery,
			params.Paginator.Limit,
		)
	}

	// if no token is provided, first page
	// newest to oldest
	if params.Paginator.NextToken == "" && params.Paginator.PrevToken == "" {
		if params.Filter != "" {
			filterQuery = fmt.Sprintf("WHERE %s", params.Filter)
		}

		paginationQuery = fmt.Sprintf(`
            %s
            ORDER BY usrs.serial_id DESC, usrs.id DESC
            LIMIT %d
            `,
			filterQuery,
			params.Paginator.Limit,
		)
	}

	slog.Debug("repository.user.SelectAll", "filter_query", filterQuery)

	var whereQuery string
	if filterQuery != "" && paginationQuery != "" {
		whereQuery = paginationQuery
	} else if filterQuery != "" && paginationQuery == "" {
		whereQuery = filterQuery
	} else if filterQuery == "" && paginationQuery != "" {
		whereQuery = paginationQuery
	} else {
		whereQuery = ""
	}

	var sortQuery string
	if params.Sort != "" {
		sortQuery = fmt.Sprintf("ORDER BY %s", params.Sort)
	}

	// assemble the query
	query := fmt.Sprintf(`
        WITH usrs AS (
            SELECT %s FROM users usrs %s
        )
        SELECT * FROM usrs %s
        `,
		fieldsStr,
		whereQuery,
		sortQuery,
	)
	slog.Debug("repository.user.SelectAll", "query", prettyPrint(query))

	// execute the query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("repository.user.SelectAll", "error", err)
		span.SetStatus(codes.Error, "failed to select all users")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User

		scanFields := make([]interface{}, 0)

		if params.Fields[0] == "" {
			scanFields = []interface{}{&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.SerialID}
		} else {
			for _, field := range params.Fields {
				switch field {
				case "id":
					scanFields = append(scanFields, &u.ID)
				case "first_name":
					scanFields = append(scanFields, &u.FirstName)
				case "last_name":
					scanFields = append(scanFields, &u.LastName)
				case "email":
					scanFields = append(scanFields, &u.Email)
				case "created_at":
					scanFields = append(scanFields, &u.CreatedAt)
				case "updated_at":
					scanFields = append(scanFields, &u.UpdatedAt)

				default:
					slog.Warn("repository.user.SelectAll", "message", "field not found", "field", field)
				}
			}

			// always scan the serial id because it is used for pagination
			scanFields = append(scanFields, &u.SerialID)
		}

		if err := rows.Scan(scanFields...); err != nil {

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				span.SetStatus(codes.Error, pgErr.Error())
				span.RecordError(pgErr)
				slog.Error("repository.users.SelectAll", "error", pgErr.Error())
				r.metrics.repositoryCalls.Add(ctx, 1,
					metric.WithAttributes(
						append(metricCommonAttributes, attribute.String("successful", "false"))...,
					),
				)

				// return the pgErr as an error without the (SQLSTATE XXXXX) suffix
				// remove the SQLSTATE XXXXX suffix
				// https://github.com/jackc/pgconn/blob/1860f4e57204614f40d05a5c76a43e8d80fde9da/errors.go#L52
				errMessage := strings.Split(pgErr.Message, "(SQLSTATE")[0]
				return nil, fmt.Errorf("%s", errMessage)
			}

			slog.Error("repository.user.SelectAll", "error", err)
			span.SetStatus(codes.Error, "failed to scan user")
			span.RecordError(err)
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)
			return nil, err
		}

		users = append(users, &u)
	}

	outLen := len(users)
	if outLen == 0 {
		slog.Warn("repository.user.SelectAll", "message", "no users found")
		return &model.SelectAllUserQueryOutput{
			Items:     make([]*model.User, 0),
			Paginator: paginator.Paginator{},
		}, nil
	}

	slog.Debug("repository.users.SelectAll", "next_id", users[outLen-1].ID, "next_serial_id", users[outLen-1].SerialID)
	slog.Debug("repository.users.SelectAll", "prev_id", users[0].ID, "prev_serial_id", users[0].SerialID)

	nextToken, prevToken := paginator.GetTokens(
		outLen,
		params.Paginator.Limit,
		users[0].ID,
		users[0].SerialID,
		users[outLen-1].ID,
		users[outLen-1].SerialID,
	)

	ret := &model.SelectAllUserQueryOutput{
		Items: users,
		Paginator: paginator.Paginator{
			Size:      outLen,
			Limit:     params.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	span.SetStatus(codes.Ok, "users selected successfully")
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return ret, nil
}
