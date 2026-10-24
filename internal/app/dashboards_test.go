//go:build unit

package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// The Grafana dashboards are provisioned from dev-env/ and nothing validates
// them before Grafana loads them.
//
// # Why a test and not a linter
//
// grafana/dashboard-linter is the obvious tool and neither of its two
// distributions can be used here: it publishes no container image, and
// `go install` refuses it because its go.mod carries replace directives. A
// Makefile target that cannot run is worse than none.
//
// A test is a better home anyway. `make check-alerts` has to be remembered;
// this runs in `make test` and in CI, which is where a broken dashboard should
// be caught -- the alternative is finding out from an empty panel, and an empty
// panel looks exactly like a service that is not being used.
//
// # What it checks
//
// Only failures that are silent in Grafana. A panel pointing at a datasource
// uid that does not exist renders "Datasource not found" in one panel while
// every other panel on the dashboard works, which is easy to miss on a
// dashboard with twenty of them.

// dashboardDir is where the provisioned dashboards live.
func dashboardDir(t *testing.T) string {
	t.Helper()

	return filepath.Join(repoRootFrom(t), "dev-env", "configuration", "grafana", "dashboard")
}

// datasourceUIDs reads the uids the provisioning file actually defines.
func datasourceUIDs(t *testing.T) map[string]bool {
	t.Helper()

	path := filepath.Join(repoRootFrom(t), "dev-env", "configuration", "grafana", "datasource", "grafana-ds.yaml")

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	uids := map[string]bool{}

	for line := range strings.SplitSeq(string(body), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "uid:"); ok {
			uids[strings.TrimSpace(after)] = true
		}
	}

	if len(uids) == 0 {
		t.Fatal("parsed no datasource uids out of grafana-ds.yaml; the parser has lost track of the format")
	}

	return uids
}

type dashboardPanel struct {
	Datasource *struct {
		UID string `json:"uid"`
	} `json:"datasource"`
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Expr        string           `json:"expr"`
	Query       string           `json:"query"`
	Targets     []dashboardPanel `json:"targets"`
	Panels      []dashboardPanel `json:"panels"`
	ID          int              `json:"id"`
}

type dashboard struct {
	UID    string           `json:"uid"`
	Title  string           `json:"title"`
	Panels []dashboardPanel `json:"panels"`
}

func dashboardFiles(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dashboardDir(t), "*.json"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no dashboards found; this test is looking in the wrong place")
	}

	return files
}

// flatten yields a panel and every panel nested inside a collapsed row.
func flatten(panels []dashboardPanel) []dashboardPanel {
	out := make([]dashboardPanel, 0, len(panels))

	for _, p := range panels {
		out = append(out, p)
		out = append(out, flatten(p.Panels)...)
	}

	return out
}

// A panel whose datasource does not exist renders "Datasource not found" in
// that panel alone, while the rest of the dashboard works.
func TestDashboardsOnlyReferenceDeclaredDatasources(t *testing.T) {
	t.Parallel()

	known := datasourceUIDs(t)

	for _, file := range dashboardFiles(t) {
		t.Run(strings.TrimSuffix(filepath.Base(file), ".json"), func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			var d dashboard
			if err := json.Unmarshal(body, &d); err != nil {
				t.Fatalf("a provisioned dashboard must be valid JSON: %v", err)
			}

			for _, p := range flatten(d.Panels) {
				for _, ds := range append([]dashboardPanel{p}, p.Targets...) {
					if ds.Datasource == nil {
						continue
					}

					uid := ds.Datasource.UID

					// A dashboard variable as the datasource is resolved at
					// render time and cannot be checked from here.
					if uid == "" || strings.HasPrefix(uid, "$") {
						continue
					}

					if !known[uid] {
						t.Errorf("panel %q (id %d) points at datasource uid %q, which grafana-ds.yaml does not define", p.Title, p.ID, uid)
					}
				}
			}
		})
	}
}

