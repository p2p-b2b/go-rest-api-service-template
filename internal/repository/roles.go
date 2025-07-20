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

// RolesRepositoryConfig is the configuration for the RolesRepository.
type RolesRepositoryConfig struct {
	DB              *pgxpool.Pool
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type rolesRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// RolesRepository is a PostgreSQL store.
type RolesRepository struct {
	db              *pgxpool.Pool
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         rolesRepositoryMetrics
}

// NewRolesRepository creates a new RolesRepository.
func NewRolesRepository(conf RolesRepositoryConfig) (*RolesRepository, error) {
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

	repo := &RolesRepository{
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
		metric.WithDescription("The number of calls to the role repository"),
	)
	if err != nil {
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (ref *RolesRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *RolesRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxPingTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Roles.PingContext")
	defer span.End()

	return ref.db.Ping(ctx)
}

// Insert a new role into the database.
func (ref *RolesRepository) Insert(ctx context.Context, input *model.InsertRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.Insert", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Insert")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Insert")
	}

	query := `
        INSERT INTO roles (id, name, description)
        VALUES ($1, $2, $3);
    `

	slog.Debug("repository.Roles.Insert", "query", prettyPrint(query, input.ID.String(), input.Name, input.Description))

	_, err := ref.db.Exec(ctx, query,
		input.ID.String(),
		input.Name,
		input.Description,
	)
	if err != nil {
		err = ref.handlePgError(err, input)
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Insert")
	}

	slog.Debug("repository.Roles.Insert", "role.id", input.ID)
	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "role inserted successfully", attribute.String("role.id", input.ID.String()))

	return nil
}

// UpdateByID updates the role with the specified ID.
func (ref *RolesRepository) UpdateByID(ctx context.Context, input *model.UpdateRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.UpdateByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UpdateByID")
	}

	span.SetAttributes(attribute.String("role.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UpdateByID")
	}

	var args []string

	if input.Name != nil && *input.Name != "" {
		args = append(args, fmt.Sprintf("name='%s'", *input.Name))
	}

	if input.Description != nil && *input.Description != "" {
		args = append(args, fmt.Sprintf("description='%s'", *input.Description))
	}

	updatedAt, err := time.Now().In(time.FixedZone("UTC", 0)).MarshalText()
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UpdateByID")
	}

	args = append(args, fmt.Sprintf("updated_at='%s'", updatedAt))

	fields := strings.Join(args, ", ")

	queryString := fmt.Sprintf(`
        UPDATE roles
        SET
            %s
        WHERE id = '%s';
        `,
		fields,
		input.ID.String(),
	)

	slog.Debug("repository.Roles.UpdateByID", "query", prettyPrint(queryString))

	result, err := ref.db.Exec(ctx, queryString)
	if err != nil {
		return ref.handlePgError(o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UpdateByID"), input)
	}

	if result.RowsAffected() == 0 {
		errorType := &model.RoleNotFoundError{RoleID: input.ID.String()}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UpdateByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "role updated successfully", attribute.String("role.id", input.ID.String()))

	return nil
}

// DeleteByID deletes the role with the specified ID.
func (ref *RolesRepository) DeleteByID(ctx context.Context, input *model.DeleteRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.DeleteByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.DeleteByID")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.DeleteByID")
	}

	queryString := `
        DELETE FROM roles WHERE id = $1;
    `

	slog.Debug("repository.Roles.Delete", "query", prettyPrint(queryString))

	result, err := ref.db.Exec(ctx, queryString, input.ID.String())
	if err != nil {
		err = ref.handlePgError(err, input)
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.DeleteByID")
	}

	if result.RowsAffected() == 0 {
		// grateful return user was deleted, security reason, but log and record error
		errorType := &model.RoleNotFoundError{RoleID: input.ID.String()}
		e := o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.DeleteByID")
		slog.Error("repository.Roles.DeleteByID", "error", e, "role.id", input.ID.String())

		return nil
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "role deleted successfully", attribute.String("role.id", input.ID.String()))

	return nil
}

