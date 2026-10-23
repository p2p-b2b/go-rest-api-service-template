package respond

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

// handlerDir is the package whose arms the table was read from.
const handlerDir = "../handler"

// arm is one place a handler answers a domain error type: the type, as
// `*domain.X`, the status it answers it with, and where.
type arm struct {
	typeName string
	site     string
	status   int
}

// TestEveryArmAgreesWithTheTable: the handlers answer each domain error in
// an arm of their own (an `errors.AsType`, `isType` or `isErr` per type and
// per method), and the table (StatusFor) says the same thing once, for what
// has to agree with the answers without writing one: IsRefusal, which tells
// a refused call from a failed one. An arm that answers a type below 500
// must find it below 500 in the table, and an arm that answers it 500 must
// find it at 500 or more, or not in the table at all; a type the arms answer
// both ways would be counted as the wrong one on half its routes.
//
// It reads the handler package's source: every `if` and `case` whose
// condition names a domain error type, through the call itself or through a
// variable it was assigned to, and the first status the arm then answers.
// Two shapes keep the status out of the arm's body and are read too: a
// guard that returns when the error is none of its types, and answers after
// it; and a function that returns the status instead of writing it. An arm
// whose status cannot be read fails the test, so that a new shape is read
// rather than skipped.
//
// A function that answers whether an error is one of its types
// (`isBadRequest`) is read as the types it names, wherever a condition calls
// it. An error the handler package builds itself and answers without an arm
// (a path id that is not a uuid, a page token that does not parse) must be
// in the table, so that it is not taken for a failure; its status is the
// builder's call site's, and is not read.
//
// What is not an arm is not held: a handler that answers the error of
// getUserIDFromContext (an InvalidJWTError the middleware should have made
// impossible) with a 500, a 400 or a 401 does it without naming the type.
func TestEveryArmAgreesWithTheTable(t *testing.T) {
	arms, built := handlerArms(t)

	table := map[string]int{}
	for _, row := range statusTable {
		table[row.name] = row.status
	}

	types := map[string]bool{}

	for name := range built {
		if _, exempt := notRequests[name]; exempt {
			continue
		}

		if _, known := table[name]; !known {
			t.Errorf("the handler package builds %s and the table has no row for it: an error the handlers answer is not in the one table of statuses", name)
		}

		types[name] = true
	}

	for _, a := range arms {
		types[a.typeName] = true

		if a.status == 0 {
			t.Errorf("%s: the status of the arm for %s could not be read; read the new shape here", a.site, a.typeName)

			continue
		}

		want, known := table[a.typeName]
		if !known {
			want = http.StatusInternalServerError
		}

		if (a.status < http.StatusInternalServerError) != (want < http.StatusInternalServerError) {
			t.Errorf("%s answers %s with %d, and the table says %d: the two are on two sides of 500", a.site, a.typeName, a.status, want)
		}
	}

	// The table names nothing the handlers do not answer: a row no arm
	// reads and no handler builds is a guess about a status nobody gives.
	for name := range table {
		if !types[name] {
			t.Errorf("the table has a row for %s, which no handler answers", name)
		}
	}

	// The handler package has more than three hundred arms over some fifty
	// types; far fewer means the walk no longer reads them and the test
	// would pass on nothing.
	t.Logf("read %d arms over %d types", len(arms), len(types))

	if len(arms) < 300 || len(types) < 50 {
		t.Errorf("read %d arms over %d types, want the handler package", len(arms), len(types))
	}
}

// notRequests are the domain errors the handler package builds that are no
// request's answer.
var notRequests = map[string]string{
	"*domain.InvalidServiceError":         "a constructor's refusal of a missing dependency, at startup",
	"*domain.InvalidOTConfigurationError": "a constructor's refusal of a missing dependency, at startup",
	"*domain.InternalServerError":         "the 500 by its name",
}

// handlerArms reads every arm of the handler package, and the domain error
// types it builds.
func handlerArms(t *testing.T) ([]arm, map[string]bool) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(handlerDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()

	var parsed []*ast.File

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		parsed = append(parsed, file)
	}

	// The predicates first: a function of one error that returns whether
	// it is one of the types it names.
	predicates := map[string][]string{}
	built := map[string]bool{}

	for _, file := range parsed {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Body != nil && len(fn.Body.List) == 1 {
				if ret, ok := fn.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 && returnsBool(fn) {
					if types := typesIn(ret.Results[0], nil, nil); len(types) > 0 {
						predicates[fn.Name.Name] = types
					}
				}
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "domain" && strings.HasSuffix(sel.Sel.Name, "Error") {
						built["*domain."+sel.Sel.Name] = true
					}
				}
			}

			return true
		})
	}

	var arms []arm

	for _, file := range parsed {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			if _, isPredicate := predicates[fn.Name.Name]; isPredicate {
				continue
			}

			arms = append(arms, armsOf(fset, fn, predicates)...)
		}
	}

	return arms, built
}

// returnsBool reports whether fn's one result is a bool.
func returnsBool(fn *ast.FuncDecl) bool {
	results := fn.Type.Results
	if results == nil || len(results.List) != 1 {
		return false
	}

	id, ok := results.List[0].Type.(*ast.Ident)

	return ok && id.Name == "bool"
}

