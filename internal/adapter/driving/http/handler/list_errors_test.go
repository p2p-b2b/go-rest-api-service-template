//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/respond"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driving"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// unreachable is what the driver says when the database cannot be reached:
// the text names where the service connects to, and as whom.
var unreachable = errors.New("failed to connect to `user=service database=service_db`:\n\t10.0.0.7:5432 (db.internal): dial error: connection refused")

func messageIn(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("the answer is not the message envelope: %v. Body: %s", err, rec.Body.String())
	}

	return body.Message
}

// A list that fails says whose mistake it was. What the caller sent wrong is
// a 400 in the error's own words, which say what to correct. Anything else
// is the fixed 500: the cause is the operator's, on the request's log line,
// and no word of it is in the answer.
func TestAFailedListSaysNothingOfItsCause(t *testing.T) {
	t.Parallel()

	refusal := &domain.ValidationErrors{}
	refusal.AddError("sort", "field not allowed", "INVALID_SORT")

	for name, c := range map[string]struct {
		err    error
		status int
	}{
		"the_database_is_unreachable":  {err: unreachable, status: http.StatusInternalServerError},
		"a_fault_wrapped_on_the_way":   {err: errors.Join(errors.New("listing projects"), unreachable), status: http.StatusInternalServerError},
		"an_error_of_no_known_kind":    {err: errors.New("pq: relation \"projects\" does not exist"), status: http.StatusInternalServerError},
		"a_sort_that_is_refused":       {err: refusal, status: http.StatusBadRequest},
		"a_column_that_does_not_exist": {err: &domain.UndefinedColumnError{Message: "the request refers to a field that does not exist"}, status: http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			writeListError(rec, httptest.NewRequest(http.MethodGet, "/projects", nil), c.err)

			if rec.Code != c.status {
				t.Fatalf("status = %d, want %d. Body: %s", rec.Code, c.status, rec.Body.String())
			}

			message := messageIn(t, rec)

			if c.status == http.StatusBadRequest {
				if message != c.err.Error() {
					t.Errorf("a refusal says %q, want the error's words %q", message, c.err.Error())
				}

				return
			}

			if message != respond.InternalServerErrorMessage {
				t.Errorf("a 500 says %q, want %q", message, respond.InternalServerErrorMessage)
			}

			for _, leaked := range []string{"connect", "user=", "database=", "10.0.0.7", "db.internal", "relation", "listing"} {
				if strings.Contains(rec.Body.String(), leaked) {
					t.Errorf("the answer carries %q of the cause: %s", leaked, rec.Body.String())
				}
			}
		})
	}
}

// failingProjects is the projects port for the two lists, with the database
// away. Any other method panics: the embedded port is nil.
type failingProjects struct{ driving.Projects }

func (failingProjects) List(context.Context, *domain.ListProjectsInput) (*domain.ListProjectsOutput, error) {
	return nil, unreachable
}

func (failingProjects) ListByUserID(context.Context, *domain.ListProjectsInput) (*domain.ListProjectsOutput, error) {
	return nil, unreachable
}

// And through a handler, as a request reaches it: the two lists of projects,
// with the database away.
func TestAListOfProjectsWithTheDatabaseAwayIsTheFixed500(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	h, err := NewProjectsHandler(ProjectsHandlerConf{Service: failingProjects{}, OT: ot})
	if err != nil {
		t.Fatalf("NewProjectsHandler: %v", err)
	}

	caller := uuid.NewV7()

	for name, serve := range map[string]func(http.ResponseWriter, *http.Request){
		"the_callers_own": h.list,
		"another_users":   h.listByUserID,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/projects", nil)
			req.SetPathValue("user_id", uuid.NewV7().String())

			rec := httptest.NewRecorder()
			serve(rec, withCaller(req, caller))

			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500. Body: %s", rec.Code, rec.Body.String())
			}

			if message := messageIn(t, rec); message != respond.InternalServerErrorMessage {
				t.Errorf("the 500 says %q, want %q", message, respond.InternalServerErrorMessage)
			}

			if strings.Contains(rec.Body.String(), "database=") || strings.Contains(rec.Body.String(), "db.internal") {
				t.Errorf("the answer names the database: %s", rec.Body.String())
			}
		})
	}
}

