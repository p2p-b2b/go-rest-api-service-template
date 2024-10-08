package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler represents the handler for the health of the service.
type HealthHandler struct {
	service UserService
}

// NewHealthHandler returns a new instance of HealthHandler.
func NewHealthHandler(us UserService) *HealthHandler {
	return &HealthHandler{
		service: us,
	}
}

// RegisterRoutes registers the routes for the handler.
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Get)
	mux.HandleFunc("/healthz", h.Get)
	mux.HandleFunc("/status", h.Get)
}

// Get returns the health of the service
// @Summary Get the health of the service
// @Description Get the health of the service
// @Tags service.health
// @Produce json
// @Success 200 {object} service.Health
// @Failure 500 {object} APIError
// @Router /health [get]
// @Router /healthz [get]
// @Router /status [get]
func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.UserHealthCheck(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	// write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}
}
