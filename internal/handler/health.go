package handler

import (
	"encoding/json"
	"net/http"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
)

// HealthUserHandlerConfig represents the configuration used to create a new HealthHandler.
type HealthUserHandlerConfig struct {
	UserService service.UserService
}

// HealthHandler represents the handler for the health of the service.
type HealthHandler struct {
	userService service.UserService
}

// NewHealthHandler returns a new instance of HealthHandler.
func NewHealthHandler(conf *HealthUserHandlerConfig) *HealthHandler {
	return &HealthHandler{
		userService: conf.UserService,
	}
}

// Get returns the health of the service
// @Summary Get the health of the service
// @Description Get the health of the service
// @Tags health
// @Produce json
// @Success 200 {object} model.Health
// @Failure 500 {object} APIError
// @Router /health [get]
// @Router /healthz [get]
// @Router /status [get]
func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	health, err := h.userService.UserHealthCheck(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	// write the response
	if err := json.NewEncoder(w).Encode(health); err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
