package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
	"github.com/p2p-b2b/go-service-template/internal/paginator"
	"github.com/p2p-b2b/go-service-template/internal/service"
)

type UserHandlerConfig struct {
	Service service.UserService
}

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(conf *UserHandlerConfig) *UserHandler {
	return &UserHandler{
		service: conf.Service,
	}
}

// GetByID Get a user by ID
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Param query query string false "Query string"
// @Success 200 {object} model.User
// @Failure 500 {object} string
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idString := r.PathValue("id")
	if idString == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, ErrIDRequired.Error(), http.StatusBadRequest)
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// encode and write the response
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateUser Create a new user
// @Summary Create a new user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "CreateUserRequest"
// @Success 201 {object} model.CreateUserRequest
// @Failure 500 {object} string
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.FirstName == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "First name is required", http.StatusBadRequest)
		return
	}

	if user.LastName == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Last name is required", http.StatusBadRequest)
		return
	}

	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdateUser Update a user
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body model.User true "User"
// @Success 200
// @Failure 500 {object} string
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.FirstName == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "First name is required", http.StatusBadRequest)
		return
	}

	if user.LastName == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Last name is required", http.StatusBadRequest)
		return
	}

	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Age must be greater than or equal to 0", http.StatusBadRequest)
		return
	}

	u := model.UpdateUserInput(user)

	user.ID = id
	if err := h.service.Update(r.Context(), &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser Delete a user
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Param id path string true "User ID"
// @Success 200
// @Failure 500 {object} string
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	u := model.DeleteUserInput{
		ID: id,
	}

	if err := h.service.Delete(r.Context(), &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListUsers Return a paginated list of users
// @Summary List all users
// @Description List all users
// @Tags users
// @Produce json
// @Param sort query string false "Sort field"
// @Param filter query string false "Filter field"
// @Param fields query string false "Fields to return"
// @Param query query string false "Query string"
// @Param next_token query string false "Next cursor"
// @Param prev_token query string false "Previous cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} model.ListUserResponse
// @Failure 500 {object} string
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var req model.ListUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// if err is distinct from EOF
		if err != io.EOF {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			slog.Error("ListUsers", "error", err)
			return
		}
	}
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// paginator
	nextToken := r.URL.Query().Get("next_token")
	prevToken := r.URL.Query().Get("prev_token")
	limitString := r.URL.Query().Get("limit")

	sort := r.URL.Query().Get("sort")

	filterFields := r.URL.Query().Get("filter")
	var filter []string
	if len(filterFields) != 0 {
		filter = strings.Split(filterFields, ",")
	}

	fieldsFields := r.URL.Query().Get("fields")
	var fields []string
	if len(fieldsFields) != 0 {
		fields = strings.Split(fieldsFields, ",")
	}

	slog.Debug("ListUsers", "sort", sort, "filter", filter, "fields", fields, "next_token", nextToken, "prev_token", prevToken, "limit", limitString)

	// convert the limit to int
	limit := paginator.DefaultLimit
	var err error
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}

		if limit < 0 {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Limit must be greater than or equal to 0", http.StatusBadRequest)
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

	usersResponse, err := h.service.List(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// inject the parameters and server url to paginator for the next and previous links
	srvScheme := "http"
	if r.TLS != nil {
		r.URL.Scheme = "https"
	}
	serverURL := srvScheme + "://" + r.Host
	usersResponse.Paginator.Next = serverURL + r.URL.Path + "?next_token=" + usersResponse.Paginator.NextToken + "&limit=" + strconv.Itoa(limit)
	usersResponse.Paginator.Prev = serverURL + r.URL.Path + "?prev_token=" + usersResponse.Paginator.PrevToken + "&limit=" + strconv.Itoa(limit)

	// write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(usersResponse); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
