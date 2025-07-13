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

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/projects.go -source=projects.go ProjectsService

// ProjectsService represents the service for the projects.
type ProjectsService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	Create(ctx context.Context, input *model.CreateProjectInput) error
	UpdateByID(ctx context.Context, input *model.UpdateProjectInput) error
	DeleteByID(ctx context.Context, input *model.DeleteProjectInput) error
	List(ctx context.Context, input *model.ListProjectsInput) (*model.ListProjectsOutput, error)
}

// ProjectsHandlerConf represents the handler for the projects.
type ProjectsHandlerConf struct {
	Service       ProjectsService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type projectsHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// ProjectsHandler represents the handler for the projects.
type ProjectsHandler struct {
	service       ProjectsService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       projectsHandlerMetrics
}

// NewProjectsHandler creates a new projectHandler.
func NewProjectsHandler(conf ProjectsHandlerConf) (*ProjectsHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "ProjectsService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	uh := &ProjectsHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		uh.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		uh.metricsPrefix += "_"
	}

	handlerCalls, err := uh.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", uh.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the projects handler"),
	)
	if err != nil {
		return nil, err
	}

	uh.metrics.handlerCalls = handlerCalls

	return uh, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *ProjectsHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /projects", mdw.ThenFunc(ref.list))
	mux.Handle("GET /projects/{project_id}", mdw.ThenFunc(ref.getByID))
	mux.Handle("PUT /projects/{project_id}", mdw.ThenFunc(ref.updateByID))
	mux.Handle("DELETE /projects/{project_id}", mdw.ThenFunc(ref.deleteByID))
	mux.Handle("POST /projects", mdw.ThenFunc(ref.create))
}

// getByID Get a project by ID
//
//	@ID				019797e6-138a-7d04-8db3-1d4755b25db3
//	@Summary		Get project
//	@Description	Retrieve a specific project by its unique identifier
//	@Tags			Projects
//	@Param			project_id	path	string	true	"The project id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.Project
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		404	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/projects/{project_id} [get]
//	@Security		AccessToken
func (ref *ProjectsHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Projects.getByID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.getByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	project, err := ref.service.GetByID(ctx, projectID)
	if err != nil {
		var projectNotFoundError *model.ProjectNotFoundError
		if errors.As(err, &projectNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Projects.getByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, project); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Projects.getByID: called", "project", project.ID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "get project",
		attribute.String("project.id", project.ID.String()))
}

// create Create a project
//
//	@ID				019797e6-138a-7d00-98db-740f21794f11
//	@Summary		Create project
//	@Description	Create a new project with specified configuration
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreateProjectRequest	true	"Create Project Request"
//	@Success		201		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/projects [post]
//	@Security		AccessToken
func (ref *ProjectsHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Projects.create")
	defer span.End()

	claims := ctx.Value(middleware.JwtClaims).(map[string]any)
	slog.Debug("claims", "claims", claims)

	userIDstring, ok := claims["sub"].(string)
	if !ok {
		errorType := &model.InvalidUserIDError{Message: "user ID is not a uuid string"}
		e := recordError(ctx, span, errorType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	userID, err := parseUUIDQueryParams(userIDstring)
	if err != nil {
		errorType := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		e := recordError(ctx, span, errorType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if req.ID == uuid.Nil {
		var err error
		req.ID, err = uuid.NewV7()
		if err != nil {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.create")
			respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
			return
		}
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.CreateProjectInput{
		ID:          req.ID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Disabled:    false,
	}

	if err := ref.service.Create(ctx, input); err != nil {
		var projectIDAlreadyExistsError *model.ProjectIDAlreadyExistsError
		var projectNameAlreadyExistsError *model.ProjectNameAlreadyExistsError
		if errors.As(err, &projectIDAlreadyExistsError) || errors.As(err, &projectNameAlreadyExistsError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Projects.create")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.create")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Projects.create", "name", input.Name)
	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, input.ID.String()))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "Project created",
		attribute.String("project.id", input.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusCreated, model.ProjectsProjectCreatedSuccessfully)
}

// updateByID Update a project by ID
//
//	@ID				019797e6-138a-7cf8-9887-e4c44ad0ae19
//	@Summary		Update project
//	@Description	Modify an existing project by its ID
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string						true	"The project id in UUID format"	Format(uuid)
//	@Param			body		body		model.UpdateProjectRequest	true	"Update Project Request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		409			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects/{project_id} [put]
//	@Security		AccessToken
func (ref *ProjectsHandler) updateByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Projects.updateByID")
	defer span.End()

	projectID, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UpdateProjectInput{
		ID:          projectID,
		Name:        req.Name,
		Description: req.Description,
		Disabled:    req.Disabled,
	}

	if err := ref.service.UpdateByID(ctx, input); err != nil {
		var projectAlreadyExistsError *model.ProjectNameAlreadyExistsError
		var projectNameAlreadyExistsError *model.ProjectNameAlreadyExistsError
		var projectNotFoundError *model.ProjectNotFoundError

		if errors.As(err, &projectAlreadyExistsError) || errors.As(err, &projectNameAlreadyExistsError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Projects.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		if errors.As(err, &projectNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Projects.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, &model.InternalServerError{Message: "failed to update project"}, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Projects.updateByID", "Name", input.Name)
	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Project updated",
		attribute.String("project.id", input.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusOK, model.ProjectsProjectUpdatedSuccessfully)
}

// deleteByID Delete a project by id
//
//	@ID				019797e6-138a-7cf4-8694-e4611baded39
//	@Summary		Delete project
//	@Description	Remove a project permanently from the system
//	@Tags			Projects
//	@Param			project_id	path	string	true	"The project id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/projects/{project_id} [delete]
//	@Security		AccessToken
func (ref *ProjectsHandler) deleteByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Projects.deleteByID")
	defer span.End()

	id, err := parseUUIDQueryParams(r.PathValue("project_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.DeleteProjectInput{
		ID: id,
	}

	if err := ref.service.DeleteByID(ctx, input); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Projects.deleteByID", "id", input.ID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Project deleted",
		attribute.String("project.id", input.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusOK, model.ProjectsProjectDeletedSuccessfully)
}

// list Return a paginated list of Project
//
//	@ID				019797e6-138a-7cf1-b7bb-fa9c5e168c49
//	@Summary		List projects
//	@Description	Retrieve paginated list of all projects in the system
//	@Tags			Projects
//	@Produce		json
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListProjectsResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/projects [get]
//	@Security		AccessToken
func (ref *ProjectsHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Projects.list")
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
		model.ProjectPartialFields,
		model.ProjectFilterFields,
		model.ProjectSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Projects.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListProjectsInput{
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
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Projects.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Projects.list: called", "projects", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list project",
		attribute.Int("project.count", len(out.Items)))
}
