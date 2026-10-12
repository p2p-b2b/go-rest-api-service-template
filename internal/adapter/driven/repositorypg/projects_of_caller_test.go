package repositorypg

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// oneRow answers one QueryRow: the id it holds, no row at all, or an error.
// It records what it was asked.
type oneRow struct {
	err   error
	sql   string
	args  []any
	asked int
	id    uuid.UUID
}

func (r *oneRow) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	r.asked++
	r.sql, r.args = sql, args

	return r
}

func (r *oneRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	*dest[0].(*uuid.UUID) = r.id

	return nil
}

// TestProjectsOfCaller: the check a change to a user's project links makes
// of the projects in its body.
func TestProjectsOfCaller(t *testing.T) {
	t.Parallel()

	caller := uuid.NewV7()
	wanted := []uuid.UUID{uuid.NewV7(), uuid.NewV7()}

	t.Run("every_project_is_the_callers", func(t *testing.T) {
		t.Parallel()

		db := &oneRow{err: pgx.ErrNoRows}

		if err := projectsOfCaller(t.Context(), db, caller, wanted); err != nil {
			t.Fatalf("refused: %v", err)
		}

		if db.asked != 1 {
			t.Fatalf("asked %d times, want once", db.asked)
		}

		// The caller's row set of the view the projects list reads, so that
		// "linked" means here what it means there, the administrator
		// account's "every project" included. "NOT" and the join of the two
		// sides are the two terms a wrong statement drops: without the
		// first an id that IS the caller's is the one refused, without the
		// second any project of the caller's vouches for every id.
		for _, term := range []string{
			"FROM view_projects_users AS callers",
			"callers.user_id = $1",
			"UNNEST($2::uuid[])",
			"WHERE NOT EXISTS (",
			"callers.id = wanted.id",
		} {
			if !strings.Contains(db.sql, term) {
				t.Errorf("the statement has no %q:\n%s", term, db.sql)
			}
		}

		// And the statement whole, spaces apart: a term added to it ("OR
		// TRUE", "AND FALSE") keeps every term above and changes what it
		// answers. Whoever changes the statement changes this line, and
		// reads the integration test that holds what it means.
		const want = "SELECT wanted.id FROM UNNEST($2::uuid[]) AS wanted(id) " +
			"WHERE NOT EXISTS ( SELECT 1 FROM view_projects_users AS callers " +
			"WHERE callers.user_id = $1 AND callers.id = wanted.id ) LIMIT 1;"

		if got := strings.Join(strings.Fields(db.sql), " "); got != want {
			t.Errorf("the statement changed:\n got: %s\nwant: %s", got, want)
		}

		if len(db.args) != 2 || db.args[0] != caller {
			t.Fatalf("args = %v, want the caller as $1", db.args)
		}

		ids, ok := db.args[1].([]string)
		if !ok || len(ids) != 2 || ids[0] != wanted[0].String() || ids[1] != wanted[1].String() {
			t.Errorf("$2 = %v, want the body's projects", db.args[1])
		}
	})

	t.Run("a_project_that_is_not_the_callers_refuses_the_request", func(t *testing.T) {
		t.Parallel()

		err := projectsOfCaller(t.Context(), &oneRow{id: wanted[1]}, caller, wanted)

		refusal, ok := errors.AsType[*domain.ProjectNotOfCallerError](err)
		if !ok {
			t.Fatalf("got %T (%v), want the refusal the handler answers 400", err, err)
		}

		if refusal.ProjectID != wanted[1] {
			t.Errorf("the refusal names %s, want the id the caller sent, %s", refusal.ProjectID, wanted[1])
		}

		if strings.Contains(refusal.Error(), wanted[0].String()) {
			t.Errorf("the refusal names a project that is the caller's: %s", refusal.Error())
		}
	})

	t.Run("no_caller_is_refused_before_anything_is_asked", func(t *testing.T) {
		t.Parallel()

		db := &oneRow{err: pgx.ErrNoRows}

		err := projectsOfCaller(t.Context(), db, uuid.Nil(), wanted)
		if !errors.Is(err, errLinksWithoutCaller) {
			t.Fatalf("got %v, want errLinksWithoutCaller", err)
		}

		if db.asked != 0 {
			t.Errorf("the database was asked %d times with no caller", db.asked)
		}

		// A fault of the service, not the caller's mistake: none of the
		// errors the handlers answer with a 4xx.
		if _, refused := errors.AsType[*domain.ProjectNotOfCallerError](err); refused {
			t.Error("a write with no caller is answered as the caller's mistake")
		}
	})

	t.Run("a_fault_of_the_database_is_not_a_refusal", func(t *testing.T) {
		t.Parallel()

		fault := errors.New("the database is away")

		if err := projectsOfCaller(t.Context(), &oneRow{err: fault}, caller, wanted); !errors.Is(err, fault) {
			t.Fatalf("got %v, want the fault", err)
		}
	})
}

