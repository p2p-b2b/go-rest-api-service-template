package handler

import (
	"fmt"
	"net/http"

	"github.com/p2p-b2b/go-service-template/internal/version"
)

type VersionHandler struct{}

func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

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
