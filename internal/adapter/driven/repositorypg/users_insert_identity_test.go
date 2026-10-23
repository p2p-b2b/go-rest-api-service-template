package repositorypg

import (
	"slices"
	"strings"
	"testing"
)

// A first sign-in stores the account and its identity in one transaction:
// they were two writes, and an identity that failed after the account was
// stored left an account nobody could sign in to. The pull request gate has
// no database: Insert's statements are read from its body, all three on its
// transaction, the identity's last, and the identity's statement is held
// whole, as Link runs it too.
func TestAnAccountIsStoredWithItsIdentity(t *testing.T) {
	t.Parallel()

	const want = "INSERT INTO users_identities (users_id, idps_id, subject, email) SELECT $1, idps.id, $3, $4 FROM idps " +
		"WHERE idps.id = $2 AND idps.idp_types = $6 AND COALESCE(idps.issuer_url, '') = $5 FOR KEY SHARE OF idps;"
	if got := strings.Join(strings.Fields(queryLinkIdentity), " "); got != want {
		t.Errorf("the identity's statement changed:\n got: %s\nwant: %s", got, want)
	}

	runs := statementsRunBy(t, "users.go", "Insert")

	wantRuns := []statementRun{{on: "tx", statement: "query1"}, {on: "tx", statement: "query2"}, {on: "tx", statement: "queryLinkIdentity"}}
	if !slices.Equal(runs, wantRuns) {
		t.Errorf("Insert runs %v, want the account, its roles and its identity, all in its transaction: %v", runs, wantRuns)
	}

	if link := statementsRunBy(t, "users_identities.go", "Link"); !slices.Equal(link, []statementRun{{on: "db", statement: "queryLinkIdentity"}}) {
		t.Errorf("Link runs %v, want the one identity statement", link)
	}
}
