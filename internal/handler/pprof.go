package handler

import (
	"net/http"
	"net/http/pprof"
)

// PprofHandler handles the pprof routes
type PprofHandler struct{}

// NewPprofHandler creates a new PprofHandler
func NewPprofHandler() *PprofHandler {
	return &PprofHandler{}
}

// RegisterRoutes registers the routes for the handler
func (h *PprofHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("GET /debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	mux.HandleFunc("GET /debug/pprof/block", pprof.Handler("block").ServeHTTP)
	mux.HandleFunc("GET /debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	mux.HandleFunc("GET /debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("GET /debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	mux.HandleFunc("GET /debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}
