package handler

import (
	"encoding/json"
	"net/http"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

// VersionHandler represents the handler for the version of the service.
type VersionHandler struct{}

// NewVersionHandler returns a new instance of VersionHandler.
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// Get returns the version of the service
// @Summary Get the version of the service
// @Description Get the version of the service
// @Tags version
// @Produce json
// @Success 200 {object} version.VersionInfo
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

	if err := json.NewEncoder(w).Encode(v); err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}
}
