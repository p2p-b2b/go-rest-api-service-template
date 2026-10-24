//go:build integration

package integration

import (
	"net/http"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

// TestAMissingProjectIsA404ForTheAdministratorAccount holds the flag to a
// project that is there. It admitted to any id, so every list under a
// project that does not exist answered the administrator account 200 and an
// empty list, while a create under the same path answered 404. Each route is
// asked under a project that is there as well, so that a 404 is the missing
// project's and not the route's.
func TestAMissingProjectIsA404ForTheAdministratorAccount(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	project := createProjectInDB(t, generateRandomName(t, "missing-proof"), "the project that is there")
	t.Cleanup(func() { deleteProjectByIDFromDB(t, project.ID) })

	routes := map[string]*apiEndpoint{
		"products":         productsListEndpoint,
		"users":            newAPIEndpoint(http.MethodGet, "/projects/{project_id}/users"),
		"resources_limits": resourcesLimitsProjectEndpoint,
	}

	call := func(t *testing.T, ep *apiEndpoint) (int, string) {
		t.Helper()

		resp, err := sendHTTPRequest(t, ctx, ep, nil, map[string]string{"Authorization": "Bearer " + admin.AccessToken})
		require.NoError(t, err)
		defer resp.Body.Close()

		return resp.StatusCode, readResponseBody(t, resp)
	}

	t.Run("the_nil_project_is_a_400", func(t *testing.T) {
		t.Parallel()

		// It parses, and is refused as an invalid input; it was a 500.
		status, body := call(t, productsListEndpoint.RewriteSlugs(uuid.Nil().String()))
		require.Equal(t, http.StatusBadRequest, status, body)
	})

	for name, route := range routes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status, body := call(t, route.RewriteSlugs(project.ID.String()))
			require.Equal(t, http.StatusOK, status, "under a project that is there: %s", body)

			missing := uuid.NewV7().String()
			status, body = call(t, route.RewriteSlugs(missing))
			require.Equal(t, http.StatusNotFound, status, "under a project that is not there: %s", body)
			require.Contains(t, body, missing, "the answer names the project, as a non-member's does")
		})
	}
}
