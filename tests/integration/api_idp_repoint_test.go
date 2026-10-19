//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAProviderIsNotRepointedUnderItsIdentities: an identity is (provider,
// subject), and a sign-in through a provider is "this provider says this
// subject", then a session for the account. `PUT /auth/idps/{idp_id}` could
// change the provider's issuer and kind while its identities stayed: whoever
// holds that route pointed the provider at an issuer of their own, asserted
// an existing subject, and was signed in as that account.
//
// So another issuer or type is a 409 while the provider has
// identities (a new provider instead). The client and the callback may
// change, and while accounts are linked that change is the administrator
// account's to make: set to an application and a host of one's own, a
// linked user's ordinary sign-in handed the code over.
//
// Not parallel: it changes rows other tests do not see, and reads them back.
func TestAProviderIsNotRepointedUnderItsIdentities(t *testing.T) {
	ctx := t.Context()

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	// An account with the Administrator role and not the flag: the routes
	// are the role's.
	firstName, lastName, email := generateUserData(t)
	password := generatePassword(t)
	roleOnlyID := createUserInDB(t, firstName, lastName, email, password)
	t.Cleanup(func() { deleteUserByIDFromDB(t, roleOnlyID) })
	enableUserByEmailFromDB(t, email)
	assignRoleToUserInDB(t, roleOnlyID, "Administrator")

	roleOnly := map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}

	okta, google := getIDPTypeFromDBByName(t, "Okta").ID, getIDPTypeFromDBByName(t, "Google").ID

	// An account with an identity at the provider; with the flag or not.
	linked := func(t *testing.T, idp string, flag bool) uuid.UUID {
		t.Helper()

		first, last, address := generateUserData(t)
		id := createUserInDB(t, first, last, address, generatePassword(t))

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		_, err := testDBPool.Exec(ctx, `UPDATE users SET admin = $2, verified = TRUE, disabled = FALSE WHERE id = $1`, id, flag)
		require.NoError(t, err)

		_, err = testDBPool.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
			id, idp, "sub-"+id.String(), address)
		require.NoError(t, err)

		return id
	}

	update := func(t *testing.T, idp string, headers map[string]string, body map[string]any) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), idpsUpdateEndpoint.Clone().RewriteSlugs(idp), body, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, text
	}

	stored := func(t *testing.T, idp string) (issuer, kind, client string) {
		t.Helper()

		var issuerURL *string

		require.NoError(t, testDBPool.QueryRow(context.Background(),
			`SELECT issuer_url, idp_types::text, client_id FROM idps WHERE id = $1`, idp).Scan(&issuerURL, &kind, &client))

		if issuerURL != nil {
			issuer = *issuerURL
		}

		return issuer, kind, client
	}

	const repointed = "identity provider cannot be pointed at another issuer or type: still used by at least one linked account"

	t.Run("an_issuer_of_ones_own_is_refused_under_an_identity", func(t *testing.T) {
		mine, theirs := newFakeProvider(t), newFakeProvider(t)
		idp := createOIDCIDP(t, adminTokens.AccessToken, mine, false)
		linked(t, idp, false)

		// Even by the administrator account: the identities are the
		// accounts', and the issuer is what says who they are.
		for who, headers := range map[string]map[string]string{"an_account_with_the_role": roleOnly, "the_administrator_account": admin} {
			status, text := update(t, idp, headers, map[string]any{"issuer_url": theirs.srv.URL})
			assert.Equal(t, http.StatusConflict, status, "%s: %s", who, text)
			assert.Contains(t, text, repointed, who)
		}

		issuer, _, _ := stored(t, idp)
		assert.Equal(t, mine.srv.URL, issuer, "the provider was pointed at another issuer")

		// So a sign-in still goes to the issuer the identities are of.
		redirect, err := url.Parse(startFlow(t, idpLoginEndpoint, idp, nil))
		require.NoError(t, err)
		assert.Equal(t, mine.srv.URL+"/authorize", redirect.Scheme+"://"+redirect.Host+redirect.Path)
	})

	t.Run("another_kind_is_refused_under_an_identity", func(t *testing.T) {
		f := newFakeProvider(t)
		idp := createOIDCIDP(t, adminTokens.AccessToken, f, false)
		linked(t, idp, false)

		status, text := update(t, idp, roleOnly, map[string]any{"idp_type_id": google})
		assert.Equal(t, http.StatusConflict, status, text)
		assert.Contains(t, text, repointed)

		_, kind, _ := stored(t, idp)
		assert.Equal(t, okta.String(), kind)
	})

	// A form sends the whole provider back: the issuer and the kind it
	// read are no change, and the rest is written.
	t.Run("the_issuer_and_kind_sent_back_with_a_new_name", func(t *testing.T) {
		f := newFakeProvider(t)
		idp := createOIDCIDP(t, adminTokens.AccessToken, f, false)
		linked(t, idp, false)

		name := generateRandomName(t, "Renamed")
		status, text := update(t, idp, roleOnly, map[string]any{"issuer_url": f.srv.URL, "idp_type_id": okta, "name": name})
		assert.Equal(t, http.StatusOK, status, text)

		var got string
		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT name FROM idps WHERE id = $1`, idp).Scan(&got))
		assert.Equal(t, name, got)
	})

	t.Run("a_provider_nobody_is_linked_to_is_repointed", func(t *testing.T) {
		before, after := newFakeProvider(t), newFakeProvider(t)
		idp := createOIDCIDP(t, adminTokens.AccessToken, before, false)

		status, text := update(t, idp, roleOnly, map[string]any{"issuer_url": after.srv.URL})
		assert.Equal(t, http.StatusOK, status, text)

		issuer, _, _ := stored(t, idp)
		assert.Equal(t, after.srv.URL, issuer)
	})

	// Where the sign-in goes, the client and the callback: the
	// administrator account's while accounts are linked, the role's while
	// nobody is.
	for field, value := range map[string]string{"client_id": "another-client", "callback_url": "https://elsewhere.example/auth/idp/callback"} {
		t.Run(field+"_under_an_identity_is_the_administrators", func(t *testing.T) {
			f := newFakeProvider(t)
			idp := createOIDCIDP(t, adminTokens.AccessToken, f, false)
			linked(t, idp, false)

			status, text := update(t, idp, roleOnly, map[string]any{field: value})
			assert.Equal(t, http.StatusForbidden, status, text)
			assert.Contains(t, text, field)

			status, text = update(t, idp, admin, map[string]any{field: value})
			assert.Equal(t, http.StatusOK, status, text)
		})

		t.Run(field+"_where_nobody_is_linked_is_the_roles", func(t *testing.T) {
			f := newFakeProvider(t)
			idp := createOIDCIDP(t, adminTokens.AccessToken, f, false)

			status, text := update(t, idp, roleOnly, map[string]any{field: value})
			assert.Equal(t, http.StatusOK, status, text)
		})
	}

	t.Run("the_client_is_written_by_the_administrator_account", func(t *testing.T) {
		f := newFakeProvider(t)
		idp := createOIDCIDP(t, adminTokens.AccessToken, f, false)
		linked(t, idp, true)

		status, text := update(t, idp, admin, map[string]any{"client_id": "another-client"})
		require.Equal(t, http.StatusOK, status, text)

		_, _, client := stored(t, idp)
		assert.Equal(t, "another-client", client)
	})

	// A provider of the github kind has no issuer, and its form sends an
	// empty one back: no change.
	t.Run("a_github_providers_form_under_an_identity", func(t *testing.T) {
		idp := createGithubIDP(t, adminTokens.AccessToken)
		linked(t, idp, false)

		status, text := update(t, idp, roleOnly, map[string]any{"issuer_url": "", "name": generateRandomName(t, "Renamed")})
		assert.Equal(t, http.StatusOK, status, text)
	})

	t.Run("a_provider_that_is_not_there_is_a_404", func(t *testing.T) {
		status, text := update(t, uuid.NewV7().String(), roleOnly, map[string]any{"issuer_url": newFakeProvider(t).srv.URL})
		assert.Equal(t, http.StatusNotFound, status, text)
	})
}

// createGithubIDP creates a provider of the github kind, which has no
// issuer, and deletes it when the test ends, after the accounts the test
// linked to it.
func createGithubIDP(t *testing.T, admin string) string {
	t.Helper()

	hdr := map[string]string{"Authorization": "Bearer " + admin}
	id := mustUUIDString(t)

	resp, err := sendHTTPRequest(t, t.Context(), idpsCreateEndpoint, map[string]any{
		"id": id, "idp_type_id": getIDPTypeFromDBByName(t, "Github").ID, "name": generateRandomName(t, "GithubIDP"),
		"description": "a github provider under an identity", "callback_url": "http://localhost:5173/auth/idp/" + id + "/callback",
		"client_id": "client-id", "client_secret": "client-secret", "logo": "https://example.com/logo.png",
		"auto_provision": false,
	}, hdr)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, readResponseBody(t, resp))
	resp.Body.Close()

	t.Cleanup(func() {
		// A fresh context: t.Context() is already cancelled by the time
		// cleanup runs.
		r, err := sendHTTPRequest(t, context.Background(), idpsDeleteEndpoint.Clone().RewriteSlugs(id), nil, hdr)
		if err != nil {
			t.Logf("cleanup: provider %s was not deleted: %v", id, err)

			return
		}

		if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusNotFound {
			t.Logf("cleanup: provider %s was left behind: %d %s", id, r.StatusCode, readResponseBody(t, r))
		}

		_ = r.Body.Close()
	})

	return id
}
