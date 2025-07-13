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

type PoliciesRepositoryConfig struct {
	DB              *pgxpool.Pool
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type policiesRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// PoliciesRepository is a PostgreSQL store.
type PoliciesRepository struct {
	db              *pgxpool.Pool
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         policiesRepositoryMetrics
}

// NewPoliciesRepository creates a new PoliciesRepository.
func NewPoliciesRepository(conf PoliciesRepositoryConfig) (*PoliciesRepository, error) {
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

	repo := &PoliciesRepository{
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
		metric.WithDescription("The number of calls to the policies repository"),
	)
	if err != nil {
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (ref *PoliciesRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *PoliciesRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxPingTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Policies.PingContext")
	defer span.End()

	return ref.db.Ping(ctx)
}

// Insert inserts a new policy into the repository.
func (ref *PoliciesRepository) Insert(ctx context.Context, input *model.CreatePolicyInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.Insert", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorType := &model.InvalidPolicyIDError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Insert")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Insert")
	}

	query := `
        INSERT INTO policies (id, resources_id, name, description, allowed_action, allowed_resource)
        VALUES ($1, $2, $3, $4, $5, $6);
    `

	slog.Debug("repository.Policies.Insert", "query", prettyPrint(query))

	_, err := ref.db.Exec(ctx, query,
		input.ID,
		input.ResourceID,
		input.Name,
		input.Description,
		input.AllowedAction,
		input.AllowedResource,
	)
	if err != nil {
		return o11y.RecordError(ctx, span, ref.handlePgError(err, input), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Insert")
	}

	slog.Debug("repository.Policies.Insert", "policy_id", input.ID.String())
	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policy inserted successfully", attribute.String("policy.id", input.ID.String()))

	return nil
}

// UpdateByID updates a policy in the repository.
func (ref *PoliciesRepository) UpdateByID(ctx context.Context, input *model.UpdatePolicyInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.UpdateByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UpdateByID")
	}

	span.SetAttributes(attribute.String("policy_id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UpdateByID")
	}

	args := []any{input.ID}

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

	if input.AllowedAction != nil && *input.AllowedAction != "" {
		args = append(args, *input.AllowedAction)
	} else {
		args = append(args, nil)
	}

	if input.AllowedResource != nil && *input.AllowedResource != "" {
		args = append(args, *input.AllowedResource)
	} else {
		args = append(args, nil)
	}

	updatedAt, err := time.Now().In(time.FixedZone("UTC", 0)).MarshalText()
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UpdateByID")
	}

	args = append(args, updatedAt)

	query := `
        UPDATE policies SET
            name = COALESCE(NULLIF($2, ''), name),
            description = COALESCE(NULLIF($3, ''), description),
            allowed_action = COALESCE(NULLIF($4, ''), allowed_action),
            allowed_resource = COALESCE(NULLIF($5, ''), allowed_resource),
            updated_at = COALESCE($6, updated_at)
        WHERE id = $1;
    `

	slog.Debug("repository.Policies.UpdateByID", "query", prettyPrint(query))

	result, err := ref.db.Exec(ctx, query, args...)
	if err != nil {
		return o11y.RecordError(ctx, span, ref.handlePgError(err, input), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UpdateByID")
	}

	if result.RowsAffected() == 0 {
		errorType := &model.PolicyNotFoundError{Message: "policy not found"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UpdateByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policy updated successfully", attribute.String("policy.id", input.ID.String()))

	return nil
}

// DeleteByID deletes a policy from the repository.
func (ref *PoliciesRepository) DeleteByID(ctx context.Context, input *model.DeletePolicyInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.DeleteByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorType := &model.InvalidPolicyIDError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.DeleteByID")
	}

	query := `
        DELETE FROM policies WHERE id = $1;
    `

	slog.Debug("repository.Policies.DeleteByID", "query", prettyPrint(query, input.ID.String()))

	result, err := ref.db.Exec(ctx, query, input.ID)
	if err != nil {
		return o11y.RecordError(ctx, span, ref.handlePgError(err, input), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.DeleteByID")
	}

	if result.RowsAffected() == 0 {
		// grateful return user was deleted, security reason, but log and record error
		errorType := &model.PolicyNotFoundError{Message: "policy not found"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.DeleteByID")
	}

	slog.Debug("repository.Policies.DeleteByID", "policy_id", input.ID.String())
	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policy deleted successfully", attribute.String("policy.id", input.ID.String()))

	return nil
}

// SelectByID returns the resource with the specified ID.
func (ref *PoliciesRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.Policy, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.SelectByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	span.SetAttributes(attribute.String("policy.id", id.String()))

	if id == uuid.Nil {
		errorType := &model.InvalidPolicyIDError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByID")
	}

	query := `
        SELECT
            pol.id,
            pol.name,
            pol.description,
            pol.allowed_action,
            pol.allowed_resource,
            pol.system,
            pol.created_at,
            pol.updated_at,
            array_agg(DISTINCT(ARRAY[COALESCE(res.id::varchar, '00000000-0000-0000-0000-000000000000'), COALESCE(res.name::varchar,'')])) AS resource
        FROM policies AS pol
            -- resources
            LEFT JOIN resources AS res ON pol.resources_id = res.id
        WHERE pol.id = $1
        GROUP BY pol.id, res.id;
    `

	slog.Debug("repository.Policies.SelectByID", "query", prettyPrint(query))

	row := ref.db.QueryRow(ctx, query, id)

	var element model.Policy
	var resources []string

	if err := row.Scan(
		&element.ID,
		&element.Name,
		&element.Description,
		&element.AllowedAction,
		&element.AllowedResource,
		&element.System,
		&element.CreatedAt,
		&element.UpdatedAt,
		&resources,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorType := &model.PolicyNotFoundError{Message: "policy not found"}
			return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByID")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByID")
	}

	var err error
	if len(resources) > 0 {
		element.Resource.ID, err = uuid.Parse(resources[0])
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByID")
		}

		element.Resource.Name = resources[1]
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policies selected successfully", attribute.String("policy.id", id.String()))
	return &element, nil
}

func (ref *PoliciesRepository) Select(ctx context.Context, input *model.SelectPoliciesInput) (*model.SelectPoliciesOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.Select", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "pol."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"allowed_action",
		"allowed_resource",
		"system",
		"created_at",
		"updated_at",
		"serial_id",
		"array_agg(DISTINCT(ARRAY[COALESCE(res.id::varchar, '00000000-0000-0000-0000-000000000000'), COALESCE(res.name::varchar,'')])) AS resource",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.PoliciesFilterFields)
		filterQuery = fmt.Sprintf("WHERE (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "pol.serial_id DESC, pol.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH pol AS (
            SELECT
                {{.QueryColumns}}
            FROM policies AS pol
                -- resources
                LEFT JOIN resources AS res ON pol.resources_id = res.id
            {{ .QueryWhere }}
            GROUP BY pol.id, res.id
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM pol ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "pol.serial_id DESC, pol.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("pol", tokenDirection, id, serial, filterQuery, false)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
	}

	query := tpl.String()
	slog.Debug("repository.Policies.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
	}
	defer rows.Close()

	var fetchedItems []model.Policy
	for rows.Next() {
		var item model.Policy
		var resources []string

		scanFields, err := ref.buildScanFields(&item, &resources, input.Fields)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
		}

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
		}

		if len(resources) > 0 {
			item.Resource.ID, err = uuid.Parse(resources[0])
			if err != nil {
				return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
			}

			item.Resource.Name = resources[1]
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.Select")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.Warn("repository.Policies.Select", "what", "no policies found")
		return &model.SelectPoliciesOutput{
			Items:     make([]model.Policy, 0),
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

	ret := &model.SelectPoliciesOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policies selected successfully")

	return ret, nil
}

// SelectByRoleID returns the policies with the specified role ID.
func (ref *PoliciesRepository) SelectByRoleID(ctx context.Context, roleID uuid.UUID, input *model.SelectPoliciesInput) (*model.SelectPoliciesOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.SelectByRoleID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "pol."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"allowed_action",
		"allowed_resource",
		"system",
		"created_at",
		"updated_at",
		"serial_id",
		"array_agg(DISTINCT(ARRAY[COALESCE(res.id::varchar, '00000000-0000-0000-0000-000000000000'), COALESCE(res.name::varchar,'')])) AS resource",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.PoliciesFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "pol.serial_id DESC, pol.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH pol AS (
            SELECT
                {{.QueryColumns}}
            FROM policies AS pol
                -- resources
                LEFT JOIN resources AS res ON pol.resources_id = res.id
                -- roles
                LEFT JOIN roles_policies AS rp ON pol.id = rp.policies_id
            WHERE rp.roles_id = $1
            {{ .QueryWhere }}
            GROUP BY pol.id, res.id
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM pol ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "pol.serial_id DESC, pol.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("pol", tokenDirection, id, serial, filterQuery, true)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
	}

	query := tpl.String()
	slog.Debug("repository.Policies.SelectByRoleID", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
	}
	defer rows.Close()

	var fetchedItems []model.Policy

	for rows.Next() {
		var item model.Policy
		var resources []string

		scanFields, err := ref.buildScanFields(&item, &resources, input.Fields)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
		}

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
		}

		if len(resources) > 0 {
			item.Resource.ID, err = uuid.Parse(resources[0])
			if err != nil {
				return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
			}

			item.Resource.Name = resources[1]
		}

		fetchedItems = append(fetchedItems, item)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.SelectByRoleID")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		slog.Warn("repository.Policies.SelectByRoleID", "what", "no policies found")
		return &model.SelectPoliciesOutput{
			Items:     make([]model.Policy, 0),
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

	ret := &model.SelectPoliciesOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "policies selected successfully")

	return ret, nil
}

// LinkRoles links roles to a policies.
func (ref *PoliciesRepository) LinkRoles(ctx context.Context, input *model.LinkRolesToPolicyInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.LinkRoles", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorType := &model.InvalidPolicyIDError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.LinkRoles")
	}

	span.SetAttributes(attribute.String("policy_id", input.PolicyID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.LinkRoles")
	}

	var policiesRoles bytes.Buffer
	for i, roleID := range input.RoleIDs {
		_, err := policiesRoles.WriteString(fmt.Sprintf("('%s', '%s')", input.PolicyID.String(), roleID))
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.LinkRoles")
		}

		if i < len(input.RoleIDs)-1 {
			_, err := policiesRoles.WriteString(", ")
			if err != nil {
				return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.LinkRoles")
			}
		}
	}

	query := fmt.Sprintf(`
        -- Insert roles to policies
        INSERT INTO roles_policies (policies_id, roles_id)
        VALUES %s
        ON CONFLICT (policies_id, roles_id)
        DO UPDATE SET updated_at = NOW();
    `,
		policiesRoles.String(),
	)

	slog.Debug("repository.Policies.LinkRoles", "query", prettyPrint(query))

	_, err := ref.db.Exec(ctx, query)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.LinkRoles")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "roles linked to policies")

	return nil
}

