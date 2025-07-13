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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ProjectsRepositoryConfig struct {
	DB              *pgxpool.Pool
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type projectsRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// ProjectsRepository is a PostgreSQL store.
type ProjectsRepository struct {
	db              *pgxpool.Pool
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         projectsRepositoryMetrics
}

// NewProjectsRepository creates a new ProjectsRepository.
func NewProjectsRepository(conf ProjectsRepositoryConfig) (*ProjectsRepository, error) {
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

	repo := &ProjectsRepository{
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
		metric.WithDescription("The number of calls to the project repository"),
	)
	if err != nil {
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (ref *ProjectsRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *ProjectsRepository) PingContext(ctx context.Context) error {
	ctx, span, _, cancel := ref.setupContext(ctx, "repository.Projects.PingContext", ref.maxPingTimeout)
	defer cancel()
	defer span.End()

	return ref.db.Ping(ctx)
}

// Insert a new project into the database.
func (ref *ProjectsRepository) Insert(ctx context.Context, input *model.InsertProjectInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Projects.Insert", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert")
	}

	span.SetAttributes(attribute.String("project.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert")
	}

	tx, txErr := ref.db.Begin(ctx)
	if txErr != nil {
		return o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert", "failed to begin transaction")
	}
	defer func() {
		if txErr != nil {
			if err := tx.Rollback(ctx); err != nil {
				e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert", "failed to rollback transaction")
				slog.Error("repository.Projects.Insert", "error", e)
			}
		} else {
			if err := tx.Commit(ctx); err != nil {
				e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert", "failed to commit transaction")
				if e != nil {
					slog.Error("repository.Projects.Insert", "error", e)
				}
			}
		}
	}()

	query1 := `
        INSERT INTO projects (id, name, description, disabled)
        VALUES ($1, $2, $3, $4);
    `

	slog.Debug("repository.Projects.Insert", "query", prettyPrint(query1, input.ID, input.Name, input.Description, input.Disabled))

	_, txErr = tx.Exec(ctx, query1,
		input.ID,
		input.Name,
		input.Description,
		input.Disabled,
	)
	if txErr != nil {
		txErr := o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert")
		return ref.handlePgError(txErr, input)
	}

	query2 := `
        INSERT INTO projects_users (projects_id, users_id)
        VALUES ($1, $2);
    `

	slog.Debug("repository.Projects.Insert", "query", prettyPrint(query2, input.ID, input.UserID))

	_, txErr = tx.Exec(ctx, query2, input.ID, input.UserID)
	if txErr != nil {
		txErr := o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Insert")
		return ref.handlePgError(txErr, input)
	}

	slog.Debug("repository.Projects.Insert", "project.id", input.ID)
	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "project inserted successfully", attribute.String("project.id", input.ID.String()))

	return nil
}

func (ref *ProjectsRepository) UpdateByID(ctx context.Context, input *model.UpdateProjectInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Projects.UpdateByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.UpdateByID")
	}

	span.SetAttributes(attribute.String("project.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.UpdateByID")
	}

	args := []any{input.ID, input.UserID}

	if input.Name != nil && *input.Name != "" {
		args = append(args, *input.Name)
	} else {
		args = append(args, nil)
	}

	if input.Description != nil && *input.Description != "" {
		args = append(args, *input.Description)
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
		slog.Error("repository.Projects.UpdateByID", "error", err)
		return err
	}

	args = append(args, updatedAt)

	query := `
        UPDATE projects SET
            name        = COALESCE(NULLIF($3, ''), name),
            description = COALESCE(NULLIF($4, ''), description),
            disabled    = COALESCE($5, disabled),
            updated_at  = $6
        WHERE id = $1
        AND (
            -- 2. Add a security check that must pass
            -- Condition A: The user is an admin
            (SELECT admin FROM users WHERE id = $2) = TRUE
            OR
            -- Condition B: The user is assigned to this specific project
            EXISTS (
                SELECT 1
                FROM projects_users
                WHERE projects_id = $1
                AND users_id = $2
            )
        );
    `

	slog.Debug("repository.Projects.UpdateByID", "query", prettyPrint(query, args...))

	result, err := ref.db.Exec(ctx, query, args...)
	if err != nil {
		return ref.handlePgError(o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.UpdateByID"), input)
	}

	if result.RowsAffected() == 0 {
		errorType := &model.ProjectNotFoundError{ID: input.ID}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.UpdateByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "project updated successfully", attribute.String("project.id", input.ID.String()))

	return nil
}

// DeleteByID deletes the project with the specified ID.
func (ref *ProjectsRepository) DeleteByID(ctx context.Context, input *model.DeleteProjectInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Projects.DeleteByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.DeleteByID")
	}

	span.SetAttributes(attribute.String("project.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.DeleteByID")
	}

	query := `
        DELETE FROM projects
        WHERE id = $1
        AND (
            -- 2. Add a security check that must pass
            -- Condition A: The user is an admin
            (SELECT admin FROM users WHERE id = $2) = TRUE
            OR
            -- Condition B: The user is assigned to this specific project
            EXISTS (
                SELECT 1
                FROM projects_users
                WHERE projects_id = $1
                AND users_id = $2
            )
        );
    `

	_, err := ref.db.Exec(ctx, query, input.ID, input.UserID)
	if err != nil {
		e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.DeleteByID")
		return ref.handlePgError(e, input)
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "project deleted successfully", attribute.String("project.id", input.ID.String()))

	return nil
}

