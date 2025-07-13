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

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/users.go -source=users.go UsersService

// UsersService represents the service for the user.
type UsersService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	Create(ctx context.Context, input *model.CreateUserInput) error
	UpdateByID(ctx context.Context, input *model.UpdateUserInput) error
	DeleteByID(ctx context.Context, input *model.DeleteUserInput) error

	List(ctx context.Context, input *model.ListUsersInput) (*model.ListUsersOutput, error)
	ListByRoleID(ctx context.Context, roleID uuid.UUID, input *model.ListUsersInput) (*model.ListUsersOutput, error)

	SelectAuthz(ctx context.Context, userID uuid.UUID) (map[string]any, error)
	LinkRoles(ctx context.Context, input *model.LinkRolesToUserInput) error
	UnLinkRoles(ctx context.Context, input *model.UnLinkRolesFromUsersInput) error
}

// UsersHandlerConf represents the configuration for the user handler.
type UsersHandlerConf struct {
	Service       UsersService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type usersHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// UsersHandler represents the handler for the user.
type UsersHandler struct {
	service       UsersService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       usersHandlerMetrics
}

// NewUsersHandler creates a new UsersHandler.
func NewUsersHandler(conf UsersHandlerConf) (*UsersHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "UsersService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	handler := &UsersHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the user handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *UsersHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /users", mdw.ThenFunc((ref.list)))
	mux.Handle("POST /users", mdw.ThenFunc(ref.create))
	mux.Handle("GET /users/{user_id}", mdw.ThenFunc(ref.getByID))
	mux.Handle("PUT /users/{user_id}", mdw.ThenFunc(ref.updateByID))
	mux.Handle("DELETE /users/{user_id}", mdw.ThenFunc(ref.deleteByID))

	// link/unlink roles to user
	mux.Handle("POST /users/{user_id}/roles", mdw.ThenFunc(ref.linkRoles))
	mux.Handle("DELETE /users/{user_id}/roles", mdw.ThenFunc(ref.unLinkRoles))

	// select authz
	mux.Handle("GET /users/{user_id}/authz", mdw.ThenFunc(ref.selectAuthz))

	// list users by role
	mux.Handle("GET /roles/{role_id}/users", mdw.ThenFunc(ref.listByRoleID))
}

// getByID Get a user by ID
//
//	@ID				0198042a-f9c5-75df-b843-b92a4d5c590e
//	@Summary		Get user
//	@Description	Retrieve a specific user account by its unique identifier
//	@Tags			Users
//	@Produce		json
//	@Param			user_id	path		string	true	"The user ID in UUID format"	Format(uuid)
//	@Success		200		{object}	model.User
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		404		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id} [get]
//	@Security		AccessToken
func (ref *UsersHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.getByID")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.getByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	out, err := ref.service.GetByID(ctx, userID)
	if err != nil {
		var UserNotFoundError *model.UserNotFoundError
		if errors.As(err, &UserNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Users.getByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.getByID", "user.email", out.Email)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, model.UsersUserFound,
		attribute.String("user.id", out.ID.String()),
		attribute.String("user.email", out.Email),
	)
}

// create Create a new user
//
//	@ID				0198042a-f9c5-75e3-acf6-6901bb33ae65
//	@Summary		Create user
//	@Description	Create a new user account with specified configuration
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreateUserRequest	true	"Create user request"
//	@Success		201		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users [post]
//	@Security		AccessToken
func (ref *UsersHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.create")
	defer span.End()

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if req.ID == uuid.Nil {
		var err error
		req.ID, err = uuid.NewV7()
		if err != nil {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.create")
			respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
			return
		}
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.CreateUserInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	if err := ref.service.Create(ctx, input); err != nil {
		var userAlreadyExistsError *model.UserAlreadyExistsError
		var userEmailAlreadyExistsError *model.UserEmailAlreadyExistsError
		var invalidEmailError *model.InvalidEmailError

		if errors.As(err, &userAlreadyExistsError) ||
			errors.As(err, &userEmailAlreadyExistsError) ||
			errors.As(err, &invalidEmailError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Users.create")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.create")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.create", "user.email", input.Email)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, input.ID.String()))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "User created",
		attribute.String("user.id", input.ID.String()),
		attribute.String("user.email", input.Email))

	respond.WriteJSONMessage(w, r, http.StatusCreated, model.UsersUserCreatedSuccessfully)
}

// updateByID Update a user
//
//	@ID				0198042a-f9c5-75e7-8cb9-231bee55c64e
//	@Summary		Update user
//	@Description	Modify an existing user account by its ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string					true	"The user ID in UUID format"	Format(uuid)
//	@Param			body	body		model.UpdateUserRequest	true	"Update user request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id} [put]
//	@Security		AccessToken
func (ref *UsersHandler) updateByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.updateByID")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UpdateUserInput{
		ID:        userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Disabled:  req.Disabled,
	}

	if err := ref.service.UpdateByID(ctx, input); err != nil {
		var userAlreadyExistsError *model.UserAlreadyExistsError
		var userEmailAlreadyExistsError *model.UserEmailAlreadyExistsError
		var userNotFoundError *model.UserNotFoundError

		if errors.As(err, &userAlreadyExistsError) ||
			errors.As(err, &userEmailAlreadyExistsError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Users.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		if errors.As(err, &userNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Users.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.updateByID", "user.email", input.Email)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "User updated",
		attribute.String("user.id", input.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.UsersUserUpdatedSuccessfully)
}

// deleteByID Delete a user
//
//	@ID				0198042a-f9c5-75eb-b683-6c1847af7108
//	@Summary		Delete user
//	@Description	Remove a user account permanently from the system
//	@Tags			Users
//	@Param			user_id	path	string	true	"The user ID in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/users/{user_id} [delete]
//	@Security		AccessToken
func (ref *UsersHandler) deleteByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.deleteByID")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.DeleteUserInput{
		ID: userID,
	}

	if err := ref.service.DeleteByID(ctx, input); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.deleteByID", "id", input.ID)

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "User deleted",
		attribute.String("user.id", userID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.UsersUserDeletedSuccessfully)
}

// list Return a paginated list of users
//
//	@ID				0198042a-f9c5-75ef-8ea1-29ecbbe01a2e
//	@Summary		List users
//	@Description	Retrieve paginated list of all users in the system
//	@Tags			Users
//	@Produce		json
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListUsersResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/users [get]
//	@Security		AccessToken
func (ref *UsersHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.list")
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
		model.UsersPartialFields,
		model.UsersFilterFields,
		model.UsersSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListUsersInput{
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
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.listUsers: called", "users", len(out.Items))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "List users",
		attribute.Int("users.count", len(out.Items)))
}

// linkRoles Link roles to user
//
//	@ID				0198042a-f9c5-75f3-985f-d30e67bb3688
//	@Summary		Link roles to user
//	@Description	Associate multiple roles with a user within a specific project
//	@Tags			Users,Roles
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string							true	"The user ID in UUID format"	Format(uuid)
//	@Param			user	body		model.LinkRolesToUserRequest	true	"Link Roles Request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id}/roles [post]
//	@Security		AccessToken
func (ref *UsersHandler) linkRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.linkRoles")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.LinkRolesToUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.LinkRolesToUserInput{
		UserID:  userID,
		RoleIDs: req.RoleIDs,
	}

	if err := ref.service.LinkRoles(ctx, input); err != nil {
		var userAlreadyExistsError *model.UserAlreadyExistsError
		var userNotFoundError *model.UserNotFoundError
		if errors.As(err, &userAlreadyExistsError) || errors.As(err, &userNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Users.linkRoles")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.linkRoles", "user.id", userID.String())

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Roles linked to user",
		attribute.String("user.id", userID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.UsersRoleLinkedToUserSuccessfully)
}

