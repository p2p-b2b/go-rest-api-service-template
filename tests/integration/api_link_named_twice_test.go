//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestALinkThatNamesAnIdTwiceLinksItOnce: the bulk links are one statement,
// INSERT ... SELECT FROM UNNEST(...) ON CONFLICT ... DO UPDATE, and Postgres
// refuses to update one row twice in one statement (SQLSTATE 21000). A body
// that named an id twice was a 500. Naming it twice is naming it: one link,
// a 200. (A user's projects had it fixed already; these are the four others.)
func TestALinkThatNamesAnIdTwiceLinksItOnce(t *testing.T) {
	ctx := t.Context()

	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	headers := map[string]string{"Authorization": "Bearer " + admin.AccessToken}

	firstName, lastName, email := generateUserData(t)
	user := createUserInDB(t, firstName, lastName, email, generatePassword(t))
	t.Cleanup(func() { deleteUserByIDFromDB(t, user) })

	project := createProjectInDB(t, generateRandomName(t, "Twice"), "a link named twice")
	t.Cleanup(func() { deleteProjectByIDFromDB(t, project.ID) })

	role, policy := uuid.NewV7(), uuid.NewV7()

	_, err := testDBPool.Exec(ctx, `INSERT INTO roles (id, name, description) VALUES ($1, $2, 'a link named twice')`, role, "twice-"+role.String()[:8])
	require.NoError(t, err)

	var resource uuid.UUID
	require.NoError(t, testDBPool.QueryRow(ctx, `SELECT id FROM resources WHERE action = 'GET' ORDER BY id LIMIT 1`).Scan(&resource))

	_, err = testDBPool.Exec(ctx,
		`INSERT INTO policies (id, resources_id, name, description, allowed_action, allowed_resource)
		 VALUES ($1, $2, $3, 'a link named twice', 'GET', '/version')`,
		policy, resource, "twice-"+policy.String()[:8])
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = testDBPool.Exec(context.Background(), `DELETE FROM roles WHERE id = $1`, role)
		_, _ = testDBPool.Exec(context.Background(), `DELETE FROM policies WHERE id = $1`, policy)
	})

	links := func(t *testing.T, statement string, args ...any) int {
		t.Helper()

		var n int
		require.NoError(t, testDBPool.QueryRow(context.Background(), statement, args...).Scan(&n))

		return n
	}

	for _, c := range []struct {
		name     string
		endpoint *apiEndpoint
		slug     string
		body     map[string]any
		count    string
		args     []any
	}{
		{
			name: "a_projects_users", endpoint: projectsLinkUsersEndpoint, slug: project.ID.String(),
			body:  map[string]any{"user_ids": []uuid.UUID{user, user}},
			count: `SELECT count(*) FROM projects_users WHERE projects_id = $1 AND users_id = $2`, args: []any{project.ID, user},
		},
		{
			name: "a_roles_users", endpoint: rolesLinkUsersEndpoint, slug: role.String(),
			body:  map[string]any{"user_ids": []uuid.UUID{user, user}},
			count: `SELECT count(*) FROM users_roles WHERE roles_id = $1 AND users_id = $2`, args: []any{role, user},
		},
		{
			name: "a_roles_policies", endpoint: rolesLinkPoliciesEndpoint, slug: role.String(),
			body:  map[string]any{"policy_ids": []uuid.UUID{policy, policy}},
			count: `SELECT count(*) FROM roles_policies WHERE roles_id = $1 AND policies_id = $2`, args: []any{role, policy},
		},
		{
			name: "a_users_roles", endpoint: usersLinkRolesEndpoint, slug: user.String(),
			body:  map[string]any{"role_ids": []uuid.UUID{role, role}},
			count: `SELECT count(*) FROM users_roles WHERE users_id = $1 AND roles_id = $2`, args: []any{user, role},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), c.endpoint.Clone().RewriteSlugs(c.slug), c.body, headers)
			require.NoError(t, err)

			text := readResponseBody(t, response)
			response.Body.Close()

			assert.Equal(t, http.StatusOK, response.StatusCode, "a body that names an id twice: %s", text)
			assert.Equal(t, 1, links(t, c.count, c.args...), "the link is not there once")
		})
	}
}
