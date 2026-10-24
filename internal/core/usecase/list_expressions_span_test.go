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

			// The input's own Validate, as a statement of the function's
			// body: not a Validate of another value, not one in a closure or
			// a branch that may not run.
			var validated, attribute token.Pos

			for _, stmt := range fn.Body.List {
				var call ast.Expr

				switch st := stmt.(type) {
				case *ast.IfStmt:
					if assign, ok := st.Init.(*ast.AssignStmt); ok && len(assign.Rhs) == 1 {
						call = assign.Rhs[0]
					}
				case *ast.ExprStmt:
					call = st.X
				}

				if isInputValidate(call) && validated == token.NoPos {
					validated = stmt.Pos()
				}
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || attribute != token.NoPos {
					return true
				}

				if f, ok := call.Fun.(*ast.SelectorExpr); ok && f.Sel.Name == "String" && len(call.Args) > 0 && isExpressionKey(call.Args[0]) {
					attribute = call.Pos()
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

// isInputValidate reports whether expr is a call of input.Validate().
func isInputValidate(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	f, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || f.Sel.Name != "Validate" {
		return false
	}

	receiver, ok := f.X.(*ast.Ident)

	return ok && receiver.Name == "input"
}

// isExpressionKey reports whether an attribute's key is one of the three
// list expressions, spelled as a literal or as the domain's field name.
func isExpressionKey(key ast.Expr) bool {
	switch k := key.(type) {
	case *ast.BasicLit:
		return k.Value == `"sort"` || k.Value == `"filter"` || k.Value == `"fields"`
	case *ast.SelectorExpr:
		return k.Sel.Name == "FieldSort" || k.Sel.Name == "FieldFilter" || k.Sel.Name == "FieldFields"
	}

	return false
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