// armsOf reads the arms of one function.
func armsOf(fset *token.FileSet, fn *ast.FuncDecl, predicates map[string][]string) []arm {
	// A variable an `errors.AsType[*domain.X](err)` was assigned to, in a
	// statement of its own (`_, isX := ...`), names X wherever a condition
	// reads it. One assigned in an if's init names X in that if only, and
	// is read there.
	inits := map[ast.Node]bool{}
	bound := map[string]string{}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if s, ok := n.(*ast.IfStmt); ok && s.Init != nil {
			inits[s.Init] = true
		}

		return true
	})

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if inits[n] {
			return false
		}

		if as, ok := n.(*ast.AssignStmt); ok && len(as.Lhs) == 2 && len(as.Rhs) == 1 {
			if name := domainTypeOf(as.Rhs[0]); name != "" {
				if id, ok := as.Lhs[1].(*ast.Ident); ok && id.Name != "_" {
					bound[id.Name] = name
				}
			}
		}

		return true
	})

	var arms []arm

	add := func(types []string, status int, at ast.Node) {
		for _, name := range types {
			arms = append(arms, arm{typeName: name, status: status, site: fmt.Sprintf("%s:%s", fset.Position(at.Pos()), fn.Name.Name)})
		}
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		for i, stmt := range block.List {
			switch s := stmt.(type) {
			case *ast.IfStmt:
				types := typesIn(s.Cond, bound, predicates)
				if as, ok := s.Init.(*ast.AssignStmt); ok && len(as.Rhs) == 1 {
					if name := domainTypeOf(as.Rhs[0]); name != "" {
						types = append(types, name)
					}
				}

				if len(types) == 0 {
					continue
				}

				status := statusIn(s.Body)
				if status == 0 && negated(s.Cond) {
					// `if !isX && !isY { return false }`, then the answer.
					status = statusIn(&ast.BlockStmt{List: block.List[i+1:]})
				}

				add(types, status, s)
			case *ast.SwitchStmt:
				for _, c := range s.Body.List {
					clause := c.(*ast.CaseClause)
					for _, e := range clause.List {
						if types := typesIn(e, bound, predicates); len(types) > 0 {
							add(types, statusIn(&ast.BlockStmt{List: clause.Body}), clause)
						}
					}
				}
			}
		}

		return true
	})

	return arms
}

// domainTypeOf is X when e is `errors.AsType[*domain.X](…)`,
// `isType[*domain.X](…)` or `isErr[*domain.X](…)`.
func domainTypeOf(e ast.Expr) string {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return ""
	}

	index, ok := call.Fun.(*ast.IndexExpr)
	if !ok {
		return ""
	}

	star, ok := index.Index.(*ast.StarExpr)
	if !ok {
		return ""
	}

	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "domain" {
		return ""
	}

	return "*domain." + sel.Sel.Name
}

// typesIn is every domain error type a condition names: in a call, through a
// variable it was bound to, or through a predicate that names them.
func typesIn(cond ast.Expr, bound map[string]string, predicates map[string][]string) []string {
	var types []string

	ast.Inspect(cond, func(n ast.Node) bool {
		e, ok := n.(ast.Expr)
		if !ok {
			return true
		}

		if name := domainTypeOf(e); name != "" {
			types = append(types, name)

			return false
		}

		if call, ok := e.(*ast.CallExpr); ok {
			if fn, ok := call.Fun.(*ast.Ident); ok {
				if named, ok := predicates[fn.Name]; ok {
					types = append(types, named...)

					return false
				}
			}
		}

		if id, ok := e.(*ast.Ident); ok {
			if name, ok := bound[id.Name]; ok {
				types = append(types, name)
			}
		}

		return true
	})

	return types
}

// negated reports whether a condition holds when none of its types is the
// error: a guard that leaves its arm's answer for after it.
func negated(cond ast.Expr) bool {
	switch c := cond.(type) {
	case *ast.UnaryExpr:
		return c.Op == token.NOT
	case *ast.BinaryExpr:
		return c.Op == token.LAND && negated(c.X) && negated(c.Y)
	case *ast.ParenExpr:
		return negated(c.X)
	}

	return false
}

// statusIn is the first status the statements answer: the one a respond call
// is given, 500 for the fixed internal error, or one returned as a value.
func statusIn(n ast.Node) int {
	status := 0

	ast.Inspect(n, func(m ast.Node) bool {
		if status != 0 {
			return false
		}

		switch x := m.(type) {
		case *ast.ReturnStmt:
			for _, r := range x.Results {
				if s := httpStatus(r); s != 0 {
					status = s
				}
			}
		case *ast.CallExpr:
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "respond" {
				return true
			}

			if sel.Sel.Name == "WriteInternalError" {
				status = http.StatusInternalServerError

				return false
			}

			for _, a := range x.Args {
				if s := httpStatus(a); s != 0 {
					status = s

					return false
				}
			}
		}

		return true
	})

	return status
}

// httpStatus is the value of `http.StatusX`, or 0.
func httpStatus(e ast.Expr) int {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return 0
	}

	if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "http" {
		return 0
	}

	return httpStatuses[sel.Sel.Name]
}

// httpStatuses are the statuses the handlers answer with, by name.
var httpStatuses = func() map[string]int {
	m := map[string]int{}

	for code := 100; code < 600; code++ {
		if text := http.StatusText(code); text != "" {
			m["Status"+strings.NewReplacer(" ", "", "-", "", "'", "").Replace(text)] = code
		}
	}

	// The names whose text does not spell them.
	m["StatusRequestEntityTooLarge"] = http.StatusRequestEntityTooLarge
	m["StatusRequestURITooLong"] = http.StatusRequestURITooLong
	m["StatusTeapot"] = http.StatusTeapot

	return m
}()
