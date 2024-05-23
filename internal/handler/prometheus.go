package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler represents the handler for the health of the service.
type PrometheusHandler struct{}

// PrometheusHandler returns a new instance of Prometheus.
func NewPrometheusHandler() *PrometheusHandler {
	return &PrometheusHandler{}
}

// RegisterRoutes registers the routes for the handler.
func (h *PrometheusHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /metrics", promhttp.Handler())

}
