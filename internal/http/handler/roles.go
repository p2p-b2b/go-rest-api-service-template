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

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/roles.go -source=roles.go RolesService

// RolesService represents the service for the roles.
type RolesService interface {
	List(ctx context.Context, input *model.ListRolesInput) (*model.ListRolesOutput, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, input *model.ListRolesInput) (*model.ListRolesOutput, error)
	ListByPolicyID(ctx context.Context, policyID uuid.UUID, input *model.ListRolesInput) (*model.ListRolesOutput, error)

	Create(ctx context.Context, input *model.CreateRoleInput) error

	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	UpdateByID(ctx context.Context, input *model.UpdateRoleInput) error
	DeleteByID(ctx context.Context, input *model.DeleteRoleInput) error

	// link/unlink policies to/from a role
	LinkPolicies(ctx context.Context, input *model.LinkPoliciesToRoleInput) error
	UnlinkPolicies(ctx context.Context, input *model.UnlinkPoliciesFromRoleInput) error

	// link/unlink users to/from a role
	LinkUsers(ctx context.Context, input *model.LinkUsersToRoleInput) error
	UnlinkUsers(ctx context.Context, input *model.UnlinkUsersFromRoleInput) error
}

// RolesHandlerConf represents the handler for the roles.
type RolesHandlerConf struct {
	Service       RolesService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type rolesHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// RolesHandler represents the handler for the roles.
type RolesHandler struct {
	service       RolesService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       rolesHandlerMetrics
}

// NewRolesHandler creates a new roleHandler.
func NewRolesHandler(conf RolesHandlerConf) (*RolesHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "RolesService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	handler := &RolesHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the roles handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *RolesHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /roles", mdw.ThenFunc(ref.list))
	mux.Handle("GET /roles/{role_id}", mdw.ThenFunc(ref.getByID))
	mux.Handle("POST /roles", mdw.ThenFunc(ref.create))
	mux.Handle("PUT /roles/{role_id}", mdw.ThenFunc(ref.updateByID))
	mux.Handle("DELETE /roles/{role_id}", mdw.ThenFunc(ref.deleteByID))

	// link/unlink role to users
	mux.Handle("POST /roles/{role_id}/users", mdw.ThenFunc(ref.linkUsers))
	mux.Handle("DELETE /roles/{role_id}/users", mdw.ThenFunc(ref.unLinkUsers))

	// Link and unlink policies to/from a role
	mux.Handle("POST /roles/{role_id}/policies", mdw.ThenFunc(ref.linkPolicies))
	mux.Handle("DELETE /roles/{role_id}/policies", mdw.ThenFunc(ref.unLinkPolicies))

	// list roles by user id
	mux.Handle("GET /users/{user_id}/roles", mdw.ThenFunc(ref.listByUserID))

	// list roles by policy id
	mux.Handle("GET /policies/{policy_id}/roles", mdw.ThenFunc(ref.listByPolicyID))
}

// getByID Get a role by its ID
//
//	@ID				0198042a-f9c5-76e1-a650-772c826f079e
//	@Summary		Get role
//	@Description	Retrieve a specific role by its unique identifier
//	@Tags			Roles
//	@Param			role_id	path	string	true	"The role id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.Role
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		404	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/roles/{role_id} [get]
//	@Security		AccessToken
func (ref *RolesHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.getByID")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.getByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	out, err := ref.service.GetByID(ctx, roleID)
	if err != nil {
		var roleNotFoundError *model.RoleNotFoundError
		if errors.As(err, &roleNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Roles.getByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.getByID: called", "role.id", out.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "get role",
		attribute.String("role.id", out.ID.String()))
}

// create Create a role
//
//	@ID				0198042a-f9c5-76e5-8fe5-b93a07311c47
//	@Summary		Create role
//	@Description	Create a new role with specified permissions and access levels
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreateRoleRequest	true	"Create role request"
//	@Success		201		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/roles [post]
//	@Security		AccessToken
func (ref *RolesHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.create")
	defer span.End()

	var req model.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if req.ID == uuid.Nil {
		var err error
		req.ID, err = uuid.NewV7()
		if err != nil {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.create")
			respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
			return
		}
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.CreateRoleInput{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := ref.service.Create(ctx, input); err != nil {
		var errNameExists *model.RoleNameAlreadyExistsError
		var errIDExists *model.RoleIDAlreadyExistsError
		var errInvalidByteSequenceError *model.InvalidByteSequenceError

		if errors.As(err, &errIDExists) || errors.As(err, &errNameExists) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Roles.create")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		if errors.As(err, &errInvalidByteSequenceError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.create")
			respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.create")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.create", "name", input.Name)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "Role created", attribute.String("role.id", input.ID.String()))

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, input.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusCreated, model.RolesRoleCreatedSuccessfully)
}

