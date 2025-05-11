package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
		return nil, &model.InvalidDBConfigurationError{Message: "invalid database configuration. It is nil"}
	}

	if conf.MaxPingTimeout < 10*time.Millisecond {
		return nil, &model.InvalidDBMaxPingTimeoutError{Message: "invalid max ping timeout. It must be greater than 10 millisecond"}
	}

	if conf.MaxQueryTimeout < 10*time.Millisecond {
		return nil, &model.InvalidDBMaxQueryTimeoutError{Message: "invalid max query timeout. It must be greater than 10 millisecond"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "invalid OpenTelemetry configuration. It is nil"}
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
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Users.Insert", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InputIsInvalidError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Insert")
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Insert")
	}

	query := `
        INSERT INTO users (id, first_name, last_name, email, password_hash, disabled)
        VALUES ($1, $2, $3, $4, $5, $6);
    `

	slog.Debug("repository.Users.Insert", "query", prettyPrint(query))

	_, err := ref.db.Exec(ctx, query,
		input.ID,
		input.FirstName,
		input.LastName,
		input.Email,
		input.PasswordHash,
		input.Disabled,
	)
	if err != nil {
		return ref.handlePgError(o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Insert"), input)
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "user inserted successfully", attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersRepository) UpdateByID(ctx context.Context, input *model.UpdateUserInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Users.UpdateByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InputIsInvalidError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.UpdateByID")
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.UpdateByID")
	}

	args := []any{input.ID}

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

	updatedAt, err := time.Now().In(time.FixedZone("UTC", 0)).MarshalText()
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.UpdateByID", "failed to marshal time")
	}

	args = append(args, updatedAt)

	query := `
        UPDATE users SET
            first_name    = COALESCE(NULLIF($2, ''), first_name),
            last_name     = COALESCE(NULLIF($3, ''), last_name),
            email         = COALESCE(NULLIF($4, ''), email),
            password_hash = COALESCE(NULLIF($5, ''), password_hash),
            disabled      = COALESCE($6, disabled),
            updated_at    = COALESCE($7, updated_at)
        WHERE id = $1;
    `

	slog.Debug("repository.Users.UpdateByID", "query", prettyPrint(query))

	result, err := ref.db.Exec(ctx, query, args...)
	if err != nil {
		return ref.handlePgError(o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.UpdateByID", "query failed"), input)
	}

	if result.RowsAffected() == 0 {
		errorType := &model.UserNotFoundError{ID: input.ID.String()}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.UpdateByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "user updated successfully", attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersRepository) DeleteByID(ctx context.Context, input *model.DeleteUserInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Users.DeleteByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InputIsInvalidError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.DeleteByID")
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.DeleteByID")
	}

	query := `
        DELETE FROM users WHERE id = $1
    `

	slog.Debug("repository.Users.DeleteByID", "query", prettyPrint(query, input.ID.String()))

	result, err := ref.db.Exec(ctx, query, input.ID)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.DeleteByID", "query failed")
	}

	if result.RowsAffected() == 0 {
		// grateful return user was deleted, security reason, but log and record error
		errorType := &model.UserNotFoundError{ID: input.ID.String()}
		err := o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.DeleteByID")
		if err != nil {
			return nil
		}

		return nil
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "user deleted successfully", attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Users.SelectByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	if id == uuid.Nil {
		errorType := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByID")
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
		if errors.Is(err, sql.ErrNoRows) {
			errorType := &model.UserNotFoundError{ID: id.String()}
			return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByID")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "user selected successfully", attribute.String("user.id", id.String()))

	return &item, nil
}

func (ref *UsersRepository) SelectByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Users.SelectByEmail", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	if email == "" {
		errorType := &model.InvalidUserEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByEmail", "email is empty")
	}

	if len(email) < model.ValidUserEmailMinLength || len(email) > model.ValidUserEmailMaxLength {
		errorType := &model.InvalidUserEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByEmail")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		errorType := &model.InvalidUserEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByEmail")
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.UserNotFoundError{Email: email}
		}
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.SelectByEmail", "scan failed")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "user selected successfully", attribute.String("user.email", email))

	return &item, nil
}

