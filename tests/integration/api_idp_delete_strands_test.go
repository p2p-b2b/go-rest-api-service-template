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

// TestAnIdentityProvidersDeleteStrandsNobody: an identity provider's delete
// takes its identities with it (the foreign key cascades), and asked nothing
// first. An account that signs in through that provider and has no password
// was left with no way in at all, for nobody to sign in to; a provider
// created again does not give it back, since its sign-in finds an account
// with the address that is not linked. Recovery by e-mail sends such an
// account nothing.
//
// The rule: the delete is a 409 while an account's only way in is an
// identity at that provider. The administrator disables the provider instead,
// or deletes those accounts, or makes them local accounts first. Unlinking
// the last identity of such an account was already refused; this is the same
// act for all of them at once.
//
// Not parallel: it reads the rows it changes.
func TestAnIdentityProvidersDeleteStrandsNobody(t *testing.T) {
	ctx := t.Context()

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	// A provider, and an account with an identity at it.
	provider := func(t *testing.T) uuid.UUID {
		t.Helper()

		return uuid.MustParse(createOIDCIDP(t, adminTokens.AccessToken, newFakeProvider(t), false))
	}

	account := func(t *testing.T, local bool, providers ...uuid.UUID) uuid.UUID {
		t.Helper()

		firstName, lastName, email := generateUserData(t)
		id := createUserInDB(t, firstName, lastName, email, generatePassword(t))

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		_, err := testDBPool.Exec(ctx, `UPDATE users SET local_account = $2, verified = TRUE, disabled = FALSE WHERE id = $1`, id, local)
		require.NoError(t, err)

		for _, idp := range providers {
			_, err := testDBPool.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
				id, idp, "sub-"+id.String(), email)
			require.NoError(t, err)
		}

		return id
	}

	remove := func(t *testing.T, idp uuid.UUID) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), idpsDeleteEndpoint.Clone().RewriteSlugs(idp.String()), nil, admin)
		require.NoError(t, err)

		body := readResponseBody(t, response)
		response.Body.Close()

		return response.StatusCode, body
	}

	exists := func(t *testing.T, idp uuid.UUID) bool {
		t.Helper()

		var found bool

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT EXISTS (SELECT 1 FROM idps WHERE id = $1)`, idp).Scan(&found))

		return found
	}

	identities := func(t *testing.T, user uuid.UUID) int {
		t.Helper()

		var n int

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT count(*) FROM users_identities WHERE users_id = $1`, user).Scan(&n))

		return n
	}

	const kept = "identity provider cannot be deleted: still used by at least one account that has no other way to sign in"

	t.Run("an_account_whose_only_way_in_it_is_keeps_it", func(t *testing.T) {
		idp := provider(t)
		federated := account(t, false, idp)

		status, body := remove(t, idp)
		assert.Equal(t, http.StatusConflict, status, body)
		assert.Contains(t, body, kept)
		assert.True(t, exists(t, idp), "the provider was deleted under an account whose only way in it is")
		assert.Equal(t, 1, identities(t, federated), "the account lost its identity")

		// Asked again, it is refused again: nothing was changed by the first.
		status, body = remove(t, idp)
		assert.Equal(t, http.StatusConflict, status, body)

		// Made a local account, it has its password: the provider goes, and
		// the identity with it.
		_, err := testDBPool.Exec(ctx, `UPDATE users SET local_account = TRUE WHERE id = $1`, federated)
		require.NoError(t, err)

		status, body = remove(t, idp)
		require.Equal(t, http.StatusOK, status, body)
		assert.False(t, exists(t, idp))
		assert.Zero(t, identities(t, federated))
	})

	t.Run("a_disabled_account_counts_too", func(t *testing.T) {
		// Enabled again, it would have no way in.
		idp := provider(t)
		federated := account(t, false, idp)

		_, err := testDBPool.Exec(ctx, `UPDATE users SET disabled = TRUE WHERE id = $1`, federated)
		require.NoError(t, err)

		status, body := remove(t, idp)
		assert.Equal(t, http.StatusConflict, status, body)
		assert.True(t, exists(t, idp))
	})

	t.Run("an_account_with_a_password_does_not_keep_it", func(t *testing.T) {
		idp := provider(t)
		local := account(t, true, idp)

		status, body := remove(t, idp)
		require.Equal(t, http.StatusOK, status, body)
		assert.False(t, exists(t, idp))
		assert.Zero(t, identities(t, local), "the identity outlived its provider")
	})

	t.Run("an_account_with_an_identity_elsewhere_does_not_keep_it", func(t *testing.T) {
		idp, other := provider(t), provider(t)
		federated := account(t, false, idp, other)

		status, body := remove(t, idp)
		require.Equal(t, http.StatusOK, status, body)
		assert.Equal(t, 1, identities(t, federated), "the identity at the other provider is the account's way in")

		// And that other one is its only way in now.
		status, body = remove(t, other)
		assert.Equal(t, http.StatusConflict, status, body)
		assert.True(t, exists(t, other))
	})

	t.Run("one_account_among_several_keeps_it", func(t *testing.T) {
		idp := provider(t)
		account(t, true, idp)
		account(t, false, idp)
		account(t, true, idp)

		status, body := remove(t, idp)
		assert.Equal(t, http.StatusConflict, status, body)
		assert.True(t, exists(t, idp))
	})

	t.Run("a_provider_nobody_signs_in_through_is_deleted", func(t *testing.T) {
		idp := provider(t)

		status, body := remove(t, idp)
		require.Equal(t, http.StatusOK, status, body)
		assert.False(t, exists(t, idp))
	})

	t.Run("a_provider_that_is_not_there_is_a_404", func(t *testing.T) {
		status, body := remove(t, uuid.NewV7())
		assert.Equal(t, http.StatusNotFound, status, body)
		assert.NotContains(t, body, "still used")
	})

	// A disable takes nothing away that enabling does not give back, and is
	// what an administrator does instead: it is allowed under such an
	// account. (That it is refused under the last administrator account is
	// held by the unit tests: this suite makes such accounts all the time.)
	t.Run("the_same_provider_is_disabled_instead", func(t *testing.T) {
		idp := provider(t)
		account(t, false, idp)

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), idpsUpdateEndpoint.Clone().RewriteSlugs(idp.String()), map[string]any{"enabled": false}, admin)
		require.NoError(t, err)

		body := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusOK, response.StatusCode, body)

		var enabled bool

		require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT enabled FROM idps WHERE id = $1`, idp).Scan(&enabled))
		assert.False(t, enabled)
	})
}

// TestAnIdentityAtADisabledProviderIsNoWayIn: "can sign in" counted a
// federated administrator account that has an identity at a provider,
// whether or not the provider is enabled. Disabled, it is no way in, and the
// account must not count as one that can still sign in when the guard asks
// whether another is the last. And the count without one provider is what
// is left when that provider is disabled or deleted.
//
// Not parallel, and that is the point: the suite's parallel tests create
// accounts with the flag all the time, and this counts them.
func TestAnIdentityAtADisabledProviderIsNoWayIn(t *testing.T) {
	ctx := t.Context()

	users, err := repositorypg.NewUsersRepository(repositorypg.UsersRepositoryConfig{
		DB:              testDBPool,
		OT:              newTestOpenTelemetry(t),
		MaxPingTimeout:  5 * time.Second,
		MaxQueryTimeout: 10 * time.Second,
	})
	require.NoError(t, err)

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	idp := uuid.MustParse(createOIDCIDP(t, adminTokens.AccessToken, newFakeProvider(t), false))
	other := uuid.MustParse(createOIDCIDP(t, adminTokens.AccessToken, newFakeProvider(t), false))

	count := func() int {
		t.Helper()

		n, err := users.CountEnabledAdministrators(ctx)
		require.NoError(t, err)

		return n
	}

	// Whether the account is one of those the count counts: the guard of
	// the administrator account asks it about its target.
	signsIn := func(account uuid.UUID) bool {
		t.Helper()

		among, err := users.IsEnabledAdministrator(ctx, account)
		require.NoError(t, err)

		return among
	}

	identities, err := repositorypg.NewUsersIdentitiesRepository(repositorypg.UsersIdentitiesRepositoryConfig{
		DB:              testDBPool,
		OT:              newTestOpenTelemetry(t),
		MaxPingTimeout:  5 * time.Second,
		MaxQueryTimeout: 10 * time.Second,
	})
	require.NoError(t, err)

	// What each identity of the account says of its provider: an account's
	// own unlink reads it.
	enabledAt := func(account uuid.UUID) map[uuid.UUID]bool {
		t.Helper()

		items, err := identities.SelectByUserID(ctx, account)
		require.NoError(t, err)

		at := map[uuid.UUID]bool{}
		for _, item := range items {
			at[item.IDPID] = item.IDPEnabled
		}

		return at
	}

	without := func(provider uuid.UUID) int {
		t.Helper()

		n, err := users.CountEnabledAdministratorsWithout(ctx, provider)
		require.NoError(t, err)

		return n
	}

	before := count()

	// An administrator account that signs in through the provider alone.
	firstName, lastName, email := generateUserData(t)
	federated := createUserInDB(t, firstName, lastName, email, generatePassword(t))

	t.Cleanup(func() { deleteUserByIDFromDB(t, federated) })

	_, err = testDBPool.Exec(ctx, `UPDATE users SET admin = TRUE, disabled = FALSE, verified = TRUE, local_account = FALSE WHERE id = $1`, federated)
	require.NoError(t, err)

	require.Equal(t, before, count(), "a federated account with no identity has no way in")
	require.False(t, signsIn(federated), "and is not one of the accounts counted")

	_, err = testDBPool.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
		federated, idp, "sub-"+federated.String(), email)
	require.NoError(t, err)

	require.Equal(t, before+1, count(), "with an identity at an enabled provider it can sign in")
	require.True(t, signsIn(federated))
	require.Equal(t, map[uuid.UUID]bool{idp: true}, enabledAt(federated))
	assert.Equal(t, before, without(idp), "without that provider it cannot")
	assert.Equal(t, before+1, without(other), "without another provider it still can")

	// Disabled, the provider is no way in.
	_, err = testDBPool.Exec(ctx, `UPDATE idps SET enabled = FALSE WHERE id = $1`, idp)
	require.NoError(t, err)

	assert.Equal(t, before, count(), "an identity at a disabled provider was counted as a way in")
	assert.Equal(t, before, without(idp))
	assert.False(t, signsIn(federated), "the guard would refuse to lock it as one that can sign in")
	assert.Equal(t, map[uuid.UUID]bool{idp: false}, enabledAt(federated), "the identity does not say its provider is disabled")

	// A second identity, at an enabled provider, is one.
	_, err = testDBPool.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
		federated, other, "sub-other-"+federated.String(), email)
	require.NoError(t, err)

	assert.Equal(t, before+1, count())
	assert.True(t, signsIn(federated))
	assert.Equal(t, map[uuid.UUID]bool{idp: false, other: true}, enabledAt(federated))
	assert.Equal(t, before, without(other), "the disabled provider's identity does not stand in for the other")
	assert.Equal(t, before+1, without(idp))

	// An ordinary account, and a disabled one with the flag, are not counted.
	_, err = testDBPool.Exec(ctx, `UPDATE users SET disabled = TRUE WHERE id = $1`, federated)
	require.NoError(t, err)
	assert.False(t, signsIn(federated))
	assert.False(t, signsIn(uuid.NewV7()), "an account that is not there is not one of them")

	// And a provider that was not named is not "no provider".
	_, err = users.CountEnabledAdministratorsWithout(ctx, uuid.Nil())
	require.Error(t, err)
}
