package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
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
// @Param user body model.CreateUserInput true "CreateUserInput"
// @Success 201 {object} model.CreateUserInput
// @Failure 500 {object} string
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user model.CreateUserInput
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

// ListUsers List all users
// @Summary List all users
// @Description List all users
// @Tags users
// @Produce json
// @Success 200 {object} model.ListUserOutput
// @Failure 500 {object} string
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	next := r.URL.Query().Get("next")
	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	filter := r.URL.Query().Get("filter")
	fields := r.URL.Query().Get("fields")

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}

	params := &model.ListUserInput{
		Next:   next,
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
		Order:  order,
		Filter: filter,
		Fields: fields,
	}

	users, err := h.service.List(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// write the response
	if err := json.NewEncoder(w).Encode(users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
