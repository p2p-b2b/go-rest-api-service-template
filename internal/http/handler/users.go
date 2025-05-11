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
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, input *model.CreateUserInput) error
	UpdateByID(ctx context.Context, input *model.UpdateUserInput) error
	DeleteByID(ctx context.Context, input *model.DeleteUserInput) error
	List(ctx context.Context, input *model.ListUsersInput) (*model.ListUsersOutput, error)
}

// UsersHandler represents the http handler for the user.
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
		return nil, &InvalidServiceError{Name: "UsersService", Reason: "service is required"}
	}

	if conf.OT == nil {
		return nil, &InvalidOpenTelemetryError{Name: "UsersHandler", Reason: "open telemetry is required"}
	}

	uh := &UsersHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		uh.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		uh.metricsPrefix += "_"
	}

	handlerCalls, err := uh.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", uh.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the user handler"),
	)
	if err != nil {
		slog.Error("handler.Users.registerMetrics", "error", err)
		return nil, err
	}
	uh.metrics.handlerCalls = handlerCalls

	return uh, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *UsersHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /users", mdw.ThenFunc((ref.list)))
	mux.Handle("GET /users/{user_id}", mdw.ThenFunc((ref.getByID)))
	mux.Handle("PUT /users/{user_id}", mdw.ThenFunc((ref.updateByID)))
	mux.Handle("POST /users", mdw.ThenFunc((ref.create)))
	mux.Handle("DELETE /users/{user_id}", mdw.ThenFunc((ref.deleteByID)))
}

// getByID Get a user by ID
//
//	@Id				b823ba3c-3b83-4eaa-bdf7-ce1b05237f23
//	@Summary		Get a user by ID
//	@Description	Get a user by ID
//	@Tags			Users
//	@Produce		json
//	@Param			user_id	path		string	true	"The user ID in UUID format"	Format(uuid)
//	@Success		200		{object}	model.User
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		404		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id} [get]
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
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "User found",
		attribute.String("user.id", out.ID.String()),
		attribute.String("user.email", out.Email),
	)
}

// create Create a new user
//
//	@ID				8a1488b0-2d2c-42a0-a57a-6560aaf3ec76
//	@Summary		Create user
//	@Description	Create new user from scratch.
//	@Description	If the id is not provided, it will be generated automatically.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreateUserRequest	true	"Create user request"
//	@Success		201		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users [post]
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
		req.ID = uuid.New()
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	user := &model.CreateUserInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	if err := ref.service.Create(ctx, user); err != nil {
		var userAlreadyExistsError *model.UserAlreadyExistsError
		var userEmailAlreadyExistsError *model.UserEmailAlreadyExistsError
		var invalidUserEmailError *model.InvalidUserEmailError

		if errors.As(err, &userAlreadyExistsError) ||
			errors.As(err, &userEmailAlreadyExistsError) ||
			errors.As(err, &invalidUserEmailError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Users.create")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.create")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.create", "user.email", user.Email)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, user.ID.String()))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "User created",
		attribute.String("user.id", user.ID.String()),
		attribute.String("user.email", user.Email))

	respond.WriteJSONMessage(w, r, http.StatusCreated, model.UserUserCreatedSuccessfully)
}

// updateByID Update a user
//
//	@Id				a7979074-e16c-4aec-86e0-e5a154bbfc51
//	@Summary		Update a user
//	@Description	Update a user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string					true	"The user ID in UUID format"	Format(uuid)
//	@Param			body	body		model.UpdateUserRequest	true	"User update request"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/users/{user_id} [put]
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

	user := model.UpdateUserInput{
		ID:        userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Disabled:  req.Disabled,
	}

	if err := ref.service.UpdateByID(ctx, &user); err != nil {
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

	slog.Debug("handler.Users.updateByID", "user.email", user.Email)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "User updated",
		attribute.String("user.id", user.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.UserUserUpdatedSuccessfully)
}

// deleteByID Delete a user
//
//	@Id				48e60e0a-ea1c-46d4-8729-c47dd82a4e93
//	@Summary		Delete a user
//	@Description	Delete a user
//	@Tags			Users
//	@Param			user_id	path	string	true	"The user ID in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/users/{user_id} [delete]
func (ref *UsersHandler) deleteByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Users.deleteByID")
	defer span.End()

	userID, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	user := model.DeleteUserInput{
		ID: userID,
	}

	if err := ref.service.DeleteByID(ctx, &user); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Users.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Users.deleteByID", "id", user.ID)

	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "User deleted",
		attribute.String("user.id", userID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.UserUserDeletedSuccessfully)
}

// list Return a paginated list of users
//
//	@ID				b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5
//	@Summary		List users
//	@Description	List users with pagination and filtering
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
		model.UserPartialFields,
		model.UserFilterFields,
		model.UserSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Users.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	sParams := &model.ListUsersInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.List(ctx, sParams)
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
