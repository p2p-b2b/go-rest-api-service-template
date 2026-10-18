//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheAdministratorsLockIsAFieldOfItsOwn: an account an administrator
// disabled is not opened by its verification link, and the row remembers
// that in `disabled_by_admin`. It was set by a CHANGE of `disabled` alone,
// because a form sends every field back. So a registration that was never
// enabled could not be shut: it is disabled already, "disabled: true" is no
// change, and its link still opened it. And no answer of the API said which
// of the two a disabled, unverified account is.
//
// The rule: the lock is a field of its own, read
// and written. True locks an account whatever `disabled` is; false unlocks:
// enabled if the address is proven, waiting for its link if not. A value
// equal to the stored one changes nothing, so a form that sends back what it
// read is safe.
//
// Not parallel: it reads the mail of the accounts it makes.
func TestTheAdministratorsLockIsAFieldOfItsOwn(t *testing.T) {
	ctx := t.Context()

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	type row struct{ verified, disabled, locked bool }

	state := func(t *testing.T, email string) (uuid.UUID, row) {
		t.Helper()

		var (
			id uuid.UUID
			r  row
		)

		require.NoError(t, testDBPool.QueryRow(context.Background(),
			`SELECT id, verified, disabled, disabled_by_admin FROM users WHERE email = $1`, email).Scan(&id, &r.verified, &r.disabled, &r.locked))

		return id, r
	}

	// A registration through the public route, with its mail arrived:
	// unverified, disabled, and not locked.
	register := func(t *testing.T) (email, password string) {
		t.Helper()

		firstName, lastName, email := generateUserData(t)
		password = generatePassword(t)

		t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), authRegisterEndpoint, map[string]any{
			"email": email, "first_name": firstName, "last_name": lastName, "password": password,
		})
		require.NoError(t, err)

		body := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusCreated, response.StatusCode, body)
		require.Positive(t, mailsTo(t, email, 10*time.Second), "no verification mail arrived: this case proves nothing")

		_, r := state(t, email)
		require.Equal(t, row{verified: false, disabled: true, locked: false}, r, "a registration, before its link is followed")

		return email, password
	}

	update := func(t *testing.T, email string, headers map[string]string, body map[string]any) (int, string) {
		t.Helper()

		id, _ := state(t, email)

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersUpdateEndpoint.RewriteSlugs(id.String()), body, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, text
	}

	must := func(t *testing.T, email string, body map[string]any) {
		t.Helper()

		status, text := update(t, email, admin, body)
		require.Equal(t, http.StatusOK, status, text)
	}

	follow := func(t *testing.T, email string) {
		t.Helper()

		confirmed, err := confirmVerification(t, verificationTokenFromEmail(t, verifyEmailAddress, email))
		require.NoError(t, err)

		text := readResponseBody(t, confirmed)
		confirmed.Body.Close()

		require.Equal(t, http.StatusOK, confirmed.StatusCode, "the link was refused: this case proves nothing. %s", text)
	}

	signsIn := func(t *testing.T, email, password string) bool {
		t.Helper()

		response, _ := attemptLogin(t, email, password, "")
		defer response.Body.Close()

		return response.StatusCode == http.StatusOK
	}

	// What a read answers of the account.
	answered := func(t *testing.T, email string) (disabled, locked *bool) {
		t.Helper()

		id, _ := state(t, email)

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersGetEndpoint.RewriteSlugs(id.String()), nil, admin)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusOK, response.StatusCode, text)

		var user struct {
			Disabled        *bool `json:"disabled"`
			DisabledByAdmin *bool `json:"disabled_by_admin"`
		}

		require.NoError(t, json.Unmarshal([]byte(text), &user), text)

		return user.Disabled, user.DisabledByAdmin
	}

	// The case the field exists for.
	t.Run("a_registration_that_was_never_enabled_is_shut_by_the_lock", func(t *testing.T) {
		email, password := register(t)

		// A read says it is not locked: the field is answered, and false.
		_, notLocked := answered(t, email)
		require.NotNil(t, notLocked, "a read does not answer the lock of an account that has none")
		require.False(t, *notLocked)

		// What could be sent before: no change, nothing remembered.
		must(t, email, map[string]any{"disabled": true})

		_, r := state(t, email)
		require.Equal(t, row{verified: false, disabled: true, locked: false}, r, "`disabled: true` on a registration is no change")

		must(t, email, map[string]any{"disabled_by_admin": true})

		_, r = state(t, email)
		require.Equal(t, row{verified: false, disabled: true, locked: true}, r)

		// And a read says which of the two it is.
		disabled, locked := answered(t, email)
		require.NotNil(t, disabled)
		require.NotNil(t, locked, "a read does not answer the lock")
		assert.True(t, *disabled)
		assert.True(t, *locked)

		follow(t, email)

		_, r = state(t, email)
		assert.True(t, r.verified, "the link proves the address")
		assert.True(t, r.disabled, "a registration an administrator locked was opened by its verification link")
		assert.False(t, signsIn(t, email, password))

		// Unlocked, with its address proven by now: enabled.
		must(t, email, map[string]any{"disabled_by_admin": false})

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, locked: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	t.Run("unlocked_before_its_address_is_proven_it_waits_for_its_link", func(t *testing.T) {
		email, password := register(t)

		must(t, email, map[string]any{"disabled_by_admin": true})
		must(t, email, map[string]any{"disabled_by_admin": false})

		_, r := state(t, email)
		require.Equal(t, row{verified: false, disabled: true, locked: false}, r, "unlocking does not open an account whose address is not proven")
		require.False(t, signsIn(t, email, password))

		follow(t, email)

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, locked: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	// A form that knows both fields sends both back with every edit.
	t.Run("a_form_that_sends_back_what_it_read_changes_nothing", func(t *testing.T) {
		email, _ := register(t)

		must(t, email, map[string]any{"first_name": "Renamed", "disabled": true, "disabled_by_admin": false})

		_, r := state(t, email)
		require.Equal(t, row{verified: false, disabled: true, locked: false}, r, "a rename locked a registration")

		must(t, email, map[string]any{"disabled_by_admin": true})
		must(t, email, map[string]any{"first_name": "Again", "disabled": true, "disabled_by_admin": true})
		must(t, email, map[string]any{"last_name": "Surname"})

		_, r = state(t, email)
		assert.Equal(t, row{verified: false, disabled: true, locked: true}, r, "an edit that changes neither field took the lock off")
	})

	t.Run("an_enabled_account_is_locked_and_unlocked", func(t *testing.T) {
		email, password := register(t)
		follow(t, email)
		require.True(t, signsIn(t, email, password))

		must(t, email, map[string]any{"disabled_by_admin": true})

		_, r := state(t, email)
		assert.Equal(t, row{verified: true, disabled: true, locked: true}, r)
		assert.False(t, signsIn(t, email, password))

		must(t, email, map[string]any{"disabled_by_admin": false})

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, locked: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	t.Run("two_changes_in_two_directions_are_refused_whole", func(t *testing.T) {
		email, _ := register(t)

		status, text := update(t, email, admin, map[string]any{"first_name": "Renamed", "disabled": false, "disabled_by_admin": true})
		assert.Equal(t, http.StatusBadRequest, status, text)

		_, r := state(t, email)
		assert.Equal(t, row{verified: false, disabled: true, locked: false}, r, "part of a refused body was written")

		var firstName string

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT first_name FROM users WHERE email = $1`, email).Scan(&firstName))
		assert.NotEqual(t, "Renamed", firstName, "part of a refused body was written")
	})

	// "Enable it" and "it is locked" in one body, on an account that is
	// locked: the lock is the value the row already has, and is still what
	// the body says. Taken for "no change" and dropped, the `disabled` that
	// was left enabled the account and took the lock off, and the same body
	// sent again locked it.
	t.Run("a_locked_account_is_not_enabled_by_a_body_that_says_it_is_locked", func(t *testing.T) {
		email, password := register(t)
		must(t, email, map[string]any{"disabled_by_admin": true})

		for range 2 {
			status, text := update(t, email, admin, map[string]any{"disabled": false, "disabled_by_admin": true})
			assert.Equal(t, http.StatusBadRequest, status, text)

			_, r := state(t, email)
			assert.Equal(t, row{verified: false, disabled: true, locked: true}, r, "a body that says the account is locked enabled it")
		}

		assert.False(t, signsIn(t, email, password))
	})

	// The mirror: "disable it" and "not by an administrator", on an enabled
	// account.
	t.Run("an_enabled_account_is_not_disabled_by_a_body_that_says_it_is_not_locked", func(t *testing.T) {
		email, password := register(t)
		follow(t, email)

		status, text := update(t, email, admin, map[string]any{"disabled": true, "disabled_by_admin": false})
		assert.Equal(t, http.StatusBadRequest, status, text)

		_, r := state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, locked: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	// Unlocking leaves an unverified account waiting for its link. An
	// administrator who wants it open says so in the same body, as an
	// administrator always could with `disabled: false`.
	t.Run("an_unverified_account_is_unlocked_and_enabled_by_saying_both", func(t *testing.T) {
		email, password := register(t)
		must(t, email, map[string]any{"disabled_by_admin": true})
		must(t, email, map[string]any{"disabled_by_admin": false, "disabled": false})

		_, r := state(t, email)
		assert.Equal(t, row{verified: false, disabled: false, locked: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	// A locked account is a disabled one, and the table holds a statement
	// written by hand to that: enabled and marked, a client would read the
	// account as locked while it signs in.
	t.Run("the_table_refuses_an_enabled_account_that_is_marked", func(t *testing.T) {
		email, _ := register(t)
		must(t, email, map[string]any{"disabled_by_admin": true})

		_, err := testDBPool.Exec(context.Background(), `UPDATE users SET disabled = FALSE WHERE email = $1`, email)
		require.Error(t, err, "the table took an enabled account that is marked as locked")
		assert.Contains(t, err.Error(), "users_locked_is_disabled")
	})

	t.Run("the_users_list_answers_it_and_filters_by_it", func(t *testing.T) {
		email, _ := register(t)
		must(t, email, map[string]any{"disabled_by_admin": true})

		list := func(t *testing.T, filter string) []struct {
			Email           string `json:"email"`
			DisabledByAdmin *bool  `json:"disabled_by_admin"`
		} {
			t.Helper()

			e := newAPIEndpoint(http.MethodGet, "/users")
			e.SetQueryParams(map[string]string{"limit": "100", "filter": filter, "fields": "email,disabled_by_admin", "sort": "disabled_by_admin DESC"})

			response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), e, nil, admin)
			require.NoError(t, err)

			text := readResponseBody(t, response)
			response.Body.Close()

			require.Equal(t, http.StatusOK, response.StatusCode, text)

			var page struct {
				Items []struct {
					Email           string `json:"email"`
					DisabledByAdmin *bool  `json:"disabled_by_admin"`
				} `json:"items"`
			}

			require.NoError(t, json.Unmarshal([]byte(text), &page), text)

			return page.Items
		}

		locked := list(t, "disabled_by_admin = true AND email = '"+email+"'")
		require.Len(t, locked, 1, "the locked account is not in the list of locked accounts")
		require.NotNil(t, locked[0].DisabledByAdmin)
		assert.True(t, *locked[0].DisabledByAdmin)

		assert.Empty(t, list(t, "disabled_by_admin = false AND email = '"+email+"'"))
	})

	// The lock is an account's sign-in, as `disabled` is: on an account with
	// the `admin` flag it is the flag's to set.
	t.Run("the_lock_on_an_administrator_account_is_an_administrators", func(t *testing.T) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		managerID := createUserInDB(t, firstName, lastName, email, password)
		assignRoleToUserInDB(t, managerID, "UserManager")

		t.Cleanup(func() { deleteUserByIDFromDB(t, managerID) })

		manager := map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}

		_, _, flagged := generateUserData(t)
		flaggedID := createUserInDB(t, "Flag", "Holder", flagged, generatePassword(t))

		t.Cleanup(func() { deleteUserByIDFromDB(t, flaggedID) })

		_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = TRUE, verified = TRUE, disabled = FALSE, local_account = TRUE WHERE id = $1`, flaggedID)
		require.NoError(t, err)

		status, text := update(t, flagged, manager, map[string]any{"disabled_by_admin": true})
		assert.Equal(t, http.StatusForbidden, status, text)

		_, r := state(t, flagged)
		assert.Equal(t, row{verified: true, disabled: false, locked: false}, r, "an account without the flag locked one that has it")

		// An ordinary account is the role's to lock.
		ordinary, _ := register(t)

		status, text = update(t, ordinary, manager, map[string]any{"disabled_by_admin": true})
		assert.Equal(t, http.StatusOK, status, text)
	})
}

// TestEveryUserAnswerCarriesTheLock: the field is in every answer that
// carries a user, written at each place one is built: GET /users/{user_id}
// and GET /users are TestTheAdministratorsLockIsAFieldOfItsOwn's; here are
// /me, /me/authz, a role's users and a project's users. The lists read a
// locked account, and answer `true`; /me is read by an account that may act,
// and answers `false`.
func TestEveryUserAnswerCarriesTheLock(t *testing.T) {
	ctx := t.Context()

	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	headers := map[string]string{"Authorization": "Bearer " + admin.AccessToken}

	get := func(t *testing.T, endpoint *apiEndpoint) string {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), endpoint, nil, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusOK, response.StatusCode, text)

		return text
	}

	// A locked account, in a role and in a project.
	firstName, lastName, email := generateUserData(t)
	locked := createUserInDB(t, firstName, lastName, email, generatePassword(t))
	t.Cleanup(func() { deleteUserByIDFromDB(t, locked) })

	_, err := testDBPool.Exec(ctx, `UPDATE users SET disabled = TRUE, disabled_by_admin = TRUE WHERE id = $1`, locked)
	require.NoError(t, err)

	var role uuid.UUID
	require.NoError(t, testDBPool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'ReadOnly'`).Scan(&role))
	assignRoleToUserInDB(t, locked, "ReadOnly")

	project := createProjectInDB(t, generateRandomName(t, "LockProject"), "the lock in a project's users")
	t.Cleanup(func() { deleteProjectByIDFromDB(t, project.ID) })

	_, err = testDBPool.Exec(ctx, `INSERT INTO projects_users (projects_id, users_id) VALUES ($1, $2)`, project.ID, locked)
	require.NoError(t, err)

	lockOf := func(t *testing.T, text string) *bool {
		t.Helper()

		var list struct {
			Items []struct {
				ID              uuid.UUID `json:"id"`
				DisabledByAdmin *bool     `json:"disabled_by_admin"`
			} `json:"items"`
		}

		require.NoError(t, json.Unmarshal([]byte(text), &list), text)

		for _, item := range list.Items {
			if item.ID == locked {
				return item.DisabledByAdmin
			}
		}

		t.Fatalf("the locked account is not in the list: %s", text)

		return nil
	}

	only := func(endpoint *apiEndpoint) *apiEndpoint {
		endpoint.SetQueryParam("filter", "id = '"+locked.String()+"'")

		return endpoint
	}

	t.Run("a_roles_users", func(t *testing.T) {
		lock := lockOf(t, get(t, only(rolesListUserLinkedEndpoint.Clone().RewriteSlugs(role.String()))))
		require.NotNil(t, lock, "a role's users answer without the lock")
		assert.True(t, *lock)
	})

	t.Run("a_projects_users", func(t *testing.T) {
		lock := lockOf(t, get(t, only(newAPIEndpoint(http.MethodGet, "/projects/{project_id}/users").RewriteSlugs(project.ID.String()))))
		require.NotNil(t, lock, "a project's users answer without the lock")
		assert.True(t, *lock)
	})

	for name, endpoint := range map[string]*apiEndpoint{"me": meGetEndpoint, "me_authz": meAuthzEndpoint} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, get(t, endpoint.Clone()), `"disabled_by_admin":false`, "the answer carries no lock")
		})
	}
}
