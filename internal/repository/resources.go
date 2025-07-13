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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ResourcesRepositoryConfig struct {
	DB              *pgxpool.Pool
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type resourcesRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// ResourcesRepository is a PostgreSQL store.
type ResourcesRepository struct {
	db              *pgxpool.Pool
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         resourcesRepositoryMetrics
}

// NewResourcesRepository creates a new ResourcesRepository.
func NewResourcesRepository(conf ResourcesRepositoryConfig) (*ResourcesRepository, error) {
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

	repo := &ResourcesRepository{
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
		metric.WithDescription("The number of calls to the resources repository"),
	)
	if err != nil {
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (ref *ResourcesRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *ResourcesRepository) PingContext(ctx context.Context) error {
	ctx, span, _, cancel := ref.setupContext(ctx, "repository.Resources.PingContext", ref.maxPingTimeout)
	defer cancel()
	defer span.End()

	return ref.db.Ping(ctx)
}

// SelectByID returns the resource with the specified ID.
func (ref *ResourcesRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.Resource, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Resources.SelectByID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if id == uuid.Nil {
		errorType := &model.InvalidResourceIDError{ID: id}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.SelectByID", "id is nil")
	}

	span.SetAttributes(attribute.String("Resources.id", id.String()))

	query := `
        SELECT
            id,
            name,
            description,
            action,
            resource,
            system,
            created_at,
            updated_at
        FROM resources
        WHERE id = $1;
    `

	slog.Debug("repository.Resources.SelectByID", "query", prettyPrint(query, id.String()))

	row := ref.db.QueryRow(ctx, query, id)

	var element model.Resource
	if err := row.Scan(
		&element.ID,
		&element.Name,
		&element.Description,
		&element.Action,
		&element.Resource,
		&element.System,
		&element.CreatedAt,
		&element.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorType := &model.ResourceNotFoundError{ID: id.String()}
			return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.SelectByID")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.SelectByID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "resources selected successfully", attribute.String("Resources.id", id.String()))
	return &element, nil
}

func (ref *ResourcesRepository) Select(ctx context.Context, input *model.SelectResourcesInput) (*model.SelectResourcesOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Resources.Select", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorType := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "res."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"action",
		"resource",
		"system",
		"created_at",
		"updated_at",
		"serial_id",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.ResourcesFilterFields)
		filterQuery = fmt.Sprintf("WHERE (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "res.serial_id DESC, res.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH res AS (
            SELECT
                {{.QueryColumns}}
            FROM resources AS res
            {{ .QueryWhere }}
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM res ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "res.serial_id DESC, res.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select", "failed to get paginator direction")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("res", tokenDirection, id, serial, filterQuery, false)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Resources.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select", "failed to select all resources")
	}
	defer rows.Close()

	var fetchedItems []model.Resource
	for rows.Next() {
		var item model.Resource

		scanFields := ref.buildScanFields(&item, input.Fields)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select", "failed to scan resources")
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Resources.Select", "failed to scan resources")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		return &model.SelectResourcesOutput{
			Items:     make([]model.Resource, 0),
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

	ret := &model.SelectResourcesOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "resources selected successfully")

	return ret, nil
}

// Helper functions for common patterns

// setupContext creates a context with timeout and starts a span with standard attributes.
// Returns the new context, span, and common metric attributes.
func (ref *ResourcesRepository) setupContext(ctx context.Context, operation string, timeout time.Duration) (context.Context, trace.Span, []attribute.KeyValue, context.CancelFunc) {
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

func (ref *ResourcesRepository) buildScanFields(item *model.Resource, requestedFields string) []any {
	scanFields := make([]any, 0)

	if requestedFields == "" {
		// All fields were requested
		return []any{
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Action,
			&item.Resource,
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
		case "action":
			scanFields = append(scanFields, &item.Action)
		case "resource":
			scanFields = append(scanFields, &item.Resource)
		case "system":
			scanFields = append(scanFields, &item.System)
		case "created_at":
			scanFields = append(scanFields, &item.CreatedAt)
		case "updated_at":
			scanFields = append(scanFields, &item.UpdatedAt)

		default:
			slog.Warn("repository.Resources.buildScanFields", "what", "field not found", "field", field)
		}
	}

	// Always include ID and SerialID fields for pagination
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)
	return scanFields
}
