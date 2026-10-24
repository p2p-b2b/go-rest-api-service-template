//go:build unit

package usecase

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// A delete removes the row, gives the usage slot back, and drops what is
// cached of the row. The decrement can fail after the row is gone (a fault of
// the database, or of the counter's signature), and the function returns its
// error. Three deletes dropped the cache after the decrement: a decrement
// that failed returned with a deleted user, provider or project still in the
// cache, for the cache's lifetime (GET /users/{user_id} answered the deleted
// account).
//
// This reads the package: in every function that both gives a usage slot back
// and drops something from the cache, the cache is dropped first.
func TestADeleteDropsItsCacheBeforeItGivesItsSlotBack(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the package: %v", err)
	}

	fset := token.NewFileSet()
	found := map[string]bool{}

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

			var decrement, drop token.Pos

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// The first decrement and the LAST invalidation: an unrelated
				// key dropped before the decrement does not excuse the row's
				// own dropped after it.
				if f, ok := call.Fun.(*ast.SelectorExpr); ok {
					switch {
					case f.Sel.Name == "DecrementUsage" && decrement == token.NoPos:
						decrement = call.Pos()
					case f.Sel.Name == "Invalidate":
						drop = call.Pos()
					}
				}

				return true
			})

			if decrement == token.NoPos || drop == token.NoPos {
				continue
			}

			where := name + ":" + fn.Name.Name
			found[where] = true

			if drop > decrement {
				t.Errorf("%s gives its usage slot back at line %d and drops its cache at line %d: a decrement that fails returns with the deleted row still cached",
					where, fset.Position(decrement).Line, fset.Position(drop).Line)
			}
		}
	}

	// The three that were wrong, by name: a delete that stops calling one of
	// the two is not looked at above, and this says so.
	for _, where := range []string{"projects.go:DeleteByID", "users.go:DeleteByID", "idps.go:DeleteByID"} {
		if !found[where] {
			t.Errorf("%s no longer both gives a usage slot back and drops its cache: found %v", where, found)
		}
	}
}