// handlerFunctions parses the package's own files and calls visit for every
// function that has a body.
func handlerFunctions(t *testing.T, fset *token.FileSet, visit func(fn *ast.FuncDecl)) (constants map[string]bool) {
	t.Helper()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list the package: %v", err)
	}

	constants = map[string]bool{}

	var functions []*ast.FuncDecl

	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		file, err := parser.ParseFile(fset, name, body, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Body != nil {
					functions = append(functions, d)
				}
			case *ast.GenDecl:
				// The package's own constants: a name for a fixed sentence.
				if d.Tok != token.CONST {
					continue
				}

				for _, spec := range d.Specs {
					for _, name := range spec.(*ast.ValueSpec).Names {
						constants[name.Name] = true
					}
				}
			}
		}
	}

	for _, fn := range functions {
		visit(fn)
	}

	return constants
}

// Every handler of this package writes an error's text only at a status the
// call names, and at a 5xx it writes a sentence of the service's.
//
// A message written at a computed status is a message that is sometimes a
// 500's: the lists wrote `e.Error()` at whatever their error mapped to, and
// compiled, and passed every test, because no test failed a list for a
// reason that was not the caller's. writeListError is the place that chooses
// between the two for them, and it answers a 500 through
// respond.WriteInternalError.
//
// "A sentence of the service's" is read from the argument itself: a string,
// a constant, or a name that was given one of those. Whatever is computed
// from an error (its text, a Sprintf of it, a variable that holds either) is
// not, however it is spelled.
func TestNoHandlerWritesAnErrorsTextAtA500(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	source := func(n ast.Node) string {
		var b strings.Builder

		if err := printer.Fprint(&b, fset, n); err != nil {
			t.Fatalf("print: %v", err)
		}

		return b.String()
	}

	// The writers, and where each takes its status and what it says.
	type writer struct{ status, says int }

	writers := map[string]writer{
		"respond.WriteJSONMessage":         {status: 2, says: 3},
		"respond.WriteJSONMessageWithCode": {status: 2, says: 4},
		"respond.WriteJSONData":            {status: 1, says: 2},
	}

	type found struct {
		fn   *ast.FuncDecl
		call *ast.CallExpr
		w    writer
	}

	var calls []found

	constants := handlerFunctions(t, fset, func(fn *ast.FuncDecl) {
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if w, isWriter := writers[source(call.Fun)]; isWriter && len(call.Args) > w.says {
					calls = append(calls, found{fn: fn, call: call, w: w})
				}
			}

			return true
		})
	})

	// fixed reports whether e is a sentence of the service's: a string, a
	// constant of this package or of another (domain.X, respond.X), or a
	// name that this function only ever gave such a value.
	var fixed func(fn *ast.FuncDecl, e ast.Expr, depth int) bool

	fixed = func(fn *ast.FuncDecl, e ast.Expr, depth int) bool {
		switch v := e.(type) {
		case *ast.BasicLit:
			return v.Kind == token.STRING
		case *ast.SelectorExpr:
			pkg, ok := v.X.(*ast.Ident)

			return ok && (pkg.Name == "domain" || pkg.Name == "respond")
		case *ast.Ident:
			if constants[v.Name] {
				return true
			}

			if depth > 3 {
				return false
			}

			given, all := 0, true

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				assign, ok := n.(*ast.AssignStmt)
				if !ok || len(assign.Lhs) != len(assign.Rhs) {
					return true
				}

				for i, lhs := range assign.Lhs {
					if name, ok := lhs.(*ast.Ident); ok && name.Name == v.Name {
						given++

						all = all && fixed(fn, assign.Rhs[i], depth+1)
					}
				}

				return true
			})

			return given > 0 && all
		}

		return false
	}

	// branchOn is the condition of the innermost `case` the call stands in.
	branchOn := func(fn *ast.FuncDecl, call *ast.CallExpr) string {
		condition := ""

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			clause, ok := n.(*ast.CaseClause)
			if !ok || call.Pos() < clause.Pos() || call.End() > clause.End() {
				return true
			}

			parts := make([]string, len(clause.List))
			for i, e := range clause.List {
				parts[i] = source(e)
			}

			condition = strings.Join(parts, ", ") // an inner clause is visited later and wins

			return true
		})

		return condition
	}

	messages := 0

	for _, c := range calls {
		status, says := source(c.call.Args[c.w.status]), c.call.Args[c.w.says]
		where := fset.Position(c.call.Pos()).String() + " (" + c.fn.Name.Name + ")"
		isData := source(c.call.Fun) == "respond.WriteJSONData"

		if !isData {
			messages++
		}

		// The status is one the call names. Data may go out at a status
		// that was worked out (the health answer is its payload at 200 or
		// at 503), as long as it is not built from an error.
		if !strings.HasPrefix(status, "http.Status") {
			if isData {
				if strings.Contains(source(says), "Error()") || strings.Contains(source(says), "Sprint") {
					t.Errorf("%s answers at the computed status %s with data built from an error", where, status)
				}

				continue
			}

			if c.fn.Name.Name != "writeListError" {
				t.Errorf("%s answers at the computed status %s: name the status, or answer through writeListError", where, status)
			}

			continue
		}

		if !serverFault[status] {
			continue
		}

		// At a 5xx, data is not an error's text in an envelope of its own.
		if isData {
			if strings.Contains(source(says), "Error()") || strings.Contains(source(says), "Sprint") {
				t.Errorf("%s answers %s with data built from an error: answer through respond.WriteInternalError", where, status)
			}

			continue
		}

		if fixed(c.fn, says, 0) {
			continue
		}

		// What is left is an error's text. It is meant in the places below,
		// each a branch on a type whose sentence the service writes.
		if typ, meant := typedFault[c.fn.Name.Name+" "+status]; meant && strings.Contains(branchOn(c.fn, c.call), "isType[*domain."+typ+"]") && source(says) == "e.Error()" {
			continue
		}

		t.Errorf("%s writes %s at %s: at a 5xx the message is a sentence of the service's; answer through respond.WriteInternalError", where, source(says), status)
	}

	// The scan read the package: its handlers write some hundreds of
	// messages. Far fewer means it no longer finds them.
	if messages < 100 {
		t.Errorf("read %d messages, want the package's hundreds: this scan no longer reads them", messages)
	}
}

