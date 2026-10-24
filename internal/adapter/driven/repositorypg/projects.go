package repositorypg

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

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
	"github.com/slashdevops/go-rest-api-service-template/pkg/cslog"

	"go.opentelemetry.io/otel/attribute"
)

type ProjectsRepositoryConfig struct {
	DB              *pgxpool.Pool
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
}

// ProjectsRepository is a PostgreSQL store.
type ProjectsRepository struct {
	db              *pgxpool.Pool
	ot              *o11y.OpenTelemetry
	metrics         *o11y.LayerMetrics
	metricsMetadata o11y.Metadata
	metricsPrefix   string
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
}

// NewProjectsRepository creates a new ProjectsRepository.
func NewProjectsRepository(conf ProjectsRepositoryConfig) (*ProjectsRepository, error) {
	if conf.DB == nil {
		return nil, &domain.InvalidDBConfigurationError{Message: "invalid database configuration. It is nil"}
	}

	if conf.MaxPingTimeout < domain.ValidDatabaseMinPingTimeout || conf.MaxPingTimeout > domain.ValidDatabaseMaxPingTimeout {
		return nil, &domain.InvalidDBMaxPingTimeoutError{
			Message: fmt.Sprintf(
				"invalid max ping timeout. It must be between %d and %d",
				domain.ValidDatabaseMinPingTimeout, domain.ValidDatabaseMaxPingTimeout,
			),
		}
	}

	if conf.MaxQueryTimeout < domain.ValidDatabaseMinQueryTimeout || conf.MaxQueryTimeout > domain.ValidDatabaseMaxQueryTimeout {
		return nil, &domain.InvalidDBMaxQueryTimeoutError{
			Message: fmt.Sprintf(
				"invalid max query timeout. It must be between %d and %d",
				domain.ValidDatabaseMinQueryTimeout, domain.ValidDatabaseMaxQueryTimeout,
			),
		}
	}

	if conf.OT == nil {
		return nil, &domain.InvalidOTConfigurationError{Message: "invalid OpenTelemetry configuration. It is nil"}
	}

	ref := &ProjectsRepository{
		db:              conf.DB,
		maxPingTimeout:  conf.MaxPingTimeout,
		maxQueryTimeout: conf.MaxQueryTimeout,
		ot:              conf.OT,
		metricsMetadata: o11y.Metadata{
			Layer:  AppLayer,
			Domain: "Projects",
			Action: "NewProjectsRepository",
		},
	}
	if conf.MetricsPrefix != "" {
		ref.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		ref.metricsPrefix += "_"
	}

	metrics, err := o11y.NewLayerMetrics(ref.ot.Metrics.Meter, ref.metricsPrefix)
	if err != nil {
		return nil, err
	}

	ref.metrics = metrics

	return ref, nil
}

// DriverName returns the name of the driver.
func (ref *ProjectsRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *ProjectsRepository) PingContext(ctx context.Context) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxPingTimeout, ref.metricsMetadata, "PingContext")
	defer cancel()
	defer span.End()

	if err := ref.db.Ping(ctx); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "database ping successful")

	return nil
}

// Insert a new project into the database.
func (ref *ProjectsRepository) Insert(ctx context.Context, input *domain.InsertProjectInput) (err error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "Insert")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("project.id", input.ID.String()))

	tx, err := ref.db.Begin(ctx)
	if err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}
	defer endTx(ctx, tx, &err, func(e error) error {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(e, input), ref.metrics, attrs)
	})

	query1 := `
        INSERT INTO projects (id, name, description, disabled)
        VALUES ($1, $2, $3, $4);
    `

	cslog.Trace(ctx, "repository.Projects.Insert",
		"query", prettyPrint(
			query1,
			input.ID.String(),
			input.Name,
			input.Description,
			input.Disabled,
		),
	)

	_, err = tx.Exec(ctx, query1,
		input.ID,
		input.Name,
		input.Description,
		input.Disabled,
	)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
	}

	query2 := `
        INSERT INTO projects_users (projects_id, users_id)
        VALUES ($1, $2);
    `

	cslog.Trace(ctx, "repository.Projects.Insert",
		"query", prettyPrint(
			query2,
			input.ID.String(),
			input.UserID.String(),
		),
	)

	_, err = tx.Exec(ctx, query2, input.ID, input.UserID)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
	}

	slog.DebugContext(ctx, "repository.Projects.Insert", "project.id", input.ID)
	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "project inserted successfully", attribute.String("project.id", input.ID.String()))

	return nil
}

