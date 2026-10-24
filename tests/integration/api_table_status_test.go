//go:build integration

package integration

import (
	"context"
	"io"
	"net/http"
	"testing"

	"uuid"

	"github.com/stretchr/testify/require"
)

// A refusal answered 500 is counted as a refusal on the metric and read as a
// fault by every client. Two routes did it for an error the status table
// calls a 404: a policy update to a grant no endpoint serves (the create
// answers that 404), and a role linked to a policy that is not there (the
// grant guard reads each policy to learn what it grants).
func TestARefusalTheTableKnowsIsNotA500(t *testing.T) {
	ctx := t.Context()
	admin := getAdminUserTokens(t)
	header := map[string]string{"Authorization": "Bearer " + admin.AccessToken}

	roleID := uuid.NewV7()
	created, err := sendHTTPRequest(t, ctx, newAPIEndpoint(http.MethodPost, "/roles"), map[string]any{
		"id":          roleID.String(),
		"name":        "table-status-" + roleID.String(),
		"description": "a role the status table test links to",
	}, header)
	require.NoError(t, err)
	created.Body.Close()
	require.Equal(t, http.StatusCreated, created.StatusCode)

	policyID := createTestPolicyForRoleTest(t, admin.AccessToken, "table-status-")

	t.Cleanup(func() {
		for _, endpoint := range []*apiEndpoint{
			newAPIEndpoint(http.MethodDelete, "/roles/{role_id}").RewriteSlugs(roleID.String()),
			newAPIEndpoint(http.MethodDelete, "/policies/{policy_id}").RewriteSlugs(policyID.String()),
		} {
			if resp, err := sendHTTPRequest(t, context.Background(), endpoint, nil, header); err == nil {
				resp.Body.Close()
			}
		}
	})

	t.Run("a_policy_update_to_a_grant_no_endpoint_serves_is_a_404", func(t *testing.T) {
		resp, err := sendHTTPRequest(t, ctx, newAPIEndpoint(http.MethodPut, "/policies/{policy_id}").RewriteSlugs(policyID.String()),
			map[string]any{"allowed_resource": "/nowhere/at/all"}, header)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusNotFound, resp.StatusCode, "body: %s", body)
	})

	t.Run("a_role_linked_to_a_policy_that_is_not_there_is_a_404", func(t *testing.T) {
		resp, err := sendHTTPRequest(t, ctx, newAPIEndpoint(http.MethodPost, "/roles/{role_id}/policies").RewriteSlugs(roleID.String()),
			map[string]any{"policy_ids": []string{uuid.NewV7().String()}}, header)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusNotFound, resp.StatusCode, "body: %s", body)
	})
}
