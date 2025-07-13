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

// ProductsRepositoryConfig is the configuration for the ProductsRepository.
type ProductsRepositoryConfig struct {
	DB              *pgxpool.Pool
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type productsRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// ProductsRepository is a PostgreSQL store.
type ProductsRepository struct {
	db              *pgxpool.Pool
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
	ot              *o11y.OpenTelemetry
	metricsPrefix   string
	metrics         productsRepositoryMetrics
}

// NewProductsRepository creates a new ProductsRepository.
func NewProductsRepository(conf ProductsRepositoryConfig) (*ProductsRepository, error) {
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

	repo := &ProductsRepository{
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
		metric.WithDescription("The number of calls to the product repository"),
	)
	if err != nil {
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (ref *ProductsRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *ProductsRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxPingTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Products.PingContext")
	defer span.End()

	return ref.db.Ping(ctx)
}

func (ref *ProductsRepository) Insert(ctx context.Context, input *model.InsertProductInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.Insert", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Insert")
	}

	span.SetAttributes(attribute.String("product.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Insert")
	}

	tx, txErr := ref.db.Begin(ctx)
	if txErr != nil {
		return o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Insert", "failed to begin transaction")
	}
	defer func() {
		if txErr != nil {
			if err := tx.Rollback(ctx); err != nil {
				e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Insert", "failed to rollback transaction")
				slog.Error("repository.Embeddings.InsertByProjectID", "error", e)
			}
		} else {
			if err := tx.Commit(ctx); err != nil {
				e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Insert", "failed to commit transaction")
				if e != nil {
					slog.Error("repository.Products.Insert", "error", e)
				}
			}
		}
	}()

	query1 := `
        INSERT INTO products (id, projects_id, name, description)
        VALUES ($1, $2, $3, $4);
    `

	slog.Debug("repository.Products.Insert", "query", prettyPrint(query1, input.ID, input.ProjectID, input.Name, input.Description))

	_, txErr = tx.Exec(ctx, query1,
		input.ID,
		input.ProjectID,
		input.Name,
		input.Description,
	)
	if txErr != nil {
		return ref.handlePgError(txErr, input)
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "product inserted successfully", attribute.String("product.id", input.ID.String()))

	return nil
}

func (ref *ProductsRepository) Update(ctx context.Context, input *model.UpdateProductInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.Update", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update")
	}

	span.SetAttributes(attribute.String("product.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update")
	}

	args := []any{input.ID, input.ProjectID}
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

	updatedAt, err := time.Now().In(time.FixedZone("UTC", 0)).MarshalText()
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update", "failed to marshal updated_at time")
	}
	args = append(args, updatedAt)

	tx, txErr := ref.db.Begin(ctx)
	if txErr != nil {
		return o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update", "failed to begin transaction")
	}

	defer func() {
		if txErr != nil {
			if err := tx.Rollback(ctx); err != nil {
				e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update", "failed to rollback transaction")
				slog.Error("repository.Products.Update", "error", e)
			}
		} else {
			if err := tx.Commit(ctx); err != nil {
				e := o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update", "failed to commit transaction")
				slog.Error("repository.Products.Update", "error", e)
			}
		}
	}()

	query1 := `
        UPDATE products
        SET
            name = COALESCE($3, name),
            description = COALESCE($4, description),
            updated_at = COALESCE($5, updated_at)
        WHERE id = $1 AND projects_id = $2;
    `

	slog.Debug("repository.Products.UpdateByID", "query", prettyPrint(query1, args...))

	result, txErr := tx.Exec(ctx, query1, args...)
	if txErr != nil {
		return ref.handlePgError(o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update"), input)
	}

	if result.RowsAffected() == 0 {
		txErr = &model.ProductNotFoundError{ProductID: input.ID.String()}
		return o11y.RecordError(ctx, span, txErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Update")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "product updated successfully", attribute.String("product.id", input.ID.String()))

	return nil
}

func (ref *ProductsRepository) Delete(ctx context.Context, input *model.DeleteProductInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.Delete", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input cannot be nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Delete")
	}

	span.SetAttributes(attribute.String("input.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Delete")
	}

	query := `
        DELETE FROM products
        WHERE id = $1 AND projects_id = $2;
    `

	slog.Debug("repository.Products.Delete", "query", prettyPrint(query, input.ID, input.ProjectID))

	result, err := ref.db.Exec(ctx, query, input.ID, input.ProjectID)
	if err != nil {
		return o11y.RecordError(ctx, span, ref.handlePgError(err, input), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Delete")
	}

	if result.RowsAffected() == 0 {
		// grateful return user was deleted, security reason, but log and record error
		errorType := &model.ProductNotFoundError{ProductID: input.ID.String()}
		e := o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Delete")
		slog.Error("repository.Products.Delete", "error", e, "product.id", input.ID.String())

		return nil
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "product deleted successfully", attribute.String("product.id", input.ID.String()))

	return nil
}

func (ref *ProductsRepository) SelectByIDByProjectID(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (*model.Product, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.SelectByIDByProjectID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if id == uuid.Nil {
		invalidErr := &model.InvalidProductIDError{Message: "invalid product ID"}
		return nil, o11y.RecordError(ctx, span, invalidErr, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByIDByProjectID")
	}

	query := `
        SELECT
            p.id,
            p.name,
            p.description,
            p.created_at,
            p.updated_at,
            array_agg(
                DISTINCT(
                    ARRAY[
                        COALESCE(prj.id::varchar, '00000000-0000-0000-0000-000000000000'),
                        COALESCE(prj.name, '')
                    ]
                )
            ) AS projects,
            array_agg(
                DISTINCT(
                    ARRAY[
                        COALESCE(ppp.payment_processor_product_id::varchar, '00000000-0000-0000-0000-000000000000'),
                        COALESCE(pp.id::varchar, '00000000-0000-0000-0000-000000000000'),
                        COALESCE(pp.name, ''),
                        COALESCE(ppt.id::varchar, '00000000-0000-0000-0000-000000000000'),
                        COALESCE(ppt.name, '')
                    ]
                )
            ) AS payment_processors
        FROM products AS p
            LEFT JOIN projects prj ON prj.id = p.projects_id
            LEFT JOIN products_payment_processors ppp ON ppp.product_id = p.id
            LEFT JOIN payment_processors pp ON pp.id = ppp.payment_processor_id
            LEFT JOIN payment_processor_types ppt ON ppt.id = pp.payment_processor_types_id
        WHERE p.id = $1 AND p.projects_id = $2
        GROUP BY p.id;
    `

	slog.Debug("repository.Products.SelectByID", "query", prettyPrint(query, id, projectID))

	row := ref.db.QueryRow(ctx, query, id, projectID)

	var element model.Product
	var productPaymentProcessorResponse []string
	var projects []string

	if err := row.Scan(
		&element.ID,
		&element.Name,
		&element.Description,
		&element.CreatedAt,
		&element.UpdatedAt,
		&projects,
		&productPaymentProcessorResponse,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorType := &model.ProductNotFoundError{ProductID: id.String()}
			return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByIDByProjectID")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByIDByProjectID", "failed to scan product")
	}

	// PostgreSQL -> {{f282315d-1e65-43fd-8f12-a9c27be60c9e, "Project Name"}}
	// Go -> [f282315d-1e65-43fd-8f12-a9c27be60c9e, Project Name]
	for i := 0; i < len(projects); i += 2 {
		id, err := uuid.Parse(projects[i])
		if err != nil {
			return nil, o11y.RecordError(ctx, span, &model.InvalidProjectIDError{Message: "invalid project ID in projects array"}, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByIDByProjectID", "failed to parse project ID")
		}

		element.Projects = &model.Project{
			ID:   id,
			Name: projects[i+1],
		}
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "product selected successfully", attribute.String("product.id", id.String()))

	return &element, nil
}

func (ref *ProductsRepository) SelectByProjectID(ctx context.Context, projectID uuid.UUID, input *model.SelectProductsInput) (*model.SelectProductsOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.SelectByProjectID", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "p."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"serial_id",
		"array_agg(DISTINCT(ARRAY[COALESCE(prj.id::varchar, '00000000-0000-0000-0000-000000000000'),COALESCE(prj.name, '')])) AS projects",
		"array_agg(DISTINCT(ARRAY[COALESCE(ppp.payment_processor_product_id::varchar, '00000000-0000-0000-0000-000000000000'),COALESCE(pp.id::varchar, '00000000-0000-0000-0000-000000000000'),COALESCE(pp.name, ''),COALESCE(ppt.id::varchar, '00000000-0000-0000-0000-000000000000'),COALESCE(ppt.name, '')])) AS payment_processors",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.ProductsFilterFields)
		filterQuery = fmt.Sprintf("AND (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "p.serial_id DESC, p.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH p AS (
            SELECT
                {{.QueryColumns}}
            FROM products AS p
                LEFT JOIN projects prj ON prj.id = p.projects_id
                LEFT JOIN products_payment_processors ppp ON ppp.product_id = p.id
                LEFT JOIN payment_processors pp ON pp.id = ppp.payment_processor_id
                LEFT JOIN payment_processor_types ppt ON ppt.id = pp.payment_processor_types_id
            WHERE p.projects_id = $1
            {{ .QueryWhere }}
            GROUP BY p.id
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM p ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "p.serial_id DESC, p.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID", "invalid token")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("p", tokenDirection, id, serial, filterQuery, true)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Products.SelectByProjectID", "query", prettyPrint(query, projectID))

	// execute the query
	rows, err := ref.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID", "failed to select all products")
	}
	defer rows.Close()

	var fetchedItems []model.Product
	for rows.Next() {
		var item model.Product
		var projects []string
		var productPaymentProcessorResponse []string

		scanFields := ref.buildScanFields(&item,
			&projects,
			&productPaymentProcessorResponse,
			input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID", "failed to scan product")
		}

		// PostgreSQL -> {{f282315d-1e65-43fd-8f12-a9c27be60c9e, "Project Name"}}
		// Go -> [f282315d-1e65-43fd-8f12-a9c27be60c9e, Project Name]
		for i := 0; i < len(projects); i += 2 {
			id, err := uuid.Parse(projects[i])
			if err != nil {
				return nil, o11y.RecordError(ctx, span, &model.InvalidProjectIDError{Message: "invalid project ID in projects array"}, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID", "failed to parse project ID")
			}

			item.Projects = &model.Project{
				ID:   id,
				Name: projects[i+1],
			}
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.SelectByProjectID", "failed to scan rows")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		return &model.SelectProductsOutput{
			Items:     make([]model.Product, 0),
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

	ret := &model.SelectProductsOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "products selected successfully")

	return ret, nil
}

func (ref *ProductsRepository) Select(ctx context.Context, input *model.SelectProductsInput) (*model.SelectProductsOutput, error) {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.Select", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select")
	}

	// if no fields are provided, select all fields
	sqlFieldsPrefix := "p."
	fieldsArray := []string{
		"id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"serial_id",
		"array_agg(DISTINCT(ARRAY[COALESCE(prj.id::varchar, '00000000-0000-0000-0000-000000000000'), COALESCE(prj.name, '')])) AS projects",
		"array_agg(DISTINCT(ARRAY[COALESCE(ppp.payment_processor_product_id::varchar, '00000000-0000-0000-0000-000000000000'), COALESCE(pp.id::varchar, '00000000-0000-0000-0000-000000000000'),COALESCE(pp.name, ''), COALESCE(ppt.id::varchar, '00000000-0000-0000-0000-000000000000'), COALESCE(ppt.name, '')])) AS payment_processors",
	}

	fieldsStr := buildFieldSelection(sqlFieldsPrefix, fieldsArray, input.Fields)

	var filterQuery string
	if input.Filter != "" {
		filterSentence := injectPrefixToFields(sqlFieldsPrefix, input.Filter, model.ProductsFilterFields)
		filterQuery = fmt.Sprintf("WHERE (%s)", filterSentence)
	}

	var sortQuery string
	if input.Sort == "" {
		sortQuery = "p.serial_id DESC, p.id DESC"
	} else {
		sortQuery = input.Sort
	}

	// query template
	queryTemplate := `
        WITH p AS (
            SELECT
                {{.QueryColumns}}
            FROM products AS p
                LEFT JOIN projects prj ON prj.id = p.projects_id
                LEFT JOIN products_payment_processors ppp ON ppp.product_id = p.id
                LEFT JOIN payment_processors pp ON pp.id = ppp.payment_processor_id
                LEFT JOIN payment_processor_types ppt ON ppt.id = pp.payment_processor_types_id
            {{ .QueryWhere }}
            GROUP BY p.id
            ORDER BY {{.QueryInternalSort}}
            LIMIT {{.QueryLimit}}
        ) SELECT * FROM p ORDER BY {{.QueryExternalSort}}
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
	queryValues.QueryInternalSort = "p.serial_id DESC, p.id DESC"
	queryValues.QueryExternalSort = sortQuery

	tokenDirection, id, serial, err := model.GetPaginatorDirection(input.Paginator.NextToken, input.Paginator.PrevToken)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select", "invalid token")
	}

	queryValues.QueryWhere, queryValues.QueryInternalSort = buildPaginationCriteria("p", tokenDirection, id, serial, filterQuery, false)

	// render the template on query variable
	var tpl bytes.Buffer
	t := template.Must(template.New("query").Parse(queryTemplate))
	err = t.Execute(&tpl, queryValues)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select", "failed to render query template")
	}

	query := tpl.String()
	slog.Debug("repository.Products.Select", "query", prettyPrint(query))

	// execute the query
	rows, err := ref.db.Query(ctx, query)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select", "failed to select all products")
	}
	defer rows.Close()

	var fetchedItems []model.Product
	for rows.Next() {
		var item model.Product
		var projects []string
		var productPaymentProcessorResponse []string

		scanFields := ref.buildScanFields(&item,
			&projects,
			&productPaymentProcessorResponse,
			input.Fields)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select", "failed to scan product")
		}

		// PostgreSQL -> {{f282315d-1e65-43fd-8f12-a9c27be60c9e, "Project Name"}}
		// Go -> [f282315d-1e65-43fd-8f12-a9c27be60c9e, Project Name]
		for i := 0; i < len(projects); i += 2 {
			id, err := uuid.Parse(projects[i])
			if err != nil {
				return nil, o11y.RecordError(ctx, span, &model.InvalidProjectIDError{Message: "invalid project ID in projects array"}, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select", "failed to parse project ID")
			}

			item.Projects = &model.Project{
				ID:   id,
				Name: projects[i+1],
			}
		}

		fetchedItems = append(fetchedItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, rows.Err(), ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.Select", "failed to scan rows")
	}

	hasMore := len(fetchedItems) > input.Paginator.Limit
	displayItems := fetchedItems
	if hasMore {
		displayItems = fetchedItems[:input.Paginator.Limit]
	}

	outLen := len(displayItems)
	if outLen == 0 {
		return &model.SelectProductsOutput{
			Items:     make([]model.Product, 0),
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

	ret := &model.SelectProductsOutput{
		Items: displayItems,
		Paginator: model.Paginator{
			Size:      outLen,
			Limit:     input.Paginator.Limit,
			NextToken: nextToken,
			PrevToken: prevToken,
		},
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.repositoryCalls, metricCommonAttributes, "products selected successfully")

	return ret, nil
}

func (ref *ProductsRepository) LinkToPaymentProcessor(ctx context.Context, input *model.LinkProductToPaymentProcessorInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.LinkToPaymentProcessor", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input cannot be nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.LinkToPaymentProcessor")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.LinkToPaymentProcessor")
	}

	query := `
		INSERT INTO products_payment_processors (product_id, payment_processor_id, payment_processor_product_id)
		VALUES ($1, $2, $3)
	`

	slog.Debug("repository.Products.LinkToPaymentProcessor", "query", prettyPrint(query, input.ProductID, input.PaymentProcessorID, input.PaymentProcessorProductID))

	_, err := ref.db.Exec(ctx, query, input.ProductID, input.PaymentProcessorID, input.PaymentProcessorProductID)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.LinkToPaymentProcessor")
	}

	return nil
}

func (ref *ProductsRepository) UnlinkFromPaymentProcessor(ctx context.Context, input *model.UnlinkProductFromPaymentProcessorInput) error {
	ctx, span, metricCommonAttributes, cancel := ref.setupContext(ctx, "repository.Products.UnlinkFromPaymentProcessor", ref.maxQueryTimeout)
	defer cancel()
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input cannot be nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.UnlinkFromPaymentProcessor")
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.UnlinkFromPaymentProcessor")
	}

	query := `
		DELETE FROM products_payment_processors
		WHERE product_id = $1 AND payment_processor_id = $2 AND payment_processor_product_id = $3
	`

	slog.Debug("repository.Products.UnlinkFromPaymentProcessor", "query", prettyPrint(query, input.ProductID, input.PaymentProcessorID, input.PaymentProcessorProductID))

	_, err := ref.db.Exec(ctx, query, input.ProductID, input.PaymentProcessorID, input.PaymentProcessorProductID)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.repositoryCalls, metricCommonAttributes, "repository.Products.UnlinkFromPaymentProcessor")
	}

	return nil
}

// handlePgError maps PostgreSQL errors to domain-specific errors.
// Returns the appropriate domain error or the original error if no mapping exists.
func (ref *ProductsRepository) handlePgError(err error, input any) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.Message, "_pkey") {
				switch v := input.(type) {
				case *model.InsertProductInput:
					return &model.ProductIDAlreadyExistsError{ID: v.ID.String()}
				case *model.UpdateProductInput:
					return &model.ProductIDAlreadyExistsError{ID: v.ID.String()}
				case uuid.UUID:
					return &model.ProductIDAlreadyExistsError{ID: v.String()}
				}
			}

			if strings.Contains(pgErr.Message, "name") {
				switch v := input.(type) {
				case *model.InsertProductInput:
					return &model.ProductNameAlreadyExistsError{Name: v.Name}
				case *model.UpdateProductInput:
					if v.Name != nil {
						return &model.ProductNameAlreadyExistsError{Name: *v.Name}
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

// Helper functions for common patterns

// setupContext creates a context with timeout and starts a span with standard attributes.
// Returns the new context, span, and common metric attributes.
func (ref *ProductsRepository) setupContext(ctx context.Context, operation string, timeout time.Duration) (context.Context, trace.Span, []attribute.KeyValue, context.CancelFunc) {
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

func (ref *ProductsRepository) buildScanFields(item *model.Product,
	projects *[]string,
	productPaymentProcessorResponse *[]string,
	requestedFields string,
) []any {
	scanFields := make([]any, 0)

	if requestedFields == "" {
		// All fields were requested
		return []any{
			&item.ID,
			&item.Name,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.SerialID,
			projects,
			productPaymentProcessorResponse,
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
		case "created_at":
			scanFields = append(scanFields, &item.CreatedAt)
		case "updated_at":
			scanFields = append(scanFields, &item.UpdatedAt)
		case "projects":
			scanFields = append(scanFields, projects)
		case "payment_processors":
			scanFields = append(scanFields, productPaymentProcessorResponse)

		default:
			slog.Warn("repository.Products.buildScanFields", "what", "field not found", "field", field)
		}
	}

	// Always include ID and SerialID fields for pagination
	if !idFound {
		scanFields = append(scanFields, &item.ID)
	}

	scanFields = append(scanFields, &item.SerialID)
	return scanFields
}