func (ref *ProjectsRepository) UpdateByID(ctx context.Context, input *domain.UpdateProjectInput) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "UpdateByID")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("project.id", input.ID.String()))

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

	query := `
        UPDATE projects SET
            name        = COALESCE(NULLIF($3, ''), name),
            description = COALESCE(NULLIF($4, ''), description),
            disabled    = COALESCE($5, disabled),
            updated_at  = CURRENT_TIMESTAMP
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

	cslog.Trace(ctx, "sql", "query", prettyPrint(query, args...))

	result, err := ref.db.Exec(ctx, query, args...)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
	}

	if result.RowsAffected() == 0 {
		errorType := &domain.ProjectNotFoundError{ID: input.ID}
		return o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "project updated successfully", attribute.String("project.id", input.ID.String()))

	return nil
}

// DeleteByID deletes the project with the specified ID.
func (ref *ProjectsRepository) DeleteByID(ctx context.Context, input *domain.DeleteProjectInput) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "DeleteByID")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("project.id", input.ID.String()))

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

	result, err := ref.db.Exec(ctx, query, input.ID, input.UserID)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
	}

	// Nothing matched: either no such project, or the caller may not delete it.
	// Reporting that matters beyond the status code — the use-case decrements
	// the owner's resource counter after a successful delete, so silently
	// returning nil here let a caller lower their own usage by deleting ids that
	// do not exist, and then create past their limit indefinitely.
	//
	// The handler already maps this error to the 404 its annotations declare;
	// until now that branch was unreachable.
	if result.RowsAffected() == 0 {
		errorType := &domain.ProjectNotFoundError{Message: fmt.Sprintf("project with ID %s not found", input.ID.String())}
		return o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "project deleted successfully", attribute.String("project.id", input.ID.String()))

	return nil
}

// querySelectMembership is [ProjectsRepository.SelectMembership]'s
// statement: the project $1, the caller $2.
const querySelectMembership = `
    SELECT
        COALESCE((SELECT admin FROM users WHERE id = $2), FALSE)
            AND EXISTS (SELECT 1 FROM projects WHERE id = $1) AS is_admin,
        EXISTS (
            SELECT 1 FROM projects_users
            WHERE projects_id = $1 AND users_id = $2
        ) AS is_member;
