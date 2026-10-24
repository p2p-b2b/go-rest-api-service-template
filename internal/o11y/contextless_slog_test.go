package o11y_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// requestPathRoots are where the packages a request passes through live:
// every package under them is walked, at any depth.
//
// It was a list of packages written by hand, and three that a request passes
// through were missing from it, each logging without its context on every
// request it refused: the identity providers' adapter, the token verifier,
// and the middleware's validator. A list is right only on the day it is
// written; a walk takes in a package the day it is added.
//
// internal/app is deliberately outside them. Its logging is startup and
// shutdown -- configuration, migrations, the seeded-administrator check, the
// shutdown sequence -- and none of it happens inside a request, so there is
// no span for those lines to belong to. Requiring a context there would mean
// inventing context.Background() at the call site, which carries no span and
// would turn a real rule into a ritual.
var requestPathRoots = []string{
	filepath.Join("internal", "core"),
	filepath.Join("internal", "adapter"),
}

// notInTheRequestPath are packages under requestPathRoots that a request
// never runs in. The HTTP server's package starts, stops and watches the
// server: it is the process's lifecycle, as internal/app is.
var notInTheRequestPath = map[string]bool{
	filepath.Join("internal", "adapter", "driving", "http", "server"): true,
}

// requestPathPackages are the packages under requestPathRoots, relative to
// root, that hold Go source.
func requestPathPackages(t *testing.T, root string) []string {
	t.Helper()

	var packages []string

	for _, top := range requestPathRoots {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return err
			}

			pkg, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			if notInTheRequestPath[pkg] {
				return filepath.SkipDir
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
					packages = append(packages, pkg)

					break
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", top, err)
		}
	}

	// The walk must reach the packages the hand-written list named, and the
	// three it missed: fewer would be a walk that reads nothing.
	for _, want := range []string{
		filepath.Join("internal", "core", "usecase"),
		filepath.Join("internal", "adapter", "driven", "repositorypg"),
		filepath.Join("internal", "adapter", "driving", "http", "handler"),
		filepath.Join("internal", "adapter", "driven", "oauthidp"),
		filepath.Join("internal", "adapter", "driven", "tokenjwt"),
		filepath.Join("internal", "adapter", "driving", "http", "jwtvalidator"),
		filepath.Join("internal", "adapter", "driven", "notifieremail", "templates"),
	} {
		if !slices.Contains(packages, want) {
			t.Fatalf("the walk did not reach %s: %v", want, packages)
		}
	}

	return packages
}

// contextlessSlogExemptions names the FUNCTIONS that legitimately log without a
// context, because they do not run inside a request and so have no span for a
// record to belong to.
//
// It is keyed on the function rather than the file on purpose. A file-level
// exemption silently covers every call added to that file afterwards, including
// the ones this test exists to catch -- `users.go` holds both a scan helper and
// the repository methods, and exempting the file would stop guarding the
// methods. The key is the package's name and the function's
// ("usecase.dummyPasswordHash"): the walk covers every package under
// requestPathRoots, and a bare name would exempt a function of that name in
// any of them.
//
// Adding a name here is a claim that the function cannot be reached from a
// request. Check it before adding one.
var contextlessSlogExemptions = map[string]bool{
	// Maps a requested field name to the struct field to scan into. Pure column
	// plumbing, called from the row loop, needing no context.
	"repositorypg.buildScanFields": true,

	// A sync.OnceValue initialiser: it runs at first use, once per process,
	// not inside any request.
	"usecase.dummyPasswordHash": true,

	// Emitted while the CORS middleware is being CONSTRUCTED, at startup, to
	// say that a wildcard origin and credentials cannot be combined. There is
	// no request yet.
	"middleware.Cors": true,

	// Filters the rate-limit rule set when it is loaded from the database, not
	// per request. A rule that cannot be enforced is reported once, on reload.
	"usecase.EnforceableRateLimits": true,
}

