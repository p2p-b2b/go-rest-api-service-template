package o11y_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWhoeverRecordsOnTheLayerInstrumentsSaysTheOutcome: the shared pair is
// recorded by recordResult, and by two adapters that mirror their own
// results onto it themselves (cache, mail). The dashboards' error ratios are
// app_outcome="failed" over all calls, so a recorder that sets no
// app.outcome is a layer whose failures no ratio can see: measured before
// this guard, 12 458 cache calls and 100 mail calls carried no outcome.
//
// A recorder is a function that adds to a LayerMetrics' counter or records
// on its histogram. Each must set the attribute in its own body: a second
// recorder in a file that already has one is not excused by its neighbour.
func TestWhoeverRecordsOnTheLayerInstrumentsSaysTheOutcome(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	fset := token.NewFileSet()

	var recorders []string

	err := filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			var records, saysOutcome bool

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.CallExpr:
					// x.Counter.Add(...) or x.Histogram.Record(...): the two
					// fields of o11y.LayerMetrics.
					sel, ok := node.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					field, ok := sel.X.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					if (field.Sel.Name == "Counter" && sel.Sel.Name == "Add") || (field.Sel.Name == "Histogram" && sel.Sel.Name == "Record") {
						records = true
					}
				case *ast.SelectorExpr:
					if node.Sel.Name == "AttrOutcome" {
						saysOutcome = true
					}
				case *ast.Ident:
					if node.Name == "AttrOutcome" {
						saysOutcome = true
					}
				}

				return true
			})

			if !records {
				continue
			}

			recorders = append(recorders, rel+":"+fn.Name.Name)

			if !saysOutcome {
				t.Errorf("%s: %s records on the layer instruments and sets no app.outcome", rel, fn.Name.Name)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking the tree: %v", err)
	}

	// recordResult, the cache adapter's recordLayer and the mail adapter's.
	if len(recorders) < 3 {
		t.Errorf("found %d recorders of the shared pair (%v), want at least 3: the guard is reading nothing", len(recorders), recorders)
	}
}

// TestARepositoryRecordsTheErrorItReturns: a repository maps the driver's
// error to the service's own (a unique violation is "already exists") and
// records what happened. Recorded before it was mapped, the call was counted
// with the driver's error -- which no table knows, so a failure -- and then
// returned the mapped one, which is answered 409: one refusal, counted as a
// failure at the repository layer. Measured on the suite: two inserts.
func TestARepositoryRecordsTheErrorItReturns(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	dir := filepath.Join(root, "internal", "adapter", "driven", "repositorypg")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	var read int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		read++

		lines := strings.Split(string(source), "\n")
		for i := 1; i < len(lines); i++ {
			// `return ref.handlePgError(x, …)` straight after a RecordError
			// of the unmapped error.
			if !strings.Contains(lines[i], "return ") || !strings.Contains(lines[i], "ref.handlePgError(") {
				continue
			}

			if strings.Contains(lines[i-1], "o11y.RecordError(") && !strings.Contains(lines[i-1], "handlePgError(") {
				t.Errorf("%s:%d records the driver's error and returns a mapped one; map it first:\n%s\n%s",
					entry.Name(), i+1, lines[i-1], lines[i])
			}
		}
	}

	if read < 15 {
		t.Errorf("read %d repository files: the guard is reading nothing", read)
	}
}