`

// SelectMembership answers the one question every project-scoped request
// has to ask before touching anything: may this caller act on this project?
//
// The OPA policy authorises on (user, method, path) and expands the "*" in a
// permission to "any uuid", so a grant on /projects/*/embedding_configs opened
// every project. Membership is data the database owns, so it is asked here,
// once per request, from the middleware. Two facts in one round trip: the
// admin flag, which admits the caller to every project that exists, and the
// link row.
//
// The flag admits to a project that is there, and to nothing else. It
// admitted to any id, so every list under a project that does not exist
// answered the administrator account 200 and an empty list (and the
// project's resource limits, for a project with no row), while a create
// under the same path answered 404. A project that is not there is no
// membership, and the middleware answers its 404, the one a non-member gets,
// for everybody. A member's link row says the project is there.
func (ref *ProjectsRepository) SelectMembership(ctx context.Context, id, userID uuid.UUID) (domain.ProjectMembership, error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "SelectMembership")
	defer cancel()
	defer span.End()

	span.SetAttributes(
		attribute.String("project.id", id.String()),
		attribute.String("user.id", userID.String()),
	)

	if id == uuid.Nil() || userID == uuid.Nil() {
		return domain.ProjectMembershipNone, o11y.RecordError(ctx, span, start, &domain.InvalidInputError{Message: "project ID and user ID are required"}, ref.metrics, attrs)
	}

	var isAdmin, isMember bool
	if err := ref.db.QueryRow(ctx, querySelectMembership, id, userID).Scan(&isAdmin, &isMember); err != nil {
		return domain.ProjectMembershipNone, o11y.RecordError(ctx, span, start, ref.handlePgError(err, nil), ref.metrics, attrs)
	}

	membership := domain.ProjectMembershipNone
	switch {
	case isAdmin:
		membership = domain.ProjectMembershipAdmin
	case isMember:
		membership = domain.ProjectMembershipMember
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "membership resolved",
		attribute.String("project.membership", membership.String()),
	)

	return membership, nil
}

// SelectByIDByUserID returns the project with the specified ID.
func (ref *ProjectsRepository) SelectByIDByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Project, error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "SelectByIDByUserID")
	defer cancel()
	defer span.End()

	span.SetAttributes(attribute.String("project.id", id.String()))

	if !domain.IsUUIDV7(id) {
		errorType := &domain.InvalidProjectIDError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	if userID == uuid.Nil() {
		errorType := &domain.InvalidUserIDError{Message: "userID is nil"}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	if id == uuid.Nil() && userID == uuid.Nil() {
		errorType := &domain.InvalidInputError{Message: "both project ID and user ID are nil"}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
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
        FROM view_projects_users AS vp
        WHERE vp.id = $1 AND vp.user_id = $2;
    `

	cslog.Trace(ctx, "repository.Projects.SelectByIDByUserID",
		"query", prettyPrint(
			query,
			id.String(),
			userID.String(),
		),
	)

	row := ref.db.QueryRow(ctx, query, id.String(), userID.String())

	var item domain.Project

	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Disabled,
		&item.System,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(err, id), ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "project selected successfully", attribute.String("project.id", id.String()))
	return &item, nil
}

