package handler

import (
	"encoding/json"
	"net/http"

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
// @Tags user
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.User
// @Router /user/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	// write the response
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
// @Tags user
// @Accept json
// @Produce json
// @Param user body model.User true "User"
// @Success 201 {object} model.User
// @Router /user [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
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

	if user.Age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Age must be greater than or equal to 0", http.StatusBadRequest)
		return
	}
	u := model.CreateUserInput(user)

	if err := h.service.Create(r.Context(), &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdateUser Update a user
// @Summary Update a user
// @Description Update a user
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body model.User true "User"
// @Success 200
// @Router /user/{id} [put]
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

	if user.Age < 0 {
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
// @Tags user
// @Param id path string true "User ID"
// @Success 200
// @Router /user/{id} [delete]
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

// ListUsers handles the HTTP GET - / endpoint.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.service.List(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// write the response
	if err := json.NewEncoder(w).Encode(users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
