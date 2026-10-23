package repositorypg

import (
	"strings"
	"testing"
)

// GET /products and GET /resources_limits have no project in their path, so
// the membership check holds them to nothing, and each answered every row of
// the deployment to whoever held its grant. Each is held to who asks in its
// statement: the products of the projects the projects list answers the
// caller, and the usage rows of the caller's projects and its own, every row
// for the administrator account.
//
// The unit suite has no database, so this reads each statement: its scope
// whole, spaces apart (a term added inside it, "OR TRUE", answers every row),
// the caller as its one argument, and a list asked for with no caller
// refused before anything is read.
func TestTheBareListsAreHeldToTheCaller(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		file  string
		scope string
		alias string
	}{
		{
			file: "products.go",
			scope: "FROM products AS prd LEFT JOIN projects AS prjs ON prd.projects_id = prjs.id " +
				"WHERE EXISTS ( SELECT 1 FROM view_projects_users AS callers " +
				"WHERE callers.user_id = $1 AND callers.id = prd.projects_id ) {{ .QueryWhere }} " +
				"GROUP BY prd.id, prjs.id ORDER BY {{.QueryInternalSort}} LIMIT {{.QueryLimit}} )",
			alias: "prd",
		},
		{
			file: "resources_limits.go",
			scope: "AND rls.scope_id = rl.scope_id WHERE ( " +
				"EXISTS (SELECT 1 FROM users AS caller WHERE caller.id = $1 AND caller.admin IS TRUE) " +
				"OR (rls.scope_type = 'user' AND rls.scope_id = $1) " +
				"OR (rls.scope_type = 'project' AND EXISTS ( SELECT 1 FROM projects_users AS membership " +
				"WHERE membership.projects_id = rls.scope_id AND membership.users_id = $1 )) ) {{ .QueryWhere }} " +
				"ORDER BY {{.QueryInternalSort}} LIMIT {{.QueryLimit}} )",
			alias: "rls",
		},
	} {
		t.Run(strings.TrimSuffix(tc.file, ".go"), func(t *testing.T) {
			t.Parallel()

			source := functionSource(t, tc.file, "Select")

			if !strings.Contains(strings.Join(strings.Fields(source), " "), tc.scope) {
				t.Errorf("the statement's scope changed; want, spaces apart:\n%s", tc.scope)
			}

			// The statement that runs is the template as it was rendered.
			if got := strings.Count(source, "query ="); got != 0 {
				t.Errorf("the rendered statement is assigned to %d times after it is rendered", got)
			}

			if got := strings.Count(source, "ref.db.Query("); got != 1 {
				t.Fatalf("the list runs %d statements, want one: read each one", got)
			}

			if !strings.Contains(source, "ref.db.Query(ctx, query, input.CallerID)") {
				t.Error("the statement is not run with the caller as $1")
			}

			// The filter and the pager attach to the scope; they never open
			// a WHERE of their own, which would follow the scope's: the
			// filter is written with AND, and the pager is told a WHERE is
			// there already.
			if strings.Contains(source, `"WHERE (%s)"`) {
				t.Error("the filter opens a WHERE of its own")
			}

			pager := "buildPaginationCriteria(\"" + tc.alias + "\", tokenDirection, id, serial, filterQuery, true)"
			if !strings.Contains(source, pager) {
				t.Errorf("the pager is not told the statement has a WHERE; want %s", pager)
			}

			refusal := strings.Index(source, "if input.CallerID == uuid.Nil() {")
			query := strings.Index(source, "ref.db.Query(")

			if refusal < 0 {
				t.Fatal("a list asked for with no caller is not refused")
			}

			if refusal > query {
				t.Error("the list is read before the caller is looked at")
			}
		})
	}
}