// SelectByUserID lists the projects of input.UserID that are
// input.CallerID's too.
//
// The route it serves, GET /users/{user_id}/projects, has no project in its
// path, so no membership check runs for it, and the list is of somebody
// else's projects: it is scoped to who asks. view_projects_users has every
// project for an account with the `admin` flag and an account's memberships
// for anyone else, so the administrator account reads a user's whole list, a
// caller asking about themselves their own, and anyone else the projects
// they share with that user.
//
// A list with no caller is refused, never answered unscoped.
func (ref *ProjectsRepository) SelectByUserID(ctx context.Context, input *domain.SelectProjectsInput) (*domain.SelectProjectsOutput, error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "SelectByUserID")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if input.CallerID == uuid.Nil() {
		errorValue := &domain.InvalidInputError{Message: "a user's projects were asked for with no caller"}
		return nil, o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
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
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, domain.ProjectFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "vp.serial_id DESC, vp.id DESC"
	} else {
		sortQuery = injectPrefixToSortFields(sqlFieldsPrefix, input.Sort, domain.ProjectSortFields)
	}

	// query template
	queryTemplate := `
        WITH vp AS (
            SELECT
                {{.QueryColumns}}
            FROM view_projects_users AS vp
            WHERE vp.user_id = $1
                AND EXISTS (
                    SELECT 1 FROM view_projects_users AS callers
                    WHERE callers.user_id = $2 AND callers.id = vp.id
                )
                {{ .QueryWhere }}
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM vp ORDER BY {{.QueryExternalSort}}
    `

	// struct to hold the query values
	var queryValues struct {
		QueryColumns      template.HTML
		QueryWhere        template.HTML
		QueryInternalSort string
		QueryExternalSort string
		QueryLimit        int
	}

	// default values
	queryValues.QueryColumns = template.HTML(fieldsStr)
	queryValues.QueryWhere = template.HTML(filterQuery)
	queryValues.QueryLimit = input.Paginator.Limit + 1 // Fetch one extra item
	queryValues.QueryInternalSort = "vp.serial_id DESC, vp.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := domain.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("vp", tokenDirection, id, serial, filterQuery, true)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	query := tpl.String()
	cslog.Trace(ctx, "sql", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query, input.UserID, input.CallerID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(filterRefusal(filterQuery, query, err), input), ref.metrics, attrs)
	}
	defer rows.Close()

	var fetchedItems []domain.Project
	for rows.Next() {
		var item domain.Project

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(filterRefusal(filterQuery, query, err), input), ref.metrics, attrs)
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.WarnContext(ctx, "no projects found")
		return &domain.SelectProjectsOutput{
			Items:     make([]domain.Project, 0),
			Paginator: domain.Paginator{},
		}, nil
	}

	repoFoundMoreForNextQuery := false
	repoFoundMoreForPrevQuery := false

	switch tokenDirection {
	case domain.TokenDirectionNext: // Used 'next' token to get current page
		repoFoundMoreForPrevQuery = true // Came from a previous page
		repoFoundMoreForNextQuery = hasMore
	case domain.TokenDirectionPrev: // Used 'prev' token to get current page
		repoFoundMoreForNextQuery = true // Came from a next page
		repoFoundMoreForPrevQuery = hasMore
	default: // Initial load (tokenDirection == domain.TokenDirectionInvalid)
		repoFoundMoreForNextQuery = hasMore
		// repoFoundMoreForPrevQuery remains false, GetTokens will handle it
	}

	// Calculate min/max serial_id from entire result set to ensure correct pagination boundaries
	// regardless of external sorting order (e.g., "name ASC")
	minSerialID := displayItems[0].SerialID
	maxSerialID := displayItems[0].SerialID
	minSerialItem := displayItems[0]
	maxSerialItem := displayItems[0]

	for _, item := range displayItems {
		if item.SerialID < minSerialID {
			minSerialID = item.SerialID
			minSerialItem = item
		}
		if item.SerialID > maxSerialID {
			maxSerialID = item.SerialID
			maxSerialItem = item
		}
	}

	nextToken, prevToken := domain.GetTokens(
		outLen,
		maxSerialItem.ID, // Use item with MAX serial for prev token
		maxSerialItem.SerialID,
		minSerialItem.ID, // Use item with MIN serial for next token
		minSerialItem.SerialID,
		tokenDirection,
		repoFoundMoreForNextQuery,
		repoFoundMoreForPrevQuery,
	)
	ret := &domain.SelectProjectsOutput{
		Items: displayItems,
		Paginator: domain.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "projects selected successfully")

	return ret, nil
}

