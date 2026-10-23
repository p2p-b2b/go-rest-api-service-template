package repositorypg

import (
	"context"
	"errors"
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
