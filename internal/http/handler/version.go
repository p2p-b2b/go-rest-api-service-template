package handler

import (
	"net/http"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

// VersionHandler represents the handler for the version of the service.
type VersionHandler struct{}

// NewVersionHandler returns a new instance of VersionHandler.
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// RegisterRoutes registers the routes for the version of the service.
func (ref *VersionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /version", ref.get)
}

// get returns the version of the service
//
// @Summary Get the version of the service
// @Description Get the version of the service
// @Tags Version
// @Produce json
// @Success 200 {object} Version
// @Failure 500 {object} respond.HTTPMessage
// @Router /version [get]
func (ref *VersionHandler) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	v := Version{
		Version:       version.Version,
		BuildDate:     version.BuildDate,
		GitCommit:     version.GitCommit,
		GitBranch:     version.GitBranch,
		GoVersion:     version.GoVersion,
		GoVersionArch: version.GoVersionArch,
		GoVersionOS:   version.GoVersionOS,
	}

	if err := respond.WriteJSONData(w, http.StatusOK, v); err != nil {
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}
}
