package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/products.go -source=products.go ProductsService

// ProductsService represents the service for the products.
type ProductsService interface {
	GetByIDByProjectID(ctx context.Context, projectID uuid.UUID, id uuid.UUID) (*model.Product, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID, input *model.ListProductsInput) (*model.ListProductsOutput, error)
	List(ctx context.Context, input *model.ListProductsInput) (*model.ListProductsOutput, error)

	CreateByProjectID(ctx context.Context, input *model.CreateProductInput) error
	UpdateByIDByProjectID(ctx context.Context, input *model.UpdateProductInput) error
	DeleteByIDByProjectID(ctx context.Context, input *model.DeleteProductInput) error

	LinkToPaymentProcessor(ctx context.Context, input *model.LinkProductToPaymentProcessorInput) error
	UnlinkFromPaymentProcessor(ctx context.Context, input *model.UnlinkProductFromPaymentProcessorInput) error
}

// ProductsHandlerConf represents the handler for the products.
type ProductsHandlerConf struct {
	Service       ProductsService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type productsHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// ProductsHandler represents the handler for the products.
type ProductsHandler struct {
	service       ProductsService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       productsHandlerMetrics
}

// NewProductsHandler creates a new productHandler.
func NewProductsHandler(conf ProductsHandlerConf) (*ProductsHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "ProductsService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	handler := &ProductsHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the products handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *ProductsHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /products", mdw.ThenFunc(ref.list))
	mux.Handle("GET /projects/{project_id}/products", mdw.ThenFunc(ref.listByProjectID))
	mux.Handle("GET /projects/{project_id}/products/{product_id}", mdw.ThenFunc(ref.getByIDByProjectID))
	mux.Handle("POST /projects/{project_id}/products", mdw.ThenFunc(ref.createByProjectID))
	mux.Handle("PUT /projects/{project_id}/products/{product_id}", mdw.ThenFunc(ref.updateByIDByProjectID))
	mux.Handle("DELETE /projects/{project_id}/products/{product_id}", mdw.ThenFunc(ref.deleteByIDByProjectID))

	mux.Handle("POST /projects/{project_id}/products/{product_id}/payment_processor", mdw.ThenFunc(ref.linkToPaymentProcessor))
	mux.Handle("DELETE /projects/{project_id}/products/{product_id}/payment_processor", mdw.ThenFunc(ref.unlinkFromPaymentProcessor))
}

