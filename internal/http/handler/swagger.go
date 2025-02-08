package handler

import (
	"net/http"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
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
func (ref *SwaggerHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /swagger/", mdw.ThenFunc(ref.hf))
}
