package handler

import (
	"net/http"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

// VersionHandler represents the handler for the version of the service.
type VersionHandler struct{}

// NewVersionHandler returns a new instance of VersionHandler.
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// RegisterRoutes registers the routes for the version of the service.
func (h *VersionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /version", h.Get)
}

// Get returns the version of the service
// @Summary Get the version of the service
// @Description Get the version of the service
// @Tags version
// @Produce json
// @Success 200 {object} version.VersionInfo
// @Failure 500 {object} APIResponse
// @Router /version [get]
func (h *VersionHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	v := version.VersionInfo{
		Version:       version.Version,
		BuildDate:     version.BuildDate,
		GitCommit:     version.GitCommit,
		GitBranch:     version.GitBranch,
		GoVersion:     version.GoVersion,
		GoVersionArch: version.GoVersionArch,
		GoVersionOS:   version.GoVersionOS,
	}

	if err := WriteJSONData(w, http.StatusOK, v); err != nil {
		WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}
}
