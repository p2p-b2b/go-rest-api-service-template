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

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driven/repositorypg"
)

// TestAnAdministratorAccountIsNotAUserManagersToLock: an account with the
// `admin` flag is what a deployment is operated from, and the seeded
// UserManager role manages accounts. With only that role an account disabled
// the seeded administrator (200), which locks the one account with the flag
// out until somebody writes to the database. The same grants make a local
// account a federated one, delete an account, and take its roles away.
//
// The rule: an account with the flag is disabled, made federated,
// stripped of a role or deleted only by an account with the flag, and the
// last that can still sign in by nobody. Every other account stays the
// role's to manage.
//
// "The last one" is not tested here: the suite's tests create accounts with
// the flag all the time, so there is never a last one to find. The unit
// tests hold it with a counted double.
//
// Not parallel: it reads the rows it changes.
func TestAnAdministratorAccountIsNotAUserManagersToLock(t *testing.T) {
	ctx := t.Context()

	// An account, its address, and the headers it signs in with.
	account := func(role string, flag bool) (uuid.UUID, string, map[string]string) {
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

		return id, email, map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}
	}

	_, _, manager := account("UserManager", false)
	_, _, roleOnly := account("Administrator", false)

	// An account with the flag that is not the seeded one, with a role to
	// take away; an ordinary one, the same; and a caller with the flag.
	flagged, flaggedEmail, _ := account("ReadOnly", true)
	ordinary, _, _ := account("ReadOnly", false)
	_, _, administrator := account("Administrator", true)

	var readOnly uuid.UUID

	require.NoError(t, testDBPool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'ReadOnly'`).Scan(&readOnly))

	send := func(t *testing.T, endpoint *apiEndpoint, headers map[string]string, body map[string]any, slugs ...string) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), endpoint.RewriteSlugs(slugs...), body, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, text
	}

	state := func(t *testing.T, user uuid.UUID) (exists, disabled, local, hasRole bool) {
		t.Helper()

		err := testDBPool.QueryRow(context.Background(), `
            SELECT disabled, local_account,
                   EXISTS (SELECT 1 FROM users_roles WHERE users_id = users.id AND roles_id = $2)
            FROM users WHERE id = $1`, user, readOnly).Scan(&disabled, &local, &hasRole)

		return err == nil, disabled, local, hasRole
	}

	const administratorAccount = "the administrator account"

	t.Run("a_user_manager_does_not_disable_it", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"disabled": true}, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, administratorAccount)

		_, disabled, _, _ := state(t, flagged)
		assert.False(t, disabled, "an account with the flag was disabled by an account without it")
	})

	// The body the frontend's form sends: the names, the address it read,
	// and `disabled`, together. The address is no change; the lock is one.
	t.Run("nor_with_the_whole_form", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager,
			map[string]any{"first_name": "Whole", "last_name": "Form", "email": flaggedEmail, "disabled": true}, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, body)

		_, disabled, _, _ := state(t, flagged)
		assert.False(t, disabled, "an account with the flag was disabled by an account without it, through a body that also carried its address")
	})

	t.Run("nor_make_it_a_federated_account", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"local_account": false}, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, body)

		_, _, local, _ := state(t, flagged)
		assert.True(t, local, "an account with the flag lost its password login to an account without the flag")
	})

	t.Run("it_renames_it_sending_the_rest_back_as_it_is", func(t *testing.T) {
		// What a client that sends the whole form does.
		status, body := send(t, usersUpdateEndpoint, manager,
			map[string]any{"first_name": "Renamed", "email": flaggedEmail, "disabled": false, "local_account": true}, flagged.String())
		require.Equal(t, http.StatusOK, status, body)
	})

	t.Run("nor_delete_it", func(t *testing.T) {
		status, body := send(t, usersDeleteEndpoint, manager, nil, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, administratorAccount)

		exists, _, _, _ := state(t, flagged)
		assert.True(t, exists, "an account with the flag was deleted by an account without it")
	})

	t.Run("nor_take_a_role_from_it", func(t *testing.T) {
		status, body := send(t, usersUnlinkRolesEndpoint, manager, map[string]any{"role_ids": []uuid.UUID{readOnly}}, flagged.String())
		assert.Equal(t, http.StatusForbidden, status, body)

		_, _, _, hasRole := state(t, flagged)
		assert.True(t, hasRole, "an account with the flag was stripped of a role by an account without the flag")
	})

	t.Run("nor_through_the_role", func(t *testing.T) {
		// The seeded UserManager does not hold this route; the Administrator
		// role does, and the role is not the flag.
		status, body := send(t, rolesUnlinkUsersEndpoint, roleOnly, map[string]any{"user_ids": []uuid.UUID{ordinary, flagged}}, readOnly.String())
		assert.Equal(t, http.StatusForbidden, status, body)

		_, _, _, hasRole := state(t, flagged)
		assert.True(t, hasRole)

		_, _, _, ordinaryHasRole := state(t, ordinary)
		assert.True(t, ordinaryHasRole, "part of a refused body was done")
	})

	t.Run("every_other_account_stays_the_roles_to_manage", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"disabled": true}, ordinary.String())
		require.Equal(t, http.StatusOK, status, body)

		status, body = send(t, usersUnlinkRolesEndpoint, manager, map[string]any{"role_ids": []uuid.UUID{readOnly}}, ordinary.String())
		require.Equal(t, http.StatusOK, status, body)

		status, body = send(t, usersDeleteEndpoint, manager, nil, ordinary.String())
		require.Equal(t, http.StatusOK, status, body)

		exists, _, _, _ := state(t, ordinary)
		assert.False(t, exists)
	})

	t.Run("an_account_with_the_flag_does_all_of_it", func(t *testing.T) {
		status, body := send(t, usersUpdateEndpoint, administrator, map[string]any{"disabled": true}, flagged.String())
		require.Equal(t, http.StatusOK, status, body)

		status, body = send(t, usersUpdateEndpoint, administrator, map[string]any{"disabled": false}, flagged.String())
		require.Equal(t, http.StatusOK, status, body)

		status, body = send(t, usersUnlinkRolesEndpoint, administrator, map[string]any{"role_ids": []uuid.UUID{readOnly}}, flagged.String())
		require.Equal(t, http.StatusOK, status, body)

		_, _, _, hasRole := state(t, flagged)
		assert.False(t, hasRole)
	})

	// The other direction: what an administrator locked is unlocked by
	// an administrator. Enabling an account an administrator disabled, and
	// making one an administrator federated a local account again (its
	// stored password and its recovery work again), were still the role's.
	t.Run("nor_unlock_one_an_administrator_locked", func(t *testing.T) {
		locked, lockedEmail, _ := account("ReadOnly", true)

		_, err := testDBPool.Exec(ctx, `UPDATE users SET disabled = TRUE, local_account = FALSE WHERE id = $1`, locked)
		require.NoError(t, err)

		status, body := send(t, usersUpdateEndpoint, manager, map[string]any{"disabled": false}, locked.String())
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.Contains(t, body, administratorAccount)

		status, body = send(t, usersUpdateEndpoint, manager, map[string]any{"local_account": true}, locked.String())
		assert.Equal(t, http.StatusForbidden, status, body)

		// With the whole form, as the frontend sends it.
		status, body = send(t, usersUpdateEndpoint, manager,
			map[string]any{"first_name": "Whole", "last_name": "Form", "email": lockedEmail, "disabled": false}, locked.String())
		assert.Equal(t, http.StatusForbidden, status, body)

		_, disabled, local, _ := state(t, locked)
		assert.True(t, disabled, "an account with the flag that was disabled was enabled by an account without the flag")
		assert.False(t, local, "an account with the flag that was federated got its password login back from an account without the flag")

		// The values sent back as they are, with a new name: no change.
		status, body = send(t, usersUpdateEndpoint, manager,
			map[string]any{"first_name": "Renamed", "disabled": true, "local_account": false}, locked.String())
		require.Equal(t, http.StatusOK, status, body)

		// An account with the flag unlocks it.
		status, body = send(t, usersUpdateEndpoint, administrator, map[string]any{"disabled": false, "local_account": true}, locked.String())
		require.Equal(t, http.StatusOK, status, body)

		_, disabled, local, _ = state(t, locked)
		assert.False(t, disabled)
		assert.True(t, local)
	})

	// The verification of an address enables the account it verifies, and
	// names no caller. An account an administrator created is never
	// verified, is in use all the same, and whoever asks for a verification
	// link for its address gets one mailed there. With the flag and
	// disabled by an administrator, it enabled itself through that link.
	t.Run("nor_does_its_own_verification_link_unlock_it", func(t *testing.T) {
		locked, lockedEmail, _ := account("ReadOnly", true)

		_, err := testDBPool.Exec(ctx, `UPDATE users SET disabled = TRUE, verified = FALSE WHERE id = $1`, locked)
		require.NoError(t, err)

		// The public route: anybody may ask for the link.
		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), authReVerifyEndpoint, map[string]any{"email": lockedEmail})
		require.NoError(t, err)
		response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)

		require.Positive(t, mailsTo(t, lockedEmail, 10*time.Second), "no verification mail arrived: this case proves nothing")

		confirmed, err := confirmVerification(t, verificationTokenFromEmail(t, verifyEmailAddress, lockedEmail))
		require.NoError(t, err)
		confirmed.Body.Close()
		require.Equal(t, http.StatusOK, confirmed.StatusCode, "the link was refused: this case proves nothing")

		var verified bool

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT verified FROM users WHERE id = $1`, locked).Scan(&verified))
		assert.True(t, verified, "the link proves the address")

		_, disabled, _, _ := state(t, locked)
		assert.True(t, disabled, "an account with the flag that an administrator disabled enabled itself through its verification link")
	})

	// The seeded administrator is deleted by nobody: its link to the
	// Administrator role is guarded by a trigger, and a delete takes that
	// link with it. That refusal reached the caller as a 500.
	t.Run("the_seeded_administrator_is_deleted_by_nobody_and_that_is_a_403", func(t *testing.T) {
		for who, headers := range map[string]map[string]string{"a user manager": manager, "an account with the flag": administrator} {
			status, body := send(t, usersDeleteEndpoint, headers, nil, seededAdministrator.String())
			assert.Equal(t, http.StatusForbidden, status, "%s: %s", who, body)
		}

		exists, _, _, _ := state(t, seededAdministrator)
		require.True(t, exists, "the seeded administrator is gone")
	})

	// Nor is it stripped of the Administrator role, from either side: the
	// same trigger refuses, and an account with the flag, which the guard
	// lets through, was answered that refusal as a 500.
	t.Run("the_seeded_administrator_keeps_its_role_and_that_is_a_403", func(t *testing.T) {
		for route, call := range map[string]func() (int, string){
			"from_the_users_side": func() (int, string) {
				return send(t, usersUnlinkRolesEndpoint, administrator, map[string]any{"role_ids": []string{administratorRoleID}}, seededAdministrator.String())
			},
			"from_the_roles_side": func() (int, string) {
				return send(t, rolesUnlinkUsersEndpoint, administrator, map[string]any{"user_ids": []uuid.UUID{seededAdministrator}}, administratorRoleID)
			},
		} {
			status, body := call()
			assert.Equal(t, http.StatusForbidden, status, "%s: %s", route, body)
			assert.Contains(t, body, "the bootstrap administrator", route)
		}

		var linked bool
		require.NoError(t, testDBPool.QueryRow(t.Context(), `SELECT EXISTS (SELECT 1 FROM users_roles WHERE users_id = $1 AND roles_id = $2)`, seededAdministrator, administratorRoleID).Scan(&linked))
		require.True(t, linked, "the seeded administrator lost the Administrator role")
	})
}

