package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/query"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -package=mocks -destination=../../mocks/handler/users.go -source=users.go UserService

// UserService represents a service for managing users.
type UserService interface {
	// UserHealthCheck verifies a connection to the repository is still alive.
	UserHealthCheck(ctx context.Context) (model.Health, error)

	// GetUserByID returns the user with the specified ID.
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// CreateUser inserts a new user into the database.
	CreateUser(ctx context.Context, user *model.UserParamsInput) error

	// UpdateUser updates the user.
	UpdateUser(ctx context.Context, user *model.UserParamsInput) error

	// DeleteUser deletes the user.
	DeleteUser(ctx context.Context, user *model.UserParamsInput) error

	// ListUsers returns a list of users.
	ListUsers(ctx context.Context, params *model.ListUserRequest) (*model.ListUserResponse, error)
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
		slog.Error("failed to create user_handler.calls metric", "error", err)
		return err
	}
	h.metrics.handlerCalls = handlerCalls

	return nil
}

// RegisterRoutes registers the routes for the user.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/{uid}", h.GetByID)
	mux.HandleFunc("PUT /users/{uid}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{uid}", h.DeleteUser)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users", h.ListUsers)
}

// GetByID Get a user by ID
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param uid path string true "The user ID in UUID format"
// @Param fields query string false "Fields to return. Example: id,first_name,last_name"
// @Success 200 {object} model.User
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users/{uid} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users"),
		attribute.String("function", "GetByID"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	uidString := r.PathValue("uid")
	if uidString == "" {
		span.SetStatus(codes.Error, ErrUserIDRequired.Error())
		span.RecordError(ErrUserIDRequired)
		slog.Error("handler.GetByID", "error", ErrUserIDRequired.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrUserIDRequired.Error())
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(uidString)
	if err != nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("handler.GetByID", "error", ErrInvalidUserID.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)
		WriteError(w, r, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	if id == uuid.Nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("handler.GetByID", "error", ErrInvalidUserID.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	user, err := h.service.GetUserByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, ErrInternalServerError.Error())
		span.RecordError(err)
		slog.Error("handler.GetByID", "error", ErrInternalServerError.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)),
			),
		)
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("code", fmt.Sprintf("%d", http.StatusOK)),
		),
	)

	// encode and write the response
	if err := json.NewEncoder(w).Encode(user); err != nil {
		span.SetStatus(codes.Error, ErrEncodingPayload.Error())
		span.RecordError(ErrEncodingPayload)
		slog.Error("handler.GetByID", "error", ErrEncodingPayload.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)),
			),
		)
		WriteError(w, r, http.StatusInternalServerError, ErrEncodingPayload.Error())
		return
	}

	slog.Debug("handler.GetByID", "user", user)
	span.SetStatus(codes.Ok, "User found")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("code", fmt.Sprintf("%d", http.StatusOK)),
		),
	)
}

// CreateUser Create a new user
// @Summary Create a new user.
// @Description Create a new user from scratch.
// @Description If the id is not provided, it will be generated automatically.
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "CreateUserRequest"
// @Success 201 {object} string
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users"),
		attribute.String("function", "CreateUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.CreateUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.CreateUser", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := &model.UserParamsInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	slog.Debug("handler.CreateUser", "user", user)

	if err := h.service.CreateUser(ctx, user); err != nil {
		if errors.Is(err, service.ErrUserIDAlreadyExists) {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(service.ErrUserIDAlreadyExists)
			slog.Error("handler.CreateUser", "error", service.ErrUserIDAlreadyExists.Error())
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusConflict)))...,
				),
			)

			WriteError(w, r, http.StatusConflict, service.ErrUserIDAlreadyExists.Error())
			return
		}

		span.SetStatus(codes.Error, ErrInternalServerError.Error())
		span.RecordError(ErrInternalServerError)
		slog.Error("handler.CreateUser", "error", ErrInternalServerError.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	span.SetStatus(codes.Ok, "User created")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusCreated)))...,
		),
	)
}

