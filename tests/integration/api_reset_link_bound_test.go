//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var authResetEndpoint = newAPIEndpoint(http.MethodPost, "/auth/password/reset")

// TestAResetLinkIsBoundToTheAccountAsItIsNow: a reset link is bound to the
// account and to the address it was mailed to, and held to the account as
// it is when the link is followed. A link mailed to an address the account
// no longer has set the password until it expired, and so did one mailed
// before an administrator disabled the account. Each refusal is the one 401
// of a link that is not valid, and says nothing of why.
//
// Each case asks for a real link through the public recovery route, reads
// it from the mail server, changes the account through the API as the
// administrator account, and follows it.
func TestAResetLinkIsBoundToTheAccountAsItIsNow(t *testing.T) {
	ctx := t.Context()

	administrator := getAdminUserTokens(t)
	asAdministrator := map[string]string{"Authorization": "Bearer " + administrator.AccessToken}

	const chosen = "Quartz9Falcon!"

	// A local account, verified and enabled, with a registration's role: the
	// one a reset link is mailed to.
	account := func(t *testing.T) (uuid.UUID, string) {
		t.Helper()

		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id := createUserInDB(t, firstName, lastName, email, password)

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		_, err := testDBPool.Exec(ctx, `UPDATE users SET verified = TRUE, disabled = FALSE, disabled_by_admin = FALSE WHERE id = $1`, id)
		require.NoError(t, err)

		// The role a registration is given (auto_assign): it holds the reset
		// route's grant, which the reset's chain asks for.
		assignRoleToUserInDB(t, id, "AuthenticatedUser")

		return id, email
	}

	follow := func(t *testing.T, token string) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), authResetEndpoint,
			map[string]any{"password": chosen}, map[string]string{"Authorization": "Bearer " + token})
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, text
	}

	change := func(t *testing.T, id uuid.UUID, body map[string]any) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), usersUpdateEndpoint.RewriteSlugs(id.String()), body, asAdministrator)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusOK, response.StatusCode, "the change was refused: this case proves nothing. %s", text)
	}

	hash := func(t *testing.T, id uuid.UUID) string {
		t.Helper()

		var h string

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&h))

		return h
	}

	t.Run("the_account_it_was_mailed_to", func(t *testing.T) {
		_, email := account(t)
		token := resetTokenFromEmail(t, email)

		status, body := follow(t, token)
		require.Equal(t, http.StatusOK, status, body)

		assert.NotEmpty(t, loginForMembership(t, email, chosen).AccessToken, "the new password does not sign in")
	})

	refused := func(t *testing.T, name string, body func() map[string]any) {
		t.Run(name, func(t *testing.T) {
			id, email := account(t)
			token := resetTokenFromEmail(t, email)
			before := hash(t, id)

			change(t, id, body())

			status, text := follow(t, token)
			assert.Equal(t, http.StatusUnauthorized, status, text)
			assert.Contains(t, text, "the link is not valid for the account")
			assert.Equal(t, before, hash(t, id), "the password was written")

			// And it is spent: a second try is the used link's 401.
			again, _ := follow(t, token)
			assert.Equal(t, http.StatusUnauthorized, again)
		})
	}

	refused(t, "its_address_changed_since", func() map[string]any {
		return map[string]any{"email": fmt.Sprintf("corrected-%s@example.com", uuid.NewV7().String()[:8])}
	})
	refused(t, "an_administrator_disabled_it_since", func() map[string]any {
		return map[string]any{"disabled": true}
	})
	refused(t, "it_is_no_longer_a_local_account", func() map[string]any {
		return map[string]any{"local_account": false}
	})
}

// resetTokenFromEmail asks the public recovery route for a reset link for
// address and returns its token, from the mail that arrives for it.
func resetTokenFromEmail(t *testing.T, address string) string {
	t.Helper()

	response, err := sendHTTPRequest(t, context.Background(), authRecoverEndpoint, map[string]any{"email": address}, nil)
	require.NoError(t, err)

	text := readResponseBody(t, response)
	response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode, text)
	require.Positive(t, mailsTo(t, address, 10*time.Second), "no reset mail arrived")

	search := newMailAPIEndpoint(http.MethodGet, "/search")
	search.requestURL.RawQuery = "query=" + url.QueryEscape("to:"+address)

	found, err := sendHTTPRequest(t, context.Background(), search, nil)
	require.NoError(t, err)

	var messages struct {
		Messages []struct {
			ID string `json:"ID"`
		} `json:"messages"`
	}

	require.NoError(t, json.Unmarshal([]byte(readResponseBody(t, found)), &messages))
	found.Body.Close()
	require.NotEmpty(t, messages.Messages)

	raw, err := sendHTTPRequest(t, context.Background(), newMailAPIEndpoint(http.MethodGet, "/message/"+messages.Messages[0].ID+"/raw"), nil)
	require.NoError(t, err)

	content, err := io.ReadAll(raw.Body)
	raw.Body.Close()
	require.NoError(t, err)

	match := regexp.MustCompile(`\?token=([A-Za-z0-9._-]+)`).FindSubmatch(content)
	require.NotNil(t, match, "no link in the reset mail")

	return string(match[1])
}
