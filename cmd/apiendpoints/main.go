package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

var (
	appName = "apiendpoints"

	showVersion     bool
	showLongVersion bool
	showHelp        bool
	swaggerJSONFile = config.Field[string]{Value: "./docs/swagger.json"}
)

func init() {
	// Version, Help and debug flags
	flag.BoolVar(&showVersion, "version", false, "Show the version information")
	flag.BoolVar(&showLongVersion, "version.long", false, "Show the long version information")
	flag.BoolVar(&showHelp, "help", false, "Show this help message")

	// Swagger JSON file
	flag.StringVar(&swaggerJSONFile.Value, "swagger.file", swaggerJSONFile.Value, "Path to the swagger.json file")

	// Parse the command line arguments
	flag.Parse()

	flag.Usage = func() {
		_, err := fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\nOptions:\n", appName)
		if err != nil {
			slog.Error("failed to print usage", "error", err)
			os.Exit(1)
		}

		flag.PrintDefaults()
	}

	// implement the version flag
	if showVersion {
		_, err := fmt.Printf("%s version: %s\n", appName, version.Version)
		if err != nil {
			slog.Error("failed to print version", "error", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	// implement the long version flag
	if showLongVersion {
		_, err := fmt.Printf("%s version: %s,  Git Commit: %s, Build Date: %s, Go Version: %s, OS/Arch: %s/%s\n",
			appName,
			version.Version,
			version.GitCommit,
			version.BuildDate,
			version.GoVersion,
			version.GoVersionOS,
			version.GoVersionArch,
		)
		if err != nil {
			slog.Error("failed to print long version", "error", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	// implement the help flag
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	// Open the swagger.json file
	file, err := os.Open(swaggerJSONFile.Value)
	if err != nil {
		slog.Error("failed to open swagger.json file", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	// decode the swagger.json file
	var sw swagger
	if err := json.NewDecoder(file).Decode(&sw); err != nil {
		slog.Error("failed to decode swagger.json file", "error", err)
		os.Exit(1)
	}

	excludes := map[string]string{
		"/ui":            "GET",
		"/users/health":  "GET",
		"/version":       "GET",
		"/health/status": "GET",
	}

	var records []Record
	idMaxWidth := len("ID")
	summaryMaxWidth := len("Summary")
	descriptionMaxWidth := len("Description")
	methodMaxWidth := len("Method")
	pathMaxWidth := len("Path")
	systemMaxWidth := len("System")

	for path, methods := range sw.Paths {
		for method, data := range methods {
			// skip the excludes paths
			if val, ok := excludes[path]; ok {
				if val == strings.ToUpper(method) {
					continue
				}
			}

			var id uuid.UUID
			if data.OperationID == "" {
				id, err = uuid.NewV7()
				if err != nil {
					panic(err)
				}
			} else {
				id, err = uuid.Parse(data.OperationID)
				if err != nil {
					panic(err)
				}
			}

			if len(id.String()) > idMaxWidth {
				idMaxWidth = len(id.String())
			}

			if len(data.Summary) > summaryMaxWidth {
				summaryMaxWidth = len(data.Summary)
			}

			if len(data.Description) > descriptionMaxWidth {
				descriptionMaxWidth = len(data.Description)
			}

			if len(method) > methodMaxWidth {
				methodMaxWidth = len(method)
			}

			if len(path) > pathMaxWidth {
				pathMaxWidth = len(path)
			}

			if len("TRUE") > systemMaxWidth {
				systemMaxWidth = len("TRUE")
			}

			record := Record{
				ID:          "'" + id.String() + "'",
				Summary:     "'" + data.Summary + "'",
				Description: "'" + data.Description + "'",
				Method:      "'" + strings.ToUpper(method) + "'",
				Path:        "'" + path + "'",
				System:      "TRUE",
			}

			records = append(records, record)
		}
	}

	// sort the records by path
	sort.Slice(records, func(i, j int) bool {
		return records[i].Path < records[j].Path
	})

	idMaxWidth += 2
	summaryMaxWidth += 2
	descriptionMaxWidth += 2 // Add padding for description
	methodMaxWidth += 2
	pathMaxWidth += 2
	systemMaxWidth += 2 // Consistent padding

	for i, record := range records {
		if i == len(records)-1 {
			_, err := fmt.Printf("(%-*s, %-*s, %-*s, %-*s, %-*s, %-*s);\n",
				idMaxWidth, record.ID,
				summaryMaxWidth, record.Summary,
				descriptionMaxWidth, record.Description,
				methodMaxWidth, record.Method,
				pathMaxWidth, record.Path,
				systemMaxWidth, record.System,
			)
			if err != nil {
				slog.Error("failed to print record", "error", err)
				os.Exit(1)
			}
		} else {
			_, err := fmt.Printf("(%-*s, %-*s, %-*s, %-*s, %-*s, %-*s),\n",
				idMaxWidth, record.ID,
				summaryMaxWidth, record.Summary,
				descriptionMaxWidth, record.Description,
				methodMaxWidth, record.Method,
				pathMaxWidth, record.Path,
				systemMaxWidth, record.System,
			)
			if err != nil {
				slog.Error("failed to print record", "error", err)
				os.Exit(1)
			}
		}
	}
}

type Record struct {
	ID          string
	Summary     string
	Description string
	Method      string
	Path        string
	System      string
}

// swagger is the struct that represents the swagger.json file
type swagger struct {
	Swagger string `json:"swagger"`
	Info    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"info"`
	Paths map[string]map[string]struct {
		Description string   `json:"description"`
		Consumes    []string `json:"consumes"`
		Produces    []string `json:"produces"`
		Tags        []string `json:"tags"`
		Summary     string   `json:"summary"`
		OperationID string   `json:"operationId"`
		Responses   map[string]struct {
			Description string `json:"description"`
			Schema      struct {
				Type string `json:"type"`
			} `json:"schema"`
		} `json:"responses"`
	} `json:"paths"`
}