// UnlinkRoles unlinks roles from a policies.
func (ref *PoliciesRepository) UnlinkRoles(ctx context.Context, input *model.UnlinkRolesFromPolicyInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Policies.UnlinkRoles", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorType := &model.InvalidPolicyIDError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UnlinkRoles")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UnlinkRoles")
	}

	var rolesIn bytes.Buffer
	for i, roleID := range input.RoleIDs {
		if i > 0 {
			if _, err := rolesIn.WriteString(", "); err != nil {
				return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UnlinkRoles")
			}
		}

		if _, err := rolesIn.WriteString(fmt.Sprintf("('%s', '%s')", input.PolicyID.String(), roleID)); err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UnlinkRoles")
		}
	}

	query := fmt.Sprintf(`
        -- Delete roles from policies
        DELETE FROM roles_policies
        WHERE policies_id = '%s' AND roles_id IN %s;
    `,
		input.PolicyID.String(),
		rolesIn.String(),
	)

	slog.Debug("repository.Policies.UnlinkRoles", "query", prettyPrint(query))

	_, err := ref.db.Exec(ctx, query)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Policies.UnlinkRoles")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "roles unlinked from policies")

	return nil
}

// Helper functions for common patterns

// setupContext creates a context with timeout and starts a span with standard attributes.
// Returns the new context, span, and common metric attributes.
func (ref *PoliciesRepository) setupContext(ctx context.Context, operation string, timeout time.Duration) (context.Context, trace.Span, []attribute.KeyValue, context.CancelFunc) {
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
func (ref *PoliciesRepository) handlePgError(err error, input any) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.Message, "_pkey") {
				switch v := input.(type) {
				case *model.CreatePolicyInput:
					return &model.PolicyIDAlreadyExistsError{ID: v.ID}
				case *model.UpdatePolicyInput:
					return &model.PolicyIDAlreadyExistsError{ID: v.ID}
				case uuid.UUID:
					return &model.PolicyIDAlreadyExistsError{ID: v}
				}
			}
			if strings.Contains(pgErr.Message, "_name") {
				switch v := input.(type) {
				case *model.CreatePolicyInput:
					return &model.PolicyNameAlreadyExistsError{Name: v.Name}
				case *model.UpdatePolicyInput:
					if v.Name != nil {
						return &model.PolicyNameAlreadyExistsError{Name: *v.Name}
					}
				}
			}
		case "23503": // Foreign key violation
			if strings.Contains(pgErr.Message, "resources_id_fkey") {
				switch v := input.(type) {
				case *model.CreatePolicyInput:
					return &model.ResourceIDNotFoundError{ID: v.ResourceID.String()}
				}
			}
		case "P0001": // Raised exception
			if strings.Contains(pgErr.Message, "updated") || strings.Contains(pgErr.Message, "deleted") {
				switch v := input.(type) {
				case *model.UpdatePolicyInput:
					return &model.SystemPolicyError{PolicyID: v.ID.String()}
				case *model.DeletePolicyInput:
					return &model.SystemPolicyError{PolicyID: v.ID.String()}
				case uuid.UUID:
					return &model.SystemPolicyError{PolicyID: v.String()}
				}
			}
		}
	}

	return err
}

// buildScanFields creates the scan targets for the result rows based on the requested fields.
func (ref *PoliciesRepository) buildScanFields(item *model.Policy, resources *[]string, requestedFields string) ([]any, error) {
	scanFields := make([]any, 0)

	if requestedFields == "" {
		// All fields were requested
		return []any{
			&item.ID,
			&item.Name,
			&item.Description,
			&item.AllowedAction,
			&item.AllowedResource,
			&item.System,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.SerialID,
			&resources,
		}, nil
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
		case "allowed_action":
			scanFields = append(scanFields, &item.AllowedAction)
		case "allowed_resource":
			scanFields = append(scanFields, &item.AllowedResource)
		case "system":
			scanFields = append(scanFields, &item.System)
		case "created_at":
			scanFields = append(scanFields, &item.CreatedAt)
		case "updated_at":
			scanFields = append(scanFields, &item.UpdatedAt)

		default:
			slog.Warn("repository.Policies.buildScanFields", "what", "field not found", "field", field)
		}
	}

	// always select id and serial_id for pagination
	// if id is not selected, it will be added to the scanFields
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)

	return scanFields, nil
}