// getByIDByProjectID Get a product by its ID
//
//	@ID				0198042a-f9c5-7603-99b1-7c20ee58542b
//	@Summary		Get product
//	@Description	Retrieve a specific product by its unique identifier
//	@Tags			Products,Projects
//	@Param			project_id	path	string	true	"The project id in UUID format"	Format(uuid)
//	@Param			product_id	path	string	true	"The product id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.Product
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		404	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products/{product_id} [get]
//	@Security		AccessToken
func (ref *ProductsHandler) getByIDByProjectID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.getByIDByProjectID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.getByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	productID, err := parseUUIDQueryParams(r.PathValue("product_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.getByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	out, err := ref.service.GetByIDByProjectID(ctx, projectID, productID)
	if err != nil {
		var productNotFoundError *model.ProductNotFoundError
		if errors.As(err, &productNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Products.getByIDByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.getByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.getByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Products.getByID: called", "product.id", out.ID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "get product",
		attribute.String("product.id", out.ID.String()))
}

// createByProjectID Create a product
//
//	@ID				0198042a-f9c5-7606-8aab-1c2db5b81a89
//	@Summary		Create product
//	@Description	Create a new product with specified configuration
//	@Tags			Products,Projects
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string						true	"The project id in UUID format"	Format(uuid)
//	@Param			body		body		model.CreateProductRequest	true	"Create product request"
//	@Success		201			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		409			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products [post]
//	@Security		AccessToken
func (ref *ProductsHandler) createByProjectID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.createByProjectID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.createByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.createByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	// Generate a new UUID if not provided
	if req.ID == uuid.Nil {
		var err error
		req.ID, err = uuid.NewV7()
		if err != nil {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.createByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
			return
		}
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.createByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.CreateProductInput{
		ID:          req.ID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := ref.service.CreateByProjectID(ctx, input); err != nil {
		var errNameExists *model.ProductNameAlreadyExistsError
		var errIDExists *model.ProductIDAlreadyExistsError
		var errInvalidByteSequenceError *model.InvalidByteSequenceError

		if errors.As(err, &errIDExists) || errors.As(err, &errNameExists) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Products.createByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		if errors.As(err, &errInvalidByteSequenceError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.createByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.createByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, input.ID.String()))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "Product created", attribute.String("product.id", input.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusCreated, model.ProductsProductCreatedSuccessfully)
}

// updateByIDByProjectID Update a product
//
//	@ID				0198042a-f9c5-7607-b75a-532912a6f35d
//	@Summary		Update product
//	@Description	Modify an existing product by its ID
//	@Tags			Products,Projects
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string						true	"The project id in UUID format"	Format(uuid)
//	@Param			product_id	path		string						true	"The model id in UUID format"	Format(uuid)
//	@Param			body		body		model.UpdateProductRequest	true	"Update product request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		409			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products/{product_id} [put]
//	@Security		AccessToken
func (ref *ProductsHandler) updateByIDByProjectID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.updateByIDByProjectID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.updateByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	productID, err := parseUUIDQueryParams(r.PathValue("product_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.updateByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.updateByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.updateByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := model.UpdateProductInput{
		ID:          productID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := ref.service.UpdateByIDByProjectID(ctx, &input); err != nil {
		var errProductNameExists *model.ProductNameAlreadyExistsError
		var errProductIDExists *model.ProductIDAlreadyExistsError
		var errProductNotFound *model.ProductNotFoundError
		var errInvalidMessageFormatError *model.InvalidMessageFormatError // bad request

		if errors.As(err, &errProductNameExists) || errors.As(err, &errProductIDExists) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Products.updateByIDByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		if errors.As(err, &errProductNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Products.updateByIDByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		if errors.As(err, &errInvalidMessageFormatError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.updateByIDByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.updateByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Product updated",
		attribute.String("product.id", input.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.ProductsProductUpdatedSuccessfully)
}

// deleteByIDByProjectID Delete a product
//
//	@ID				0198042a-f9c5-760a-99c8-1f68d597d300
//	@Summary		Delete product
//	@Description	Remove a product permanently from the system
//	@Tags			Products,Projects
//	@Param			project_id	path	string	true	"The project id in UUID format"	Format(uuid)
//	@Param			product_id	path	string	true	"The product id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products/{product_id} [delete]
//	@Security		AccessToken
func (ref *ProductsHandler) deleteByIDByProjectID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.deleteByIDByProjectID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.deleteByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	productID, err := parseUUIDQueryParams(r.PathValue("product_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.deleteByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := model.DeleteProductInput{
		ID:        productID,
		ProjectID: projectID,
	}

	if err := ref.service.DeleteByIDByProjectID(ctx, &input); err != nil {
		var errProductNotFound *model.ProductNotFoundError
		if errors.As(err, &errProductNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Products.deleteByIDByProjectID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.deleteByIDByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Product deleted",
		attribute.String("product.id", input.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.ProductsProductDeletedSuccessfully)
}

// listByProjectID Retrieves a paginated list of products by project ID
//
//	@ID				0198042a-f9c5-760e-9d2f-94cce8243e5a
//	@Summary		List products by project
//	@Description	Retrieve paginated list of products for a specific project
//	@Tags			Products,Projects
//	@Produce		json
//	@Param			project_id	path		string	true	"The project id in UUID format"															Format(uuid)
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListProductsResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products [get]
//	@Security		AccessToken
func (ref *ProductsHandler) listByProjectID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.listByProjectID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.listByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	// parse the query parameters
	params := map[string]any{
		"sort":      r.URL.Query().Get("sort"),
		"filter":    r.URL.Query().Get("filter"),
		"fields":    r.URL.Query().Get("fields"),
		"nextToken": r.URL.Query().Get("next_token"),
		"prevToken": r.URL.Query().Get("prev_token"),
		"limit":     r.URL.Query().Get("limit"),
	}

	sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(
		params,
		model.ProductsPartialFields,
		model.ProductsFilterFields,
		model.ProductsSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.listByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListProductsInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.ListByProjectID(ctx, projectID, input)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.listByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.listByProjectID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Products.listByProjectID: called", "products.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list product",
		attribute.Int("products.count", len(out.Items)))
}

// list Retrieves a paginated list of all products
//
//	@ID				0198042a-f9c5-7612-a055-58177eca0772
//	@Summary		List products
//	@Description	Retrieve paginated list of all products in the system
//	@Tags			Products
//	@Produce		json
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListProductsResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/products [get]
//	@Security		AccessToken
func (ref *ProductsHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.list")
	defer span.End()

	// parse the query parameters
	params := map[string]any{
		"sort":      r.URL.Query().Get("sort"),
		"filter":    r.URL.Query().Get("filter"),
		"fields":    r.URL.Query().Get("fields"),
		"nextToken": r.URL.Query().Get("next_token"),
		"prevToken": r.URL.Query().Get("prev_token"),
		"limit":     r.URL.Query().Get("limit"),
	}

	sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(
		params,
		model.ProductsPartialFields,
		model.ProductsFilterFields,
		model.ProductsSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListProductsInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.List(ctx, input)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Products.list: called", "products.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list product",
		attribute.Int("products.count", len(out.Items)))
}

// linkToPaymentProcessor Links a product to a payment processor
//
//	@ID				0198042a-f9c5-7616-8c3b-e4f19d83a033
//	@Summary		Link product to payment processor
//	@Description	Associate a product with a payment processor to enable billing and invoicing
//	@Tags			Products,Projects
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string										true	"The project id in UUID format"	Format(uuid)
//	@Param			product_id	path		string										true	"The product id in UUID format"	Format(uuid)
//	@Param			body		body		model.LinkProductToPaymentProcessorRequest	true	"Link product to payment processor request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products/{product_id}/payment_processor [post]
//	@Security		AccessToken
func (ref *ProductsHandler) linkToPaymentProcessor(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.linkToPaymentProcessor")
	defer span.End()

	productID, err := parseUUIDQueryParams(r.PathValue("product_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.linkToPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.LinkProductToPaymentProcessorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.linkToPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.linkToPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.LinkProductToPaymentProcessorInput{
		ProductID:                 productID,
		PaymentProcessorID:        req.PaymentProcessorID,
		PaymentProcessorProductID: req.PaymentProcessorProductID,
	}

	if err := ref.service.LinkToPaymentProcessor(ctx, input); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.linkToPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	respond.WriteJSONMessage(w, r, http.StatusOK, "product linked to payment processor successfully")
}

// unlinkFromPaymentProcessor Unlinks a product from a payment processor
//
//	@ID				0198042a-f9c5-761a-bd02-da039b52bea2
//	@Summary		Unlink product from payment processor
//	@Description	Remove the association between a product and a payment processor
//	@Tags			Products,Projects
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string											true	"The project id in UUID format"	Format(uuid)
//	@Param			product_id	path		string											true	"The product id in UUID format"	Format(uuid)
//	@Param			body		body		model.UnlinkProductFromPaymentProcessorRequest	true	"Unlink product from payment processor request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects/{project_id}/products/{product_id}/payment_processor [delete]
//	@Security		AccessToken
func (ref *ProductsHandler) unlinkFromPaymentProcessor(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Products.unlinkFromPaymentProcessor")
	defer span.End()

	productID, err := parseUUIDQueryParams(r.PathValue("product_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.unlinkFromPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UnlinkProductFromPaymentProcessorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.unlinkFromPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Products.unlinkFromPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UnlinkProductFromPaymentProcessorInput{
		ProductID:                 productID,
		PaymentProcessorID:        req.PaymentProcessorID,
		PaymentProcessorProductID: req.PaymentProcessorProductID,
	}

	if err := ref.service.UnlinkFromPaymentProcessor(ctx, input); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Products.unlinkFromPaymentProcessor")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	respond.WriteJSONMessage(w, r, http.StatusOK, "product unlinked from payment processor successfully")
}