func (ref *ProjectsRepository) Select(ctx context.Context, input *domain.SelectProjectsInput) (*domain.SelectProjectsOutput, error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "Select")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if error := input.Validate(); error != nil {
		return nil, o11y.RecordError(ctx, span, start, error, ref.metrics, attrs)
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
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, domain.ProjectFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "vp.serial_id DESC, vp.id DESC"
	} else {
		sortQuery = injectPrefixToSortFields(sqlFieldsPrefix, input.Sort, domain.ProjectSortFields)
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
		QueryInternalSort string
		QueryExternalSort string
		QueryLimit        int
	}

	// default values
	queryValues.QueryColumns = template.HTML(fieldsStr)
	queryValues.QueryWhere = template.HTML(filterQuery)
	queryValues.QueryLimit = input.Paginator.Limit + 1 // Fetch one extra item
	queryValues.QueryInternalSort = "vp.serial_id DESC, vp.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := domain.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("vp", tokenDirection, id, serial, filterQuery, true)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	query := tpl.String()
	cslog.Trace(ctx, "repository.Projects.Select",
		"query", prettyPrint(
			query, input.UserID.String(),
		),
	)

	// execute the query
	rows, err := ref.db.Query(ctx, query, input.UserID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(filterRefusal(filterQuery, query, err), input), ref.metrics, attrs)
	}
	defer rows.Close()

	var fetchedItems []domain.Project
	for rows.Next() {
		var item domain.Project

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(filterRefusal(filterQuery, query, err), input), ref.metrics, attrs)
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.WarnContext(ctx, "no projects found")
		return &domain.SelectProjectsOutput{
			Items:     make([]domain.Project, 0),
			Paginator: domain.Paginator{},
		}, nil
	}

	repoFoundMoreForNextQuery := false
	repoFoundMoreForPrevQuery := false

	switch tokenDirection {
	case domain.TokenDirectionNext: // Used 'next' token to get current page
		repoFoundMoreForPrevQuery = true // Came from a previous page
		repoFoundMoreForNextQuery = hasMore
	case domain.TokenDirectionPrev: // Used 'prev' token to get current page
		repoFoundMoreForNextQuery = true // Came from a next page
		repoFoundMoreForPrevQuery = hasMore
	default: // Initial load (tokenDirection == domain.TokenDirectionInvalid)
		repoFoundMoreForNextQuery = hasMore
		// repoFoundMoreForPrevQuery remains false, GetTokens will handle it
	}

	// Calculate min/max serial_id from entire result set to ensure correct pagination boundaries
	// regardless of external sorting order (e.g., "name ASC")
	minSerialID := displayItems[0].SerialID
	maxSerialID := displayItems[0].SerialID
	minSerialItem := displayItems[0]
	maxSerialItem := displayItems[0]

	for _, item := range displayItems {
		if item.SerialID < minSerialID {
			minSerialID = item.SerialID
			minSerialItem = item
		}
		if item.SerialID > maxSerialID {
			maxSerialID = item.SerialID
			maxSerialItem = item
		}
	}

	nextToken, prevToken := domain.GetTokens(
		outLen,
		maxSerialItem.ID, // Use item with MAX serial for prev token
		maxSerialItem.SerialID,
		minSerialItem.ID, // Use item with MIN serial for next token
		minSerialItem.SerialID,
		tokenDirection,
		repoFoundMoreForNextQuery,
		repoFoundMoreForPrevQuery,
	)
	ret := &domain.SelectProjectsOutput{
		Items: displayItems,
		Paginator: domain.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "projects selected successfully")

	return ret, nil
}

func (ref *ProjectsRepository) LinkUsers(ctx context.Context, input *domain.LinkUsersToProjectInput) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "LinkUsers")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	// Prepare arrays for UNNEST
	projectIDs := make([]string, len(input.UserIDs))
	usersIDs := make([]string, len(input.UserIDs))
	for i, userID := range input.UserIDs {
		projectIDs[i] = input.ProjectID.String()
		usersIDs[i] = userID.String()
	}

	query := `
        -- insert the new users. DISTINCT: a body may name one twice, and
        -- ON CONFLICT DO UPDATE refuses to touch one row twice in one
        -- statement (SQLSTATE 21000), which was a 500.
        INSERT INTO projects_users (projects_id, users_id)
        SELECT DISTINCT * FROM UNNEST($1::uuid[], $2::uuid[])
        ON CONFLICT (projects_id, users_id)
        DO UPDATE SET updated_at = NOW();
    `

	cslog.Trace(ctx, "sql", "query", prettyPrint(query, projectIDs, usersIDs))

	_, err := ref.db.Exec(ctx, query, projectIDs, usersIDs)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "projects linked successfully")

	return nil
}

func (ref *ProjectsRepository) UnlinkUsers(ctx context.Context, input *domain.UnlinkUsersFromProjectInput) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "UnlinkUsers")
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	// Prepare arrays for UNNEST
	userIDs := make([]string, len(input.UserIDs))
	for i, userID := range input.UserIDs {
		userIDs[i] = userID.String()
	}

	query := `
      -- Unlink users from projects
      DELETE FROM projects_users
      WHERE projects_id = $1 AND users_id IN (SELECT unnest($2::uuid[]));
    `

	cslog.Trace(ctx, "sql", "query", prettyPrint(query, input.ProjectID.String(), userIDs))

	_, err := ref.db.Exec(ctx, query, input.ProjectID.String(), userIDs)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err, input), ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "projects unlinked successfully")

	return nil
}