// Two panels sharing an id makes Grafana drop one of them on save, and the
// loss is silent.
func TestDashboardPanelIDsAreUnique(t *testing.T) {
	t.Parallel()

	for _, file := range dashboardFiles(t) {
		t.Run(strings.TrimSuffix(filepath.Base(file), ".json"), func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			var d dashboard
			if err := json.Unmarshal(body, &d); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			seen := map[int]string{}

			for _, p := range flatten(d.Panels) {
				if p.ID == 0 {
					continue
				}

				if first, clash := seen[p.ID]; clash {
					t.Errorf("panel id %d is used by both %q and %q", p.ID, first, p.Title)
				}

				seen[p.ID] = p.Title
			}
		})
	}
}

// Every dashboard needs a stable uid: provisioning keys on it, so a dashboard
// that loses or changes one is re-created as a duplicate rather than updated,
// and every link and bookmark to the old one breaks.
func TestDashboardsHaveAStableUID(t *testing.T) {
	t.Parallel()

	seen := map[string]string{}

	for _, file := range dashboardFiles(t) {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		var d dashboard
		if err := json.Unmarshal(body, &d); err != nil {
			t.Fatalf("%s: invalid JSON: %v", filepath.Base(file), err)
		}

		if d.UID == "" {
			t.Errorf("%s has no uid", filepath.Base(file))

			continue
		}

		if first, clash := seen[d.UID]; clash {
			t.Errorf("uid %q is used by both %s and %s", d.UID, first, filepath.Base(file))
		}

		seen[d.UID] = filepath.Base(file)
	}
}

// Loki is in the pod so that logs are visible from a dashboard. A dashboard set
// that quietly loses its log panels leaves the third signal invisible, which is
// the state this whole change exists to end.
//
// It asserts the PROPERTY -- some dashboard queries Loki, and the one named for
// logs is among them -- rather than naming a file. The previous version named
// go-rest-api-service-template-observability.json and broke the moment that dashboard was
// renamed, which is a test failing for the wrong reason.
func TestSomeDashboardQueriesLoki(t *testing.T) {
	t.Parallel()

	var queryingLoki []string

	for _, file := range dashboardFiles(t) {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		var d dashboard
		if err := json.Unmarshal(body, &d); err != nil {
			t.Fatalf("%s: invalid JSON: %v", filepath.Base(file), err)
		}

		for _, p := range flatten(d.Panels) {
			if p.Datasource != nil && p.Datasource.UID == "loki" {
				queryingLoki = append(queryingLoki, filepath.Base(file))

				break
			}
		}
	}

	if len(queryingLoki) == 0 {
		t.Fatal("no dashboard queries Loki; logs are the third signal and a dashboard is where they are seen")
	}

	// The dashboard named for logs must be one of them, or "Logs" is a title
	// over a page of metrics.
	var hasLogsDashboard bool

	for _, name := range queryingLoki {
		if strings.Contains(name, "logs") {
			hasLogsDashboard = true
		}
	}

	if !hasLogsDashboard {
		t.Errorf("the logs dashboard does not query Loki; dashboards that do: %v", queryingLoki)
	}
}

// Every dashboard scopes its queries to the selected service.
//
// One Grafana can serve this service and the template it came from, or two
// replicas of one deployment. A panel that omits the selector sums them
// together and says nothing about having done so -- and a dashboard where some
// panels are scoped and some are not is worse than one where none are, because
// the two disagree and neither is marked.
func TestDashboardQueriesAreScopedToTheSelectedService(t *testing.T) {
	t.Parallel()

	for _, file := range dashboardFiles(t) {
		t.Run(strings.TrimSuffix(filepath.Base(file), ".json"), func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			var d dashboard
			if err := json.Unmarshal(body, &d); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			for _, p := range flatten(d.Panels) {
				for _, target := range p.Targets {
					expr := target.Expr
					if expr == "" || !strings.Contains(expr, "_") {
						continue
					}

					// target_info is the build-info table: one series per
					// instance is the point of it.
					if strings.Contains(expr, "target_info") {
						continue
					}

					if !strings.Contains(expr, "service_name") {
						t.Errorf("panel %q has a query that is not scoped to $service_name: %s", p.Title, expr)
					}
				}
			}
		})
	}
}

