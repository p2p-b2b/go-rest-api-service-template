package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -package=mocks -destination=../../../mocks/handler/users.go -source=users.go UserService

// UserService represents the service for the user.
type UserService interface {
	HealthCheck(ctx context.Context) (service.Health, error)
	GetByID(ctx context.Context, id uuid.UUID) (*service.User, error)
	GetByEmail(ctx context.Context, email string) (*service.User, error)
	Create(ctx context.Context, input *service.CreateUserInput) error
	Update(ctx context.Context, input *service.UpdateUserInput) error
	Delete(ctx context.Context, input *service.DeleteUserInput) error
	List(ctx context.Context, input *service.ListUserInput) (*service.ListUsersOutput, error)
}

// UserHandler represents the handler for the user.
type UserHandlerConf struct {
	Service       UserService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type userHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// UserHandler represents the handler for the user.
type UserHandler struct {
	service       UserService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       userHandlerMetrics
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(conf UserHandlerConf) (*UserHandler, error) {
	if conf.Service == nil {
		slog.Error("service is required")
		return nil, ErrInvalidService
	}

	if conf.OT == nil {
		slog.Error("open telemetry is required")
		return nil, ErrInvalidOpenTelemetry
	}

	uh := &UserHandler{
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
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/health", h.getHealth)

	mux.HandleFunc("GET /users", h.listUsers)
	mux.HandleFunc("POST /users", h.createUser)
	mux.HandleFunc("GET /users/{user_id}", h.getByID)
	mux.HandleFunc("PUT /users/{user_id}", h.updateUser)
	mux.HandleFunc("DELETE /users/{user_id}", h.deleteUser)
}

// getHealth returns the health of the service
//
// @Summary Get the health of the service
// @Description Get the health of the service
// @Tags Users,Health
// @Produce json
// @Success 200 {object} Health
// @Failure 500 {object} respond.HTTPMessage
// @Router /users/health [get]
func (h *UserHandler) getHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	sHealth, err := h.service.HealthCheck(ctx)
	if err != nil {
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	health := &Health{
		Status: sHealth.Status.String(),
		Checks: make([]Check, len(sHealth.Checks)),
	}

	for i, sCheck := range sHealth.Checks {
		health.Checks[i] = Check{
			Name:   sCheck.Name,
			Kind:   sCheck.Kind,
			Status: sCheck.Status.String(),
			Data:   sCheck.Data,
		}
	}

	if err := respond.WriteJSONData(w, http.StatusOK, health); err != nil {
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	slog.Debug("handler.Users.getHealth: called")
}

// getByID Get a user by ID
//
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags Users
// @Produce json
// @Param user_id path string true "The user ID in UUID format" Format(uuid)
// @Success 200 {object} User
// @Failure 400 {object} respond.HTTPMessage
// @Failure 404 {object} respond.HTTPMessage
// @Failure 500 {object} respond.HTTPMessage
// @Router /users/{user_id} [get]
func (h *UserHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.Users.getByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.Users.getByID"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.Users.getByID"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	id, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.getByID", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	sUser, err := h.service.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.getByID", "error", err.Error())

		if errors.Is(err, service.ErrUserNotFound) {
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("code", fmt.Sprintf("%d", http.StatusNotFound)),
				),
			)

			respond.WriteJSONMessage(w, r, http.StatusNotFound, err.Error())
			return
		}

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)),
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// create user from service.User
	user := &User{
		ID:        sUser.ID,
		FirstName: sUser.FirstName,
		LastName:  sUser.LastName,
		Email:     sUser.Email,
		Disabled:  sUser.Disabled,
		CreatedAt: sUser.CreatedAt,
		UpdatedAt: sUser.UpdatedAt,
	}

	if err := respond.WriteJSONData(w, http.StatusOK, user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.getByID", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)),
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.Users.getByID", "email", user.Email)
	span.SetStatus(codes.Ok, "User found")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("code", fmt.Sprintf("%d", http.StatusOK)),
		),
	)
}

// createUser Create a new user
//
// @Summary Create a new user.
// @Description Create a new user from scratch.
// @Description If the id is not provided, it will be generated automatically.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "CreateUserRequest" Format(json)
// @Success 201 {object} respond.HTTPMessage
// @Failure 400 {object} respond.HTTPMessage
// @Failure 409 {object} respond.HTTPMessage
// @Failure 500 {object} respond.HTTPMessage
// @Router /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.Users.createUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.Users.createUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.Users.createUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("handler.Users.createUser", "error", err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}

	if err := req.Validate(); err != nil {
		slog.Error("handler.Users.createUser", "error", err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := &service.CreateUserInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	if err := h.service.Create(ctx, user); err != nil {
		slog.Error("handler.Users.createUser", "error", err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		if errors.Is(err, service.ErrUserIDAlreadyExists) ||
			errors.Is(err, service.ErrUserEmailAlreadyExists) {

			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusConflict)))...,
				),
			)

			respond.WriteJSONMessage(w, r, http.StatusConflict, err.Error())
			return
		}

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.Users.createUser", "email", user.Email)
	span.SetStatus(codes.Ok, "User created")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusCreated)))...,
		),
	)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, user.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusCreated, "User created")
}