// SelectByID returns the role and its policies with the specified ID.
func (ref *RolesRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.SelectByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if id == uuid.Nil {
		invalidErr := &model.InvalidRoleIDError{Message: "invalid role ID"}
		return nil, o11y.RecordError(ctx, span, invalidErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByID")
	}

	query := `
        SELECT
            rls.id,
            rls.name,
            rls.description,
            rls.system,
            rls.auto_assign,
            rls.created_at,
            rls.updated_at
        FROM roles AS rls
        WHERE rls.id = $1
        GROUP BY rls.id;
    `

	slog.Debug("repository.Roles.SelectByID", "query", prettyPrint(query))

	row := ref.db.QueryRow(ctx, query, id.String())

	var item model.Role

	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.System,
		&item.AutoAssign,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorType := &model.RoleNotFoundError{RoleID: id.String()}
			return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByID")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "role selected successfully", attribute.String("role.id", id.String()))

	return &item, nil
}

func (ref *RolesRepository) Select(ctx context.Context, input *model.SelectRolesInput) (*model.SelectRolesOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.Select", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Select")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "rls."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"system",
		"auto_assign",
		"created_at",
		"updated_at",
		"serial_id",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.RolesFilterFields)
		filterQuery = fmt.Sprintf("WHERE (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "rls.serial_id DESC, rls.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH rls AS (
            SELECT
                {{.QueryColumns}}
            FROM roles AS rls
                {{ .QueryWhere }}
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM rls ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "rls.serial_id DESC, rls.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Select", "invalid token")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("rls", tokenDirection, id, serial, filterQuery, false)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Select", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Roles.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Select", "failed to select all roles")
	}
	defer rows.Close()

	var fetchedItems []model.Role
	for rows.Next() {
		var item model.Role

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Select", "failed to scan role")
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.Select", "failed to scan rows")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		return &model.SelectRolesOutput{
			Items:     make([]model.Role, 0),
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

	ret := &model.SelectRolesOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "roles selected successfully")

	return ret, nil
}

// SelectByPolicyID selects the roles by policy ID.
func (ref *RolesRepository) SelectByPolicyID(ctx context.Context, policyID uuid.UUID, input *model.SelectRolesInput) (*model.SelectRolesOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.SelectByPolicyID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "rls."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"system",
		"auto_assign",
		"created_at",
		"updated_at",
		"serial_id",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.RolesFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "rls.serial_id DESC, rls.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH rls AS (
            SELECT
                {{.QueryColumns}}
            FROM roles AS rls
                -- policies
                LEFT JOIN roles_policies AS rp ON rls.id = rp.roles_id
                LEFT JOIN policies AS p ON rp.policies_id = p.id
            WHERE p.id = $1
            {{ .QueryWhere }}
            GROUP BY rls.id
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM rls ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "rls.serial_id DESC, rls.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID", "invalid token")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("rls", tokenDirection, id, serial, filterQuery, true)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Roles.SelectByPolicyID", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query, policyID.String())
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID", "failed to select all roles")
	}
	defer rows.Close()

	var fetchedItems []model.Role
	for rows.Next() {
		var item model.Role

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID", "failed to scan role")
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByPolicyID", "failed to scan rows")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.Warn("repository.Roles.SelectByPolicyID", "what", "no roles found")
		return &model.SelectRolesOutput{
			Items:     make([]model.Role, 0),
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

	ret := &model.SelectRolesOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "roles selected successfully", attribute.String("policy.id", policyID.String()))

	return ret, nil
}

