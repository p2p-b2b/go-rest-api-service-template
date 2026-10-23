//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bareListScope is what one test of a bare list needs: the administrator
// account, two projects, and a caller that is a member of one of them and
// holds the role the list's grant was written for.
type bareListScope struct {
	asAdministrator map[string]string
	member          map[string]string
	mine, theirs    uuid.UUID
	memberID        uuid.UUID
	otherID         uuid.UUID
}

func newBareListScope(t *testing.T, role string) bareListScope {
	t.Helper()

	administrator := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, administrator.UserID) })

	mine := createProjectInDB(t, generateRandomName(t, "mine"), "the project the caller is a member of")
	t.Cleanup(func() { deleteProjectByIDFromDB(t, mine.ID) })

	theirs := createProjectInDB(t, generateRandomName(t, "theirs"), "a project the caller is not a member of")
	t.Cleanup(func() { deleteProjectByIDFromDB(t, theirs.ID) })

	firstName, lastName, email := generateUserData(t)
	password := generatePassword(t)
	memberID := createUserInDB(t, firstName, lastName, email, password)
	t.Cleanup(func() { deleteUserByIDFromDB(t, memberID) })
	assignRoleToUserInDB(t, memberID, role)

	otherFirst, otherLast, otherEmail := generateUserData(t)
	otherID := createUserInDB(t, otherFirst, otherLast, otherEmail, generatePassword(t))
	t.Cleanup(func() { deleteUserByIDFromDB(t, otherID) })

	_, err := testDBPool.Exec(t.Context(), `INSERT INTO projects_users (projects_id, users_id) VALUES ($1, $2)`, mine.ID, memberID)
	require.NoError(t, err)

	return bareListScope{
		asAdministrator: map[string]string{"Authorization": "Bearer " + administrator.AccessToken},
		member:          map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken},
		mine:            mine.ID,
		theirs:          theirs.ID,
		memberID:        memberID,
		otherID:         otherID,
	}
}

// listItems asks for a list filtered to this test's own rows and returns
// its items as they were answered.
func listItems[T any](t *testing.T, path, filter string, headers map[string]string) []T {
	t.Helper()

	endpoint := newAPIEndpoint(http.MethodGet, path)
	endpoint.SetQueryParam("filter", filter)
	endpoint.SetQueryParam("limit", "100")

	response, err := sendHTTPRequest(t, context.WithoutCancel(t.Context()), endpoint, nil, headers)
	require.NoError(t, err)

	text := readResponseBody(t, response)
	response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode, text)

	var page struct {
		Items []T `json:"items"`
	}

	require.NoError(t, json.Unmarshal([]byte(text), &page), text)

	return page.Items
}

// TestTheProductsListIsTheCallers: GET /products answered every project's
// products to whoever held its grant, and the seeded ProjectMember role,
// whose work is inside the projects it is a member of, holds it. It answers
// what the projects list answers now: every project's products to the
// administrator account, the products of the projects one is a member of to
// anyone else.
func TestTheProductsListIsTheCallers(t *testing.T) {
	scope := newBareListScope(t, "ProjectMember")

	// One product of the same name in each project: the name is this test's
	// own, so the filter finds these two and nothing else.
	probe := "bare-list-" + uuid.NewV7().String()

	for _, project := range []uuid.UUID{scope.mine, scope.theirs} {
		_, err := testDBPool.Exec(t.Context(), `INSERT INTO products (projects_id, name, description) VALUES ($1, $2, 'a probe')`, project, probe)
		require.NoError(t, err)
	}

	type item struct {
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	}

	listed := func(t *testing.T, headers map[string]string) []string {
		t.Helper()

		var got []string
		for _, product := range listItems[item](t, "/products", fmt.Sprintf("name = '%s'", probe), headers) {
			got = append(got, product.Project.ID)
		}

		return got
	}

	t.Run("a_member_is_answered_its_projects_products", func(t *testing.T) {
		assert.ElementsMatch(t, []string{scope.mine.String()}, listed(t, scope.member),
			"another project's product was answered to a member")
	})

	t.Run("the_administrator_account_is_answered_every_projects_products", func(t *testing.T) {
		assert.ElementsMatch(t, []string{scope.mine.String(), scope.theirs.String()}, listed(t, scope.asAdministrator))
	})
}

// TestTheUsageListIsTheCallers: GET /resources_limits answered every usage
// row of the deployment, with each project's and each user's id, to whoever
// held its grant (the seeded ReadOnly role is enough). It is held to who
// asks now: the administrator account is answered every row; anyone else a
// project's rows for the projects they are a member of, and their own
// user's rows. The system's rows, the deployment's totals, are the
// administrator account's alone.
//
// One usage row in each scope, under a resource type of this test's own so
// that the counters the service keeps are not touched.
func TestTheUsageListIsTheCallers(t *testing.T) {
	scope := newBareListScope(t, "ReadOnly")

	// The whole uuid: its first digits are its timestamp's, and repeat for a
	// minute.
	probe := "usage-scope-" + uuid.NewV7().String()

	rows := map[string]string{
		"project:" + scope.mine.String():   "project",
		"project:" + scope.theirs.String(): "project",
		"user:" + scope.memberID.String():  "user",
		"user:" + scope.otherID.String():   "user",
		"system:" + uuid.Nil().String():    "system",
	}

	t.Cleanup(func() {
		_, _ = testDBPool.Exec(context.Background(), `DELETE FROM resources_usage WHERE resource_type = $1`, probe)
	})

	for key, scopeType := range rows {
		_, err := testDBPool.Exec(t.Context(), `INSERT INTO resources_usage (scope_type, scope_id, resource_type, usage) VALUES ($1, $2, $3, 1)`,
			scopeType, key[len(scopeType)+1:], probe)
		require.NoError(t, err)
	}

	type item struct {
		ScopeType string `json:"scope_type"`
		ScopeID   string `json:"scope_id"`
	}

	listed := func(t *testing.T, headers map[string]string) []string {
		t.Helper()

		var got []string
		for _, row := range listItems[item](t, "/resources_limits", fmt.Sprintf("resource_type = '%s'", probe), headers) {
			got = append(got, row.ScopeType+":"+row.ScopeID)
		}

		return got
	}

	t.Run("a_member_is_answered_its_projects_and_its_own_rows", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"project:" + scope.mine.String(), "user:" + scope.memberID.String()}, listed(t, scope.member),
			"another project's row, another user's or the system's was answered to a member")
	})

	t.Run("the_administrator_account_is_answered_every_row", func(t *testing.T) {
		var want []string
		for key := range rows {
			want = append(want, key)
		}

		assert.ElementsMatch(t, want, listed(t, scope.asAdministrator))
	})
}
