package handler

type Version struct {
	Version       string `json:"version"`
	BuildDate     string `json:"build_date"`
	GitCommit     string `json:"git_commit"`
	GitBranch     string `json:"git_branch"`
	GoVersion     string `json:"go_version"`
	GoVersionArch string `json:"go_version_arch"`
	GoVersionOS   string `json:"go_version_os"`
}