// updateUser Update a user
//
// @Summary Update a user
// @Description Update a user
// @Tags Users
// @Accept json
// @Produce json
// @Param user_id path string true "The user ID in UUID format" Format(uuid)
// @Param user body UpdateUserRequest true "User" Format(json)
// @Success 200 {object} respond.HTTPMessage
// @Failure 400 {object} respond.HTTPMessage
// @Failure 409 {object} respond.HTTPMessage
// @Failure 500 {object} respond.HTTPMessage
// @Router /users/{user_id} [put]
func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.Users.updateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.Users.updateUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.Users.updateUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	id, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.updateUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.updateUser", "error", err.Error())

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.updateUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := service.UpdateUserInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Disabled:  req.Disabled,
	}

	if err := h.service.Update(ctx, &user); err != nil {
		span.SetStatus(codes.Error, ErrInternalServerError.Error())
		span.RecordError(ErrInternalServerError)
		slog.Error("handler.Users.updateUser", "error", ErrInternalServerError.Error())

		if errors.Is(err, service.ErrUserEmailAlreadyExists) ||
			errors.Is(err, service.ErrUserNotFound) {
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusConflict)))...,
				),
			)

			respond.WriteJSONMessage(w, r, http.StatusConflict, err.Error())
			return
		}

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.Users.updateUser", "email", user.Email)
	span.SetStatus(codes.Ok, "User updated")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))
	respond.WriteJSONMessage(w, r, http.StatusOK, "User updated")
}

// deleteUser Delete a user
//
// @Summary Delete a user
// @Description Delete a user
// @Tags Users
// @Param user_id path string true "The user ID in UUID format" Format(uuid)
// @Produce json
// @Success 204 {object} respond.HTTPMessage
// @Failure 400 {object} respond.HTTPMessage
// @Failure 500 {object} respond.HTTPMessage
// @Router /users/{user_id} [delete]
func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.Users.deleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.Users.deleteUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.Users.deleteUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	id, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.deleteUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := service.DeleteUserInput{
		ID: id,
	}

	if err := h.service.Delete(ctx, &user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.deleteUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.Users.deleteUser", "id", user.ID)
	span.SetStatus(codes.Ok, "User deleted")
	span.SetAttributes(attribute.String("user.id", id.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)

	respond.WriteJSONMessage(w, r, http.StatusNoContent, "User deleted")
}

// listUsers Return a paginated list of users
//
// @Summary List all users
// @Description List all users
// @Tags Users
// @Produce json
// @Param sort query string false "Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC" Format(string)
// @Param filter query string false "Filter field. Example: id=1 AND first_name='John'" Format(string)
// @Param fields query string false "Fields to return. Example: id,first_name,last_name" Format(string)
// @Param next_token query string false "Next cursor" Format(string)
// @Param prev_token query string false "Previous cursor" Format(string)
// @Param limit query int false "Limit" Format(int)
// @Success 200 {object} ListUsersResponse
// @Failure 400 {object} respond.HTTPMessage
// @Failure 500 {object} respond.HTTPMessage
// @Router /users [get]
func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.Users.listUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.Users.listUsers"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.Users.listUsers"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
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
		repository.UserPartialFields,
		repository.UserFilterFields,
		repository.UserSortFields,
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.listUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	sParams := &service.ListUserInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: paginator.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	sUsers, err := h.service.List(ctx, sParams)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.listUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	users := &ListUsersResponse{
		Items:     make([]*User, len(sUsers.Items)),
		Paginator: sUsers.Paginator,
	}

	for i, sUser := range sUsers.Items {
		users.Items[i] = &User{
			ID:        sUser.ID,
			FirstName: sUser.FirstName,
			LastName:  sUser.LastName,
			Email:     sUser.Email,
			Disabled:  sUser.Disabled,
			CreatedAt: sUser.CreatedAt,
			UpdatedAt: sUser.UpdatedAt,
		}
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	users.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, users); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.Users.listUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.Users.listUsers: called", "users", len(users.Items))
	span.SetStatus(codes.Ok, "list users")
	span.SetAttributes(attribute.Int("users.count", len(users.Items)))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)
}