func (ref *UsersRepository) Select(ctx context.Context, input *model.SelectUsersInput) (*model.SelectUsersOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Users.Select", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InputIsInvalidError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "usrs."
	fieldsArray := []string{
		"id",
		"first_name",
		"last_name",
		"email",
		"password_hash",
		"disabled",
		"created_at",
		"updated_at",
		"serial_id",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

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

	// query template
	queryTemplate := `
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
		QueryColumns      template.HTML
		QueryWhere        template.HTML
		QueryLimit        int
		QueryInternalSort string
		QueryExternalSort string
	}

	// default values
	queryValues.QueryColumns = template.HTML(fieldsStr)
	queryValues.QueryWhere = template.HTML(filterQuery)
	queryValues.QueryLimit = input.Paginator.Limit + 1 // Fetch one extra item
	queryValues.QueryInternalSort = "usrs.serial_id DESC, usrs.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select", "invalid token")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("usrs", tokenDirection, id, serial, filterQuery)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Users.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select", "failed to select all users")
	}
	defer rows.Close()

	var fetchedItems []model.User
	for rows.Next() {
		var item model.User

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select", "failed to scan user")
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Users.Select", "failed to scan all users")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		return &model.SelectUsersOutput{
			Items:     make([]model.User, 0),
			Paginator: model.Paginator{},
		}, nil
	}

	repoFoundMoreForNextQuery := false
	repoFoundMoreForPrevQuery := false

	switch tokenDirection {
	case model.TokenDirectionNext: // Used 'next' token to get current page
		repoFoundMoreForPrevQuery = true // Came from a previous page
		repoFoundMoreForNextQuery = hasMore
	case model.TokenDirectionPrev: // Used 'prev' token to get current page
		repoFoundMoreForNextQuery = true // Came from a next page
		repoFoundMoreForPrevQuery = hasMore
	default: // Initial load (tokenDirection == model.TokenDirectionInvalid)
		repoFoundMoreForNextQuery = hasMore
		// repoFoundMoreForPrevQuery remains false, GetTokens will handle it
	}

	nextToken, prevToken := model.GetTokens(
		outLen,
		displayItems[0].ID,
		displayItems[0].SerialID,
		displayItems[outLen-1].ID,
		displayItems[outLen-1].SerialID,
		tokenDirection,
		repoFoundMoreForNextQuery,
		repoFoundMoreForPrevQuery,
	)

	ret := &model.SelectUsersOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "users selected successfully")

	return ret, nil
}

// Helper functions for common patterns

// setupContext creates a context with timeout and starts a span with standard attributes.
// Returns the new context, span, and common metric attributes.
func (ref *UsersRepository) setupContext(ctx context.Context, operation string, timeout time.Duration) (context.Context, trace.Span, []attribute.KeyValue, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("driver", ref.DriverName()),
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes, cancel
}

// handlePgError maps PostgreSQL errors to domain-specific errors.
// Returns the appropriate domain error or the original error if no mapping exists.
func (ref *UsersRepository) handlePgError(err error, input any) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.Message, "_pkey") {
				switch v := input.(type) {
				case *model.InsertUserInput:
					return &model.UserAlreadyExistsError{ID: v.ID.String()}
				case *model.UpdateUserInput:
					return &model.UserAlreadyExistsError{ID: v.ID.String()}
				case uuid.UUID:
					return &model.UserAlreadyExistsError{ID: v.String()}
				}
			}

			if strings.Contains(pgErr.Message, "_email") {
				switch v := input.(type) {
				case *model.InsertUserInput:
					return &model.UserEmailAlreadyExistsError{Email: v.Email}
				case *model.UpdateUserInput:
					if v.Email != nil {
						return &model.UserEmailAlreadyExistsError{Email: *v.Email}
					}
				}
			}
		case "22021": // invalid byte sequence for encoding
			return &model.InvalidByteSequenceError{Message: pgErr.Message}
		case "08P01": // invalid message format
			return &model.InvalidMessageFormatError{Message: pgErr.Message}
		}
	}

	return err
}

// buildScanFields creates the scan targets for the result rows based on the requested fields.
func (ref *UsersRepository) buildScanFields(item *model.User, requestedFields string) []any {
	scanFields := make([]any, 0)

	if requestedFields == "" {
		// All fields were requested
		return []any{
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
	}

	var idFound bool
	inputFields := strings.SplitSeq(requestedFields, ",")

	for field := range inputFields {
		field = strings.TrimSpace(field)

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
			slog.Warn("repository.Users.buildScanFields", "what", "field not found", "field", field)
		}
	}

	// Always include ID and SerialID fields for pagination
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)
	return scanFields
}
