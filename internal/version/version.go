package version

import "runtime"

var (
	// Version is the current version of the application
	Version = "0.0.1"

	// BuildDate is the date the application was built
	BuildDate = "1970-01-01T00:00:00Z"

	// GitCommit is the commit hash the application was built from
	GitCommit = ""

	// GitBranch is the branch the application was built from
	GitBranch = ""

	// BuildUser is the user that built the application
	BuildUser = ""

	// GoVersion is the version of Go used to build the application
	GoVersion = runtime.Version()

	// GoVersionArch is the architecture of Go used to build the application
	GoVersionArch = runtime.GOARCH

	// GoVersionOS is the operating system of Go used to build the application
	GoVersionOS = runtime.GOOS
)

// VersionInfo represents the version information of the application.
type VersionInfo struct {
	Version       string `json:"version"`
	BuildDate     string `json:"buildDate"`
	GitCommit     string `json:"gitCommit"`
	GitBranch     string `json:"gitBranch"`
	GoVersion     string `json:"goVersion"`
	GoVersionArch string `json:"goVersionArch"`
	GoVersionOS   string `json:"goVersionOS"`
}
