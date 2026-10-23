package repositorypg

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// A provider may be pointed at another issuer or type while nobody is linked
// to it, and a sign-in already in flight, or one that read the provider from
// a cached copy, would then store an identity under the provider as it is
// now, for a subject another issuer asserted. Both sides take the provider's
// row: the identity's insert reads it with the issuer and kind the sign-in
// used, FOR KEY SHARE, in the statement that inserts; the update takes it FOR
// UPDATE, which conflicts with KEY SHARE, before its own statement. The pull
// request gate has no database: the statements are held whole, and the
// update's order is read from its body.
func TestAnIdentityIsStoredOnlyUnderTheProviderItSignedInWith(t *testing.T) {
	t.Parallel()

	flat := func(s string) string { return strings.Join(strings.Fields(s), " ") }

	for statement, want := range map[string]string{
		queryLinkIdentity: "INSERT INTO users_identities (users_id, idps_id, subject, email) SELECT $1, idps.id, $3, $4 FROM idps " +
			"WHERE idps.id = $2 AND idps.idp_types = $6 AND COALESCE(idps.issuer_url, '') = $5 FOR KEY SHARE OF idps;",
		queryLockIDPForUpdate: "SELECT TRUE FROM idps WHERE id = $1 FOR UPDATE;",
	} {
		if got := flat(statement); got != want {
			t.Errorf("a statement changed:\n got: %s\nwant: %s", got, want)
		}
	}

	runs := statementsRunBy(t, "idps.go", "UpdateByID")

	want := []statementRun{{on: "tx", statement: "queryLockIDPForUpdate"}, {on: "tx", statement: "queryUpdateIDP"}}
	if !slices.Equal(runs, want) {
		t.Errorf("UpdateByID runs %v, want the provider's lock and then the update, both in its transaction: %v", runs, want)
	}

	// The issuer is $5 and the type $6 at both inserts: two uuids or two
	// strings swapped would compile, and refuse every sign-in.
	for file, call := range map[string]string{
		"users_identities.go": "queryLinkIdentity, input.UserID, input.IDPID, input.Subject, input.Email, input.IssuerURL, input.IDPTypeID)",
		"users.go":            "queryLinkIdentity, identity.UserID, identity.IDPID, identity.Subject, identity.Email, identity.IssuerURL, identity.IDPTypeID)",
	} {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(b), call) {
			t.Errorf("%s does not insert the identity with its arguments in the statement's order: want %q", file, call)
		}
	}
}
