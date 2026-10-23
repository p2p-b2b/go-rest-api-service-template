//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/payload"
)

// refuseIdentity makes the database refuse to store the identity of one
// subject, in the way `raise` says, until the function it returns is called.
//
// The identity's statement of a first sign-in cannot be made to fail from
// the outside: the identity is new, so nothing it could collide with exists
// before the request, and the ways it does fail (a fault, a link made by a
// concurrent request, the provider deleted meanwhile) happen inside the
// account's transaction. A trigger is the one place a test can stand.
//
// It fires for the one subject and nothing else, so no other test's link can
// meet it.
func refuseIdentity(t *testing.T, subject, raise string) (allow func()) {
	t.Helper()

	require.NotContains(t, subject, "'", "the subject is written into the trigger's condition")

	// The end of a version 7 uuid is its random part; the start is the time.
	id := strings.ReplaceAll(mustUUIDString(t), "-", "")
	name := "test_refuse_identity_" + id[len(id)-16:]
	quoted := pgx.Identifier{name}.Sanitize()

	// Not the test's context: the trigger has to go when the test has ended,
	// and by then that context is cancelled.
	ctx := context.Background()

	_, err := testDBPool.Exec(ctx, fmt.Sprintf(`
        CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $body$
        BEGIN
            %s;
        END
        $body$;`, quoted, raise))
	require.NoError(t, err)

	_, err = testDBPool.Exec(ctx, fmt.Sprintf(`
        CREATE TRIGGER %s BEFORE INSERT ON users_identities
        FOR EACH ROW WHEN (NEW.subject = '%s') EXECUTE FUNCTION %s();`, quoted, subject, quoted))
	require.NoError(t, err)

	allowed := false

	allow = func() {
		if allowed {
			return
		}

		allowed = true

		_, err := testDBPool.Exec(ctx, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON users_identities;`, quoted))
		assert.NoError(t, err)

		_, err = testDBPool.Exec(ctx, fmt.Sprintf(`DROP FUNCTION IF EXISTS %s();`, quoted))
		assert.NoError(t, err)
	}

	t.Cleanup(allow)

	return allow
}

// usersCounted is the system's count of accounts, the one the users limit is
// held against. The row is written by the first account the API creates, so
// a database that has none yet counts none.
func usersCounted(t *testing.T) int {
	t.Helper()

	var usage int

	err := testDBPool.QueryRow(context.Background(),
		`SELECT usage FROM resources_usage WHERE scope_type = 'system' AND resource_type = 'users'`).Scan(&usage)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0
	}

	require.NoError(t, err)

	return usage
}

// TestIDPFlowAFirstSignInThatCannotStoreItsIdentityLeavesNoAccount: a first
// sign-in through a provider stores the account and the identity that says
// whose it is.
//
// They were two writes. When the second failed, the first stayed: an account
// with no password and no identity, which nobody can sign in to, holding the
// address and a slot of the users limit, and the person's next sign-in was
// refused as the takeover the flow refuses, every time, until an
// administrator deleted the account. Now the two are one transaction: an
// identity that cannot be stored stores no account. The sign-in that failed
// still says why, and the next one works.
//
// Not parallel: it reads the count of accounts before and after.
func TestIDPFlowAFirstSignInThatCannotStoreItsIdentityLeavesNoAccount(t *testing.T) {
	admin := getAdminUserTokens(t)
	f := newFakeProvider(t)
	idpID := createOIDCIDP(t, admin.AccessToken, f, true)

	cases := map[string]struct {
		raise  string
		status int
	}{
		// Anything the database answers that the service has no words for.
		"a_fault": {
			raise:  `RAISE EXCEPTION 'staged by the test: the identity cannot be stored'`,
			status: http.StatusInternalServerError,
		},
		// What a concurrent request that stored the identity first leaves
		// this one with: the primary key's violation, by its name.
		"the_identity_was_linked_meanwhile": {
			raise:  `RAISE EXCEPTION USING ERRCODE = 'unique_violation', CONSTRAINT = 'pk_users_identities', MESSAGE = 'staged by the test'`,
			status: http.StatusConflict,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, email := generateUserData(t)
			t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

			subject := "sub-" + mustUUIDString(t)
			allow := refuseIdentity(t, subject, c.raise)

			before := usersCounted(t)

			state, code := f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person(subject, email, true))

			status, body := callback(t, idpID, state, code)
			require.Equal(t, c.status, status, body)

			// Nothing of the sign-in is left: no account, and the slot the
			// account took is back. (The identity was never stored: that is
			// what the trigger is for.)
			var accounts int

			require.NoError(t, testDBPool.QueryRow(t.Context(), `SELECT count(*) FROM users WHERE email = $1`, email).Scan(&accounts))

			assert.Zero(t, accounts, "the account was left behind, with no password and no identity")
			assert.Equal(t, before, usersCounted(t), "the account's slot in the users limit was not given back")

			// And the same person signs in the next time.
			allow()

			state, code = f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person(subject, email, true))

			status, body = callback(t, idpID, state, code)
			require.Equal(t, http.StatusOK, status, "the sign-in after the failed one was refused: %s", body)

			var out payload.IDPCallbackResponse

			require.NoError(t, json.Unmarshal([]byte(body), &out))
			require.NotNil(t, out.Login, "a login answers a session")
			assert.NotEmpty(t, out.Login.AccessToken)

			assert.Equal(t, before+1, usersCounted(t), "one account, counted once")
		})
	}
}
