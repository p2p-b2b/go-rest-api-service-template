package handler

import (
	"encoding/json"
	"net/http"

	"github.com/p2p-b2b/go-service-template/internal/service"
)

type HealthUserHandlerConfig struct {
	Service service.UserService
}

type HealthHandler struct {
	service service.UserService
}

func NewHealthHandler(conf *HealthUserHandlerConfig) *HealthHandler {
	return &HealthHandler{
		service: conf.Service,
	}
}

// Get returns the health of the service
// @Summary Get the health of the service
// @Description Get the health of the service
// @Tags health
// @Produce json
// @Success 200 {object} model.Health
// @Router /health [get]
// @Router /healthz [get]
// @Router /status [get]
func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health, err := h.service.HealthCheck(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// write the response
	if err := json.NewEncoder(w).Encode(health); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
