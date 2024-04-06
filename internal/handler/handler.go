package handler

import (
	"fmt"
	"net/http"

	"github.com/wereweare/go-service-template/internal/version"
)

func GetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"version":"%s","buildDate":"%s","gitCommit":"%s","gitBranch":"%s","goVersion":"%s"}`, version.Version, version.BuildDate, version.GitCommit, version.GitBranch, version.GoVersion)
}