// handlePgError maps PostgreSQL errors to domain-specific errors.
// Returns the appropriate domain error or the original error if no mapping exists.
func (ref *ProjectsRepository) handlePgError(err error, input any) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.Message, "_pkey") {
				switch v := input.(type) {
				case *domain.InsertProjectInput:
					return &domain.ProjectAlreadyExistsError{ID: v.ID}
				case *domain.UpdateProjectInput:
					return &domain.ProjectAlreadyExistsError{ID: v.ID}
				case *domain.LinkUsersToProjectInput: //
					return &domain.ProjectAlreadyExistsError{ID: v.ProjectID, Message: "Project already exists"}
				case uuid.UUID:
					return &domain.ProjectAlreadyExistsError{ID: v}
				}
			}

			if strings.Contains(pgErr.Message, "projects_name") || strings.Contains(pgErr.Message, "name") {
				switch v := input.(type) {
				case *domain.InsertProjectInput:
					return &domain.ProjectAlreadyExistsError{Name: v.Name}
				case *domain.UpdateProjectInput:
					if v.Name != nil {
						return &domain.ProjectAlreadyExistsError{Name: *v.Name}
					}
				}
			}

			if strings.Contains(pgErr.Message, "_users_id_fkey") {
				switch v := input.(type) {
				case *domain.UpdateProjectInput:
					return &domain.UserNotFoundError{ID: v.ID}
				case uuid.UUID:
					return &domain.UserNotFoundError{ID: v}
				}
			}

			if strings.Contains(pgErr.Message, "_projects_id_fkey") {
				switch v := input.(type) {
				case *domain.UpdateProjectInput:
					return &domain.ProjectNotFoundError{ID: v.ID}
				case uuid.UUID:
					return &domain.ProjectNotFoundError{ID: v}
				}
			}
		case "23503": // Foreign key violation
			if strings.Contains(pgErr.Message, "_fkey") {
				switch v := input.(type) {
				case *domain.LinkUsersToProjectInput: // works too for model.Unlink
					return &domain.ProjectNotFoundError{ID: v.ProjectID, Message: "Project not found"}
				}
			}
		case "P0001": // Raised exception
			if strings.Contains(pgErr.Message, "updated") {
				switch v := input.(type) {
				case *domain.UpdateProjectInput:
					return &domain.SystemProjectError{ID: v.ID}
				case uuid.UUID:
					return &domain.SystemProjectError{ID: v}
				}
			}

			if strings.Contains(pgErr.Message, "deleted") {
				switch v := input.(type) {
				case *domain.DeleteProjectInput:
					return &domain.SystemProjectError{ID: v.ID}
				case uuid.UUID:
					return &domain.SystemProjectError{ID: v}
				}
			}
		case "22021": // invalid byte sequence for encoding
			return &domain.InvalidByteSequenceError{Message: pgInvalidByteSequence}
		case "08P01": // invalid message format
			return &domain.InvalidMessageFormatError{Message: pgInvalidMessageFormat}
		case "42703": // undefined column
			return &domain.UndefinedColumnError{Message: pgUndefinedColumn}
		case "42804": // datatype mismatch
			return &domain.DatatypeMismatchError{Message: pgDatatypeMismatch}
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return &domain.ProjectNotFoundError{}
	}

	return err
}

func (ref *ProjectsRepository) buildScanFields(item *domain.Project, requestedFields string) []any {
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
			slog.Warn("field not found while building the scan list", "field", field)
		}
	}

	// Always include ID and SerialID fields for pagination
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)
	return scanFields
}
