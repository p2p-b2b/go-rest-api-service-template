package repositorypg

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

// fakeTx records how a transaction was ended. The embedded nil pgx.Tx
// panics on any other method, so endTx touching one fails the test.
type fakeTx struct {
	pgx.Tx
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (f *fakeTx) Commit(context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeTx) Rollback(context.Context) error {
	f.rolledBack = true
	return f.rollbackErr
}

func TestEndTx(t *testing.T) {
	errCommit := errors.New("commit failed")
	errMethod := errors.New("insert failed")
	recorded := func(e error) error { return errors.Join(errors.New("recorded"), e) }

	tests := []struct {
		name         string
		methodErr    error
		commitErr    error
		rollbackErr  error
		wantCommit   bool
		wantRollback bool
		wantRecorded bool
		wantErr      error
	}{
		{name: "success_commits", wantCommit: true},
		{name: "failed_commit_is_returned", commitErr: errCommit, wantCommit: true, wantRecorded: true, wantErr: errCommit},
		{name: "method_error_rolls_back", methodErr: errMethod, wantRollback: true, wantErr: errMethod},
		{name: "closed_tx_rollback_is_quiet", methodErr: errMethod, rollbackErr: pgx.ErrTxClosed, wantRollback: true, wantErr: errMethod},
		{name: "failed_rollback_keeps_method_error", methodErr: errMethod, rollbackErr: errors.New("rollback failed"), wantRollback: true, wantErr: errMethod},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &fakeTx{commitErr: tt.commitErr, rollbackErr: tt.rollbackErr}
			err := tt.methodErr

			endTx(t.Context(), tx, &err, recorded)

			assert.Equal(t, tt.wantCommit, tx.committed, "committed")
			assert.Equal(t, tt.wantRollback, tx.rolledBack, "rolled back")
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantRecorded, strings.HasPrefix(err.Error(), "recorded"), "passed through record")
		})
	}
}

// TestEndTxReachesTheCaller pins the reason endTx takes a pointer: the
// failed COMMIT must replace the nil the method returned.
func TestEndTxReachesTheCaller(t *testing.T) {
	errCommit := errors.New("commit failed")
	tx := &fakeTx{commitErr: errCommit}

	insert := func() (err error) {
		defer endTx(t.Context(), tx, &err, func(e error) error { return e })
		return nil
	}

	assert.ErrorIs(t, insert(), errCommit)
}

// TestNoTransactionIsEndedInADeferredClosure: a transaction is ended by
// endTx, which reads the error the caller receives. A deferred closure that
// commits "when err is nil" reads a local instead, and commits on every
// error path that returns without setting it.
//
// UsersRepository.Insert and ProjectsRepository.Insert did, and logged a
// failed COMMIT and answered nil: a 201 for a row that was never stored.
// Three shapes fail here, anywhere in the package:
//
//   - a Commit written inside a deferred function literal;
//   - `defer tx.Commit(ctx)`, which commits whatever the method returns;
//   - endTx given a pointer to anything but the method's named result, which
//     is the same mistake behind the helper.
//
// A deferred Rollback is not among them: after a commit it does nothing, and
// it decides nothing. What this cannot see: a closure stored in a variable
// and deferred, and a deferred helper that commits. Neither exists today.
func TestNoTransactionIsEndedInADeferredClosure(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	fset := token.NewFileSet()

	var files, endings int

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}

		files++

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			results := map[string]bool{}

			if fn.Type.Results != nil {
				for _, field := range fn.Type.Results.List {
					for _, ident := range field.Names {
						results[ident.Name] = true
					}
				}
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.DeferStmt:
					if sel, ok := node.Call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Commit" {
						t.Errorf("%s: a deferred Commit commits whatever the method returns; end the transaction with endTx", fset.Position(node.Pos()))
					}

					closure, ok := node.Call.Fun.(*ast.FuncLit)
					if !ok {
						return true
					}

					ast.Inspect(closure.Body, func(inner ast.Node) bool {
						call, ok := inner.(*ast.CallExpr)
						if !ok {
							return true
						}

						if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Commit" {
							t.Errorf("%s: a deferred closure commits; end the transaction with endTx, which reads the returned error", fset.Position(call.Pos()))
						}

						return true
					})
				case *ast.CallExpr:
					callee, ok := node.Fun.(*ast.Ident)
					if !ok || callee.Name != "endTx" || len(node.Args) < 3 {
						return true
					}

					endings++

					pointer, ok := node.Args[2].(*ast.UnaryExpr)
					target, isIdent := ast.Expr(nil), false

					if ok && pointer.Op == token.AND {
						target = pointer.X
						_, isIdent = target.(*ast.Ident)
					}

					if !isIdent || !results[target.(*ast.Ident).Name] {
						t.Errorf("%s: endTx is given something other than a pointer to %s's named result; it would decide on a local",
							fset.Position(node.Pos()), fn.Name.Name)
					}
				}

				return true
			})
		}
	}

	// The package has eighteen files and four methods that end a
	// transaction with endTx; fewer means the walk no longer finds them and
	// the test would pass on nothing.
	if files < 15 || endings < 4 {
		t.Errorf("read %d files and %d endTx calls, want the repository package", files, endings)
	}
}
