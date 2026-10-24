package middleware

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// o11y.NoteFailure records its caller's caller as where a failure arose.
// Called straight from a middleware's closure that is net/http's
// HandlerFunc.ServeHTTP, and from a deferred recover the runtime's panic, so
// the access line named neither the middleware nor anything of the service.
// The package calls it through noteFailure alone, which makes the caller's
// caller the middleware.
func TestTheMiddlewaresNoteAFailureThroughTheirHelper(t *testing.T) {
	t.Parallel()

	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, src, 0)
		if err != nil {
			t.Fatal(err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Name.Name == "noteFailure" {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "NoteFailure" {
					if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "o11y" {
						t.Errorf("%s: %s calls o11y.NoteFailure itself; call noteFailure", path, fn.Name.Name)
					}
				}

				return true
			})
		}
	}
}

// Where a failure the middleware answered arose, as the access line says it:
// the middleware's own function, for a panic.
func TestAMiddlewaresFailureIsSaidToAriseInIt(t *testing.T) {
	t.Parallel()

	held := func(t *testing.T, ctx context.Context) o11y.Failure {
		t.Helper()

		failure, ok := o11y.FailureFrom(ctx)
		if !ok {
			t.Fatal("the failure was not handed to the access line")
		}

		return failure
	}

	t.Run("a_panic", func(t *testing.T) {
		t.Parallel()

		ctx := o11y.WithFailure(t.Context())
		h := Logging(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("a handler's bug")
		}))

		func() {
			defer func() { _ = recover() }()

			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(ctx, http.MethodGet, "/things", nil))
		}()

		if failure := held(t, ctx); !strings.Contains(failure.Func, "middleware.Logging") {
			t.Errorf("the panic is said to arise in %q, want the middleware", failure.Func)
		}
	})
}