// updateByID Update a role
//
//	@ID				0198042a-f9c5-76e9-922d-2411530cd8f8
//	@Summary		Update role
//	@Description	Modify an existing role by its ID
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string					true	"The model id in UUID format"	Format(uuid)
//	@Param			body	body		model.UpdateRoleRequest	true	"Update role request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/roles/{role_id} [put]
//	@Security		AccessToken
func (ref *RolesHandler) updateByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.updateByID")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UpdateRoleInput{
		ID:          roleID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := ref.service.UpdateByID(ctx, input); err != nil {
		var errRoleNameExists *model.RoleNameAlreadyExistsError
		var errRoleIDExists *model.RoleIDAlreadyExistsError
		var errRoleNotFound *model.RoleNotFoundError
		var errInvalidMessageFormatError *model.InvalidMessageFormatError // bad request

		if errors.As(err, &errRoleNameExists) || errors.As(err, &errRoleIDExists) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Roles.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		if errors.As(err, &errRoleNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Roles.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		if errors.As(err, &errInvalidMessageFormatError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.updateByID", "role.id", input.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Role updated",
		attribute.String("role.id", input.ID.String()))

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))
	respond.WriteJSONMessage(w, r, http.StatusOK, model.RolesRoleUpdatedSuccessfully)
}

// deleteByID Delete a role
//
//	@ID				0198042a-f9c5-76ed-99a5-84923071fa6b
//	@Summary		Delete role
//	@Description	Remove a role permanently from the system
//	@Tags			Roles
//	@Param			role_id	path	string	true	"The role id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/roles/{role_id} [delete]
//	@Security		AccessToken
func (ref *RolesHandler) deleteByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.deleteByID")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.DeleteRoleInput{
		ID: roleID,
	}

	if err := ref.service.DeleteByID(ctx, input); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.deleteByID", "id", input.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Role deleted",
		attribute.String("role.id", input.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.RolesRoleDeletedSuccessfully)
}

// list Retrieves a paginated list of all the roles in the system
//
//	@ID				0198042a-f9c5-76f1-9cf8-37e45b647fc0
//	@Summary		List roles
//	@Description	Retrieve paginated list of all roles in the system
//	@Tags			Roles
//	@Produce		json
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListRolesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/roles [get]
//	@Security		AccessToken
func (ref *RolesHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.list")
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
		model.RolesPartialFields,
		model.RolesFilterFields,
		model.RolesSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListRolesInput{
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
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.list: called", "roles.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list role",
		attribute.Int("roles.count", len(out.Items)))
}

// linkUsers Link users to a role
//
//	@ID				0198042a-f9c5-76f5-8ff6-b4479bdaa6b6
//	@Summary		Link users to role
//	@Description	Associate multiple users with a specific role for authorization
//	@Tags			Roles,Users
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string							true	"The role id in UUID format"	Format(uuid)
//	@Param			body	body		model.LinkUsersToRoleRequest	true	"Link users to role request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/roles/{role_id}/users [post]
//	@Security		AccessToken
func (ref *RolesHandler) linkUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.linkUsers")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.linkUsers")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.LinkUsersToRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.linkUsers")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.linkUsers")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.LinkUsersToRoleInput{
		RoleID:  roleID,
		UserIDs: req.UserIDs,
	}

	if err := ref.service.LinkUsers(ctx, input); err != nil {
		var errRoleNotFound *model.RoleNotFoundError
		if errors.As(err, &errRoleNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Roles.linkUsers")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.linkUsers")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.linkUsers", "role.id", input.RoleID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Users linked to role",
		attribute.String("role.id", input.RoleID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.RolesUsersLinkedSuccessfully)
}

// unLinkUsers Unlink users from a role
//
//	@ID				0198042a-f9c5-76f9-9394-170db55f62f4
//	@Summary		Unlink users from role
//	@Description	Remove user associations from a specific role
//	@Tags			Roles,Users
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string								true	"The Embeddings Role ID in UUID format"	Format(uuid)
//	@Param			body	body		model.UnlinkUsersFromRoleRequest	true	"UnLink users from role request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/roles/{role_id}/users [delete]
//	@Security		AccessToken
func (ref *RolesHandler) unLinkUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.unLinkUsers")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.unLinkUsers")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UnlinkUsersFromRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.unLinkUsers")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.unLinkUsers")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UnlinkUsersFromRoleInput{
		RoleID:  roleID,
		UserIDs: req.UserIDs,
	}

	if err := ref.service.UnlinkUsers(ctx, input); err != nil {
		var errRoleNotFound *model.RoleNotFoundError
		if errors.As(err, &errRoleNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Roles.unLinkUsers")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.unLinkUsers")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.unLinkUsers", "role.id", input.RoleID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Users unlinked from role",
		attribute.String("role.id", input.RoleID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.RolesUsersUnlinkedSuccessfully)
}