// typedFault are the 5xx answers that carry an error's text on purpose, as
// "<function> <status>" and the type the answer branches on. The type is one
// whose text the service writes itself, from fixed sentences: no word of the
// provider's, or of a library's, is in it.
var typedFault = map[string]string{
	// A sign-in through an identity provider that could not be reached, at
	// its start (the discovery document): "identity provider unreachable:
	// the identity provider is not reachable". The caller is told whose
	// fault it is, and no more. The callback's 503 is one fixed sentence.
	"writeStartError http.StatusServiceUnavailable": "IDPUnreachableError",
}

// serverFault are the statuses a handler answers when the fault is the
// service's.
var serverFault = map[string]bool{
	"http.StatusInternalServerError": true,
	"http.StatusNotImplemented":      true,
	"http.StatusBadGateway":          true,
	"http.StatusServiceUnavailable":  true,
	"http.StatusGatewayTimeout":      true,
}

// A handler that reads the list parameters answers its list's failure
// through the helper, with the failure it recorded, and stops there.
//
// The helper is right in one place and the lists are many: a list that
// writes its own answer again, that hands the helper the bare error (the
// failure is then recorded by nobody), or that goes on after it, compiles
// and passes every test that does not drive that one list.
func TestEveryListAnswersItsFailureThroughTheHelper(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	source := func(n ast.Node) string {
		var b strings.Builder

		if err := printer.Fprint(&b, fset, n); err != nil {
			t.Fatalf("print: %v", err)
		}

		return b.String()
	}

	const recorded = "writeListError(w, r, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs))"

	lists, answered := 0, 0

	handlerFunctions(t, fset, func(fn *ast.FuncDecl) {
		readsListParameters, throughHelper, ownHelper := false, 0, false

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				switch source(node.Fun) {
				case "parseListQueryParams", "ref.listInput":
					// The products handler reads them through a helper of
					// its own, shared by its two lists.
					readsListParameters = true
				case "ref.writeQueryError":
					// The rate-limit rules have a helper of their own, which
					// names its 400s and answers the rest through
					// respond.WriteInternalError.
					ownHelper = true
				}
			case *ast.BlockStmt:
				for i, stmt := range node.List {
					expr, ok := stmt.(*ast.ExprStmt)
					if !ok {
						continue
					}

					call, ok := expr.X.(*ast.CallExpr)
					if !ok || source(call.Fun) != "writeListError" {
						continue
					}

					throughHelper++
					answered++

					where := fset.Position(call.Pos()).String() + " (" + fn.Name.Name + ")"

					if got := source(call); got != recorded {
						t.Errorf("%s answers with %s, want %s: the failure the handler recorded", where, got, recorded)
					}

					if i+1 >= len(node.List) {
						t.Errorf("%s goes on after its answer: nothing returns", where)
					} else if _, returns := node.List[i+1].(*ast.ReturnStmt); !returns {
						t.Errorf("%s goes on after its answer", where)
					}
				}
			}

			return true
		})

		// The two that read the parameters and run no list.
		if fn.Name.Name == "parseListQueryParams" || fn.Name.Name == "listInput" || !readsListParameters {
			return
		}

		lists++

		if throughHelper != 1 && !ownHelper {
			t.Errorf("%s reads the list parameters and answers its list's failure through the helper %d times, want once", fn.Name.Name, throughHelper)
		}
	})

	if lists < 10 || answered < lists-1 {
		t.Errorf("read %d lists and %d answers through the helper: this scan no longer reads them", lists, answered)
	}
}

