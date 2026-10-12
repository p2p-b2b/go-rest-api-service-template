//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seededAdministrator is the account migration 00003 creates, with the
// `admin` flag: its projects are every project.
var seededAdministrator = uuid.MustParse("019822af-b448-73fb-89a1-447e8f8d1cde")

// TestAnotherUsersProjectsAreOnlyTheCallersToo: GET /users/{user_id}/projects
// answers the projects of the user in its path, and has no project in its
// path for a membership check to hold it to. So it is scoped to who asks: a
// caller sees a project of another user only when the caller is linked to
// that project too. The list is the path's user's projects that are also the
// caller's: all of them for the administrator account and for a caller asking
// about themselves, the shared ones for anyone else.
//
// Not parallel: it counts on the projects it creates.
func TestAnotherUsersProjectsAreOnlyTheCallersToo(t *testing.T) {
	ctx := t.Context()

	shared := createProjectInDB(t, generateRandomName(t, "shared"), "a project both are members of")
	theirs := createProjectInDB(t, generateRandomName(t, "theirs"), "a project only the other user is a member of")
	mine := createProjectInDB(t, generateRandomName(t, "mine"), "a project only the caller is a member of")

	account := func(role string, projects ...uuid.UUID) (uuid.UUID, map[string]string) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id := createUserInDB(t, firstName, lastName, email, password)
		assignRoleToUserInDB(t, id, role)

		for _, project := range projects {
			_, err := testDBPool.Exec(ctx, `INSERT INTO projects_users (projects_id, users_id) VALUES ($1, $2)`, project, id)
			require.NoError(t, err)
		}

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		return id, map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}
	}

	// Registered before the accounts' cleanups, so it runs after them.
	t.Cleanup(func() {
		for _, project := range []uuid.UUID{shared.ID, theirs.ID, mine.ID} {
			deleteProjectByIDFromDB(t, project)
		}
	})

	callerID, caller := account("ReadOnly", shared.ID, mine.ID)
	otherID, other := account("ReadOnly", shared.ID, theirs.ID)
	_, outsider := account("ReadOnly")

	// The Administrator role grants every route and names no project; the
	// `admin` flag is another thing, and only the flag is "every project".
	_, roleOnly := account("Administrator", shared.ID)

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	type page struct {
		Message string `json:"message"`
		Items   []struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
		} `json:"items"`
		Paginator struct {
			NextToken string `json:"next_token"`
		} `json:"paginator"`
	}

	// The projects of `user` as `headers` is answered them, narrowed to the
	// three this test made (the Default Project and other tests' are there
	// for whoever may see them).
	list := func(t *testing.T, user uuid.UUID, headers map[string]string, params map[string]string) page {
		t.Helper()

		e := projectsListByUserIDEndpoint.RewriteSlugs(user.String())

		q := map[string]string{"limit": "100"}
		for k, v := range params {
			q[k] = v
		}

		e.SetQueryParams(q)

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), e, nil, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		require.Equal(t, http.StatusOK, response.StatusCode, text)

		var p page

		require.NoError(t, json.Unmarshal([]byte(text), &p), text)

		return p
	}

	made := map[uuid.UUID]string{shared.ID: "shared", theirs.ID: "theirs", mine.ID: "mine"}

	ofThisTest := func(p page) []string {
		var out []string

		for _, item := range p.Items {
			if name, ok := made[item.ID]; ok {
				out = append(out, name)
			}
		}

		return out
	}

	t.Run("another_users_projects_are_the_ones_the_caller_is_linked_to", func(t *testing.T) {
		got := list(t, otherID, caller, nil)

		assert.ElementsMatch(t, []string{"shared"}, ofThisTest(got),
			"the other user's projects, as answered to a caller who shares one of them")

		for _, item := range got.Items {
			assert.NotEqual(t, theirs.ID, item.ID, "a project the caller is not linked to, by name: %s", item.Name)
		}
	})

	t.Run("the_seeded_administrators_list_is_not_every_project", func(t *testing.T) {
		// An account with the `admin` flag has every project; what another
		// account reads of them is still only its own.
		got := list(t, seededAdministrator, caller, nil)

		assert.ElementsMatch(t, []string{"shared", "mine"}, ofThisTest(got),
			"the administrator's projects, as answered to a member of two of them")
	})

	t.Run("a_filter_stays_inside_it", func(t *testing.T) {
		filtered := list(t, otherID, caller, map[string]string{"filter": "name = '" + theirs.Name + "'"})
		assert.Empty(t, filtered.Items, "a filter found a project the caller is not linked to")

		own := list(t, otherID, caller, map[string]string{"filter": "name = '" + shared.Name + "'"})
		assert.Len(t, own.Items, 1)
	})

	t.Run("a_sorted_walk_stays_inside_it", func(t *testing.T) {
		var walked []string

		params := map[string]string{"sort": "name ASC", "limit": "1"}

		for range 50 {
			got := list(t, seededAdministrator, caller, params)
			walked = append(walked, ofThisTest(got)...)

			for _, item := range got.Items {
				assert.NotEqual(t, theirs.ID, item.ID)
			}

			if got.Paginator.NextToken == "" {
				break
			}

			params["next_token"] = got.Paginator.NextToken
		}

		assert.ElementsMatch(t, []string{"shared", "mine"}, walked)
	})

	t.Run("a_caller_reads_their_own_list_whole", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"shared", "mine"}, ofThisTest(list(t, callerID, caller, nil)))
		assert.ElementsMatch(t, []string{"shared", "theirs"}, ofThisTest(list(t, otherID, other, nil)))
	})

	t.Run("a_member_of_nothing_reads_nobodys_projects", func(t *testing.T) {
		assert.Empty(t, list(t, otherID, outsider, nil).Items)
		assert.Empty(t, list(t, seededAdministrator, outsider, nil).Items)
	})

	t.Run("the_administrator_role_without_the_flag_is_answered_like_anyone", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"shared"}, ofThisTest(list(t, otherID, roleOnly, nil)))
		assert.ElementsMatch(t, []string{"shared"}, ofThisTest(list(t, seededAdministrator, roleOnly, nil)))
	})

	t.Run("the_administrator_account_reads_a_users_whole_list", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"shared", "theirs"}, ofThisTest(list(t, otherID, admin, nil)))
		assert.ElementsMatch(t, []string{"shared", "mine"}, ofThisTest(list(t, callerID, admin, nil)))
	})
}
