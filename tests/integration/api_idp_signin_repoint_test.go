//go:build integration

package integration

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// changedDuringSignIn is the start of the refusal of a sign-in whose
// provider is no longer the one it signed in with.
const changedDuringSignIn = "the identity provider changed while signing in"

// heldSignIn is a first sign-in at a fake provider, held at the provider's
// token endpoint, the provider already read, until it is released.
type heldSignIn struct {
	done    chan struct{}
	release func()
	body    string
	status  int
}

func holdASignIn(t *testing.T, provider *fakeProvider, idp, email string) *heldSignIn {
	t.Helper()

	state, code := provider.authorize(t, startFlow(t, idpLoginEndpoint, idp, nil), person("sub-"+email, email, true))

	hold, reached := make(chan struct{}), make(chan struct{})

	provider.mu.Lock()
	provider.hold, provider.reached = hold, reached
	provider.mu.Unlock()

	signIn := &heldSignIn{done: make(chan struct{}), release: sync.OnceFunc(func() { close(hold) })}

	go func() {
		defer close(signIn.done)

		signIn.status, signIn.body = callback(t, idp, state, code)
	}()

	// Registered after the provider's server, so run before it is closed:
	// Close waits for its handlers, and a held one would never end.
	t.Cleanup(func() {
		signIn.release()

		select {
		case <-signIn.done:
		case <-time.After(15 * time.Second):
		}
	})

	select {
	case <-reached:
	case <-time.After(15 * time.Second):
		t.Fatal("the sign-in never reached the provider's token endpoint")
	}

	return signIn
}

func (s *heldSignIn) answered() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *heldSignIn) wait(t *testing.T) (int, string) {
	t.Helper()

	select {
	case <-s.done:
	case <-time.After(15 * time.Second):
		t.Fatal("the sign-in did not answer")
	}

	return s.status, s.body
}

// assertNothingStored: neither an identity under the provider nor the
// account the sign-in would have provisioned.
func assertNothingStored(t *testing.T, idp, email string) {
	t.Helper()

	var identities, accounts int
	require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT count(*) FROM users_identities WHERE idps_id = $1`, idp).Scan(&identities))
	require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE email = $1`, email).Scan(&accounts))
	assert.Zero(t, identities, "an identity is stored under the repointed provider")
	assert.Zero(t, accounts, "the account was stored without its identity")
}

// TestASignInStoresNothingUnderAProviderRepointedMeanwhile: a provider may be
// pointed at another issuer while nobody is linked to it. A first sign-in
// that was already under way, with the provider as it read it, stored the
// account's identity under the provider as it is now: a subject the old
// issuer proved, at a provider that now trusts another issuer, which then
// signs in as that account by asserting it. The identity is stored only
// while the provider's row still has the issuer the sign-in used, so this
// sign-in stores nothing, account included, and is refused (401).
func TestASignInStoresNothingUnderAProviderRepointedMeanwhile(t *testing.T) {
	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	before, after := newFakeProvider(t), newFakeProvider(t)
	idp := createOIDCIDP(t, admin.AccessToken, before, true)

	_, _, email := generateUserData(t)
	t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

	signIn := holdASignIn(t, before, idp, email)

	// Nobody is linked to the provider: pointing it elsewhere is allowed.
	response, err := sendHTTPRequest(t, context.WithoutCancel(t.Context()), idpsUpdateEndpoint.Clone().RewriteSlugs(idp),
		map[string]any{"issuer_url": after.srv.URL}, map[string]string{"Authorization": "Bearer " + admin.AccessToken})
	require.NoError(t, err)

	text := readResponseBody(t, response)
	response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode, text)

	signIn.release()

	status, body := signIn.wait(t)
	assert.Equal(t, http.StatusUnauthorized, status, "the sign-in stored an identity under a provider pointed elsewhere: %s", body)
	assert.Contains(t, body, changedDuringSignIn, "a 401 that is not this refusal")

	assertNothingStored(t, idp, email)
}