// The start of a sign-in through an identity provider answers 503 when the
// provider cannot be reached, as its routes document, and names nothing but
// that. It answered 500: the 503 was on the type of another failure, which
// building the authorization URL never returns.
func TestAnUnreachableProviderAtTheStartIsA503(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	metrics, err := o11y.NewLayerMetrics(ot.Metrics.Meter, "")
	if err != nil {
		t.Fatalf("NewLayerMetrics: %v", err)
	}

	h := &AuthnIDPsHandler{metrics: metrics}

	for name, c := range map[string]struct {
		err    error
		status int
		says   string
	}{
		"the_provider_is_not_reachable": {
			err:    &domain.IDPUnreachableError{Message: "the identity provider is not reachable"},
			status: http.StatusServiceUnavailable,
			says:   "identity provider unreachable: the identity provider is not reachable",
		},
		"anything_else_is_the_fixed_500": {
			err:    errors.New("Get \"https://login.example.com/.well-known/openid-configuration\": dial tcp 10.0.0.9:443: i/o timeout"),
			status: http.StatusInternalServerError,
			says:   respond.InternalServerErrorMessage,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, span := ot.Traces.Tracer.Start(ctx, "test")
			defer span.End()

			rec := httptest.NewRecorder()
			h.writeStartError(rec, httptest.NewRequest(http.MethodGet, "/auth/idp/x/login", nil), span, time.Now(), nil, c.err)

			if rec.Code != c.status {
				t.Fatalf("status = %d, want %d. Body: %s", rec.Code, c.status, rec.Body.String())
			}

			if message := messageIn(t, rec); message != c.says {
				t.Errorf("the answer says %q, want %q", message, c.says)
			}

			if strings.Contains(rec.Body.String(), "login.example.com") || strings.Contains(rec.Body.String(), "10.0.0.9") {
				t.Errorf("the answer names where the provider is: %s", rec.Body.String())
			}
		})
	}
}