func (ref *RolesRepository) SelectByUserID(ctx context.Context, userID uuid.UUID, input *model.SelectRolesInput) (*model.SelectRolesOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.SelectByUserID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "rls."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"system",
		"auto_assign",
		"created_at",
		"updated_at",
		"serial_id",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.RolesFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "rls.serial_id DESC, rls.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH rls AS (
            SELECT
                {{.QueryColumns}}
            FROM roles AS rls
                -- users
                LEFT JOIN users_roles AS ur ON rls.id = ur.roles_id
                LEFT JOIN users AS u ON ur.users_id = u.id
            WHERE u.id = $1
            {{ .QueryWhere }}
            GROUP BY rls.id
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM rls ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "rls.serial_id DESC, rls.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID", "invalid token")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("rls", tokenDirection, id, serial, filterQuery, true)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Roles.SelectByUserID", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query, userID.String())
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID", "failed to select all roles")
	}
	defer rows.Close()

	var fetchedItems []model.Role
	for rows.Next() {
		var item model.Role

		scanFields := ref.buildScanFields(&item, input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID", "failed to scan role")
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.SelectByUserID", "failed to scan rows")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.Warn("repository.Roles.SelectByUserID", "what", "no roles found")
		return &model.SelectRolesOutput{
			Items:     make([]model.Role, 0),
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

	ret := &model.SelectRolesOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "roles selected successfully", attribute.String("user.id", userID.String()))

	return ret, nil
}

// LinkUsers links the users to the role.
func (ref *RolesRepository) LinkUsers(ctx context.Context, input *model.LinkUsersToRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.LinkUsers", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.LinkUsers")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.LinkUsers")
	}

	// Prepare arrays for UNNEST
	roleIDs := make([]string, len(input.UserIDs))
	userIDs := make([]string, len(input.UserIDs))
	for i, userID := range input.UserIDs {
		roleIDs[i] = input.RoleID.String() // Ensure RoleID is converted to string (e.g., if it's a UUID)
		userIDs[i] = userID.String()       // Ensure userID is converted to string (e.g., if it's a UUID)
	}

	query := `
        -- insert the new users
        INSERT INTO users_roles (roles_id, users_id)
        SELECT * FROM UNNEST($1::uuid[], $2::uuid[]) -- Use appropriate type casting for your UUIDs
        ON CONFLICT (roles_id, users_id)
        DO UPDATE SET updated_at = NOW();
    `

	slog.Debug("repository.Roles.LinkUsers", "query", prettyPrint(query), "roleIDs", roleIDs, "userIDs", userIDs)

	// Pass the arrays as parameters
	_, err := ref.db.Exec(ctx, query, roleIDs, userIDs)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.LinkUsers", "failed to link users")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "users linked successfully")

	return nil
}

// UnlinkUsers unlinks the users from the role.
func (ref *RolesRepository) UnlinkUsers(ctx context.Context, input *model.UnlinkUsersFromRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.UnlinkUsers", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UnlinkUsers")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UnlinkUsers")
	}

	// Prepare the user IDs for the parameterized query.
	// Assuming UserIDs are UUIDs, convert them to string slices.
	userIDs := make([]string, len(input.UserIDs))
	for i, userID := range input.UserIDs {
		userIDs[i] = userID.String() // Ensure your UUID type has a .String() method
	}

	// Use a parameterized query with ANY() for the IN clause.
	query := `
        DELETE FROM users_roles
        WHERE roles_id = $1 AND users_id IN (SELECT unnest($2::uuid[]));
    `

	slog.Debug("repository.Roles.UnlinkUsers", "query", prettyPrint(query), "roleID", input.RoleID.String(), "userIDs", userIDs)

	// Execute the query with parameters.
	// Ensure input.RoleID is converted to its string representation if it's a UUID type.
	_, err := ref.db.Exec(ctx, query, input.RoleID.String(), userIDs)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UnlinkUsers", "failed to unlink users")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "users unlinked successfully")

	return nil
}

// LinkPolicies links the policies to the role.
func (ref *RolesRepository) LinkPolicies(ctx context.Context, input *model.LinkPoliciesToRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.LinkPolicies", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.LinkPolicies")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.LinkPolicies")
	}

	// Prepare arrays for UNNEST
	roleIDs := make([]string, len(input.PolicyIDs))
	policyIDs := make([]string, len(input.PolicyIDs))
	for i, policyID := range input.PolicyIDs {
		roleIDs[i] = input.RoleID.String() // Ensure RoleID is converted to string (e.g., if it's a UUID)
		policyIDs[i] = policyID.String()   // Ensure PolicyID is converted to string (e.g., if it's a UUID)
	}

	query := `
        -- insert the new policies
        INSERT INTO roles_policies (roles_id, policies_id)
        SELECT * FROM UNNEST($1::uuid[], $2::uuid[]) -- Use appropriate type casting for your UUIDs
        ON CONFLICT (roles_id, policies_id)
        DO UPDATE SET updated_at = NOW();
    `

	slog.Debug("repository.Roles.LinkPolicies", "query", prettyPrint(query), "roleIDs", roleIDs, "policyIDs", policyIDs)

	// Pass the arrays as parameters
	_, err := ref.db.Exec(ctx, query, roleIDs, policyIDs)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.LinkPolicies", "failed to link policies")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policies linked successfully")

	return nil
}

