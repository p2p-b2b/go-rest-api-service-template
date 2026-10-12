//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mailsTo counts the messages the mail server holds for an address, looking
// for up to `wait`: a mail is sent after the request that asked for it is
// answered.
func mailsTo(t *testing.T, address string, wait time.Duration) int {
	t.Helper()

	var found struct {
		Messages []struct {
			Subject string `json:"Subject"`
		} `json:"messages"`
	}

	for deadline := time.Now().Add(wait); ; {
		search := newMailAPIEndpoint(http.MethodGet, "/search")
		search.requestURL.RawQuery = "query=" + url.QueryEscape("to:"+address)

		response, err := sendHTTPRequest(t, context.Background(), search, nil)
		require.NoError(t, err)

		body := readResponseBody(t, response)
		response.Body.Close()

		require.NoError(t, json.Unmarshal([]byte(body), &found), body)

		if len(found.Messages) > 0 || time.Now().After(deadline) {
			return len(found.Messages)
		}

		time.Sleep(200 * time.Millisecond)
	}
}

// TestAnAccountAnAdministratorDisabledIsNotOpenedByItsLink: following a
// verification link proves the address and enables the account, because a
// registration is disabled for no other reason than that its link has not
// been followed.
//
// "Disabled until the link is followed" and "disabled by an administrator"
// are the same two values, verified false and disabled true, and the link
// must open only the first.
//
// So an account an administrator creates is stored verified, and the link
// does not apply to it; and a `disabled` an administrator sets is remembered
// (users.disabled_by_admin), so the link proves the address of such an
// account and opens nothing.
//
// What is remembered is a CHANGE of `disabled`. The users form sends every
// field back, so "disabled: true" arrives with every rename of a
// registration whose link is still out, and that is nobody disabling it.
//
// Not parallel: it reads the mail of the accounts it makes.
func TestAnAccountAnAdministratorDisabledIsNotOpenedByItsLink(t *testing.T) {
	ctx := t.Context()

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	// The row, as the database holds it.
	type row struct{ verified, disabled, byAdmin bool }

	state := func(t *testing.T, email string) (uuid.UUID, row) {
		t.Helper()

		var (
			id uuid.UUID
			r  row
		)

		require.NoError(t, testDBPool.QueryRow(context.Background(),
			`SELECT id, verified, disabled, disabled_by_admin FROM users WHERE email = $1`, email).Scan(&id, &r.verified, &r.disabled, &r.byAdmin))

		return id, r
	}

	// A registration through the public route, with its mail arrived.
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
		require.Equal(t, row{verified: false, disabled: true, byAdmin: false}, r, "a registration, before its link is followed")

		return email, password
	}

	update := func(t *testing.T, email string, body map[string]any) {
		t.Helper()

		id, _ := state(t, email)

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersUpdateEndpoint.RewriteSlugs(id.String()), body, admin)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusOK, response.StatusCode, text)
	}

	// Following the link the registration mailed. It answers 200 whether
	// or not the account is opened: that is not for whoever holds a link
	// to be told.
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

	t.Run("a_registration_is_opened_by_its_link", func(t *testing.T) {
		email, password := register(t)

		follow(t, email)

		_, r := state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, byAdmin: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	t.Run("an_account_an_administrator_disabled_is_verified_and_stays_disabled", func(t *testing.T) {
		email, password := register(t)

		// An administrator opens it by hand, and it is in use, unverified.
		update(t, email, map[string]any{"disabled": false})

		_, r := state(t, email)
		require.Equal(t, row{verified: false, disabled: false, byAdmin: false}, r)
		require.True(t, signsIn(t, email, password), "an account an administrator enabled does not sign in: this case proves nothing")

		// And then disables it.
		update(t, email, map[string]any{"disabled": true})

		_, r = state(t, email)
		require.Equal(t, row{verified: false, disabled: true, byAdmin: true}, r, "the two values of a registration whose link is still out, and the mark that says it is not one")

		// The mark outlives every later edit that does not enable: the
		// whole form sent back with a new name, which is what the users
		// page sends, and a body that carries no `disabled` at all.
		update(t, email, map[string]any{"first_name": "Renamed", "disabled": true, "local_account": true})

		_, r = state(t, email)
		require.Equal(t, row{verified: false, disabled: true, byAdmin: true}, r, "a rename with the form sent back took the administrator's mark off")

		update(t, email, map[string]any{"last_name": "Surname"})

		_, r = state(t, email)
		require.Equal(t, row{verified: false, disabled: true, byAdmin: true}, r, "an edit that does not mention `disabled` took the administrator's mark off")

		// The link its registration was mailed.
		follow(t, email)

		_, r = state(t, email)
		assert.True(t, r.verified, "the link proves the address")
		assert.True(t, r.disabled, "an account an administrator disabled enabled itself through its verification link")
		assert.False(t, signsIn(t, email, password), "an account an administrator disabled signed in")

		// It is enabled by an administrator, and the mark goes with it.
		update(t, email, map[string]any{"disabled": false})

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, byAdmin: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	// The users form sends every field back. A rename of a registration
	// arrives with "disabled: true", and must not be read as an
	// administrator disabling it: its link would then open nothing, and the
	// person who registered would be locked out by somebody correcting a
	// name. The same bytes are what an administrator who did mean to shut
	// it would send; such a registration is deleted, or enabled and then
	// disabled.
	t.Run("a_registration_renamed_with_the_whole_form_is_still_opened_by_its_link", func(t *testing.T) {
		email, password := register(t)

		update(t, email, map[string]any{"first_name": "Renamed", "disabled": true, "local_account": true})

		_, r := state(t, email)
		require.Equal(t, row{verified: false, disabled: true, byAdmin: false}, r, "disabled as it already was is no change, and is not remembered as one")

		follow(t, email)

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, byAdmin: false}, r)
		assert.True(t, signsIn(t, email, password))
	})

	// The flow decides "not verified yet" from a copy of the account that
	// may be a cached one, twelve hours old at most. The statement decides
	// again, from the row: one that was verified meanwhile is not written a
	// second time, so a link cannot open an account that is verified and
	// disabled, whatever the copy said.
	t.Run("a_row_verified_behind_the_caches_back_is_not_written_again", func(t *testing.T) {
		// Without the service's cache the flow reads the row itself and
		// answers before the statement is asked: the case would pass and
		// prove nothing about the statement.
		requireValkeyAvailable(t)

		email, password := register(t)
		id, _ := state(t, email)

		// The service reads the account, and keeps the copy.
		read, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersGetEndpoint.RewriteSlugs(id.String()), nil, admin)
		require.NoError(t, err)
		read.Body.Close()
		require.Equal(t, http.StatusOK, read.StatusCode)

		// Verified and still disabled, as an account an administrator
		// created is, written where the service does not see it.
		_, err = testDBPool.Exec(ctx, `UPDATE users SET verified = TRUE WHERE id = $1`, id)
		require.NoError(t, err)

		confirmed, err := confirmVerification(t, verificationTokenFromEmail(t, verifyEmailAddress, email))
		require.NoError(t, err)

		text := readResponseBody(t, confirmed)
		confirmed.Body.Close()

		// And answered as an account whose copy was current is: verified
		// already, not "no such account".
		assert.Equal(t, http.StatusConflict, confirmed.StatusCode, "a link for a row that is verified already: %s", text)

		_, r := state(t, email)
		assert.Equal(t, row{verified: true, disabled: true, byAdmin: false}, r, "a verified, disabled account was opened by a link")
		assert.False(t, signsIn(t, email, password))
	})

	// A link proves the address it was mailed to. The flow compares that
	// address with its copy of the account; the statement compares it with
	// the row. An account whose address was corrected since has not proven
	// the new one, and the old link does not verify, or open, it.
	t.Run("a_link_for_an_address_the_account_no_longer_has_proves_nothing", func(t *testing.T) {
		requireValkeyAvailable(t)

		email, password := register(t)
		id, _ := state(t, email)

		read, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersGetEndpoint.RewriteSlugs(id.String()), nil, admin)
		require.NoError(t, err)
		read.Body.Close()
		require.Equal(t, http.StatusOK, read.StatusCode)

		// The link is read before the address goes: the helper looks the
		// mail up by where it was sent.
		token := verificationTokenFromEmail(t, verifyEmailAddress, email)

		_, _, corrected := generateUserData(t)
		t.Cleanup(func() { deleteUserByEmailFromDB(t, corrected) })

		_, err = testDBPool.Exec(ctx, `UPDATE users SET email = $1 WHERE id = $2`, corrected, id)
		require.NoError(t, err)

		confirmed, err := confirmVerification(t, token)
		require.NoError(t, err)

		text := readResponseBody(t, confirmed)
		confirmed.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, confirmed.StatusCode, "a link mailed to an address the account no longer has: %s", text)

		_, r := state(t, corrected)
		assert.Equal(t, row{verified: false, disabled: true, byAdmin: false}, r, "an address was proven by a link that was mailed somewhere else")
		assert.False(t, signsIn(t, corrected, password))
	})

	// An account with the `admin` flag is every project. A link that is
	// mailed to an address does not open one, whatever its row says.
	t.Run("an_account_with_the_admin_flag_is_not_opened_by_a_link", func(t *testing.T) {
		email, password := register(t)
		id, _ := state(t, email)

		_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = TRUE WHERE id = $1`, id)
		require.NoError(t, err)

		follow(t, email)

		_, r := state(t, email)
		assert.True(t, r.verified, "the link proves the address")
		assert.True(t, r.disabled, "an account with the admin flag was opened by a verification link")
		assert.False(t, signsIn(t, email, password))
	})

	t.Run("an_account_an_administrator_creates_is_verified_and_no_link_applies", func(t *testing.T) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)

		t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersCreateEndpoint, map[string]any{
			"email": email, "first_name": firstName, "last_name": lastName, "password": password,
		}, admin)
		require.NoError(t, err)

		body := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusCreated, response.StatusCode, body)

		_, r := state(t, email)
		require.Equal(t, row{verified: true, disabled: true, byAdmin: true}, r, "stored verified, and disabled by an administrator, for one to enable")

		// The public route for a link, and the recovery, which mails the
		// same link to an unverified account: each answers as for any
		// address, and mails nothing.
		for name, endpoint := range map[string]*apiEndpoint{"a verification link": authReVerifyEndpoint, "a recovery": authRecoverEndpoint} {
			asked, err := sendHTTPRequest(t, context.WithoutCancel(ctx), endpoint, map[string]any{"email": email})
			require.NoError(t, err)
			asked.Body.Close()

			assert.Equal(t, http.StatusOK, asked.StatusCode, name)
		}

		assert.Zero(t, mailsTo(t, email, 3*time.Second), "a mail went to an account an administrator created and has not enabled")

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: true, byAdmin: true}, r)
		assert.False(t, signsIn(t, email, password))

		// Enabled by an administrator, it is in use, and the mark is gone.
		update(t, email, map[string]any{"disabled": false})

		_, r = state(t, email)
		assert.Equal(t, row{verified: true, disabled: false, byAdmin: false}, r)
		assert.True(t, signsIn(t, email, password))
	})
}