// UpdateUser Update a user
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param uid path string true "The user ID in UUID format"
// @Param user body model.UpdateUserRequest true "User"
// @Success 200 {object} string
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users/{uid} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.UpdateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users"),
		attribute.String("function", "UpdateUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	uidParam := r.PathValue("uid")
	if uidParam == "" {
		span.SetStatus(codes.Error, ErrUserIDRequired.Error())
		span.RecordError(ErrUserIDRequired)
		slog.Error("handler.UpdateUser", "error", ErrUserIDRequired.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrUserIDRequired.Error())
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(uidParam)
	if err != nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("handler.UpdateUser", "error", ErrInvalidUserID.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	if id == uuid.Nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("handler.UpdateUser", "error", ErrInvalidUserID.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, ErrDecodingPayload.Error())
		span.RecordError(ErrDecodingPayload)
		slog.Error("handler.UpdateUser", "error", ErrDecodingPayload.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrDecodingPayload.Error())
		return
	}

	// at least one field must be updated
	if req.FirstName == "" && req.LastName == "" && req.Email == "" {
		span.SetStatus(codes.Error, ErrAtLeastOneFieldRequired.Error())
		span.RecordError(ErrAtLeastOneFieldRequired)
		slog.Error("handler.UpdateUser", "error", ErrAtLeastOneFieldRequired.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrAtLeastOneFieldRequired.Error())
		return
	}

	user := model.UserParamsInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if err := h.service.UpdateUser(ctx, &user); err != nil {
		span.SetStatus(codes.Error, ErrInternalServerError.Error())
		span.RecordError(ErrInternalServerError)
		slog.Error("handler.UpdateUser", "error", ErrInternalServerError.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	slog.Debug("handler.UpdateUser", "user", user)
	span.SetStatus(codes.Ok, "User updated")
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)
}

// DeleteUser Delete a user
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Param uid path string true "The user ID in UUID format"
// @Success 200 {object} string
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users/{uid} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.DeleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users"),
		attribute.String("function", "DeleteUser"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	uidParam := r.PathValue("uid")
	if uidParam == "" {
		span.SetStatus(codes.Error, ErrUserIDRequired.Error())
		span.RecordError(ErrUserIDRequired)
		slog.Error("handler.DeleteUser", "error", ErrUserIDRequired.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrUserIDRequired.Error())
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(uidParam)
	if err != nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("handler.DeleteUser", "error", ErrInvalidUserID.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	if id == uuid.Nil {
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		span.RecordError(ErrInvalidUserID)
		slog.Error("handler.DeleteUser", "error", ErrInvalidUserID.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	user := model.UserParamsInput{
		ID: id,
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if err := h.service.DeleteUser(ctx, &user); err != nil {
		span.SetStatus(codes.Error, ErrInternalServerError.Error())
		span.RecordError(ErrInternalServerError)
		slog.Error("handler.DeleteUser", "error", ErrInternalServerError.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	slog.Debug("handler.DeleteUser", "user", user)
	span.SetStatus(codes.Ok, "User deleted")
	span.SetAttributes(attribute.String("user.id", id.String()))
	h.metrics.handlerCalls.Add(ctx, 1,
		metric.WithAttributes(
			append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
		),
	)
}

// ListUsers Return a paginated list of users
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
// @Success 200 {object} model.ListUserResponse
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.ot.Traces.Tracer.Start(r.Context(), "handler.users.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("component", "handler.users"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", "handler.users"),
		attribute.String("function", "ListUsers"),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")]),
	}

	// paginator
	nextToken := r.URL.Query().Get("next_token")
	prevToken := r.URL.Query().Get("prev_token")
	limitString := r.URL.Query().Get("limit")

	// sort, filter, and fields
	sort := r.URL.Query().Get("sort")
	slog.Debug("handler.ListUsers", "sort", sort)

	if !query.IsValidSort(model.UserSortFields, sort) {
		span.SetStatus(codes.Error, ErrInvalidSort.Error())
		span.RecordError(ErrInvalidSort)
		slog.Error("handler.ListUsers", "error", ErrInvalidSort.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidSort.Error())
		return
	}

	filter := r.URL.Query().Get("filter")
	slog.Debug("handler.ListUsers", "filter", filter)

	if !query.IsValidFilter(model.UserFilterFields, filter) {
		span.SetStatus(codes.Error, ErrInvalidFilter.Error())
		span.RecordError(ErrInvalidFilter)
		slog.Error("handler.ListUsers", "error", ErrInvalidFilter.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidFilter.Error())
		return
	}

	fieldsFields := r.URL.Query().Get("fields")
	slog.Debug("handler.ListUsers", "fields", fieldsFields)

	if !query.IsValidFields(model.UserFields, fieldsFields) {
		span.SetStatus(codes.Error, ErrInvalidField.Error())
		span.RecordError(ErrInvalidField)
		slog.Error("handler.ListUsers", "error", ErrInvalidField.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
			),
		)

		WriteError(w, r, http.StatusBadRequest, ErrInvalidField.Error())
		return
	}

	// list of fields after sanitize the fields (basically trim the spaces)
	fields := query.GetFields(fieldsFields)

	// convert the limit to int
	limit := paginator.DefaultLimit
	var err error
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			span.SetStatus(codes.Error, "Invalid limit")
			span.RecordError(err)
			slog.Error("handler.ListUsers", "error", "Invalid limit")
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
				),
			)

			WriteError(w, r, http.StatusBadRequest, "Invalid limit")
			return
		}

		if limit < 0 {
			span.SetStatus(codes.Error, "Limit must be greater than or equal to 0")
			span.RecordError(err)
			slog.Error("handler.ListUsers", "error", "Limit must be greater than or equal to 0")
			h.metrics.handlerCalls.Add(ctx, 1,
				metric.WithAttributes(
					append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusBadRequest)))...,
				),
			)

			WriteError(w, r, http.StatusBadRequest, "Limit must be greater than or equal to 0")
			return
		}
		if limit == 0 {
			limit = paginator.DefaultLimit
		}
	}

	params := &model.ListUserRequest{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: paginator.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	usersResponse, err := h.service.ListUsers(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.ListUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// set the next and previous page
	if usersResponse.Paginator.NextToken != "" {
		usersResponse.Paginator.NextPage = r.URL.Path + "?next_token=" + usersResponse.Paginator.NextToken + "&limit=" + strconv.Itoa(limit)
	}
	if usersResponse.Paginator.PrevToken != "" {
		usersResponse.Paginator.PrevPage = r.URL.Path + "?prev_token=" + usersResponse.Paginator.PrevToken + "&limit=" + strconv.Itoa(limit)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(usersResponse); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("handler.ListUsers", "error", err.Error())
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusInternalServerError)))...,
			),
		)

		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	} else {
		slog.Debug("handler.users.ListUsers: called", "users", len(usersResponse.Items))
		span.SetStatus(codes.Ok, "list users")
		span.SetAttributes(attribute.Int("users.count", len(usersResponse.Items)))
		h.metrics.handlerCalls.Add(ctx, 1,
			metric.WithAttributes(
				append(metricCommonAttributes, attribute.String("code", fmt.Sprintf("%d", http.StatusOK)))...,
			),
		)
	}
}
