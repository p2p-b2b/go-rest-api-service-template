//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetMailsTo counts the "Reset Password" messages the mail server holds for
// an address, looking for up to `wait`: a mail is sent after the request that
// asked for it is answered.
func resetMailsTo(t *testing.T, address string, wait time.Duration) int {
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

		resets := 0

		for _, message := range found.Messages {
			if message.Subject == "Reset Password" {
				resets++
			}
		}

		if resets > 0 || time.Now().After(deadline) {
			return resets
		}

		time.Sleep(200 * time.Millisecond)
	}
}

// TestOnlyTheAdministratorAccountChangesAnotherAccountsAddress: an account's
// e-mail address is where its password reset is sent, so an address is not
// for every holder of the update's grant to change.
//
// Only the administrator account, the one with the `admin` flag, changes an
// account's address. From anyone else a body whose address differs from the
// stored one is a 403 and changes nothing; the same address sent back is no
// change. A reset asked for an account goes to its holder, and nowhere else.
//
// Not parallel: it reads the mail server.
func TestOnlyTheAdministratorAccountChangesAnotherAccountsAddress(t *testing.T) {
	ctx := t.Context()

	account := func(role string) (uuid.UUID, string, map[string]string) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id := createUserInDB(t, firstName, lastName, email, password)
		enableUserByEmailFromDB(t, email)

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		if role == "" {
			return id, email, nil
		}

		assignRoleToUserInDB(t, id, role)

		return id, email, map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}
	}

	managerID, managerEmail, manager := account("UserManager")
	_, _, roleOnly := account("Administrator")

	// The account acted on: local, enabled, verified, with the `admin` flag,
	// as the seeded administrator is.
	holderID, holderEmail, _ := account("")

	_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = TRUE, verified = TRUE, local_account = TRUE WHERE id = $1`, holderID)
	require.NoError(t, err)

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	another := func() string { return fmt.Sprintf("another-%s@example.com", uuid.NewV7().String()[24:]) }

	update := func(t *testing.T, user uuid.UUID, headers map[string]string, body map[string]any) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersUpdateEndpoint.RewriteSlugs(user.String()), body, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, text
	}

	stored := func(t *testing.T, user uuid.UUID) (address, firstName string) {
		t.Helper()

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT email, first_name FROM users WHERE id = $1`, user).Scan(&address, &firstName))

		return address, firstName
	}

	t.Run("a_user_manager_does_not_change_another_accounts_address", func(t *testing.T) {
		mine := another()

		status, body := update(t, holderID, manager, map[string]any{"email": mine})
		assert.Equal(t, http.StatusForbidden, status, body)
		assert.NotContains(t, body, mine, "the address is in the answer")

		address, _ := stored(t, holderID)
		assert.Equal(t, holderEmail, address, "another account's address was changed by an account without the flag")

		// And so the reset still goes to the holder, and nowhere else.
		response, err := sendHTTPRequest(t, ctx, usersPasswordResetEndpoint.RewriteSlugs(holderID.String()), nil, manager)
		require.NoError(t, err)
		response.Body.Close()
		require.Equal(t, http.StatusAccepted, response.StatusCode)

		assert.Positive(t, resetMailsTo(t, holderEmail, 10*time.Second), "the reset was asked for and no mail reached the holder: this case proves nothing")
		assert.Zero(t, resetMailsTo(t, mine, 0), "a reset link for another account arrived at the caller's address")
	})

	t.Run("nothing_of_a_refused_body_is_written", func(t *testing.T) {
		_, before := stored(t, holderID)

		status, body := update(t, holderID, manager, map[string]any{"email": another(), "first_name": "Renamed"})
		assert.Equal(t, http.StatusForbidden, status, body)

		_, after := stored(t, holderID)
		assert.Equal(t, before, after, "part of a refused body was written")
	})

	// A long s for an s: an address a comparison that folds case calls the
	// stored one, that no lowering turns into it, and that Postgres stores
	// as the other address it is.
	t.Run("an_address_that_only_looks_like_the_stored_one_is_a_change", func(t *testing.T) {
		suffix := uuid.NewV7().String()[24:]
		spelled := fmt.Sprintf("holder.smith.%s@example.com", suffix)
		folded := fmt.Sprintf("holder.ſmith.%s@example.com", suffix)

		id := createUserInDB(t, "Holder", "Smith", spelled, generatePassword(t))
		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		status, body := update(t, id, manager, map[string]any{"email": folded})
		assert.Equal(t, http.StatusForbidden, status, body)

		address, _ := stored(t, id)
		assert.Equal(t, spelled, address, "an account without the flag wrote an address that is not the holder's")
	})

	t.Run("the_stored_address_sent_back_is_no_change", func(t *testing.T) {
		// What a client that sends the whole form does.
		status, body := update(t, holderID, manager, map[string]any{"email": holderEmail, "first_name": "Renamed"})
		require.Equal(t, http.StatusOK, status, body)

		address, firstName := stored(t, holderID)
		assert.Equal(t, holderEmail, address)
		assert.Equal(t, "Renamed", firstName, "the rest of the body was not written")
	})

	t.Run("a_body_without_an_address_is_as_before", func(t *testing.T) {
		status, body := update(t, holderID, manager, map[string]any{"last_name": "Surname"})
		require.Equal(t, http.StatusOK, status, body)
	})

	t.Run("not_its_own_address_either", func(t *testing.T) {
		status, body := update(t, managerID, manager, map[string]any{"email": another()})
		assert.Equal(t, http.StatusForbidden, status, body)

		address, _ := stored(t, managerID)
		assert.Equal(t, managerEmail, address)
	})

	t.Run("the_administrator_role_without_the_flag_is_held_like_anyone", func(t *testing.T) {
		status, body := update(t, holderID, roleOnly, map[string]any{"email": another()})
		assert.Equal(t, http.StatusForbidden, status, body)

		address, _ := stored(t, holderID)
		assert.Equal(t, holderEmail, address)
	})

	t.Run("the_administrator_account_changes_an_address", func(t *testing.T) {
		corrected := another()

		status, body := update(t, managerID, admin, map[string]any{"email": corrected})
		require.Equal(t, http.StatusOK, status, body)

		address, _ := stored(t, managerID)
		assert.Equal(t, corrected, address)
	})
}
