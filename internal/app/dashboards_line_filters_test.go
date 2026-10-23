//go:build unit

package app

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	otellog "go.opentelemetry.io/otel/log"

	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

var (
	// A line filter that keeps the lines containing a text: `|= "text"`, or
	// `|~ "one|another"`. The negations drop lines and name nothing.
	dashboardLineFilter = regexp.MustCompile("\\|([=~])\\s*[`\"]([^`\"]*)[`\"]")
	// A filter on a record's severity: `| severity_text = "ERROR"`,
	// `| severity_number >= 17`.
	dashboardSeverityFilter = regexp.MustCompile("\\|\\s*(severity_text|severity_number)\\s*(=~|!=|!~|>=|<=|=|>|<)\\s*(?:[`\"]([^`\"]*)[`\"]|([0-9]+))")
)

// walkDashboardStrings calls visit for every string value of a decoded JSON
// document: a panel's query, a data link's, an annotation's.
func walkDashboardStrings(v any, visit func(string)) {
	switch x := v.(type) {
	case string:
		visit(x)
	case []any:
		for _, item := range x {
			walkDashboardStrings(item, visit)
		}
	case map[string]any:
		for _, item := range x {
			walkDashboardStrings(item, visit)
		}
	}
}

// exportedSeverities are the severities a record leaves the process at, as
// the slog bridge writes them, text and number: every level the service
// logs at from o11y.ExportedLogFloor up. TRACE stays in the process.
func exportedSeverities() map[string]int {
	offset := int(otellog.SeverityDebug) - int(slog.LevelDebug)
	out := map[string]int{}

	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		if level >= o11y.ExportedLogFloor {
			out[level.String()] = int(level) + offset
		}
	}

	return out
}

// serviceSource is the module's Go source, tests left out: what the service
// can write.
func serviceSource(t *testing.T) string {
	t.Helper()

	var source strings.Builder

	for _, dir := range []string{"cmd", "internal", "pkg"} {
		root := filepath.Join("..", "..", dir)
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}

			b, err := os.ReadFile(path)
			source.Write(b)

			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	return source.String()
}

// TestDashboardFiltersNameWhatTheServiceWrites: a query that keeps the lines
// containing a text names a message the service writes, and a filter on a
// severity names one a record is exported at.
//
// Every dashboard carried a "Fatal" annotation on severity_text FATAL, a level
// the service has no way to write: it marked nothing, and a replica that
// stopped before serving looked like one that never ran. The annotation reads
// the line App.stopOnError writes now. Every dashboard carries both the
// restart and the stop annotation.
func TestDashboardFiltersNameWhatTheServiceWrites(t *testing.T) {
	t.Parallel()

	source := serviceSource(t)
	exported := exportedSeverities()

	var texts, severities int

	for _, file := range dashboardFiles(t) {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		var raw any
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatalf("%s: %v", file, err)
		}

		name := filepath.Base(file)

		for _, line := range []string{"application initialization completed", stoppedOnAnError} {
			if !strings.Contains(string(body), "|= `"+line+"`") {
				t.Errorf("%s has no annotation on %q", name, line)
			}
		}

		walkDashboardStrings(raw, func(text string) {
			if !strings.Contains(text, "{service_name=~") {
				return
			}

			query := strings.NewReplacer(`\"`, `"`, "%60", "`").Replace(text)

			for _, m := range dashboardLineFilter.FindAllStringSubmatch(query, -1) {
				alternatives := []string{m[2]}
				if m[1] == "~" {
					alternatives = strings.Split(m[2], "|")
				}

				for _, message := range alternatives {
					texts++

					if !strings.Contains(source, `"`+message) {
						t.Errorf("%s keeps the lines containing %q, and nothing writes one: %s", name, message, text)
					}
				}
			}

			for _, m := range dashboardSeverityFilter.FindAllStringSubmatch(query, -1) {
				severities++

				key, operator, value, number := m[1], m[2], m[3], m[4]

				if key == "severity_text" {
					if _, written := exported[value]; (operator == "=" || operator == "=~") && !written {
						t.Errorf("%s filters on severity_text %q, which no record is exported at: %s", name, value, text)
					}

					continue
				}

				n, err := strconv.Atoi(number)
				if err != nil {
					t.Errorf("%s compares severity_number with %q, not a number: %s", name, number, text)

					continue
				}

				if !anySeverity(exported, operator, n) {
					t.Errorf("%s filters on severity_number %s %d, which no exported record has: %s", name, operator, n, text)
				}
			}
		})
	}

	// A restart and a stop on each of five dashboards, and the severity
	// filters of the error panels. Fewer means the patterns no longer read
	// the dashboards.
	if texts < 10 || severities < 5 {
		t.Errorf("read %d line filter texts and %d severity filters, want at least 10 and 5", texts, severities)
	}
}

// anySeverity reports whether some exported severity number v satisfies
// `v operator n`.
func anySeverity(exported map[string]int, operator string, n int) bool {
	for _, v := range exported {
		switch {
		case operator == "=" && v == n, operator == "!=" && v != n,
			operator == ">" && v > n, operator == ">=" && v >= n,
			operator == "<" && v < n, operator == "<=" && v <= n:
			return true
		}
	}

	return false
}
