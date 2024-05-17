package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SwaggerHandler handles the swagger UI
type SwaggerHandler struct {
	hf http.HandlerFunc
}

// NewSwaggerHandler creates a new SwaggerHandler
func NewSwaggerHandler(url string) *SwaggerHandler {
	return &SwaggerHandler{
		hf: httpSwagger.Handler(httpSwagger.URL(url)),
	}
}

// RegisterRoutes registers the routes for the handler
func (h *SwaggerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /swagger/", h.hf)
}
