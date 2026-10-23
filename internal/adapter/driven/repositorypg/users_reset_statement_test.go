package repositorypg

import (
	"strings"
	"testing"
)

// A reset link writes its password only on the account it was mailed to, as
// the account is now, and the statement is the whole of that rule: the
// use-case's tests answer it with a double. The pull request gate has no
// database, so the statement is held whole, spaces apart: a condition
// dropped from it (the address, a local account, not disabled) still
// compiles and writes a password it should not.
func TestAResetLinkWritesOnlyTheAccountItWasMailedTo(t *testing.T) {
	t.Parallel()

	const want = "UPDATE users SET password_hash = $3, updated_at = CURRENT_TIMESTAMP " +
		"WHERE id = $1 AND email = $2 AND local_account IS TRUE AND disabled IS FALSE;"

	if got := strings.Join(strings.Fields(queryResetPassword), " "); got != want {
		t.Errorf("the statement changed:\n got: %s\nwant: %s", got, want)
	}
}
