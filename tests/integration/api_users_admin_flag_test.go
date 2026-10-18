//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheAdminFlagIsGrantedThroughTheAPI: the `admin` flag is the
// administrator account. It has every project, and it alone changes an
// account's address and another administrator account's sign-in. No route set
// it, so the account the migrations seed was the only one a deployment could
// have: one mailbox, and when it was lost, the database.
//
// The rule: PUT /users/{user_id} accepts `admin`, from an account that
// has the flag and from nobody else. The route's grant is the seeded
// UserManager's too, and the Administrator role's; neither is the flag. The
// last account with it that can sign in keeps it (held by the unit tests:
// this suite's tests make such accounts all the time, so there is never a
// last one to find), and the seeded account keeps it whoever asks.
//
// Not parallel: it reads the rows it changes.
func TestTheAdminFlagIsGrantedThroughTheAPI(t *testing.T) {
	ctx := t.Context()

	// An account, and the headers it signs in with.
	account := func(role string, flag bool) (uuid.UUID, map[string]string) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id := createUserInDB(t, firstName, lastName, email, password)

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		enableUserByEmailFromDB(t, email)

		if flag {
			_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = TRUE, verified = TRUE, local_account = TRUE WHERE id = $1`, id)
			require.NoError(t, err)
		}

		assignRoleToUserInDB(t, id, role)

		return id, map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}
	}

	managerID, manager := account("UserManager", false)
	_, roleOnly := account("Administrator", false)
	_, administrator := account("Administrator", true)
	flagged, _ := account("ReadOnly", true)

	// The account the flag is granted to, signed in once: every request it
	// makes below is with that one token.
	newcomerID, newcomer := account("ReadOnly", false)

	// A project the newcomer is not linked to.
	theirs := createProjectInDB(t, generateRandomName(t, "theirs"), "a project the newcomer is not a member of")
	t.Cleanup(func() { deleteProjectByIDFromDB(t, theirs.ID) })

	send := func(t *testing.T, endpoint *apiEndpoint, headers map[string]string, body map[string]any, slugs ...string) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), endpoint.RewriteSlugs(slugs...), body, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, text
	}

	// The flag as the row has it: nil is a flag that was never set.
	stored := func(t *testing.T, user uuid.UUID) *bool {
		t.Helper()

		var flag *bool

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT admin FROM users WHERE id = $1`, user).Scan(&flag))

		return flag
	}

	has := func(t *testing.T, user uuid.UUID) bool {
		t.Helper()

		flag := stored(t, user)

		return flag != nil && *flag
	}

	// The flag as a read of the API answers it.
	answered := func(t *testing.T, endpoint *apiEndpoint, headers map[string]string, slugs ...string) bool {
		t.Helper()

		status, body := send(t, endpoint, headers, nil, slugs...)
		require.Equal(t, http.StatusOK, status, body)

		var got struct {
			Admin *bool `json:"admin"`
		}

		require.NoError(t, json.Unmarshal([]byte(body), &got))

		return got.Admin != nil && *got.Admin
	}

	const onlyTheAdministratorAccount = "user admin can be changed only by the administrator account"

	t.Run("a_user_manager_does_not_grant_it", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"admin": true}, newcomerID.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, onlyTheAdministratorAccount)
		assert.False(t, has(t, newcomerID), "an account without the flag granted it")
	})

	t.Run("nor_to_itself", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"admin": true}, managerID.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.False(t, has(t, managerID), "an account made itself the administrator account")
	})

	// With a rename beside it, as a form sends it: nothing of the body is
	// written.
	t.Run("nor_inside_a_whole_form", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"first_name": "Granted", "admin": true}, newcomerID.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.False(t, has(t, newcomerID))

		var firstName string

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT first_name FROM users WHERE id = $1`, newcomerID).Scan(&firstName))
		assert.NotEqual(t, "Granted", firstName, "the refused body was written in part")
	})

	t.Run("the_administrator_role_without_the_flag_does_not_grant_it", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, roleOnly, map[string]any{"admin": true}, newcomerID.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, onlyTheAdministratorAccount)
		assert.False(t, has(t, newcomerID), "the role granted the flag")
	})

	t.Run("a_user_manager_does_not_take_it", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"admin": false}, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, onlyTheAdministratorAccount)
		assert.True(t, has(t, flagged), "an account without the flag took it from one that has it")
	})

	// A form sends back what it read, and that is no change: a user manager
	// renames an ordinary account with `admin: false` in the body, and an
	// account that has the flag with `admin: true`.
	t.Run("the_value_an_account_has_is_nobodys_to_refuse", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"first_name": "Renamed", "admin": false}, newcomerID.String())
		require.Equal(t, http.StatusOK, status, body)
		assert.False(t, has(t, newcomerID))

		status, body = send(t, usersUpdateEndpoint, manager, map[string]any{"first_name": "Renamed", "admin": true}, flagged.String())
		require.Equal(t, http.StatusOK, status, body)
		assert.True(t, has(t, flagged))

		// A flag that was never set is not set, and "false" for it is no
		// change either: the row is left as it is.
		_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = NULL WHERE id = $1`, newcomerID)
		require.NoError(t, err)

		status, body = send(t, usersUpdateEndpoint, manager, map[string]any{"first_name": "Renamed", "admin": false}, newcomerID.String())
		require.Equal(t, http.StatusOK, status, body)
		assert.Nil(t, stored(t, newcomerID), "an unchanged flag was written")

		_, err = testDBPool.Exec(ctx, `UPDATE users SET admin = FALSE WHERE id = $1`, newcomerID)
		require.NoError(t, err)
	})

	// The flag is every project, read from the row at each request: granted,
	// it opens them to the token the account already holds, and taken back,
	// it shuts them to that token. The account's cached copy is dropped, so
	// that /me says what the row says.
	t.Run("an_administrator_grants_it_and_takes_it_back", func(t *testing.T) {
		status, body := send(t, projectsGetEndpoint, newcomer, nil, theirs.ID.String())
		require.Equal(t, http.StatusNotFound, status, "the newcomer already reads the project: this case proves nothing. %s", body)
		require.False(t, answered(t, meGetEndpoint, newcomer), "and /me is read, so that a copy of the account is cached")

		status, body = send(t, usersUpdateEndpoint, administrator, map[string]any{"admin": true}, newcomerID.String())
		require.Equal(t, http.StatusOK, status, body)
		require.True(t, has(t, newcomerID), "the flag was not written")

		status, body = send(t, projectsGetEndpoint, newcomer, nil, theirs.ID.String())
		assert.Equal(t, http.StatusOK, status, "an account with the flag does not read a project: %s", body)
		assert.True(t, answered(t, meGetEndpoint, newcomer), "/me still answers the copy from before the grant")
		assert.True(t, answered(t, usersGetEndpoint, administrator, newcomerID.String()), "the account's read does not answer the flag")

		// The flag is not a grant of the policy engine: with ReadOnly's
		// grants the account still writes nothing.
		status, body = send(t, usersUpdateEndpoint, newcomer, map[string]any{"first_name": "Written"}, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, "the flag opened a route the account's role does not: %s", body)

		status, body = send(t, usersUpdateEndpoint, administrator, map[string]any{"admin": false}, newcomerID.String())
		require.Equal(t, http.StatusOK, status, body)
		require.False(t, has(t, newcomerID), "the flag was not taken back")

		status, body = send(t, projectsGetEndpoint, newcomer, nil, theirs.ID.String())
		assert.Equal(t, http.StatusNotFound, status, "an account whose flag was taken back still reads every project: %s", body)
		assert.False(t, answered(t, meGetEndpoint, newcomer), "/me still answers the copy from before the flag was taken back")
	})

	// An account gives its own flag up, while others can sign in, and is not
	// an administrator account at its next request.
	t.Run("an_administrator_gives_its_own_up", func(t *testing.T) {
		leavingID, leaving := account("Administrator", true)

		status, body := send(t, usersUpdateEndpoint, leaving, map[string]any{"admin": false}, leavingID.String())
		require.Equal(t, http.StatusOK, status, body)
		require.False(t, has(t, leavingID))

		status, body = send(t, usersUpdateEndpoint, leaving, map[string]any{"admin": true}, leavingID.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.False(t, has(t, leavingID), "an account took its flag back by itself")
	})

	// The seeded account keeps it, from an administrator account too. The
	// row's trigger refuses, and the answer is the guard's 403, not a 500.
	t.Run("the_seeded_account_keeps_it", func(t *testing.T) {
		require.True(t, has(t, seededAdministrator), "the seeded account has no flag: this case proves nothing")

		// The row's trigger is what this case is about: on a database whose
		// migrations were applied before it existed, the request below
		// would simply take the flag.
		var guarded bool

		require.NoError(t, testDBPool.QueryRow(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'tr_protect_bootstrap_administrator' AND NOT tgisinternal)`).Scan(&guarded))
		require.True(t, guarded, "the seeded account's trigger is not in this database: recreate it (make start-dev-env)")

		// Were the row's trigger ever missing, a statement below would take
		// the flag from the account every other test signs in as: it is
		// given back whatever this case finds.
		t.Cleanup(func() {
			_, err := testDBPool.Exec(context.Background(), `UPDATE users SET admin = TRUE WHERE id = $1`, seededAdministrator)
			assert.NoError(t, err, "the seeded account's flag was not restored")
		})

		status, body := send(t, usersUpdateEndpoint, administrator, map[string]any{"admin": false}, seededAdministrator.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, "the bootstrap administrator")
		assert.True(t, has(t, seededAdministrator), "the seeded account lost the flag")

		// Whoever writes: a statement by hand is refused by the row.
		_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = FALSE WHERE id = $1`, seededAdministrator)
		require.Error(t, err, "the table let the seeded account's flag go")
		assert.Contains(t, err.Error(), "The bootstrap administrator keeps the admin flag.")

		_, err = testDBPool.Exec(ctx, `UPDATE users SET admin = NULL WHERE id = $1`, seededAdministrator)
		require.Error(t, err, "the table let the seeded account's flag be unset")

		// And only the flag: the value it has is written like any other.
		_, err = testDBPool.Exec(ctx, `UPDATE users SET admin = TRUE WHERE id = $1`, seededAdministrator)
		require.NoError(t, err)
		assert.True(t, has(t, seededAdministrator))
	})
}
