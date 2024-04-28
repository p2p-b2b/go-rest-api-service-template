package handler

import (
	"fmt"
	"net/http"

	"github.com/p2p-b2b/go-service-template/internal/version"
)

func GetVersion(w http.ResponseWriter, r *http.Request) {
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
