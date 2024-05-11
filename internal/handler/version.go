package handler

import (
	"fmt"
	"net/http"

	"github.com/p2p-b2b/go-service-template/internal/version"
)

// VersionHandler represents the handler for the version of the service.
type VersionHandler struct{}

// NewVersionHandler returns a new instance of VersionHandler.
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// Get returns the version of the service
func (h *VersionHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, `{"version":"%s","buildDate":"%s","gitCommit":"%s","gitBranch":"%s","goVersion":"%s","goVersionArch":"%s","goVersionOS":"%s"}`,
		version.Version,
		version.BuildDate,
		version.GitCommit,
		version.GitBranch,
		version.GoVersion,
		version.GoVersionArch,
		version.GoVersionOS,
	)
	w.WriteHeader(http.StatusOK)
}
