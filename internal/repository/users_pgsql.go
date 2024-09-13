package repository

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/query"
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
func (r *PGSQLUserRepository) Insert(ctx context.Context, user *InsertUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Insert")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Insert"),
		attribute.String("user.id", user.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Insert"),
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
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.users.Insert", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		// remove the SQLSTATE XXXXX suffix
		errMessage := strings.TrimSpace(strings.Split(err.Error(), "(SQLSTATE")[0])
		return fmt.Errorf("%s", errMessage)
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
func (r *PGSQLUserRepository) Update(ctx context.Context, user *UpdateUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Update"),
		attribute.String("user.id", user.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Update"),
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

	if err := user.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.users.Update", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	var queryFields []string

	if user.FirstName != nil && *user.FirstName != "" {
		queryFields = append(queryFields, fmt.Sprintf("first_name = '%s'", *user.FirstName))
	}

	if user.LastName != nil && *user.LastName != "" {
		queryFields = append(queryFields, fmt.Sprintf("last_name = '%s'", *user.LastName))
	}

	if user.Email != nil && *user.Email != "" {
		queryFields = append(queryFields, fmt.Sprintf("email = '%s'", *user.Email))
	}

	if len(queryFields) == 0 {
		slog.Error("Update", "error", "no fields to update")
		return ErrAtLeastOneFieldMustBeUpdated
	}

	updatedAt, _ := time.Now().In(time.FixedZone("UTC", 0)).MarshalText()
	queryFields = append(queryFields, fmt.Sprintf("updated_at = '%s'", updatedAt))

	fields := strings.Join(queryFields, ", ")

	slog.Debug("repository.users.Update", "fields", fields)

	query := fmt.Sprintf(`
        UPDATE users SET %s
        WHERE id = '%s'`,
		fields,
		user.ID,
	)

	slog.Debug("repository.users.Update", "query", prettyPrint(query))

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
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

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		span.SetStatus(codes.Error, "user not found")
		span.RecordError(ErrUserNotFound)
		slog.Error("repository.user.Update", "error", ErrUserNotFound)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return ErrUserNotFound
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
func (r *PGSQLUserRepository) Delete(ctx context.Context, user *DeleteUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Delete"),
		attribute.String("user.id", user.ID.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Delete"),
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

	if err := user.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.users.Delete", "error", err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	query := fmt.Sprintf(`
        DELETE FROM users
        WHERE id = '%s'`,
		user.ID,
	)

	slog.Debug("repository.users.Delete", "query", prettyPrint(query))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
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
func (r *PGSQLUserRepository) SelectByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.SelectByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.SelectByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.SelectByID"),
	}

	if id == uuid.Nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("repository.users.SelectByID", "error", "id is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, ErrInvalidUserID
	}

	query := fmt.Sprintf(`
        SELECT id, first_name, last_name, email, created_at, updated_at
        FROM users
        WHERE id = '%s'`,
		id,
	)

	slog.Debug("repository.users.SelectByID", "query", prettyPrint(query))

	row := r.db.QueryRowContext(ctx, query)

	var u User
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
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

func (r *PGSQLUserRepository) Select(ctx context.Context, params *SelectUsersInput) (*SelectUsersOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.maxQueryTimeout)
	defer cancel()

	ctx, span := r.ot.Traces.Tracer.Start(ctx, "repository.user.Select")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Select"),
		attribute.String("user.sort", params.Sort),
		attribute.String("user.filter", params.Filter),
		attribute.Int("user.limit", params.Paginator.Limit),
		attribute.String("user.fields", strings.Join(params.Fields, ",")),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", r.DriverName()),
		attribute.String("component", "repository.user.Select"),
	}

	if params == nil {
		span.SetStatus(codes.Error, ErrFunctionParameterIsNil.Error())
		span.RecordError(ErrFunctionParameterIsNil)
		slog.Error("repository.users.Select", "error", "params is nil")
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, ErrFunctionParameterIsNil
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "usrs."
	fieldsStr := sqlFieldsPrefix + "*"
	if params.Fields[0] != "" {
		fields := make([]string, 0)
		var isIsPresent bool
		for _, field := range params.Fields {
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
	if params.Filter != "" {
		filter, err := query.PrefixFilterFields(params.Filter, sqlFieldsPrefix)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			slog.Error("repository.users.Select", "error", err)
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
	if params.Sort == "" {
		sortQuery = "usrs.serial_id DESC, usrs.id DESC"
	} else {
		sortQuery = params.Sort
	}

	// query template
	var queryTemplate string = `
        WITH {{.TableAlias}} AS (
            SELECT {{.QueryColumns}}
            FROM {{.TableName}} {{.TableAlias}}
            {{ .QueryWhere }}
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM {{.TableAlias}} ORDER BY {{.QueryExternalSort}}
    `

	// struct to hold the query values
	var queryValues struct {
		TableName         string
		TableAlias        string
		QueryColumns      string
		QueryWhere        template.HTML
		QueryLimit        int
		QueryInternalSort string
		QueryExternalSort string
	}

	// default values
	queryValues.TableName = "users"
	queryValues.TableAlias = "usrs"
	queryValues.QueryColumns = fieldsStr
	queryValues.QueryWhere = template.HTML(filterQuery)
	queryValues.QueryLimit = params.Paginator.Limit
	queryValues.QueryInternalSort = "usrs.serial_id DESC, usrs.id DESC"
	queryValues.QueryExternalSort = sortQuery

	filterQueryJoiner := "WHERE"
	if filterQuery != "" {
		filterQueryJoiner = "AND"
	}

	// if both next and prev tokens are provided, use next token
	if params.Paginator.NextToken != "" && params.Paginator.PrevToken != "" {
		slog.Warn("repository.user.Select",
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
			slog.Error("repository.user.Select", "error", err)
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
	if params.Paginator.PrevToken != "" {
		// decode the token
		id, serial, err := paginator.DecodeToken(params.Paginator.PrevToken)
		if err != nil {
			slog.Error("repository.user.Select", "error", err)
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
		slog.Error("repository.user.Select", "error", err)
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
	slog.Debug("repository.user.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("repository.user.Select", "error", err)
		span.SetStatus(codes.Error, "failed to select all users")
		span.RecordError(err)
		r.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		// remove the SQLSTATE XXXXX suffix
		errMessage := strings.TrimSpace(strings.Split(err.Error(), "(SQLSTATE")[0])
		return nil, fmt.Errorf("%s", errMessage)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User

		scanFields := make([]interface{}, 0)

		if params.Fields[0] == "" {
			scanFields = []interface{}{&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.SerialID}
		} else {
			var idFound bool

			for _, field := range params.Fields {
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
				case "created_at":
					scanFields = append(scanFields, &u.CreatedAt)
				case "updated_at":
					scanFields = append(scanFields, &u.UpdatedAt)

				default:
					slog.Warn("repository.user.Select", "message", "field not found", "field", field)
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
			slog.Error("repository.user.Select", "error", err)
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
		slog.Warn("repository.user.Select", "message", "no users found")
		return &SelectUsersOutput{
			Items:     make([]*User, 0),
			Paginator: paginator.Paginator{},
		}, nil
	}

	slog.Debug("repository.users.Select", "next_id", users[outLen-1].ID, "next_serial_id", users[outLen-1].SerialID)
	slog.Debug("repository.users.Select", "prev_id", users[0].ID, "prev_serial_id", users[0].SerialID)

	nextToken, prevToken := paginator.GetTokens(
		outLen,
		params.Paginator.Limit,
		users[0].ID,
		users[0].SerialID,
		users[outLen-1].ID,
		users[outLen-1].SerialID,
	)

	ret := &SelectUsersOutput{
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