// linkPolicies Link policies to a role
//
//	@ID				0198042a-f9c5-76fd-8012-5c9a2957e289
//	@Summary		Link policies to role
//	@Description	Associate multiple policies with a specific role for authorization
//	@Tags			Roles,Policies
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string							true	"The role id in UUID format"	Format(uuid)
//	@Param			body	body		model.LinkPoliciesToRoleRequest	true	"Link policies to role request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/roles/{role_id}/policies [post]
//	@Security		AccessToken
func (ref *RolesHandler) linkPolicies(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.linkPolicies")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.linkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.LinkPoliciesToRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.linkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.linkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.LinkPoliciesToRoleInput{
		RoleID:    roleID,
		PolicyIDs: req.PolicyIDs,
	}

	if err := ref.service.LinkPolicies(ctx, input); err != nil {
		var errRoleNotFound *model.RoleNotFoundError
		if errors.As(err, &errRoleNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Roles.linkPolicies")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.linkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.linkPolicies", "role.id", input.RoleID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Policies linked to role",
		attribute.String("role.id", input.RoleID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.RolesPoliciesLinkedSuccessfully)
}

// unLinkPolicies Unlink policies from a role
//
//	@ID				0198042a-f9c5-7700-9e40-e64f7b8c947c
//	@Summary		Unlink policies from role
//	@Description	Remove policy associations from a specific role
//	@Tags			Roles,Policies
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string								true	"The role id in UUID format"	Format(uuid)
//	@Param			body	body		model.UnlinkPoliciesFromRoleRequest	true	"UnLink policies from role request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/roles/{role_id}/policies [delete]
//	@Security		AccessToken
func (ref *RolesHandler) unLinkPolicies(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.unLinkPolicies")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.unLinkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UnlinkPoliciesFromRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.unLinkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.unLinkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UnlinkPoliciesFromRoleInput{
		RoleID:    roleID,
		PolicyIDs: req.PolicyIDs,
	}

	if err := ref.service.UnlinkPolicies(ctx, input); err != nil {
		var errRoleNotFound *model.RoleNotFoundError
		if errors.As(err, &errRoleNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Roles.unLinkPolicies")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.unLinkPolicies")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.unLinkPolicies", "role.id", input.RoleID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Policies unlinked from role",
		attribute.String("role.id", input.RoleID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.RolesPoliciesUnlinkedSuccessfully)
}

// listByUserID List roles by user ID
//
//	@ID				0198042a-f9c5-7704-b73b-55e2ec093586
//	@Summary		List roles by user
//	@Description	Retrieve paginated list of roles assigned to a specific user
//	@Tags			Roles,Users
//	@Produce		json
//	@Param			user_id		path		string	true	"The user id in UUID format"															Format(uuid)
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListRolesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/users/{user_id}/roles [get]
//	@Security		AccessToken
func (ref *RolesHandler) listByUserID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.listByUserID")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.listByUserID")
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
		model.RolesPartialFields, // Use Role fields here
		model.RolesFilterFields,  // Use Role fields here
		model.RolesSortFields,    // Use Role fields here
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.listByUserID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListRolesInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.ListByUserID(ctx, userID, input)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.listByUserID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.listByUserID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.listByUserID: called", "roles.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list role by user ID",
		attribute.Int("roles.count", len(out.Items)),
		attribute.String("user.id", userID.String()))
}

// listByPolicyID List roles by policy ID
//
//	@ID				0198042a-f9c5-7704-b73b-55e2ec093587
//	@Summary		List roles by policy
//	@Description	Retrieve paginated list of roles associated with a specific policy
//	@Tags			Roles,Policies
//	@Produce		json
//	@Param			policy_id	path		string	true	"The policy id in UUID format"															Format(uuid)
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListRolesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/policies/{policy_id}/roles [get]
//	@Security		AccessToken
func (ref *RolesHandler) listByPolicyID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Roles.listByPolicyID")
	defer span.End()

	policyID, err := parseUUIDQueryParams(r.PathValue("policy_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.listByPolicyID")
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
		model.RolesPartialFields,
		model.RolesFilterFields,
		model.RolesSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Roles.listByPolicyID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListRolesInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.ListByPolicyID(ctx, policyID, input)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.listByPolicyID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Roles.listByPolicyID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Roles.listByPolicyID: called", "roles.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list role by policy ID",
		attribute.Int("roles.count", len(out.Items)),
		attribute.String("policy.id", policyID.String()))
}