// TestAUsersProjectLinksAreCheckedBeforeTheyAreWritten reads the two writes.
// Each asks whose the body's projects are, with the caller and the body's
// ids, and returns when it is refused, before its statement runs. Written
// without it, or with the path's user as the caller, either compiles and
// passes every other unit test: the unit suite has no database to see it
// with.
//
// It reads the syntax tree, not the text: a test that looks for the call's
// spelling passes with the call's result thrown away, and with the call in a
// comment.
func TestAUsersProjectLinksAreCheckedBeforeTheyAreWritten(t *testing.T) {
	t.Parallel()

	writes := map[string]bool{"LinkProjects": true, "UnlinkProjects": true}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "users.go", nil, 0)
	if err != nil {
		t.Fatalf("parse users.go: %v", err)
	}

	// expr is the source of an expression, as written.
	expr := func(e ast.Expr) string {
		var b strings.Builder

		if err := printer.Fprint(&b, fset, e); err != nil {
			t.Fatalf("print: %v", err)
		}

		return b.String()
	}

	seen := 0

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !writes[fn.Name.Name] {
			continue
		}

		seen++

		var checked, written token.Pos

		checks, writesFound := 0, 0

		// The check is a statement of the function's own body: inside a
		// condition of its own ("unless the caller is the user") it is a
		// check that some requests skip.
		for _, stmt := range fn.Body.List {
			node, ok := stmt.(*ast.IfStmt)
			if !ok {
				continue
			}

			// if err := projectsOfCaller(ctx, ref.db, input.CallerID, input.ProjectIDs); err != nil { return ... }
			assign, ok := node.Init.(*ast.AssignStmt)
			if !ok || len(assign.Rhs) != 1 {
				continue
			}

			call, ok := assign.Rhs[0].(*ast.CallExpr)
			if !ok || expr(call.Fun) != "projectsOfCaller" {
				continue
			}

			checks++
			checked = node.Pos()

			if len(call.Args) != 4 {
				t.Fatalf("%s: the check takes %d arguments, want 4", fn.Name.Name, len(call.Args))
			}

			if got := expr(call.Args[1]); got != "ref.db" {
				t.Errorf("%s asks %s, want ref.db", fn.Name.Name, got)
			}

			if got := expr(call.Args[2]); got != "input.CallerID" {
				t.Errorf("%s holds the projects to %s, want input.CallerID: not to the path's user", fn.Name.Name, got)
			}

			if got := expr(call.Args[3]); got != "input.ProjectIDs" {
				t.Errorf("%s checks %s, want input.ProjectIDs", fn.Name.Name, got)
			}

			if got := expr(node.Cond); got != "err != nil" {
				t.Errorf("%s does not look at the refusal: %s", fn.Name.Name, got)
			}

			if len(node.Body.List) == 0 {
				t.Fatalf("%s: a refusal does nothing", fn.Name.Name)
			}

			if _, returns := node.Body.List[len(node.Body.List)-1].(*ast.ReturnStmt); !returns {
				t.Errorf("%s goes on after a refusal", fn.Name.Name)
			}
		}

		// The check is named once in the function, where it was found: a
		// second mention is a call this loop did not read.
		named := 0

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.Ident:
				if node.Name == "projectsOfCaller" {
					named++
				}
			case *ast.CallExpr:
				if expr(node.Fun) == "ref.db.Exec" {
					writesFound++
					written = node.Pos()
				}
			case *ast.AssignStmt:
				// Neither account is written to on the way to the check.
				for _, lhs := range node.Lhs {
					if got := expr(lhs); got == "input.CallerID" || got == "input.UserID" || got == "input.ProjectIDs" {
						t.Errorf("%s assigns %s: the input is the handler's", fn.Name.Name, got)
					}
				}
			}

			return true
		})

		if named != 1 {
			t.Errorf("%s names the check %d times, want once", fn.Name.Name, named)
		}

		if checks != 1 {
			t.Fatalf("%s asks whose the projects are %d times, want once, and acting on the answer", fn.Name.Name, checks)
		}

		if writesFound != 1 {
			t.Fatalf("%s has %d statements, want one: the statements changed, check each one", fn.Name.Name, writesFound)
		}

		if checked >= written {
			t.Errorf("%s writes before it asks", fn.Name.Name)
		}
	}

	if seen != len(writes) {
		t.Errorf("read %d of the %d writes of a user's project links: one was renamed or removed, correct this test", seen, len(writes))
	}
}

