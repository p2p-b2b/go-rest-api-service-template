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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

type UsersRepositoryConfig struct {
	DB              *pgxpool.Pool
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type usersRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

type UsersRepository struct {
	db              *pgxpool.Pool
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         usersRepositoryMetrics
}

func NewUsersRepository(conf UsersRepositoryConfig) (*UsersRepository, error) {
	if conf.DB == nil {
		return nil, ErrDBInvalidConfiguration
	}

	if conf.MaxPingTimeout < 10*time.Millisecond {
		return nil, ErrDBInvalidMaxPingTimeout
	}

	if conf.MaxQueryTimeout < 10*time.Millisecond {
		return nil, ErrDBInvalidMaxQueryTimeout
	}

	if conf.OT == nil {
		return nil, ErrOTInvalidConfiguration
	}

	repo := &UsersRepository{
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
		slog.Error("repository.Users.NewUsersRepository", "error", err)
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

func (ref *UsersRepository) DriverName() string {
	return sql.Drivers()[0]
}

func (ref *UsersRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxPingTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.PingContext")
	defer span.End()

	return ref.db.Ping(ctx)
}

func (ref *UsersRepository) Insert(ctx context.Context, input *model.InsertUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxQueryTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.Insert")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Insert"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Insert"),
	}

	if input == nil {
		span.SetStatus(codes.Error, model.ErrInputIsNil.Error())
		span.RecordError(model.ErrInputIsNil)
		slog.Error("repository.Users.Insert", "error", model.ErrInputIsNil.Error())
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return model.ErrInputIsNil
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.Users.Insert", "error", err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
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

	_, err := ref.db.Exec(ctx, query,
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
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if strings.Contains(pgErr.Message, "_pkey") {
					return model.ErrUserIDAlreadyExists
				}

				if strings.Contains(pgErr.Message, "_email") {
					return model.ErrUserEmailAlreadyExists
				}

				return err
			}
		}

		return err
	}

	slog.Debug("repository.Users.Insert", "user.id", input.ID)
	span.SetStatus(codes.Ok, "user inserted successfully")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	ref.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

func (ref *UsersRepository) Update(ctx context.Context, input *model.UpdateUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxQueryTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Update"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Update"),
	}

	if input == nil {
		span.SetStatus(codes.Error, model.ErrInputIsNil.Error())
		span.RecordError(model.ErrInputIsNil)
		slog.Error("repository.Users.Update", "error", "user is nil")
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return model.ErrInputIsNil
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		slog.Error("repository.Users.Update", "error", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
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
        UPDATE users SET
            first_name = COALESCE($2, first_name),
            last_name = COALESCE($3,  last_name),
            email = COALESCE($4, email),
            password_hash = COALESCE($5, password_hash),
            disabled = COALESCE($6, disabled),
            updated_at = COALESCE($7 , updated_at)
        WHERE id = $1;
    `

	result, err := ref.db.Exec(ctx, query, args...)
	if err != nil {
		slog.Error("repository.Users.Update", "error", err)
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	if result.RowsAffected() == 0 {
		span.SetStatus(codes.Error, model.ErrUserNotFound.Error())
		span.RecordError(model.ErrUserNotFound)
		slog.Error("repository.Users.Update", "error", model.ErrUserNotFound)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return model.ErrUserNotFound
	}

	span.SetStatus(codes.Ok, "user updated successfully")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	ref.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

func (ref *UsersRepository) Delete(ctx context.Context, input *model.DeleteUserInput) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxQueryTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Delete"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Delete"),
	}

	if input == nil {
		slog.Error("repository.Users.Delete", "error", "user is nil")
		span.SetStatus(codes.Error, model.ErrInputIsNil.Error())
		span.RecordError(model.ErrInputIsNil)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return model.ErrInputIsNil
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		slog.Error("repository.Users.Delete", "error", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
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

	result, err := ref.db.Exec(ctx, query, input.ID)
	if err != nil {
		span.SetStatus(codes.Error, "query failed")
		span.RecordError(err)
		slog.Error("repository.Users.Delete", "error", err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return err
	}

	if result.RowsAffected() == 0 {
		span.SetStatus(codes.Error, model.ErrUserNotFound.Error())
		span.RecordError(model.ErrUserNotFound)
		slog.Error("repository.Users.Delete", "error", model.ErrUserNotFound)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return model.ErrUserNotFound
	}

	span.SetStatus(codes.Ok, "user deleted successfully")
	span.SetAttributes(attribute.String("user.id", input.ID.String()))
	ref.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return nil
}

func (ref *UsersRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, ref.maxQueryTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.SelectByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.SelectByID"),
		attribute.String("user.id", id.String()),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.SelectByID"),
	}

	if id == uuid.Nil {
		slog.Error("repository.Users.SelectByID", "error", "id is nil")
		span.SetStatus(codes.Error, model.ErrUserInvalidID.Error())
		span.RecordError(model.ErrUserInvalidID)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, model.ErrUserInvalidID
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

	row := ref.db.QueryRow(ctx, query, id)

	var item model.User
	if err := row.Scan(
		&item.ID,
		&item.FirstName,
		&item.LastName,
		&item.Email,
		&item.PasswordHash,
		&item.Disabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		slog.Error("repository.Users.SelectByID", "error", err)
		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	span.SetAttributes(attribute.String("user.id", id.String()))
	ref.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)
	return &item, nil
}

func (ref *UsersRepository) SelectByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, ref.maxQueryTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.SelectByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.SelectByEmail"),
		attribute.String("user.email", email),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.SelectByEmail"),
	}

	if email == "" {
		slog.Error("repository.Users.SelectByEmail", "error", "email is empty")
		span.SetStatus(codes.Error, model.ErrUserInvalidEmail.Error())
		span.RecordError(model.ErrUserInvalidEmail)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, model.ErrUserInvalidEmail
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
        WHERE email = $1;
    `

	slog.Debug("repository.Users.SelectByEmail", "query", prettyPrint(query))

	row := ref.db.QueryRow(ctx, query, email)

	var item model.User
	if err := row.Scan(
		&item.ID,
		&item.FirstName,
		&item.LastName,
		&item.Email,
		&item.PasswordHash,
		&item.Disabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		slog.Error("repository.Users.SelectByEmail", "error", err)
		span.SetStatus(codes.Error, "scan failed")
		span.RecordError(err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}

		return nil, err
	}

	span.SetStatus(codes.Ok, "user selected successfully")
	span.SetAttributes(attribute.String("user.email", email))
	ref.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return &item, nil
}

func (ref *UsersRepository) Select(ctx context.Context, input *model.SelectUsersInput) (*model.SelectUsersOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, ref.maxQueryTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Users.Select")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Select"),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", "repository.Users.Select"),
	}

	if input == nil {
		span.SetStatus(codes.Error, model.ErrInputIsNil.Error())
		span.RecordError(model.ErrInputIsNil)
		slog.Error("repository.Users.Select", "error", model.ErrInputIsNil)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)
		return nil, model.ErrInputIsNil
	}

	if err := input.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("repository.Users.Select", "error", err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "usrs."
	fieldsStr := sqlFieldsPrefix + "*"

	if input.Fields != "" {
		inputFields := strings.Split(input.Fields, ",")
		fields := make([]string, 0)
		var isIsPresent bool
		for _, field := range inputFields {
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
		filterQuery = fmt.Sprintf("WHERE (%s)", input.Filter)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "usrs.serial_id DESC, usrs.id DESC"
	} else {
		sortQuery = input.Sort
	}

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

	var queryValues struct {
		QueryColumns      string
		QueryWhere        template.HTML
		QueryLimit        int
		QueryInternalSort string
		QueryExternalSort string
	}

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
		id, serial, err := model.DecodeToken(input.Paginator.NextToken)
		if err != nil {
			slog.Error("repository.Users.Select", "error", err)
			span.SetStatus(codes.Error, "invalid token")
			span.RecordError(err)
			ref.metrics.repositoryCalls.Add(ctx, 1,
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
		id, serial, err := model.DecodeToken(input.Paginator.PrevToken)
		if err != nil {
			slog.Error("repository.Users.Select", "error", err)
			span.SetStatus(codes.Error, "invalid token")
			span.RecordError(err)
			ref.metrics.repositoryCalls.Add(ctx, 1,
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
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	query := tpl.String()
	slog.Debug("repository.Users.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query)
	if err != nil {
		slog.Error("repository.Users.Select", "error", err)
		span.SetStatus(codes.Error, "failed to select all users")
		span.RecordError(err)
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, err
	}
	defer rows.Close()

	var items []model.User
	for rows.Next() {
		var item model.User

		scanFields := make([]interface{}, 0)

		if input.Fields == "" {
			scanFields = []interface{}{
				&item.ID,
				&item.FirstName,
				&item.LastName,
				&item.Email,
				&item.PasswordHash,
				&item.Disabled,
				&item.CreatedAt,
				&item.UpdatedAt,
				&item.SerialID,
			}
		} else {
			var idFound bool

			inputFields := strings.SplitSeq(input.Fields, ",")

			for field := range inputFields {
				switch field {
				case "id":
					scanFields = append(scanFields, &item.ID)
					idFound = true
				case "first_name":
					scanFields = append(scanFields, &item.FirstName)
				case "last_name":
					scanFields = append(scanFields, &item.LastName)
				case "email":
					scanFields = append(scanFields, &item.Email)
				case "password_hash":
					scanFields = append(scanFields, &item.PasswordHash)
				case "disabled":
					scanFields = append(scanFields, &item.Disabled)
				case "created_at":
					scanFields = append(scanFields, &item.CreatedAt)
				case "updated_at":
					scanFields = append(scanFields, &item.UpdatedAt)

				default:
					slog.Warn("repository.Users.Select", "what", "field not found", "field", field)
				}
			}

			// always select id and serial_id for pagination
			// if id is not selected, it will be added to the scanFields
			if !idFound {
				scanFields = append(scanFields, &item.ID)
			}

			scanFields = append(scanFields, &item.SerialID)
		}

		if err := rows.Scan(scanFields...); err != nil {
			slog.Error("repository.Users.Select", "error", err)
			span.SetStatus(codes.Error, "failed to scan user")
			span.RecordError(err)
			ref.metrics.repositoryCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("successful", "false"))...,
				),
			)

			return nil, err
		}

		items = append(items, item)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		slog.Error("repository.Users.Select", "error", rows.Err())
		span.SetStatus(codes.Error, "failed to scan user")
		span.RecordError(rows.Err())
		ref.metrics.repositoryCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("successful", "false"))...,
			),
		)

		return nil, rows.Err()
	}

	outLen := len(items)
	if outLen == 0 {
		slog.Warn("repository.Users.Select", "what", "no users found")
		return &model.SelectUsersOutput{
			Items:     make([]model.User, 0),
			Paginator: model.Paginator{},
		}, nil
	}

	slog.Debug("repository.Users.Select", "next_id", items[outLen-1].ID, "next_serial_id", items[outLen-1].SerialID)
	slog.Debug("repository.Users.Select", "prev_id", items[0].ID, "prev_serial_id", items[0].SerialID)

	nextToken, prevToken := model.GetTokens(
		outLen,
		input.Paginator.Limit,
		items[0].ID,
		items[0].SerialID,
		items[outLen-1].ID,
		items[outLen-1].SerialID,
	)

	ret := &model.SelectUsersOutput{
		Items: items,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	span.SetStatus(codes.Ok, "users selected successfully")
	ref.metrics.repositoryCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("successful", "true"))...,
		),
	)

	return ret, nil
}
