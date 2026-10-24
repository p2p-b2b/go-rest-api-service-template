package repositorypg

import (
	"strings"
	"testing"
)

// The admin flag admits to a project that is there, and the statement is the
// whole of that rule: the middleware's tests answer it with a double. The
// pull request gate has no database, so the statement is held whole, spaces
// apart: the project's EXISTS dropped from it still compiles, and every route
// under a project that is not there answers the administrator account 200.
func TestTheFlagAdmitsOnlyToAProjectThatIsThere(t *testing.T) {
	t.Parallel()

	const want = "SELECT COALESCE((SELECT admin FROM users WHERE id = $2), FALSE) " +
		"AND EXISTS (SELECT 1 FROM projects WHERE id = $1) AS is_admin, " +
		"EXISTS ( SELECT 1 FROM projects_users WHERE projects_id = $1 AND users_id = $2 ) AS is_member;"

	if got := strings.Join(strings.Fields(querySelectMembership), " "); got != want {
		t.Errorf("the statement changed:\n got: %s\nwant: %s", got, want)
	}
}
