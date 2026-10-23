package repositorypg

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// endTx ends tx according to the error the method is about to return, and
// is deferred right after Begin with errp pointing at the method's named
// error result:
//
//	func (ref *R) DeleteByID(ctx context.Context, in *domain.X) (err error) {
//		tx, err := ref.db.Begin(ctx)
//		...
//		defer endTx(ctx, tx, &err, record)
//
// A method returning an error rolls back. A method returning nil commits,
// and a failed COMMIT becomes the method's result, passed through record so
// the span and the call metrics see it. The decision reads the value the
// caller receives, never a local that each error path has to remember to
// set: the deferred closure UsersRepository.Insert and
// ProjectsRepository.Insert still end theirs with commits whenever such a
// local is nil, so an error path that returned without setting it would
// commit a partial write, and a failed COMMIT is logged and answered nil --
// a 201 for a row that was never stored.
//
// A method that panics between Begin and its return leaves the named result
// nil, and endTx, deferred, commits what it wrote: a method that can panic
// halfway through a write does not use it.
//
// A rollback that fails because the context is gone (the call ran past its
// deadline, or the request was cancelled) is not logged: the failure the
// method returns is that one, and the request's access line says it.
// pgx closes the connection, which ends the transaction.
func endTx(ctx context.Context, tx pgx.Tx, errp *error, record func(error) error) {
	if *errp != nil {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) && ctx.Err() == nil {
			slog.ErrorContext(ctx, "rollback failed", "error", err)
		}

		return
	}

	if err := tx.Commit(ctx); err != nil {
		*errp = record(err)
	}
}
