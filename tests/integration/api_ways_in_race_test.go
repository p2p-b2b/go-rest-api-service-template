//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnAccountsWaysInAreRemovedOneAtATime: two writes that each remove one
// of an account's ways in, at the same moment, both passed. A provider's
// delete read the account's other identities holding only its own
// provider's row, and an unlink checked with two reads in the use-case: an
// account with identities at two providers and no password, whose two ways
// in were removed together, had none. Both take the account's row now, so
// the second waits for the first and reads what it left: a 409, and the
// account keeps its last way in.
//
// Each case plays one of the two writes in a transaction of the test's own,
// as the service does it, and sends the other through the API.
func TestAnAccountsWaysInAreRemovedOneAtATime(t *testing.T) {
	administrator := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, administrator.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + administrator.AccessToken}

	// An account with identities at two providers, and no password once it
	// has signed in with the one it had.
	account := func(t *testing.T) (id uuid.UUID, token map[string]string, first, second string) {
		t.Helper()

		ctx := context.Background()
		first = createOIDCIDP(t, administrator.AccessToken, newFakeProvider(t), false)
		second = createOIDCIDP(t, administrator.AccessToken, newFakeProvider(t), false)

		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id = createUserInDB(t, firstName, lastName, email, password)
		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })
		enableUserByEmailFromDB(t, email)
		assignRoleToUserInDB(t, id, "AuthenticatedUser")

		token = map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}

		for _, idp := range []string{first, second} {
			_, err := testDBPool.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
				id, idp, "sub-"+uuid.NewV7().String(), email)
			require.NoError(t, err)
		}

		_, err := testDBPool.Exec(ctx, `UPDATE users SET local_account = FALSE WHERE id = $1`, id)
		require.NoError(t, err)

		return id, token, first, second
	}

	// race holds the test's write open, sends the other through the API,
	// commits once the request waits on the test's transaction or has
	// answered, and returns its answer.
	race := func(t *testing.T, write func(pgx.Tx), endpoint *apiEndpoint, headers map[string]string) (int, string) {
		t.Helper()

		type answer struct {
			err    error
			text   string
			status int
		}

		answered := make(chan answer, 1)
		sent := false

		// Run last: the request is done before the test is, whatever failed.
		t.Cleanup(func() {
			if sent {
				select {
				case <-answered:
				case <-time.After(15 * time.Second):
				}
			}
		})

		tx, err := testDBPool.Begin(context.Background())
		require.NoError(t, err)

		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })

		var pid int
		require.NoError(t, tx.QueryRow(context.Background(), `SELECT pg_backend_pid()`).Scan(&pid))

		write(tx)

		sent = true

		go func() {
			response, err := sendHTTPRequest(t, context.Background(), endpoint, nil, headers)
			if err != nil {
				answered <- answer{err: err}

				return
			}

			defer response.Body.Close()

			answered <- answer{status: response.StatusCode, text: readResponseBody(t, response)}
		}()

		var got *answer

		require.Eventually(t, func() bool {
			select {
			case a := <-answered:
				got = &a

				return true
			default:
			}

			var waiting bool

			err := testDBPool.QueryRow(context.Background(),
				`SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1 = ANY (pg_blocking_pids(pid)))`, pid).Scan(&waiting)

			return err == nil && waiting
		}, 10*time.Second, 50*time.Millisecond, "the request neither waited on the test's write nor answered")

		require.NoError(t, tx.Commit(context.Background()))

		if got == nil {
			select {
			case a := <-answered:
				got = &a
			case <-time.After(15 * time.Second):
				t.Fatal("the request did not answer after the test's write committed")
			}
		}

		sent = false

		require.NoError(t, got.err)

		return got.status, got.text
	}

	identities := func(t *testing.T, id uuid.UUID) int {
		t.Helper()

		var n int
		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT count(*) FROM users_identities WHERE users_id = $1`, id).Scan(&n))

		return n
	}

	// The locks the service takes for the delete of a provider and for an
	// unlink, in its order, and the write each then makes; the test's own
	// write leaves out the rule its statement holds, which passes here: the
	// account keeps a way in either way.
	deleteProvider := func(t *testing.T, idp string) func(pgx.Tx) {
		return func(tx pgx.Tx) {
			_, err := tx.Exec(context.Background(), `SELECT TRUE FROM idps WHERE id = $1 FOR UPDATE`, idp)
			require.NoError(t, err)

			_, err = tx.Exec(context.Background(), `
                SELECT id FROM users
                WHERE id IN (SELECT users_id FROM users_identities WHERE idps_id = $1)
                ORDER BY id FOR NO KEY UPDATE`, idp)
			require.NoError(t, err)

			_, err = tx.Exec(context.Background(), `DELETE FROM idps WHERE id = $1`, idp)
			require.NoError(t, err)
		}
	}

	unlink := func(t *testing.T, id uuid.UUID, idp string) func(pgx.Tx) {
		return func(tx pgx.Tx) {
			_, err := tx.Exec(context.Background(), `SELECT TRUE FROM users WHERE id = $1 FOR NO KEY UPDATE`, id)
			require.NoError(t, err)

			_, err = tx.Exec(context.Background(), `DELETE FROM users_identities WHERE users_id = $1 AND idps_id = $2`, id, idp)
			require.NoError(t, err)
		}
	}

	t.Run("a_providers_delete_waits_for_an_unlink", func(t *testing.T) {
		id, _, first, second := account(t)

		status, text := race(t, unlink(t, id, second), idpsDeleteEndpoint.Clone().RewriteSlugs(first), admin)

		assert.Equal(t, http.StatusConflict, status, "the delete took the account's last way in: %s", text)
		assert.Equal(t, 1, identities(t, id), "the account has no way in")
	})

	t.Run("a_providers_delete_waits_for_another_providers", func(t *testing.T) {
		id, _, first, second := account(t)

		status, text := race(t, deleteProvider(t, second), idpsDeleteEndpoint.Clone().RewriteSlugs(first), admin)

		assert.Equal(t, http.StatusConflict, status, "the delete took the account's last way in: %s", text)
		assert.Equal(t, 1, identities(t, id), "the account has no way in")
	})

	t.Run("an_unlink_waits_for_another_unlink", func(t *testing.T) {
		id, token, first, second := account(t)

		status, text := race(t, unlink(t, id, second), meIdentityUnlinkPoint.Clone().RewriteSlugs(first), token)

		assert.Equal(t, http.StatusConflict, status, "the unlink took the account's last way in: %s", text)
		assert.Equal(t, 1, identities(t, id), "the account has no way in")
	})

	t.Run("an_unlink_waits_for_a_providers_delete", func(t *testing.T) {
		id, token, first, second := account(t)

		status, text := race(t, deleteProvider(t, second), meIdentityUnlinkPoint.Clone().RewriteSlugs(first), token)

		assert.Equal(t, http.StatusConflict, status, "the unlink took the account's last way in: %s", text)
		assert.Equal(t, 1, identities(t, id), "the account has no way in")
	})
}