// TestNoContextlessSlogInTheRequestPath fails the build when a log line in the
// request path is written without its context.
//
// # What it is guarding
//
// slog.Info and slog.InfoContext differ in exactly one way that matters here:
// the OpenTelemetry bridge reads the span out of the context argument and from
// nowhere else. A context-less call therefore produces a record with no
// trace_id, and a record with no trace_id cannot be found from the trace it
// belongs to -- which is the single most useful thing the log store does.
//
// It is a whole-codebase property rather than a per-call judgement, and it
// decays silently: nothing fails, no test goes red, the line simply is not
// there when somebody clicks "logs for this span" during an incident. That is
// the same shape of failure as TestNoSharedMetadataActionWrite, and it needs
// the same kind of guard.
//
// 340 call sites were converted at once. Without this test the next one written
// by hand would be context-less, because slog.Info is what muscle memory types.
//
// # Why AST and not grep
//
// A regexp would match the spelling inside a comment, a string literal, and
// this file's own documentation. Matching the call structurally is exact.
func TestNoContextlessSlogInTheRequestPath(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)

	var offenders []string

	for _, pkg := range requestPathPackages(t, root) {
		dir := filepath.Join(root, pkg)

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", pkg, err)
		}

		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			relative := filepath.ToSlash(filepath.Join(pkg, name))
			path := filepath.Join(dir, name)

			fset := token.NewFileSet()

			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}

			for _, decl := range file.Decls {
				owner := declName(decl)
				if contextlessSlogExemptions[filepath.Base(pkg)+"."+owner] {
					continue
				}

				ast.Inspect(decl, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}

					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					ident, ok := sel.X.(*ast.Ident)
					if !ok || ident.Name != "slog" {
						return true
					}

					switch sel.Sel.Name {
					case "Debug", "Info", "Warn", "Error":
						offenders = append(offenders,
							relative+":"+
								itoa(fset.Position(call.Pos()).Line)+
								" slog."+sel.Sel.Name+" in "+owner)
					}

					return true
				})
			}
		}
	}

	if len(offenders) > 0 {
		t.Errorf("%d log call(s) in the request path do not take a context, so their records "+
			"carry no trace id and cannot be found from the trace they belong to.\n"+
			"Use slog.%sContext(ctx, ...). If the function genuinely has no context -- it does "+
			"not run inside a request -- add its file to contextlessSlogExemptions with the "+
			"reason.\n  %s",
			len(offenders), "Info", strings.Join(offenders, "\n  "))
	}
}

// TestContextlessSlogExemptionsAreAllUsed guards the guard.
//
// An exemption naming a function that no longer exists is worse than no
// exemption: it reads as a considered decision while covering nothing, and if
// the name is ever reused the blind spot comes back silently. This is the same
// failure the config guard had, where a chdir in a sibling test made its glob
// match zero files, so the check passed while examining nothing.
func TestContextlessSlogExemptionsAreAllUsed(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	seen := map[string]bool{}

	for _, pkg := range requestPathPackages(t, root) {
		dir := filepath.Join(root, pkg)

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", pkg, err)
		}

		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", name, err)
			}

			for _, decl := range file.Decls {
				seen[filepath.Base(pkg)+"."+declName(decl)] = true
			}
		}
	}

	for exempt := range contextlessSlogExemptions {
		if !seen[exempt] {
			t.Errorf("contextlessSlogExemptions names %q, which no longer exists in the request path; delete the entry", exempt)
		}
	}
}

// declName is the name a declaration is known by: the function's own name, or
// the first name a var or const block declares. It is what the exemption list
// is keyed on.
func declName(decl ast.Decl) string {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return d.Name.Name
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok && len(value.Names) > 0 {
				return value.Names[0].Name
			}
		}
	}

	return ""
}

// itoa keeps the offender list readable without pulling strconv in for one use.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte

	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	return string(digits)
}