// TestErrorPanelsCountFailuresNotRefusals: the layer instruments carry
// app.outcome (ok, refused, failed) so that an error ratio can leave out the
// requests the service refused on purpose. A panel that goes back to
// `successful="false"` counts a 404 and a dead database as the same thing
// again, and nothing but an operator misreading it would say so.
func TestErrorPanelsCountFailuresNotRefusals(t *testing.T) {
	t.Parallel()

	var onOutcome int

	for _, file := range dashboardFiles(t) {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		var d dashboard
		if err := json.Unmarshal(body, &d); err != nil {
			t.Fatalf("%s: %v", file, err)
		}

		for _, p := range flatten(d.Panels) {
			for _, target := range p.Targets {
				if !strings.Contains(target.Expr, "app_calls_total") && !strings.Contains(target.Expr, "app_call_duration_seconds") {
					continue
				}

				if strings.Contains(target.Expr, "successful") {
					t.Errorf("%s, panel %q (id %d) reads `successful` on a layer instrument; it is app_outcome now: %s",
						filepath.Base(file), p.Title, p.ID, target.Expr)
				}

				if strings.Contains(target.Expr, "app_outcome") {
					onOutcome++

					// An error ratio is the failed calls, and a breakdown is
					// by the label: nothing selects "refused" as an error.
					if strings.Contains(target.Expr, `app_outcome="refused"`) && strings.Contains(strings.ToLower(p.Title), "error") {
						t.Errorf("%s, panel %q counts refusals as errors: %s", filepath.Base(file), p.Title, target.Expr)
					}
				}
			}
		}
	}

	// Six expressions read it today: four ratios of failed calls and two
	// breakdowns. Fewer means a panel was moved off it.
	if onOutcome < 6 {
		t.Errorf("%d dashboard expressions read app_outcome, want at least 6", onOutcome)
	}
}

var (
	// A filter on a structured-metadata key of a log line: `| key = ...`.
	// A line filter (`|= "text"`, `|~ "regex"`) has no key and is not one.
	logKeyFilter = regexp.MustCompile("\\|\\s*([a-z_][a-z0-9_]*)\\s*(=~|!=|!~|>=|<=|=|>|<)\\s*(?:[`\"]([^`\"]*)[`\"])?")
	// app.layer and app.outcome on a span, whatever scope the query gives
	// them (span., resource. or none).
	traceLayerFilter   = regexp.MustCompile(`\.app\.layer\s*(=~|!=|!~|=)\s*"([^"]*)"`)
	traceOutcomeFilter = regexp.MustCompile(`\.app\.outcome\s*(=~|!=|!~|=)\s*"([^"]*)"`)
)

// logKeys are the attributes of a log line a dashboard may filter on: the
// ones the service, or the log pipeline, puts there. A filter on any other
// key matches nothing, and says nothing about having matched nothing.
var logKeys = map[string]string{
	"app_layer":       "o11y.AttrLayer, on a line written inside an operation",
	"app_domain":      "o11y.AttrDomain",
	"app_action":      "o11y.AttrAction",
	"component":       "the cache client's logger (initServices)",
	"error_type":      "the access line of a request answered 4xx or 5xx; operation_failed",
	"request_id":      "the access line",
	"route":           "the access line",
	"status":          "the access line",
	"severity_number": "the OTLP record",
	"severity_text":   "the OTLP record",
	"trace_id":        "the OTLP record",
}