// TestWhichAccountsAreAdministratorsAndCanSignIn reads the two statements the
// guard of the administrator account asks: which of a set of accounts have
// the `admin` flag, and how many accounts with the flag can still sign in.
// "Can sign in" is not "is not disabled": an account that is no longer a
// local one signs in through a provider, and one with no identity at a
// provider has no way in.
//
// Not parallel, and that is the point: the suite's parallel tests create
// accounts with the flag all the time, and this counts them.
func TestWhichAccountsAreAdministratorsAndCanSignIn(t *testing.T) {
	ctx := t.Context()

	users, err := repositorypg.NewUsersRepository(repositorypg.UsersRepositoryConfig{
		DB:              testDBPool,
		OT:              newTestOpenTelemetry(t),
		MaxPingTimeout:  5 * time.Second,
		MaxQueryTimeout: 10 * time.Second,
	})
	require.NoError(t, err)

	// One account of each kind, written as the columns are.
	create := func(admin *bool, disabled, local bool) uuid.UUID {
		firstName, lastName, email := generateUserData(t)
		id := createUserInDB(t, firstName, lastName, email, generatePassword(t))

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = $2, disabled = $3, local_account = $4 WHERE id = $1`, id, admin, disabled, local)
		require.NoError(t, err)

		return id
	}

	before, err := users.CountEnabledAdministrators(ctx)
	require.NoError(t, err)

	enabled := create(new(true), false, true)

	afterOne, err := users.CountEnabledAdministrators(ctx)
	require.NoError(t, err)
	require.Equal(t, before+1, afterOne, "an enabled local account with the flag is one that can sign in")

	disabled := create(new(true), true, true)
	stranded := create(new(true), false, false) // federated, with no identity at any provider
	ordinary := create(new(false), false, true)
	unflagged := create(nil, false, true)

	stillOne, err := users.CountEnabledAdministrators(ctx)
	require.NoError(t, err)
	assert.Equal(t, afterOne, stillOne,
		"a disabled account, a federated one with no identity, an account without the flag and one whose flag was never set are not administrators that can sign in")

	// A federated account that has an identity can.
	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	withIdentity, err := users.CountEnabledAdministrators(ctx)
	require.NoError(t, err)
	require.Equal(t, stillOne+1, withIdentity, "the suite's own administrator account is one more")

	idpID := createOIDCIDP(t, admin.AccessToken, newFakeProvider(t), false)

	_, err = testDBPool.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, 'sso@example.com')`,
		stranded, idpID, "sub-"+stranded.String())
	require.NoError(t, err)

	federated, err := users.CountEnabledAdministrators(ctx)
	require.NoError(t, err)
	assert.Equal(t, withIdentity+1, federated, "a federated account with the flag and an identity can sign in")

	// Each account is asked the same question the count answers.
	for id, want := range map[uuid.UUID]bool{enabled: true, stranded: true, disabled: false, ordinary: false, unflagged: false, uuid.NewV7(): false} {
		among, err := users.IsEnabledAdministrator(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want, among, "is %s an administrator that can sign in", id)
	}

	// An identity at a provider that is disabled is no way in: the provider
	// refuses the sign-in.
	_, err = testDBPool.Exec(ctx, `UPDATE idps SET enabled = FALSE WHERE id = $1`, idpID)
	require.NoError(t, err)

	providerOff, err := users.CountEnabledAdministrators(ctx)
	require.NoError(t, err)
	assert.Equal(t, withIdentity, providerOff, "an identity at a disabled provider is not a way in")

	among, err := users.IsEnabledAdministrator(ctx, stranded)
	require.NoError(t, err)
	assert.False(t, among, "the account whose only identity is at a disabled provider cannot sign in")

	// And which of a set have the flag.
	found, err := users.SelectAdministratorsAmong(ctx, []uuid.UUID{ordinary, enabled, unflagged, disabled, uuid.NewV7(), stranded})
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{enabled, disabled, stranded}, found,
		"the flag, whatever else the account is; not an account without it, one that never had it set, or one that is not there")

	none, err := users.SelectAdministratorsAmong(ctx, []uuid.UUID{ordinary, unflagged})
	require.NoError(t, err)
	assert.Empty(t, none)
}