// UnlinkPolicies unlinks the policies from the role.
func (ref *RolesRepository) UnlinkPolicies(ctx context.Context, input *model.UnlinkPoliciesFromRoleInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Roles.UnlinkPolicies", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UnlinkPolicies")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UnlinkPolicies")
	}

	// Prepare the policy IDs for the parameterized query.
	// Assuming PolicyIDs are UUIDs, convert them to string slices.
	policyIDs := make([]string, len(input.PolicyIDs))
	for i, policyID := range input.PolicyIDs {
		policyIDs[i] = policyID.String()
	}

	// Use parameterized query with ANY() for the IN clause.
	query := `
        DELETE FROM roles_policies
        WHERE roles_id = $1  AND policies_id IN (SELECT unnest($2::uuid[]));
    `

	slog.Debug("repository.Roles.UnlinkPolicies", "query", prettyPrint(query), "roleID", input.RoleID.String(), "policyIDs", policyIDs)

	// Execute the query with parameters.
	// Ensure input.RoleID is converted to its string representation if it's a UUID type.
	_, err := ref.db.Exec(ctx, query, input.RoleID.String(), policyIDs)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Roles.UnlinkPolicies", "failed to unlink policies")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policies unlinked successfully")

	return nil
}

// Helper functions for common patterns

// setupContext creates a context with timeout and starts a span with standard attributes.
// Returns the new context, span, and common metric attributes.
func (ref *RolesRepository) setupContext(ctx context.Context, operation string, timeout time.Duration) (context.Context, trace.Span, []attribute.KeyValue, context.CancelFunc) {
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
func (ref *RolesRepository) handlePgError(err error, input any) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.Message, "_pkey") {
				switch v := input.(type) {
				case *model.InsertRoleInput:
					return &model.RoleIDAlreadyExistsError{ID: v.ID.String()}
				case *model.UpdateRoleInput:
					return &model.RoleIDAlreadyExistsError{ID: v.ID.String()}
				case uuid.UUID:
					return &model.RoleIDAlreadyExistsError{ID: v.String()}
				}
			}

			if strings.Contains(pgErr.Message, "name") {
				switch v := input.(type) {
				case *model.InsertRoleInput:
					return &model.RoleNameAlreadyExistsError{Name: v.Name}
				case *model.UpdateRoleInput:
					if v.Name != nil {
						return &model.RoleNameAlreadyExistsError{Name: *v.Name}
					}
				}
			}
		case "P0001": // Raised exception
			if strings.Contains(pgErr.Message, "updated") || strings.Contains(pgErr.Message, "deleted") {
				switch v := input.(type) {
				case *model.UpdateRoleInput:
					return &model.SystemRoleError{RoleID: v.ID.String()}
				case *model.DeleteRoleInput:
					return &model.SystemRoleError{RoleID: v.ID.String()}
				case uuid.UUID:
					return &model.SystemRoleError{RoleID: v.String()}
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

func (ref *RolesRepository) buildScanFields(item *model.Role, requestedFields string) []any {
	scanFields := make([]any, 0)

	if requestedFields == "" {
		// All fields were requested
		return []any{
			&item.ID,
			&item.Name,
			&item.Description,
			&item.System,
			&item.AutoAssign,
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
		case "system":
			scanFields = append(scanFields, &item.System)
		case "auto_assign":
			scanFields = append(scanFields, &item.AutoAssign)
		case "created_at":
			scanFields = append(scanFields, &item.CreatedAt)
		case "updated_at":
			scanFields = append(scanFields, &item.UpdatedAt)

		default:
			slog.Warn("repository.Roles.buildScanFields", "what", "field not found", "field", field)
		}
	}

	// Always include ID and SerialID fields for pagination
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)
	return scanFields
}
