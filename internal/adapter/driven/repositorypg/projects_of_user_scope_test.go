package repositorypg

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

// functionSource is one function of a file of this package as the compiler
// reads it: printed from the syntax tree, so that a comment inside it is not
// in what is read.
func functionSource(t *testing.T, file, name string) string {
	t.Helper()

	fset := token.NewFileSet()

	parsed, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}

	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != name {
			continue
		}

		var b strings.Builder

		if err := printer.Fprint(&b, fset, fn.Body); err != nil {
			t.Fatalf("print %s: %v", name, err)
		}

		return b.String()
	}

	t.Fatalf("%s has no function %s: it was renamed or removed, correct this test", file, name)

	return ""
}

// GET /users/{user_id}/projects names two accounts: whose projects, in the
// path, and who asks. Its list is the first one's projects that are the
// second one's too, both read from the view the projects list reads, so that
// "the caller's" means every project for the administrator account and the
// memberships of anyone else.
//
// The unit suite has no database, so this reads the statement: both
// predicates are in it, each account is the argument its predicate names,
// and a list asked for with no caller is refused before anything is read.
func TestAUsersProjectsAreReadHeldToTheCaller(t *testing.T) {
	t.Parallel()

	source := functionSource(t, "projects.go", "SelectByUserID")

	for _, term := range []string{
		"FROM view_projects_users AS vp",
		"WHERE vp.user_id = $1",
		"AND EXISTS (",
		"SELECT 1 FROM view_projects_users AS callers",
		"WHERE callers.user_id = $2 AND callers.id = vp.id",
	} {
		if !strings.Contains(source, term) {
			t.Errorf("the statement has no %q", term)
		}
	}

	// And the statement's scope whole, spaces apart: a term added inside it
	// ("OR TRUE") keeps every term above and answers every project. What
	// follows it is the filter's and the pager's.
	const scope = "FROM view_projects_users AS vp WHERE vp.user_id = $1 AND EXISTS ( " +
		"SELECT 1 FROM view_projects_users AS callers " +
		"WHERE callers.user_id = $2 AND callers.id = vp.id ) {{ .QueryWhere }} " +
		"ORDER BY {{.QueryInternalSort}} LIMIT {{.QueryLimit}} )"

	if !strings.Contains(strings.Join(strings.Fields(source), " "), scope) {
		t.Errorf("the statement's scope changed; want, spaces apart:\n%s", scope)
	}

	// The statement that runs is the template as it was rendered: nothing
	// is assigned to it afterwards.
	if got := strings.Count(source, "query ="); got != 0 {
		t.Errorf("the rendered statement is assigned to %d times after it is rendered", got)
	}

	if !strings.Contains(source, "query := tpl.String()") {
		t.Error("the statement that runs is not the rendered template")
	}

	// One statement, and its two accounts in the order the predicates name
	// them. With the path's user in both places the list is that user's
	// whole list.
	if got := strings.Count(source, "ref.db.Query("); got != 1 {
		t.Fatalf("the list runs %d statements, want one: read each one", got)
	}

	if !strings.Contains(source, "ref.db.Query(ctx, query, input.UserID, input.CallerID)") {
		t.Error("the statement is not run with the path's user as $1 and the caller as $2")
	}

	// Never the user's whole list: with no caller the list is refused, and
	// the refusal comes before the statement.
	refusal := strings.Index(source, "if input.CallerID == uuid.Nil() {")
	query := strings.Index(source, "ref.db.Query(")

	if refusal < 0 {
		t.Fatal("a list asked for with no caller is not refused")
	}

	if refusal > query {
		t.Error("the list is read before the caller is looked at")
	}

	// The other list, GET /projects, has one account, the caller, and no
	// second predicate to hold it to.
	if own := functionSource(t, "projects.go", "Select"); strings.Contains(own, "AS callers") {
		t.Error("the caller's own list is held to a second account")
	}
}
