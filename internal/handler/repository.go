package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-service-template/internal/model"
	"github.com/p2p-b2b/go-service-template/internal/repository"
)

type RepositoryHandler struct {
	Repository repository.UserRepository
}

// GetUserByID handles the HTTP GET - /{id} endpoint.
func (h *RepositoryHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
	user, err := h.Repository.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// write the response
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// CreateUser handles the HTTP POST - / endpoint.
func (h *RepositoryHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
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

	if err := h.Repository.Create(r.Context(), &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdateUser handles the HTTP PUT - /{id} endpoint.
func (h *RepositoryHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
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

	user.ID = id
	if err := h.Repository.Update(r.Context(), &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser handles the HTTP DELETE - /{id} endpoint.
func (h *RepositoryHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
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

	if err := h.Repository.Delete(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListUsers handles the HTTP GET - / endpoint.
func (h *RepositoryHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.Repository.List(r.Context())
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