// unLinkRoles Unlink roles from user
//
//	@ID				0198042a-f9c5-75f7-b802-343518ee3788
//	@Summary		Unlink roles from user
//	@Description	Remove role associations from a user within a specific project
//	@Tags			Users,Roles
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string								true	"The user ID in UUID format"	Format(uuid)
//	@Param			body	body		model.UnlinkRolesFromUserRequest	true	"UnLink Roles Request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id}/roles [delete]
//	@Security		AccessToken
func (ref *UsersHandler) unLinkRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.unLinkRoles")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.unLinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UnlinkRolesFromUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.unLinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.unLinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UnLinkRolesFromUsersInput{
		UserID:  userID,
		RoleIDs: req.RoleIDs,
	}

	if err := ref.service.UnLinkRoles(ctx, input); err != nil {
		var userAlreadyExistsError *model.UserAlreadyExistsError
		var userNotFoundError *model.UserNotFoundError
		if errors.As(err, &userAlreadyExistsError) || errors.As(err, &userNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Users.unLinkRoles")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.unLinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.unLinkRoles", "user.id", userID.String())

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "Roles unlinked from user",
		attribute.String("user.id", userID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.UsersRoleUnlinkedFromUserSuccessfully)
}

// selectAuthz Get user authorization
//
//	@ID				0198042a-f9c5-75fb-b324-ec962beb2277
//	@Summary		Get user authorization
//	@Description	Retrieve user authorization permissions and roles for access control
//	@Tags			Users,Auth
//	@Produce		json
//	@Param			user_id	path		string	true	"The user ID in UUID format"	Format(uuid)
//	@Success		200		{object}	map[string]any
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		404		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id}/authz [get]
//	@Security		AccessToken
func (ref *UsersHandler) selectAuthz(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.selectAuthz")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.selectAuthz")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	out, err := ref.service.SelectAuthz(ctx, userID)
	if err != nil {
		var UserNotFoundError *model.UserNotFoundError
		if errors.As(err, &UserNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Users.selectAuthz")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.selectAuthz")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.selectAuthz")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.selectAuthz", "user.id", userID.String())

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "User authorization retrieved",
		attribute.String("user.id", userID.String()))
}

// listByRoleID List the users linked to a role
//
//	@ID				0198042a-f9c5-75ff-bbfc-224bf4342886
//	@Summary		List users by role
//	@Description	Retrieve paginated list of users associated with a specific role
//	@Tags			Users,Roles
//	@Produce		json
//	@Param			role_id		path		string	true	"The role id in UUID format"															Format(uuid)
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListUsersResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/roles/{role_id}/users [get]
//	@Security		AccessToken
func (ref *UsersHandler) listByRoleID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.listByRoleID")
	defer span.End()

	// Get project ID
	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.listByRoleID")
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
		model.UsersPartialFields,
		model.UsersFilterFields,
		model.UsersSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	span.SetAttributes(attribute.String("role.id", roleID.String()))

	input := &model.ListUsersInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.ListByRoleID(ctx, roleID, input)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.listByRoleID", "users", len(out.Items), "role_id", roleID)

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "List users by role ID",
		attribute.Int("users.count", len(out.Items)),
		attribute.String("role.id", roleID.String()))
}
