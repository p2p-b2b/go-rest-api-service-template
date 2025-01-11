package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/query"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

type UserRepositoryConfig struct {
	DB              *sql.DB
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type userRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// this implement repository.UserRepository
// UserRepository is a PostgreSQL store.
type UserRepository struct {
	db              *sql.DB
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         userRepositoryMetrics
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(conf UserRepositoryConfig) (*UserRepository, error) {
	if conf.DB == nil {
		return nil, ErrInvalidDBConfiguration
	}

	if conf.MaxPingTimeout < 10*time.Millisecond {
		return nil, ErrInvalidMaxPingTimeout
	}

	if conf.MaxQueryTimeout < 10*time.Millisecond {
		return nil, ErrInvalidMaxQueryTimeout
	}

	if conf.OT == nil {
		return nil, ErrInvalidOTConfiguration
	}

	repo := &UserRepository{
		db:              conf.DB,
		maxPingTimeout:  conf.MaxPingTimeout,
		maxQueryTimeout: conf.MaxQueryTimeout,
		ot:              conf.OT,
	}
	if conf.MetricsPrefix != "" {
		repo.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		repo.metricsPrefix += "_"
	}

	repositoryCalls, err := repo.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", repo.metricsPrefix, "repository_calls_total"),
		metric.WithDescription("The number of calls to the user repository"),
	)
	if err != nil {
		slog.Error("repository.Users.NewUserRepository", "error", err)
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (r *UserRepository) DriverName() string {
	return sql.Drivers()[0]
}

// Conn returns the connection to the repository.
func (r *UserRepository) Conn(ctx context.Context) (*sql.Conn, error) {
	return r.db.Conn(ctx)
}

// Close closes the repository, releasing any open resources.
func (r *UserRepository) Close() error {
	return r.db.Close()
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (r *UserRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxPingTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.PingContext")
	defer span.End()

	return r.db.PingContext(ctx)
}

// Insert a new user into the database.
func (r *UserRepository) Insert(ctx context.Context, input *InsertUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.Insert")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Insert"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Insert"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		slog.Error("repository.Users.Insert", "error", ErrInputIsNil.Error())
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrInputIsNil
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.Users.Insert", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	query := `
        INSERT INTO users (id, first_name, last_name, email, password_hash, disabled)
        VALUES ($1, $2, $3, $4, $5, $6);
    `

	_, err := r.db.ExecContext(ctx, query,
		input.ID,
		input.FirstName,
		input.LastName,
		input.Email,
		input.PasswordHash,
		input.Disabled,
	)
	if err != nil {
		slog.Error("repository.Users.Insert", "error", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if strings.Contains(pgErr.Message, "_pkey") {
					return ErrUserIDAlreadyExists
				}

				if strings.Contains(pgErr.Message, "_email") {
					return ErrUserEmailAlreadyExists
				}

				return err
			}
		}

		return err
	}

	slog.Debug("repository.Users.Insert", "user.id", input.ID)
	span.SetStatus(codes.Ok, "user inserted successfully")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// Update updates the user with the specified ID.
func (r *UserRepository) Update(ctx context.Context, input *UpdateUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Update"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Update"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		slog.Error("repository.Users.Update", "error", "user is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrInputIsNil
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		slog.Error("repository.Users.Update", "error", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	args := []interface{}{input.ID}

	if input.FirstName != nil && *input.FirstName != "" {
		args = append(args, *input.FirstName)
	} else {
		args = append(args, nil)
	}

	if input.LastName != nil && *input.LastName != "" {
		args = append(args, *input.LastName)
	} else {
		args = append(args, nil)
	}

	if input.Email != nil && *input.Email != "" {
		args = append(args, *input.Email)
	} else {
		args = append(args, nil)
	}

	if input.PasswordHash != nil && *input.PasswordHash != "" {
		args = append(args, *input.PasswordHash)
	} else {
		args = append(args, nil)
	}

	if input.Disabled != nil {
		args = append(args, *input.Disabled)
	} else {
		args = append(args, nil)
	}

	updatedAt, _ := time.Now().In(time.FixedZone("UTC", 0)).MarshalText()
	args = append(args, updatedAt)

	query := `
        UPDATE users
            SET
                first_name = COALESCE($2, first_name),
                last_name = COALESCE($3,  last_name),
                email = COALESCE($4, email),
                password_hash = COALESCE($5, password_hash),
                disabled = COALESCE($6, disabled),
                updated_at = COALESCE($7 , updated_at)
        WHERE id = $1;
    `

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		slog.Error("repository.Users.Update", "error", err)
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		slog.Error("repository.Users.Update", "error", err)
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("repository.Users.Update", "error", err)
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrUserNotFound.Error())
		span.RecordError(ErrUserNotFound)
		slog.Error("repository.Users.Update", "error", ErrUserNotFound)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserNotFound
	}

	span.SetStatus(codes.Ok, "user updated successfully")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// Delete deletes the user with the specified ID.
func (r *UserRepository) Delete(ctx context.Context, input *DeleteUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Delete"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Delete"),
	}

	if input == nil {
		slog.Error("repository.Users.Delete", "error", "user is nil")
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrInputIsNil
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		slog.Error("repository.Users.Delete", "error", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	query := `
        DELETE FROM users
        WHERE id = $1
    `

	result, err := r.db.ExecContext(ctx, query, input.ID)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("repository.Users.Delete", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("repository.Users.Delete", "error", err)
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrUserNotFound.Error())
		span.RecordError(ErrUserNotFound)
		slog.Error("repository.Users.Delete", "error", ErrUserNotFound)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserNotFound
	}

	span.SetStatus(codes.Ok, "user deleted successfully")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

// SelectByID returns the user with the specified ID.
func (r *UserRepository) SelectByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.SelectByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.SelectByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.SelectByID"),
	}

	if id == uuid.Nil {
		slog.Error("repository.Users.SelectByID", "error", "id is nil")
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrInvalidUserID
	}

	query := `
        SELECT
            id,
            first_name,
            last_name,
            email,
            password_hash,
            disabled,
            created_at,
            updated_at
        FROM users
        WHERE id = $1;
    `

	slog.Debug("repository.Users.SelectByID", "query", prettyPrint(query))

	row := r.db.QueryRowContext(ctx, query, id)

	var u User
	if err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.PasswordHash,
		&u.Disabled,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		slog.Error("repository.Users.SelectByID", "error", err)
		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	span.SetAttributes(attribute.String("user.id", id.String()))
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)
	return &u, nil
}

// SelectByEmail returns the user with the specified email.
func (r *UserRepository) SelectByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.SelectByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.SelectByEmail"),
		attribute.String("user.email", email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.SelectByEmail"),
	}

	if email == "" {
		slog.Error("repository.Users.SelectByEmail", "error", "email is empty")
		span.SetStatus(codes.Error, ErrInvalidUserEmail.Error())
		span.RecordError(ErrInvalidUserEmail)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrInvalidUserEmail
	}

	query := `SELECT
                id,
                first_name,
                last_name,
                email,
                password_hash,
                disabled,
                created_at,
                updated_at
              FROM users
              WHERE email = $1;`

	// slog.Debug("repository.Users.SelectByEmail", "query", prettyPrint(query))

	row := r.db.QueryRowContext(ctx, query, email)

	var u User
	if err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.PasswordHash,
		&u.Disabled,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		slog.Error("repository.Users.SelectByEmail", "error", err)
		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	span.SetAttributes(attribute.String("user.email", email))
	r.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return &u, nil
}

func (r *UserRepository) Select(ctx context.Context, input *SelectUsersInput) (*SelectUsersOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.Users.Select")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Select"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.Users.Select"),
	}

	if input == nil {
		span.SetStatus(codes.Error, ErrInputIsNil.Error())
		span.RecordError(ErrInputIsNil)
		slog.Error("repository.Users.Select", "error", ErrInputIsNil)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, ErrInputIsNil
	}

	if err := input.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.Users.Select", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "usrs."
	fieldsStr := sqlFieldsPrefix + "*"
	if input.Fields[0] != "" {
		fields := make([]string, 0)
		var isIsPresent bool
		for _, field := range input.Fields {
			fields = append(fields, sqlFieldsPrefix+field)
			if field == "id" {
				isIsPresent = true
			}
		}

		// id and serial_id are always selected because they are used for pagination
		if !isIsPresent {
			fields = append(fields, sqlFieldsPrefix+"id")
		}

		fields = append(fields, sqlFieldsPrefix+"serial_id")
		fieldsStr = strings.Join(fields, ", ")
	}

	var filterQuery string
	if input.Filter != "" {
		filter, err := query.PrefixFilterFields(input.Filter, sqlFieldsPrefix)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			slog.Error("repository.Users.Select", "error", err)
			r.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)

			return nil, err
		}

		filterQuery = fmt.Sprintf("WHERE (%s)", filter)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "usrs.serial_id DESC, usrs.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	var queryTemplate string = `
        WITH usrs AS (
            SELECT
                {{.QueryColumns}}
            FROM users AS usrs
            {{ .QueryWhere }}
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM usrs ORDER BY {{.QueryExternalSort}}
    `

	// struct to hold the query values
	var queryValues struct {
		QueryColumns      string
		QueryWhere        template.HTML
		QueryLimit        int
		QueryInternalSort string
		QueryExternalSort string
	}

	// default values
	queryValues.QueryColumns = fieldsStr
	queryValues.QueryWhere = template.HTML(filterQuery)
	queryValues.QueryLimit = input.Paginator.Limit
	queryValues.QueryInternalSort = "usrs.serial_id DESC, usrs.id DESC"
	queryValues.QueryExternalSort = sortQuery

	filterQueryJoiner := "WHERE"
	if filterQuery != "" {
		filterQueryJoiner = "AND"
	}

	// if both next and prev tokens are provided, use next token
	if input.Paginator.NextToken != "" && input.Paginator.PrevToken != "" {
		slog.Warn("repository.Users.Select",
			"message",
			"both next and prev tokens are provided, going to use next token")

		// clean the prev token
		input.Paginator.PrevToken = ""
	}

	// if next token is provided
	if input.Paginator.NextToken != "" {
		// decode the token
		id, serial, err := paginator.DecodeToken(input.Paginator.NextToken)
		if err != nil {
			slog.Error("repository.Users.Select", "error", err)
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
		queryValues.QueryInternalSort = "usrs.serial_id DESC, usrs.id DESC"
		queryValues.QueryWhere = template.HTML(fmt.Sprintf(`
                %s
                    %s (usrs.serial_id < '%d')
                    AND (usrs.id < '%s' OR usrs.serial_id < '%d')`,
			filterQuery,
			filterQueryJoiner,
			serial,
			id.String(),
			serial,
		))

	}

	// if prev token is provided
	if input.Paginator.PrevToken != "" {
		// decode the token
		id, serial, err := paginator.DecodeToken(input.Paginator.PrevToken)
		if err != nil {
			slog.Error("repository.Users.Select", "error", err)
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
		queryValues.QueryInternalSort = "usrs.serial_id ASC, usrs.id ASC"
		queryValues.QueryWhere = template.HTML(fmt.Sprintf(`
                %s
                    %s (usrs.serial_id > '%d')
                    AND (usrs.id > '%s' OR usrs.serial_id > '%d')`,
			filterQuery,
			filterQueryJoiner,
			serial,
			id.String(),
			serial,
		))
	}

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err := t.Execute(&tpl, queryValues)
	if err != nil {
		slog.Error("repository.Users.Select", "error", err)
		span.SetStatus(codes.Error, "failed to render query template")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	query := tpl.String()
	slog.Debug("repository.Users.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("repository.Users.Select", "error", err)
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

	var users []*User
	for rows.Next() {
		var u User

		scanFields := make([]interface{}, 0)

		if input.Fields[0] == "" {
			scanFields = []interface{}{
				&u.ID,
				&u.FirstName,
				&u.LastName,
				&u.Email,
				&u.PasswordHash,
				&u.Disabled,
				&u.CreatedAt,
				&u.UpdatedAt,
				&u.SerialID,
			}
		} else {
			var idFound bool

			for _, field := range input.Fields {
				switch field {
				case "id":
					scanFields = append(scanFields, &u.ID)
					idFound = true
				case "first_name":
					scanFields = append(scanFields, &u.FirstName)
				case "last_name":
					scanFields = append(scanFields, &u.LastName)
				case "email":
					scanFields = append(scanFields, &u.Email)
				case "password_hash":
					scanFields = append(scanFields, &u.PasswordHash)
				case "disabled":
					scanFields = append(scanFields, &u.Disabled)
				case "created_at":
					scanFields = append(scanFields, &u.CreatedAt)
				case "updated_at":
					scanFields = append(scanFields, &u.UpdatedAt)

				default:
					slog.Warn("repository.Users.Select", "what", "field not found", "field", field)
				}
			}

			// always select id and serial_id for pagination
			// if id is not selected, it will be added to the scanFields
			if !idFound {
				scanFields = append(scanFields, &u.ID)
			}

			scanFields = append(scanFields, &u.SerialID)
		}

		if err := rows.Scan(scanFields...); err != nil {
			slog.Error("repository.Users.Select", "error", err)
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
		slog.Warn("repository.Users.Select", "what", "no users found")
		return &SelectUsersOutput{
			Items:     make([]*User, 0),
			Paginator: paginator.Paginator{},
		}, nil
	}

	slog.Debug("repository.Users.Select", "next_id", users[outLen-1].ID, "next_serial_id", users[outLen-1].SerialID)
	slog.Debug("repository.Users.Select", "prev_id", users[0].ID, "prev_serial_id", users[0].SerialID)

	nextToken, prevToken := paginator.GetTokens(
		outLen,
		input.Paginator.Limit,
		users[0].ID,
		users[0].SerialID,
		users[outLen-1].ID,
		users[outLen-1].SerialID,
	)

	ret := &SelectUsersOutput{
		Items: users,
		Paginator: paginator.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
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
