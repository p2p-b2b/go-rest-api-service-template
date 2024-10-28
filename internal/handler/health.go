package handler

import (
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
	mux.HandleFunc("/health", h.get)
	mux.HandleFunc("/healthz", h.get)
	mux.HandleFunc("/status", h.get)
}

// get returns the health of the service
// @Summary Get the health of the service
// @Description Get the health of the service
// @Tags service.health
// @Produce json
// @Success 200 {object} service.Health
// @Failure 500 {object} APIResponse
// @Router /health [get]
// @Router /healthz [get]
// @Router /status [get]
func (h *HealthHandler) get(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.UserHealthCheck(r.Context())
	if err != nil {
		WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	if err := WriteJSONData(w, http.StatusOK, health); err != nil {
		WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}
}