// A link names two rows, and each has a foreign key. They were one branch:
// either answered "User not found", so a project that was gone was a 404
// about the user. The project's is the refusal the check gives, for a
// project that was the caller's at the check and is gone at the insert; the
// user's stays the 404.
func TestALinksTwoForeignKeysAreToldApart(t *testing.T) {
	t.Parallel()

	input := &domain.LinkProjectsToUserInput{UserID: uuid.NewV7(), CallerID: uuid.NewV7(), ProjectIDs: []uuid.UUID{uuid.NewV7()}}

	t.Run("the_project_is_gone", func(t *testing.T) {
		t.Parallel()

		err := (&UsersRepository{}).handlePgError(&pgconn.PgError{
			Code:           "23503",
			ConstraintName: constraintProjectsUsersProject,
			Message:        `insert or update on table "projects_users" violates foreign key constraint "projects_users_projects_id_fkey"`,
		}, input)

		refusal, ok := errors.AsType[*domain.ProjectNotOfCallerError](err)
		if !ok {
			t.Fatalf("got %T (%v), want the refusal of a project that is not the caller's", err, err)
		}

		if strings.Contains(refusal.Error(), input.UserID.String()) {
			t.Errorf("the refusal names the user: %s", refusal.Error())
		}
	})

	t.Run("the_user_is_not_there", func(t *testing.T) {
		t.Parallel()

		err := (&UsersRepository{}).handlePgError(&pgconn.PgError{
			Code:           "23503",
			ConstraintName: "projects_users_users_id_fkey",
			Message:        `insert or update on table "projects_users" violates foreign key constraint "projects_users_users_id_fkey"`,
		}, input)

		missing, ok := errors.AsType[*domain.UserNotFoundError](err)
		if !ok {
			t.Fatalf("got %T (%v), want a missing user", err, err)
		}

		if missing.ID != input.UserID {
			t.Errorf("the missing user is %s, want the path's %s", missing.ID, input.UserID)
		}
	})
}

// The names the repository tells the two foreign keys apart by are the ones
// the migration writes out. Renamed there, the project's falls through to
// "User not found" and no build fails.
func TestTheLinksForeignKeysAreNamedInTheMigration(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("../../../../database/migrations/00005_projects_users.sql")
	if err != nil {
		t.Fatalf("read the migration: %v", err)
	}

	migration := string(body)

	for _, name := range []string{constraintProjectsUsersProject, "projects_users_users_id_fkey"} {
		if !strings.Contains(migration, "CONSTRAINT "+name+" REFERENCES") {
			t.Errorf("migration 00005 does not declare %s", name)
		}
	}
}