// TestASignInWaitsForARepointUnderWay: the other order. The repoint is
// played by a transaction of the test's own, as the update runs it (the
// provider's row FOR UPDATE, then the new issuer), held open while a first
// sign-in stores its identity. The insert waits for it, reads the row again
// once it commits, finds another issuer, and stores nothing. The foreign key
// alone waited too, and then found the provider there.
func TestASignInWaitsForARepointUnderWay(t *testing.T) {
	ctx := t.Context()

	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	before, after := newFakeProvider(t), newFakeProvider(t)
	idp := createOIDCIDP(t, admin.AccessToken, before, true)

	_, _, email := generateUserData(t)
	t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

	signIn := holdASignIn(t, before, idp, email)

	repoint, err := testDBPool.Begin(ctx)
	require.NoError(t, err)

	// Run first: the sign-in waits on this transaction, whatever failed.
	t.Cleanup(func() { _ = repoint.Rollback(context.Background()) })

	var pid int
	require.NoError(t, repoint.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid))

	_, err = repoint.Exec(ctx, `SELECT TRUE FROM idps WHERE id = $1 FOR UPDATE`, idp)
	require.NoError(t, err)

	_, err = repoint.Exec(ctx, `UPDATE idps SET issuer_url = $2 WHERE id = $1`, idp, after.srv.URL)
	require.NoError(t, err)

	signIn.release()

	require.Eventually(t, func() bool {
		if signIn.answered() {
			return true
		}

		var waiting bool

		err := testDBPool.QueryRow(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1 = ANY (pg_blocking_pids(pid)))`, pid).Scan(&waiting)

		return err == nil && waiting
	}, 10*time.Second, 50*time.Millisecond, "the sign-in neither waited on the repoint nor answered")

	if signIn.answered() {
		status, body := signIn.wait(t)
		t.Fatalf("the sign-in answered %d while the provider's row was locked: %s", status, body)
	}

	require.NoError(t, repoint.Commit(ctx))

	status, body := signIn.wait(t)
	assert.Equal(t, http.StatusUnauthorized, status, "the sign-in stored an identity under a provider repointed while it waited: %s", body)
	assert.Contains(t, body, changedDuringSignIn, "a 401 that is not this refusal")

	assertNothingStored(t, idp, email)
}

// TestARepointWaitsForAnIdentityBeingStored: the repoint's own statement
// refuses another issuer while accounts are linked, but it read "linked" on
// a snapshot, and an UPDATE that changes no key column does not wait for an
// identity's insert (its lock is not one the insert's KEY SHARE conflicts
// with): an identity stored in the same moment was not seen, and the
// provider was pointed elsewhere under it. The update takes the row FOR
// UPDATE first now, which waits for the insert, and then sees the identity.
//
// The insert is played by a transaction of the test's own, held open until
// the update is seen waiting on it.
func TestARepointWaitsForAnIdentityBeingStored(t *testing.T) {
	ctx := t.Context()

	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	before, after := newFakeProvider(t), newFakeProvider(t)
	idp := createOIDCIDP(t, admin.AccessToken, before, false)

	firstName, lastName, email := generateUserData(t)
	account := createUserInDB(t, firstName, lastName, email, generatePassword(t))
	t.Cleanup(func() { deleteUserByIDFromDB(t, account) })

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

	insert, err := testDBPool.Begin(ctx)
	require.NoError(t, err)

	t.Cleanup(func() { _ = insert.Rollback(context.Background()) })

	var pid int
	require.NoError(t, insert.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid))

	_, err = insert.Exec(ctx, `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
		account, idp, "sub-"+account.String(), email)
	require.NoError(t, err)

	sent = true

	go func() {
		response, err := sendHTTPRequest(t, context.Background(), idpsUpdateEndpoint.Clone().RewriteSlugs(idp),
			map[string]any{"issuer_url": after.srv.URL}, map[string]string{"Authorization": "Bearer " + admin.AccessToken})
		if err != nil {
			answered <- answer{err: err}

			return
		}

		defer response.Body.Close()

		answered <- answer{status: response.StatusCode, text: readResponseBody(t, response)}
	}()

	var got *answer

	// Waiting on the insert, or already answered (the code before waited
	// on nothing).
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
	}, 10*time.Second, 50*time.Millisecond, "the repoint neither waited on the identity's insert nor answered")

	require.NoError(t, insert.Commit(ctx))

	if got == nil {
		select {
		case a := <-answered:
			got = &a
		case <-time.After(15 * time.Second):
			t.Fatal("the repoint did not answer after the identity was stored")
		}
	}

	sent = false

	require.NoError(t, got.err)
	assert.Equal(t, http.StatusConflict, got.status, "the provider was pointed elsewhere under an identity stored in the same moment: %s", got.text)

	var issuer string
	require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT issuer_url FROM idps WHERE id = $1`, idp).Scan(&issuer))
	assert.Equal(t, before.srv.URL, issuer, "the provider was pointed elsewhere")
}

// TestASignInDoesNotSignInUnderAProviderRepointedMeanwhile: the read side of
// the same race. A sign-in that read the provider at one issuer, held there
// (or reading it from a cached copy), proves a subject with that issuer; the
// provider is pointed at another issuer meanwhile, nobody being linked, and
// an account is linked to the same subject under the new one. The identity
// was found by the provider and the subject alone, and the held sign-in was
// signed in as that account. It is found only under the provider as the
// sign-in used it now: this one knows no such identity, and the insert that
// would provision it refuses the provider as changed (401). Nobody is
// signed in, and the account keeps its identity.
func TestASignInDoesNotSignInUnderAProviderRepointedMeanwhile(t *testing.T) {
	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	before, after := newFakeProvider(t), newFakeProvider(t)
	idp := createOIDCIDP(t, admin.AccessToken, before, true)

	// The account linked under the new issuer, to the subject the old one
	// is about to assert.
	firstName, lastName, victimEmail := generateUserData(t)
	victim := createUserInDB(t, firstName, lastName, victimEmail, generatePassword(t))
	t.Cleanup(func() { deleteUserByIDFromDB(t, victim) })

	_, _, email := generateUserData(t)
	t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

	// holdASignIn asserts the subject "sub-"+email.
	signIn := holdASignIn(t, before, idp, email)

	response, err := sendHTTPRequest(t, context.WithoutCancel(t.Context()), idpsUpdateEndpoint.Clone().RewriteSlugs(idp),
		map[string]any{"issuer_url": after.srv.URL}, map[string]string{"Authorization": "Bearer " + admin.AccessToken})
	require.NoError(t, err)

	text := readResponseBody(t, response)
	response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode, text)

	_, err = testDBPool.Exec(context.Background(), `INSERT INTO users_identities (users_id, idps_id, subject, email) VALUES ($1, $2, $3, $4)`,
		victim, idp, "sub-"+email, victimEmail)
	require.NoError(t, err)

	signIn.release()

	status, body := signIn.wait(t)
	assert.Equal(t, http.StatusUnauthorized, status, "a subject the old issuer proved signed in as the account linked under the new one: %s", body)
	assert.NotContains(t, body, "access_token", "a session was answered")
	assert.Contains(t, body, changedDuringSignIn, "a 401 that is not this refusal")

	var identities int
	require.NoError(t, testDBPool.QueryRow(context.Background(), `SELECT count(*) FROM users_identities WHERE idps_id = $1`, idp).Scan(&identities))
	assert.Equal(t, 1, identities, "the account's identity is the provider's only one")
}
