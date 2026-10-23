package repositorypg

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"
)

// A provider's delete raced a first sign-in through it: the identity's
// insert holds a KEY SHARE lock on the provider's row, the delete read its
// condition before the insert committed and did not read it again after
// waiting for the lock, and its cascade removed the identity just written.
// The delete takes the row FOR UPDATE first, in its transaction, so its
// statement reads the identities after any such insert. The pull request gate
// has no database: the lock is held whole, and the order is read from
// DeleteByID's body.
func TestAProvidersDeleteLocksItsRowBeforeItsCheck(t *testing.T) {
	t.Parallel()

	for statement, want := range map[string]string{
		queryLockIDPForDelete:  "SELECT TRUE FROM idps WHERE id = $1 FOR UPDATE;",
		queryLockAccountsOfIDP: "SELECT id FROM users WHERE id IN (SELECT users_id FROM users_identities WHERE idps_id = $1) ORDER BY id FOR NO KEY UPDATE;",
	} {
		if got := strings.Join(strings.Fields(statement), " "); got != want {
			t.Errorf("a lock changed:\n got: %s\nwant: %s", got, want)
		}
	}

	runs := statementsRunBy(t, "idps.go", "DeleteByID")

	want := []statementRun{{on: "tx", statement: "queryLockIDPForDelete"}, {on: "tx", statement: "queryLockAccountsOfIDP"}, {on: "tx", statement: "queryDeleteIDP"}}
	if !slices.Equal(runs, want) {
		t.Errorf("DeleteByID runs %v, want the provider's lock, the accounts' and then the delete, all in its transaction: %v", runs, want)
	}
}

// Two writes that each remove one of an account's ways in, at the same
// moment, both passed: the unlink checked in the use-case, with two reads,
// and a provider's delete read the account's other identities holding only
// its own provider's row. The unlink takes the account's row and then
// deletes with the rule in its statement, all in one transaction; the
// provider's delete takes the same rows (above). Read from Unlink's body.
func TestAnUnlinkTakesTheAccountBeforeItsRule(t *testing.T) {
	t.Parallel()

	for statement, want := range map[string]string{
		queryLockAccountForUnlink: "SELECT TRUE FROM users WHERE id = $1 FOR NO KEY UPDATE;",
		queryUnlinkIdentity: "DELETE FROM users_identities WHERE users_id = $1 AND idps_id = $2 AND EXISTS ( SELECT 1 FROM users WHERE users.id = $1 AND ( users.local_account IS TRUE " +
			"OR EXISTS ( SELECT 1 FROM users_identities AS other JOIN idps ON idps.id = other.idps_id WHERE other.users_id = $1 AND other.idps_id <> $2 AND idps.enabled ) ) );",
	} {
		if got := strings.Join(strings.Fields(statement), " "); got != want {
			t.Errorf("the unlink's statement changed:\n got: %s\nwant: %s", got, want)
		}
	}

	runs := statementsRunBy(t, "users_identities.go", "Unlink")

	want := []statementRun{{on: "tx", statement: "queryLockAccountForUnlink"}, {on: "tx", statement: "queryUnlinkIdentity"}, {on: "tx", statement: "queryIdentityExists"}}
	if !slices.Equal(runs, want) {
		t.Errorf("Unlink runs %v, want the account's lock, then the delete and the read after it, all in its transaction: %v", runs, want)
	}
}

// statementRun is a call that runs a statement: the receiver it is called on
// and the statement it is handed.
type statementRun struct{ on, statement string }

// statementsRunBy is every statement a method in file runs, in the order its
// body makes the calls.
func statementsRunBy(t *testing.T, path, method string) []statementRun {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	var body *ast.BlockStmt

	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == method && fn.Recv != nil {
			body = fn.Body
		}
	}

	if body == nil {
		t.Fatalf("%s is not in %s", method, path)
	}

	var runs []statementRun

	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (selector.Sel.Name != "Exec" && selector.Sel.Name != "QueryRow" && selector.Sel.Name != "Query") {
			return true
		}

		if statement, ok := call.Args[1].(*ast.Ident); ok {
			on := ""
			if x, ok := selector.X.(*ast.Ident); ok {
				on = x.Name
			} else if x, ok := selector.X.(*ast.SelectorExpr); ok {
				on = x.Sel.Name
			}

			runs = append(runs, statementRun{on: on, statement: statement.Name})
		}

		return true
	})

	return runs
}
