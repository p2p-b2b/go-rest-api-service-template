package repositorypg

import (
	"strings"
	"testing"
)

// A role's delete is refused while an account with the `admin` flag holds
// it, and the refusal rests on this statement alone: one that answers no
// row turns the guard off, and the use-case's tests, which answer it with a
// double, stay green. The pull request gate has no database, so the
// statement is held whole, spaces apart: which role, the flag, and a system
// role left to its trigger.
func TestTheHoldersOfARoleAreReadByTheRoleAndTheFlag(t *testing.T) {
	t.Parallel()

	const want = "SELECT users.id FROM users " +
		"JOIN users_roles ON users_roles.users_id = users.id " +
		"JOIN roles ON roles.id = users_roles.roles_id " +
		"WHERE users_roles.roles_id = $1 AND users.admin IS TRUE AND NOT roles.system " +
		"ORDER BY users.id;"

	if got := strings.Join(strings.Fields(queryAdministratorsHolding), " "); got != want {
		t.Errorf("the statement changed:\n got: %s\nwant: %s", got, want)
	}

	// And its sibling, which the unlink of a role from a set of users asks.
	if got, want := strings.Join(strings.Fields(queryAdministratorsAmong), " "),
		"SELECT id FROM users WHERE id = ANY($1::uuid[]) AND admin IS TRUE ORDER BY id;"; got != want {
		t.Errorf("the statement changed:\n got: %s\nwant: %s", got, want)
	}
}
