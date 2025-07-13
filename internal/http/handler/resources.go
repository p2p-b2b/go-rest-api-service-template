package handler

import (
	"context"
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

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/resources.go -source=resources.go ResourcesService

// ResourcesService represents the service for the resources.
type ResourcesService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Resource, error)
	List(ctx context.Context, input *model.ListResourcesInput) (*model.ListResourcesOutput, error)

	ListMatches(ctx context.Context, action, resource string, input *model.ListResourcesInput) (*model.ListResourcesOutput, error)
}

// ResourcesHandlerConf represents the configuration for the ResourcesHandler.
type ResourcesHandlerConf struct {
	Service       ResourcesService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type resourcesHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// ResourcesHandler represents the handler for the resources.
type ResourcesHandler struct {
	service       ResourcesService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       resourcesHandlerMetrics
}

// NewResourcesHandler creates a new ResourcesHandler.
func NewResourcesHandler(conf ResourcesHandlerConf) (*ResourcesHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "ResourcesService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	handler := &ResourcesHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the resources handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *ResourcesHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /resources", mdw.ThenFunc(ref.list))
	mux.Handle("GET /resources/{resource_id}", mdw.ThenFunc(ref.getByID))

	mux.Handle("GET /resources/matches", mdw.ThenFunc(ref.listMatches))
}

// getByID Get a resources by id
//
//	@ID				019791cc-06c7-7e86-ad42-b777bfcc9e40
//	@Summary		Get resource
//	@Description	Retrieve a specific resource by its identifier
//	@Tags			Resources
//	@Produce		json
//	@Param			resource_id	path		string	true	"The permission id in UUID format"	Format(uuid)
//	@Success		200			{object}	model.Resource
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		404			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/resources/{resource_id} [get]
//	@Security		AccessToken
func (ref *ResourcesHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Resources.getByID")
	defer span.End()

	resourceID, err := parseUUIDQueryParams(r.PathValue("resource_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Resources.getByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	out, err := ref.service.GetByID(ctx, resourceID)
	if err != nil {
		var resourceNotFound *model.ResourceNotFoundError
		if errors.As(err, &resourceNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Resources.getByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Resources.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Resources.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Resources.getByID", "id", out.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Resources found",
		attribute.String("resource.id", out.ID.String()))
}

// list Return a paginated list of resources
//
//	@ID				019791cc-06c7-7e8e-8d7e-cd3f9296e0fd
//	@Summary		List resources
//	@Description	Retrieve paginated list of all resources in the system
//	@Tags			Resources
//	@Produce		json
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListResourcesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/resources [get]
//	@Security		AccessToken
func (ref *ResourcesHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Resources.list")
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
		model.ResourcesPartialFields,
		model.ResourcesFilterFields,
		model.ResourcesSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Resources.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListResourcesInput{
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
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Resources.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Resources.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Resources.list: called", "resources", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list resources",
		attribute.Int("resources.count", len(out.Items)))
}

// listMatches Return a paginated list of resources matching an action and resource policy pattern
//
//	@ID				019791cc-06c7-7e92-9152-cb35902f79c4
//	@Summary		Match resources
//	@Description	Find resources that match specific action and resource policy patterns
//	@Tags			Resources
//	@Produce		json
//	@Param			action		query		string	true	"Action to filter by"																	Format(string)
//	@Param			resource	query		string	true	"Resource to filter by"																	Format(string)
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListResourcesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/resources/matches [get]
//	@Security		AccessToken
func (ref *ResourcesHandler) listMatches(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Resources.listMatches")
	defer span.End()

	action, err := model.ValidateAction(r.URL.Query().Get("action"))
	if err != nil {
		errType := &model.InvalidActionError{Action: action}
		e := recordError(ctx, span, errType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Resources.listMatches")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	resource, err := model.ValidateResource(r.URL.Query().Get("resource"))
	if err != nil {
		errType := &model.InvalidResourceError{Resource: resource}
		e := recordError(ctx, span, errType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Resources.listMatches")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	params := map[string]any{
		"sort":      r.URL.Query().Get("sort"),
		"filter":    "", // this is disabled because the filter is not supported for this endpoint
		"fields":    r.URL.Query().Get("fields"),
		"nextToken": r.URL.Query().Get("next_token"),
		"prevToken": r.URL.Query().Get("prev_token"),
		"limit":     r.URL.Query().Get("limit"),
	}

	sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(
		params,
		model.ResourcesPartialFields,
		model.ResourcesFilterFields,
		model.ResourcesSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Resources.listMatches")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListResourcesInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.ListMatches(ctx, action, resource, input)
	if err != nil {
		var errNotFound *model.ResourceNotFoundError
		var errInvalidActionError *model.InvalidActionError
		var errInvalidResourceError *model.InvalidResourceError

		if errors.As(err, &errNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Resources.listMatches")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		if errors.As(err, &errInvalidActionError) || errors.As(err, &errInvalidResourceError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Resources.listMatches")
			respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Resources.listMatches")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Resources.listMatches")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Resources.listMatches: called", "resources", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list resources by action and resource",
		attribute.Int("resources.count", len(out.Items)),
		attribute.String("action", action),
		attribute.String("resource", resource))
}