// SelectByID returns the project with the specified ID.
func (ref *ProjectsRepository) SelectByID(ctx context.Context, id, userID uuid.UUID) (*model.Project, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Projects.SelectByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	span.SetAttributes(attribute.String("project.id", id.String()))

	if id == uuid.Nil {
		errorType := &model.InvalidProjectIDError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.SelectByID")
	}

	query := `
        SELECT
            vp.id,
            vp.name,
            vp.description,
            vp.disabled,
            vp.system,
            vp.created_at,
            vp.updated_at
        FROM view_projects_users vp
        WHERE vp.id = $1 AND vp.user_id = $2;
    `

	slog.Debug("repository.Projects.SelectByID", "query", prettyPrint(query, id.String(), userID.String()))

	row := ref.db.QueryRow(ctx, query, id.String(), userID.String())

	var item model.Project

	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Disabled,
		&item.System,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorType := &model.ProjectNotFoundError{ID: id}
			return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.SelectByID")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.SelectByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "project selected successfully", attribute.String("project.id", id.String()))
	return &item, nil
}

func (ref *ProjectsRepository) Select(ctx context.Context, input *model.SelectProjectsInput) (*model.SelectProjectsOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Projects.Select", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Select")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "vp."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"disabled",
		"system",
		"created_at",
		"updated_at",
		"serial_id",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.ProjectFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "vp.serial_id DESC, vp.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH vp AS (
            SELECT
                {{.QueryColumns}}
            FROM view_projects_users AS vp
            WHERE vp.user_id = $1
                {{ .QueryWhere }}
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM vp ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "vp.serial_id DESC, vp.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Select", "failed to get paginator direction")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("vp", tokenDirection, id, serial, filterQuery, false)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Select", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Projects.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query, input.UserID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Select", "failed to select all projects")
	}
	defer rows.Close()

	var fetchedItems []model.Project
	for rows.Next() {
		var item model.Project

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Select", "failed to scan project")
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Projects.Select", "failed to scan fields")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.Warn("repository.Projects.Select", "what", "no projects found")
		return &model.SelectProjectsOutput{
			Items:     make([]model.Project, 0),
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
	ret := &model.SelectProjectsOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "projects selected successfully")

	return ret, nil
}

// Helper functions for common patterns

// setupContext creates a context with timeout and starts a span with standard attributes.
// Returns the new context, span, and common metric attributes.
func (ref *ProjectsRepository) setupContext(ctx context.Context, operation string, timeout time.Duration) (context.Context, trace.Span, []attribute.KeyValue, context.CancelFunc) {
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
func (ref *ProjectsRepository) handlePgError(err error, input any) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.Message, "_pkey") {
				switch v := input.(type) {
				case *model.InsertProjectInput:
					return &model.ProjectIDAlreadyExistsError{ID: v.ID}
				case *model.UpdateProjectInput:
					return &model.ProjectIDAlreadyExistsError{ID: v.ID}
				case uuid.UUID:
					return &model.ProjectIDAlreadyExistsError{ID: v}
				}
			}

			if strings.Contains(pgErr.Message, "projects_name") || strings.Contains(pgErr.Message, "name") {
				switch v := input.(type) {
				case *model.InsertProjectInput:
					return &model.ProjectNameAlreadyExistsError{Name: v.Name}
				case *model.UpdateProjectInput:
					if v.Name != nil {
						return &model.ProjectNameAlreadyExistsError{Name: *v.Name}
					}
				}
			}

			if strings.Contains(pgErr.Message, "_users_id_fkey") {
				switch v := input.(type) {
				case *model.UpdateProjectInput:
					return &model.UserNotFoundError{ID: v.ID.String()}
				case uuid.UUID:
					return &model.UserNotFoundError{ID: v.String()}
				}
			}

			if strings.Contains(pgErr.Message, "_projects_id_fkey") {
				switch v := input.(type) {
				case *model.UpdateProjectInput:
					return &model.ProjectNotFoundError{ID: v.ID}
				case uuid.UUID:
					return &model.ProjectNotFoundError{ID: v}
				}
			}
		case "P0001": // Raised exception
			if strings.Contains(pgErr.Message, "updated") {
				switch v := input.(type) {
				case *model.UpdateProjectInput:
					return &model.SystemProjectError{ID: v.ID}
				case uuid.UUID:
					return &model.SystemProjectError{ID: v}
				}
			}

			if strings.Contains(pgErr.Message, "deleted") {
				switch v := input.(type) {
				case *model.DeleteProjectInput:
					return &model.SystemProjectError{ID: v.ID}
				case uuid.UUID:
					return &model.SystemProjectError{ID: v}
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

func (ref *ProjectsRepository) buildScanFields(item *model.Project, requestedFields string) []any {
	scanFields := make([]any, 0)

	if requestedFields == "" {
		// All fields were requested
		return []any{
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Disabled,
			&item.System,
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
		case "name":
			scanFields = append(scanFields, &item.Name)
		case "description":
			scanFields = append(scanFields, &item.Description)
		case "disabled":
			scanFields = append(scanFields, &item.Disabled)
		case "system":
			scanFields = append(scanFields, &item.System)
		case "created_at":
			scanFields = append(scanFields, &item.CreatedAt)
		case "updated_at":
			scanFields = append(scanFields, &item.UpdatedAt)

		default:
			slog.Warn("repository.Projects.buildScanFields", "what", "field not found", "field", field)
		}
	}

	// Always include ID and SerialID fields for pagination
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)
	return scanFields
}
