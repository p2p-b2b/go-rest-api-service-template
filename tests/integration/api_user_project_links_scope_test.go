//go:build integration

package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAUsersProjectLinksChangeOnlyOnTheCallersProjects: POST and DELETE
// /users/{user_id}/projects take the projects from the body, and have no
// project in their path for a membership check to hold them to. So the ids
// are held to who asks: a caller links or unlinks a user only on a project
// the caller is linked to, every project for the administrator account.
// Another account's project and one that does not exist refuse the whole
// request, in the same words.
//
// Not parallel: it reads the links it made.
func TestAUsersProjectLinksChangeOnlyOnTheCallersProjects(t *testing.T) {
	ctx := t.Context()

	mine := createProjectInDB(t, generateRandomName(t, "mine"), "a project the caller is a member of")
	theirs := createProjectInDB(t, generateRandomName(t, "theirs"), "a project the caller is not a member of")
	missing := uuid.NewV7()

	// An account, in the projects given. With a role it signs in, and its
	// headers come back; without one it is only somebody whose links change.
	account := func(role string, projects ...uuid.UUID) (uuid.UUID, map[string]string) {
		firstName, lastName, email := generateUserData(t)
		password := generatePassword(t)
		id := createUserInDB(t, firstName, lastName, email, password)

		for _, project := range projects {
			_, err := testDBPool.Exec(ctx, `INSERT INTO projects_users (projects_id, users_id) VALUES ($1, $2)`, project, id)
			require.NoError(t, err)
		}

		t.Cleanup(func() { deleteUserByIDFromDB(t, id) })

		if role == "" {
			return id, nil
		}

		assignRoleToUserInDB(t, id, role)

		return id, map[string]string{"Authorization": "Bearer " + loginForMembership(t, email, password).AccessToken}
	}

	// Registered before the accounts' cleanups, so it runs after them.
	t.Cleanup(func() {
		deleteProjectByIDFromDB(t, mine.ID)
		deleteProjectByIDFromDB(t, theirs.ID)
	})

	// The caller manages users, is a member of one project, and has no flag.
	managerID, manager := account("UserManager", mine.ID)

	// Two more callers with the routes' grants: one with the Administrator
	// role and no `admin` flag, a member of the same one project, and one
	// that manages users and is a member of nothing.
	_, roleOnly := account("Administrator", mine.ID)
	_, memberOfNothing := account("UserManager")

	// Whose links change: one in neither project, one in the other project,
	// and one for the request that must write nothing.
	newcomer, _ := account("")
	theirMember, _ := account("", theirs.ID)
	untouched, _ := account("")

	adminTokens := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, adminTokens.UserID) })

	admin := map[string]string{"Authorization": "Bearer " + adminTokens.AccessToken}

	linked := func(t *testing.T, user, project uuid.UUID) bool {
		t.Helper()

		var is bool

		require.NoError(t, testDBPool.QueryRow(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM projects_users WHERE users_id = $1 AND projects_id = $2)`, user, project).Scan(&is))

		return is
	}

	change := func(t *testing.T, endpoint *apiEndpoint, user uuid.UUID, headers map[string]string, projects ...uuid.UUID) (int, string) {
		t.Helper()

		response, err := sendHTTPRequest(t, context.WithoutCancel(ctx), endpoint.RewriteSlugs(user.String()),
			map[string]any{"project_ids": projects}, headers)
		require.NoError(t, err)

		text := readResponseBody(t, response)
		response.Body.Close()

		message, err := messageOf(text)
		require.NoError(t, err, text)

		return response.StatusCode, message
	}

	link := func(t *testing.T, user uuid.UUID, headers map[string]string, projects ...uuid.UUID) (int, string) {
		t.Helper()

		return change(t, usersLinkProjectsEndpoint, user, headers, projects...)
	}

	unlink := func(t *testing.T, user uuid.UUID, headers map[string]string, projects ...uuid.UUID) (int, string) {
		t.Helper()

		return change(t, usersUnlinkProjectsEndpoint, user, headers, projects...)
	}

	// The words of a refusal, without the id it names.
	said := func(message string, id uuid.UUID) string { return strings.ReplaceAll(message, id.String(), "<id>") }

	t.Run("a_caller_links_a_user_to_a_project_of_theirs", func(t *testing.T) {
		status, message := link(t, newcomer, manager, mine.ID)
		require.Equal(t, http.StatusOK, status, message)
		assert.True(t, linked(t, newcomer, mine.ID))

		// Naming it twice is naming it.
		status, message = link(t, newcomer, manager, mine.ID, mine.ID)
		require.Equal(t, http.StatusOK, status, message)
	})

	t.Run("another_accounts_project_is_refused_as_one_that_does_not_exist", func(t *testing.T) {
		status, message := link(t, newcomer, manager, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.False(t, linked(t, newcomer, theirs.ID), "a user was linked to a project the caller is not linked to")
		assert.Contains(t, message, theirs.ID.String(), "the refusal names the id the caller sent")

		absentStatus, absent := link(t, newcomer, manager, missing)
		assert.Equal(t, status, absentStatus, absent)
		assert.Equal(t, said(message, theirs.ID), said(absent, missing), "another account's project and one that does not exist are told apart")
	})

	t.Run("a_caller_cannot_link_themselves_into_a_project", func(t *testing.T) {
		status, message := link(t, managerID, manager, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.False(t, linked(t, managerID, theirs.ID), "the caller made themselves a member of a project they were not linked to")
	})

	t.Run("one_project_that_is_not_theirs_refuses_the_whole_request", func(t *testing.T) {
		status, message := link(t, untouched, manager, mine.ID, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.Contains(t, message, theirs.ID.String(), "the refusal names the id that is not the caller's")
		assert.NotContains(t, message, mine.ID.String())
		assert.False(t, linked(t, untouched, mine.ID), "part of a refused request was written")
		assert.False(t, linked(t, untouched, theirs.ID))
	})

	t.Run("the_administrator_role_without_the_flag_is_held_like_anyone", func(t *testing.T) {
		status, message := link(t, untouched, roleOnly, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.False(t, linked(t, untouched, theirs.ID))

		status, message = unlink(t, theirMember, roleOnly, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.True(t, linked(t, theirMember, theirs.ID))
	})

	t.Run("a_member_of_nothing_links_nobody", func(t *testing.T) {
		for _, project := range []uuid.UUID{mine.ID, theirs.ID} {
			status, message := link(t, untouched, memberOfNothing, project)
			assert.Equal(t, http.StatusBadRequest, status, message)
			assert.False(t, linked(t, untouched, project))
		}

		status, message := unlink(t, theirMember, memberOfNothing, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.True(t, linked(t, theirMember, theirs.ID))
	})

	t.Run("a_caller_cannot_unlink_a_user_from_a_project_that_is_not_theirs", func(t *testing.T) {
		status, message := unlink(t, theirMember, manager, theirs.ID)
		assert.Equal(t, http.StatusBadRequest, status, message)
		assert.True(t, linked(t, theirMember, theirs.ID), "a user was removed from a project the caller is not linked to")

		absentStatus, absent := unlink(t, theirMember, manager, missing)
		assert.Equal(t, status, absentStatus, absent)
		assert.Equal(t, said(message, theirs.ID), said(absent, missing), "another account's project and one that does not exist are told apart")
	})

	t.Run("a_caller_unlinks_a_user_from_a_project_of_theirs", func(t *testing.T) {
		// Linked here, so that the case stands when it is run alone.
		status, message := link(t, newcomer, manager, mine.ID)
		require.Equal(t, http.StatusOK, status, message)
		require.True(t, linked(t, newcomer, mine.ID))

		status, message = unlink(t, newcomer, manager, mine.ID)
		require.Equal(t, http.StatusOK, status, message)
		assert.False(t, linked(t, newcomer, mine.ID))

		// A user who is not in it is unlinked from it all the same: the
		// project is the caller's, and there is nothing to remove.
		status, message = unlink(t, newcomer, manager, mine.ID)
		assert.Equal(t, http.StatusOK, status, message)
	})

	t.Run("a_user_that_does_not_exist_keeps_its_answers", func(t *testing.T) {
		status, message := link(t, missing, manager, mine.ID)
		assert.Equal(t, http.StatusNotFound, status, message)

		status, message = unlink(t, missing, manager, mine.ID)
		assert.Equal(t, http.StatusOK, status, message)
	})

	t.Run("the_administrator_account_links_and_unlinks_on_any_project", func(t *testing.T) {
		status, message := link(t, newcomer, admin, theirs.ID)
		require.Equal(t, http.StatusOK, status, message)
		assert.True(t, linked(t, newcomer, theirs.ID))

		status, message = unlink(t, newcomer, admin, theirs.ID)
		require.Equal(t, http.StatusOK, status, message)
		assert.False(t, linked(t, newcomer, theirs.ID))

		// And a project that does not exist is refused for it too.
		status, message = link(t, newcomer, admin, missing)
		assert.Equal(t, http.StatusBadRequest, status, message)
	})
}
