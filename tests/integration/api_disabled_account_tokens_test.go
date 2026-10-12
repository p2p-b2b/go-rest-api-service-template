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

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// TestADisabledAccountsTokensAreRefusedOnEveryRoute: a token is checked for
// its signature, its expiry and the denylist, and none of that asks about the
// account. Sign-in and refresh do look at `disabled`; a token that existed
// was not asked again, except on /me: a disabled account's access token
// kept answering on every other route until it expired.
//
// Now every authenticated route asks whether the account may act, and
// answers a disabled or deleted one 401 with the code token_revoked: the
// session is over. Logout is the exception, so that such an account can
// still end the tokens it holds.
//
// Not parallel: it disables and enables an account and reads the answers in
// between.
func TestADisabledAccountsTokensAreRefusedOnEveryRoute(t *testing.T) {
	ctx := t.Context()

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	// An account with every grant and no flag, signed in.
	signedIn := func() (uuid.UUID, map[string]string) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id := createUserInDB(t, firstName, lastName, email, password)
		enableUserByEmailFromDB(t, email)
		assignRoleToUserInDB(t, id, "Administrator")

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		return id, map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}
	}

	type answer struct {
		status int
		code   string
		text   string
	}

	ask := func(t *testing.T, endpoint *apiEndpoint, headers map[string]string, body map[string]any) answer {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), endpoint, body, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		var message struct {
			Code string `json:"code"`
		}

		_ = json.Unmarshal([]byte(text), &message)

		return answer{status: response.StatusCode, code: message.Code, text: text}
	}

	setDisabled := func(t *testing.T, id uuid.UUID, disabled bool) {
		t.Helper()

		got := ask(t, usersUpdateEndpoint.RewriteSlugs(id.String()), admin, map[string]any{"disabled": disabled})
		require.Equal(t, http.StatusOK, got.status, got.text)
	}

	users := newAPIEndpoint(http.MethodGet, "/users")

	// The routes asked: one of the general chain, one of the /me chain.
	routes := map[string]*apiEndpoint{"a_route_of_the_general_chain": users, "its_own_profile": meGetEndpoint}

	id, access := signedIn()

	t.Run("an_enabled_account_is_answered", func(t *testing.T) {
		for route, endpoint := range routes {
			got := ask(t, endpoint, access, nil)
			require.Equal(t, http.StatusOK, got.status, "%s: %s", route, got.text)
		}
	})

	t.Run("disabled_its_token_is_refused", func(t *testing.T) {
		setDisabled(t, id, true)

		for route, endpoint := range routes {
			got := ask(t, endpoint, access, nil)

			assert.Equal(t, http.StatusUnauthorized, got.status, "%s still answers a disabled account: %s", route, got.text)
			assert.Equal(t, domain.CodeTokenRevoked, got.code, "%s: the answer does not say the session is over: %s", route, got.text)
			assert.NotContains(t, got.text, "disabled", "the answer says why")
		}
	})

	t.Run("enabled_again_it_works_again", func(t *testing.T) {
		setDisabled(t, id, false)

		for route, endpoint := range routes {
			got := ask(t, endpoint, access, nil)
			assert.Equal(t, http.StatusOK, got.status, "%s: %s", route, got.text)
		}
	})

	// Logout is not asked about the account: a disabled one must still be
	// able to end the tokens it holds. And it does end them: enabled again,
	// the access token it logged out with stays refused.
	t.Run("a_disabled_account_can_still_log_out", func(t *testing.T) {
		setDisabled(t, id, true)

		got := ask(t, authLogoutEndpoint, access, map[string]any{})
		require.Equal(t, http.StatusOK, got.status, got.text)

		setDisabled(t, id, false)

		after := ask(t, users, access, nil)
		assert.Equal(t, http.StatusUnauthorized, after.status, "the logout of a disabled account answered 200 and ended nothing: %s", after.text)
	})

	// The cache is not what decides. The account's cached copy lives twelve
	// hours and is dropped, best effort, when the API changes the account.
	// Disabled where the API does not see it, the cached copy still says
	// "enabled": the check reads the database, and so do the refresh and the
	// sign-in by id.
	t.Run("disabled_behind_the_caches_back", func(t *testing.T) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		stale := createUserInDB(t, firstName, lastName, email, password)
		enableUserByEmailFromDB(t, email)
		assignRoleToUserInDB(t, stale, "Administrator")

		t.Cleanup(func() { deleteUserByIDFromDB(t, stale) })

		login := loginForMembership(t, email, password)
		itsAccess := map[string]string{"Authorization": "Bearer " + login.AccessToken}
		refresh := func() answer {
			return ask(t, authRefreshEndpoint, map[string]string{"Authorization": "Bearer " + login.RefreshToken},
				map[string]any{"refresh_token": login.RefreshToken})
		}

		// Warm whatever caches the account: its profile, its grants.
		require.Equal(t, http.StatusOK, ask(t, meGetEndpoint, itsAccess, nil).status)
		require.Equal(t, http.StatusOK, ask(t, users, itsAccess, nil).status)

		_, err := testDBPool.Exec(ctx, `UPDATE users SET disabled = TRUE WHERE id = $1`, stale)
		require.NoError(t, err)

		for route, endpoint := range routes {
			got := ask(t, endpoint, itsAccess, nil)
			assert.Equal(t, http.StatusUnauthorized, got.status, "%s: the cached copy of the account was believed: %s", route, got.text)
		}

		renewed := refresh()
		assert.NotEqual(t, http.StatusOK, renewed.status, "a disabled account renewed its session from the cached copy: %s", renewed.text)
	})

	t.Run("a_deleted_accounts_token_is_refused_the_same_way", func(t *testing.T) {
		gone, itsAccess := signedIn()

		deleted := ask(t, usersDeleteEndpoint.RewriteSlugs(gone.String()), admin, nil)
		require.Equal(t, http.StatusOK, deleted.status, deleted.text)

		for route, endpoint := range routes {
			got := ask(t, endpoint, itsAccess, nil)

			assert.Equal(t, http.StatusUnauthorized, got.status, "%s, for a deleted account: %s", route, got.text)
			assert.Equal(t, domain.CodeTokenRevoked, got.code, "%s: %s", route, got.text)
		}
	})
}
