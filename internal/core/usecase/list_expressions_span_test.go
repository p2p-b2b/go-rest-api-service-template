//go:build unit

package usecase

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"uuid"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// A list's sort, filter and fields arrive in the query string, which no body
// limit bounds, and a span is exported. The lists put the three on their
// span before the input was validated -- by the repository, after them -- so
// an expression refused for its length still reached the trace backend,
// up to the server's header limit per request. Each validates first.
//
// This reads the package: in every function that puts an expression on a
// span, a Validate call comes before it.
func TestAListsExpressionsReachTheSpanAfterTheyAreValidated(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the package: %v", err)
	}

	fset := token.NewFileSet()
	found := 0

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			var validated, attribute token.Pos

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				f, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				switch {
				case f.Sel.Name == "Validate" && validated == token.NoPos:
					validated = call.Pos()
				case f.Sel.Name == "String" && attribute == token.NoPos && len(call.Args) > 0:
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && (lit.Value == `"sort"` || lit.Value == `"filter"` || lit.Value == `"fields"`) {
						attribute = call.Pos()
					}
				}

				return true
			})

			if attribute == token.NoPos {
				continue
			}

			found++

			if validated == token.NoPos || validated > attribute {
				t.Errorf("%s:%s puts a list expression on its span at line %d before the input is validated",
					name, fn.Name.Name, fset.Position(attribute).Line)
			}
		}
	}

	if found < 6 {
		t.Errorf("found %d lists that put their expressions on a span, want at least 6: the guard is reading nothing", found)
	}
}

// selectingNothing is a projects repository a refused list must not reach.
type selectingNothing struct{ repository.Projects }

func (selectingNothing) Select(context.Context, *domain.SelectProjectsInput) (*domain.SelectProjectsOutput, error) {
	return nil, errors.New("a refused list reached the repository")
}

// The projects list, driven: an over-length sort is refused, and its span
// carries no expression.
func TestARefusedListsSpanCarriesNoExpression(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	ot := testOT(t)
	ot.Traces.Tracer = provider.Tracer("test")

	svc, err := NewProjectsService(ProjectsServiceConf{Repository: selectingNothing{}, ResourcesLimits: slotsGivenBack{}, OT: ot})
	if err != nil {
		t.Fatalf("NewProjectsService: %v", err)
	}

	long := strings.Repeat("name,", 20_000)

	_, err = svc.List(t.Context(), &domain.ListProjectsInput{UserID: uuid.NewV7(), Sort: long, Paginator: domain.Paginator{Limit: 10}})
	if _, ok := errors.AsType[*domain.ValidationErrors](err); !ok {
		t.Fatalf("got %T (%v), want the input's validation errors", err, err)
	}

	for _, span := range recorder.Ended() {
		for _, kv := range span.Attributes() {
			switch string(kv.Key) {
			case "sort", "filter", "fields":
				t.Errorf("span %q carries %q (%d bytes) for a list that was refused", span.Name(), kv.Key, len(kv.Value.AsString()))
			}
		}
	}
}
