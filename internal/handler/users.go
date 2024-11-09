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
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -package=mocks -destination=../../mocks/handler/users.go -source=users.go UserService

// UserService represents the service for the user.
type UserService interface {
	UserHealthCheck(ctx context.Context) (service.Health, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*service.User, error)
	CreateUser(ctx context.Context, input *service.CreateUserInput) error
	UpdateUser(ctx context.Context, input *service.UpdateUserInput) error
	DeleteUser(ctx context.Context, input *service.DeleteUserInput) error
	ListUsers(ctx context.Context, input *service.ListUserInput) (*service.ListUsersOutput, error)
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
func NewUserHandler(conf UserHandlerConf) *UserHandler {
	if conf.Service == nil {
		slog.Error("service is required")
		panic("service is required")
	}

	if conf.OT == nil {
		slog.Error("open telemetry is required")
		panic("open telemetry is required")
	}

	uh := &UserHandler{
		service: conf.Service,
		ot:      conf.OT,
	}
	if conf.MetricsPrefix != "" {
		uh.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		uh.metricsPrefix += "_"
	}

	if err := uh.registerMetrics(); err != nil {
		slog.Error("failed to register metrics", "error", err)
		panic(err)
	}

	return uh
}

// registerMetrics registers the metrics for the user handler.
func (h *UserHandler) registerMetrics() error {
	handlerCalls, err := h.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", h.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the user handler"),
	)
	if err != nil {
		slog.Error("handler.users.registerMetrics", "error", err)
		return err
	}
	h.metrics.handlerCalls = handlerCalls

	return nil
}

// RegisterRoutes registers the routes for the user.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.listUsers)
	mux.HandleFunc("POST /users", h.createUser)

	mux.HandleFunc("GET /users/{user_id}", h.getByID)
	mux.HandleFunc("PUT /users/{user_id}", h.updateUser)
	mux.HandleFunc("DELETE /users/{user_id}", h.deleteUser)
}

// getByID Get a user by ID
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param user_id path string true "The user ID in UUID format"
// @Success 200 {object} User
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /users/{user_id} [get]
func (h *UserHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.getByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users.getByID"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users.getByID"),
		attribute.String("http.method", r.Method),
	}

	id, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.getByID", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	sUser, err := h.service.GetUserByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.getByID", "error", err.Error())

		if errors.Is(err, service.ErrUserNotFound) {
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("code", fmt.Sprintf("%d", http.StatusNotFound)),
				),
			)

			WriteJSONMessage(w, r, http.StatusNotFound, err.Error())
			return
		}

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)),
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// create user from service.User
	user := &User{
		ID:        sUser.ID,
		FirstName: sUser.FirstName,
		LastName:  sUser.LastName,
		Email:     sUser.Email,
		CreatedAt: sUser.CreatedAt,
		UpdatedAt: sUser.UpdatedAt,
	}

	if err := WriteJSONData(w, http.StatusOK, user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.getByID", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)),
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.users.getByID", "user", user)
	span.SetStatus(codes.Ok, "User found")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("code", fmt.Sprintf("%d", http.StatusOK)),
		),
	)
}

// createUser Create a new user
// @Summary Create a new user.
// @Description Create a new user from scratch.
// @Description If the id is not provided, it will be generated automatically.
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "CreateUserRequest"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 409 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.createUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users.createUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users.createUser"),
		attribute.String("http.method", r.Method),
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.createUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.createUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := &service.CreateUserInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	slog.Debug("handler.users.createUser", "user", user)
	if err := h.service.CreateUser(ctx, user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.createUser", "error", err.Error())

		if errors.Is(err, service.ErrUserIDAlreadyExists) ||
			errors.Is(err, service.ErrUserEmailAlreadyExists) {

			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusConflict)))...,
				),
			)

			WriteJSONMessage(w, r, http.StatusConflict, err.Error())
			return
		}

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	span.SetStatus(codes.Ok, "User created")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusCreated)))...,
		),
	)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s/%s", r.Header.Get("Origin"), r.RequestURI, user.ID.String()))
	WriteJSONMessage(w, r, http.StatusCreated, "User created")
}

// updateUser Update a user
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path string true "The user ID in UUID format"
// @Param user body UpdateUserRequest true "User"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 409 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /users/{user_id} [put]
func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.updateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users.updateUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users.updateUser"),
		attribute.String("http.method", r.Method),
	}

	id, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.updateUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.updateUser", "error", err.Error())

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.updateUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := service.UpdateUserInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	if err := h.service.UpdateUser(ctx, &user); err != nil {
		span.SetStatus(codes.Error, ErrInternalServerError.Error())
		span.RecordError(ErrInternalServerError)
		slog.Error("handler.users.updateUser", "error", ErrInternalServerError.Error())

		if errors.Is(err, service.ErrUserEmailAlreadyExists) ||
			errors.Is(err, service.ErrUserNotFound) {
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusConflict)))...,
				),
			)

			WriteJSONMessage(w, r, http.StatusConflict, err.Error())
			return
		}

		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.users.updateUser", "user", user)
	span.SetStatus(codes.Ok, "User updated")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)

	// Location header is required for RESTful APIs
	w.Header().Set("Location", fmt.Sprintf("%s%s", r.Header.Get("Origin"), r.RequestURI))
	WriteJSONMessage(w, r, http.StatusOK, "User updated")
}

// deleteUser Delete a user
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Param user_id path string true "The user ID in UUID format"
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /users/{user_id} [delete]
func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.deleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users.deleteUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users.deleteUser"),
		attribute.String("http.method", r.Method),
	}

	id, err := parseUUIDQueryParams(r.PathValue("user_id"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.deleteUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := service.DeleteUserInput{
		ID: id,
	}

	if err := h.service.DeleteUser(ctx, &user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.deleteUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.users.deleteUser", "user", user)
	span.SetStatus(codes.Ok, "User deleted")
	span.SetAttributes(attribute.String("user.id", id.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)

	WriteJSONMessage(w, r, http.StatusOK, "User deleted")
}

// listUsers Return a paginated list of users
// @Summary List all users
// @Description List all users
// @Tags users
// @Produce json
// @Param sort query string false "Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"
// @Param filter query string false "Filter field. Example: id=1 AND first_name='John'"
// @Param fields query string false "Fields to return. Example: id,first_name,last_name"
// @Param next_token query string false "Next cursor"
// @Param prev_token query string false "Previous cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} ListUsersResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /users [get]
func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.listUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users.listUsers"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users.listUsers"),
		attribute.String("http.method", r.Method),
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
		slog.Error("handler.users.listUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusBadRequest, err.Error())
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

	sUsers, err := h.service.ListUsers(ctx, sParams)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.listUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
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
			CreatedAt: sUser.CreatedAt,
			UpdatedAt: sUser.UpdatedAt,
		}
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	users.Paginator.GeneratePages(location)

	if err := WriteJSONData(w, http.StatusOK, users); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.users.listUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteJSONMessage(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("handler.users.listUsers: called", "users", len(users.Items))
	span.SetStatus(codes.Ok, "list users")
	span.SetAttributes(attribute.Int("users.count", len(users.Items)))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)
}