// TestLogAndTracePanelsFilterOnWhatTheServiceWrites: a query that filters on
// an attribute names one the service puts on a log line or a span, with a
// value it can have there.
//
// A failed request writes one access line at ERROR, carrying the layer its
// error arose in, and a span carries app.outcome. A panel that filters on
// anything else is empty, and an empty error panel reads as a healthy
// service. It shipped twice: "Errors from the cache layer" and "Errors from
// the mail layer" filtered the log on app_layer cache and mail, and matched
// no record from the day they were added -- those two layers record the
// shared metric and open no operation, so no line and no span carries their
// name.
//
// Every string of a dashboard is read, not only a panel's targets: a data
// link and an annotation carry queries too.
//
// What it cannot see: whether a line with those attributes is ever written at
// the level the query asks for, and what a line filter (`|~ "text"`) matches.
// Those were checked against the running stack, and are why the cache and
// mail panels say in their descriptions which lines they read.
func TestLogAndTracePanelsFilterOnWhatTheServiceWrites(t *testing.T) {
	t.Parallel()

	// The layers that open an operation, so that a line or a span written
	// inside it carries the layer. $layer is the Layers dashboard's variable.
	layers := []string{"$layer", o11y.LayerHandler, o11y.LayerUsecase, o11y.LayerRepository}
	outcomes := []string{o11y.OutcomeOK, o11y.OutcomeRefused, o11y.OutcomeFailed}

	var logFilters, traceFilters int

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

		walkDashboardStrings(raw, func(text string) {
			// A LogQL query, wherever it is: a target, a data link (where
			// it is JSON inside a URL), an annotation.
			if strings.Contains(text, "{service_name=~") {
				query := strings.NewReplacer(`\"`, `"`, "%60", "`").Replace(text)

				for _, m := range logKeyFilter.FindAllStringSubmatch(query, -1) {
					key, operator, value := m[1], m[2], m[3]

					if _, known := logKeys[key]; !known {
						t.Errorf("%s filters the log on %q, which no line carries: %s", name, key, text)

						continue
					}

					// Only an equality names a value a line must have.
					if (operator != "=" && operator != "=~") || value == "" {
						continue
					}

					switch key {
					case "app_layer":
						logFilters++

						for layer := range strings.SplitSeq(value, "|") {
							if !slices.Contains(layers, layer) {
								t.Errorf("%s filters the log on app_layer %q, and no line carries that layer: %s", name, layer, text)
							}
						}
					case "component":
						if value != "cache" {
							t.Errorf("%s filters the log on component %q; the cache client's is the only one: %s", name, value, text)
						}
					}
				}
			}

			// A TraceQL query.
			if strings.Contains(text, ".app.layer") || strings.Contains(text, ".app.outcome") {
				for _, m := range traceLayerFilter.FindAllStringSubmatch(text, -1) {
					traceFilters++

					for layer := range strings.SplitSeq(m[2], "|") {
						if m[1] != "!=" && m[1] != "!~" && !slices.Contains(layers, layer) {
							t.Errorf("%s filters spans on app.layer %q, and no span carries that layer: %s", name, layer, text)
						}
					}
				}

				for _, m := range traceOutcomeFilter.FindAllStringSubmatch(text, -1) {
					traceFilters++

					for outcome := range strings.SplitSeq(m[2], "|") {
						if !slices.Contains(outcomes, outcome) {
							t.Errorf("%s filters spans on app.outcome %q, which is not an outcome: %s", name, outcome, text)
						}
					}
				}
			}
		})

		var d dashboard
		if err := json.Unmarshal(body, &d); err != nil {
			t.Fatalf("%s: %v", file, err)
		}

		for _, p := range flatten(d.Panels) {
			for _, target := range p.Targets {
				if target.Datasource == nil {
					continue
				}

				if target.Datasource.UID == "tempo" && !strings.Contains(target.Query, "resource.service.name") {
					t.Errorf("%s, panel %q (id %d) has a trace query that is not scoped to $service_name: %s",
						name, p.Title, p.ID, target.Query)
				}

				// $layer takes every value the metric has, cache and mail
				// among them, and a log or trace panel is empty for those
				// two. The panel has to say so, or it reads as "no errors".
				query := target.Expr + target.Query
				onLayerVariable := strings.Contains(query, "app_layer=`$layer`") || strings.Contains(query, `.app.layer = "$layer"`)

				if onLayerVariable && !strings.Contains(p.Description, "cache and mail") {
					t.Errorf("%s, panel %q (id %d) filters on $layer and does not say it is empty for cache and mail", name, p.Title, p.ID)
				}
			}
		}
	}

	// One log panel and two data links filter on a layer today, and one
	// trace panel on a layer and an outcome. Fewer means the patterns above
	// no longer read the dashboards, and the test would pass on nothing.
	if logFilters < 3 {
		t.Errorf("%d log queries filter on app_layer, want at least 3", logFilters)
	}

	if traceFilters < 2 {
		t.Errorf("%d trace filters on app.layer or app.outcome, want at least 2", traceFilters)
	}
}
