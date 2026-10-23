//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAProvidersDeleteWaitsForASignInThroughIt: an identity's insert (a
// first sign-in's, a link's) holds a KEY SHARE lock on the provider's row
// until it commits. A delete of the provider sent meanwhile read its
// condition (no account whose only way in is an identity there) before the
// insert committed, waited for the lock, and went ahead after it without
// reading the condition again: its cascade removed the identity just
// written, and the account had no way in. The delete takes the row FOR
// UPDATE first now, so it waits for the sign-in and then sees its identity:
// a 409, and the account keeps its way in.
//
// The sign-in is played by a transaction of the test's own, so that the
// delete can be seen waiting on its lock before it commits.
func TestAProvidersDeleteWaitsForASignInThroughIt(t *testing.T) {
	ctx := t.Context()

	administrator := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, administrator.UserID) })

	idp := createOIDCIDP(t, administrator.AccessToken, newFakeProvider(t), true)

	firstName, lastName, email := generateUserData(t)
	account := uuid.NewV7()

	// The delete's request runs in a goroutine that logs through t. Its
	// answer is waited for after the sign-in's transaction is ended, by a
	// cleanup registered first and so run last: on a failure the rollback
	// releases the lock, the delete answers, and the goroutine is done
	// before the test is.
	type answer struct {
		err    error
		text   string
		status int
	}

	deleted := make(chan answer, 1)
	started := false

	t.Cleanup(func() {
		if !started {
			return
		}

		select {
		case <-deleted:
		case <-time.After(15 * time.Second):
		}
	})

	signIn, err := testDBPool.Begin(ctx)
	require.NoError(t, err)

	t.Cleanup(func() { _ = signIn.Rollback(context.Background()) })

	var signInPID int
	require.NoError(t, signIn.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&signInPID))

	_, err = signIn.Exec(ctx, `
        INSERT INTO users (id, first_name, last_name, email, password_hash, disabled, verified, local_account)
        VALUES ($1, $2, $3, $4, 'no password: a federated account', FALSE, TRUE, FALSE)`,
		account, firstName, lastName, email)
	require.NoError(t, err)

	_, err = signIn.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
		account, idp, "sub-"+account.String(), email)
	require.NoError(t, err)

	started = true

	go func() {
		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), idpsDeleteEndpoint.Clone().RewriteSlugs(idp), nil,
			map[string]string{"Authorization": "Bearer " + administrator.AccessToken})
		if err != nil {
			deleted <- answer{err: err}

			return
		}

		defer response.Body.Close()

		deleted <- answer{status: response.StatusCode, text: readResponseBody(t, response)}
	}()

	// The delete is waiting on the sign-in: a backend blocked by this
	// transaction's, on a statement that reads the provider's row.
	require.Eventually(t, func() bool {
		var waiting bool

		err := testDBPool.QueryRow(context.Background(), `
            SELECT EXISTS (
                SELECT 1 FROM pg_stat_activity
                WHERE $1 = ANY (pg_blocking_pids(pid)) AND query ILIKE '%FROM idps%'
            )`, signInPID).Scan(&waiting)

		return err == nil && waiting
	}, 10*time.Second, 50*time.Millisecond, "the delete never waited on the sign-in's lock")

	require.NoError(t, signIn.Commit(ctx))

	t.Cleanup(func() { deleteUserByIDFromDB(t, account) })

	var got answer

	select {
	case got = <-deleted:
		started = false
	case <-time.After(15 * time.Second):
		t.Fatal("the delete did not answer after the sign-in committed")
	}

	require.NoError(t, got.err)
	assert.Equal(t, http.StatusConflict, got.status, "the delete went ahead under a sign-in it had waited for: %s", got.text)

	var identities int
	require.NoError(t, testDBPool.QueryRow(context.Background(),
		`SELECT count(*) FROM users_identities WHERE users_id = $1`, account).Scan(&identities))
	assert.Equal(t, 1, identities, "the account that signed in has no way in")
}
