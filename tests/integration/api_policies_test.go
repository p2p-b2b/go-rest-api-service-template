//go:build integration

package integration

import (
	// Added for readResponseBody
	"context" // Added for readResponseBody
	"strings"
	"time"

	// Added for readResponseBody
	"math/rand"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	policyCreateEndpoint      = newAPIEndpoint(http.MethodPost, "/policies")
	policyGetEndpoint         = newAPIEndpoint(http.MethodGet, "/policies/{policy_id}")
	policyListEndpoint        = newAPIEndpoint(http.MethodGet, "/policies")
	policyUpdateEndpoint      = newAPIEndpoint(http.MethodPut, "/policies/{policy_id}")
	policyDeleteEndpoint      = newAPIEndpoint(http.MethodDelete, "/policies/{policy_id}")
	policyLinkRolesEndpoint   = newAPIEndpoint(http.MethodPost, "/policies/{policy_id}/roles")
	policyUnlinkRolesEndpoint = newAPIEndpoint(http.MethodDelete, "/policies/{policy_id}/roles")
)

// Helper function to create a policy for testing
func createTestPolicy(t *testing.T, accessToken, namePrefix string) uuid.UUID {
	t.Helper()

	policyID := uuid.Must(uuid.NewV7())

	// Create a policy with a unique ID and name
	policy := map[string]any{
		"id":               policyID.String(),
		"name":             namePrefix + policyID.String(),
		"description":      "Test policy " + policyID.String(),
		"allowed_action":   "GET",
		"allowed_resource": "/users/" + policyID.String(),
	}

	accessTokenHeader := map[string]string{"Authorization": "Bearer " + accessToken}

	ctx := context.Background()
	response, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
	assert.NoError(t, err, "Failed to send request to create policy")
	defer response.Body.Close()

	assert.Equal(t, http.StatusCreated, response.StatusCode, "Expected status code 201 for policy creation. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

	// Verify the response
	createAPIResp, err := parserResponseBody[model.HTTPMessage](t, response)
	assert.NoError(t, err, "Failed to parse response body")
	assert.Equal(t, model.PoliciesPolicyCreatedSuccessfully, createAPIResp.Message, "Expected success message")

	return policyID
}

// Helper function to create a role for testing
func createTestRole(t *testing.T, accessToken, namePrefix string) uuid.UUID {
	t.Helper()

	roleID := uuid.Must(uuid.NewV7())

	role := map[string]any{
		"id":          roleID.String(),
		"name":        namePrefix + roleID.String(),
		"description": "Test role " + roleID.String(),
	}

	accessTokenHeader := map[string]string{"Authorization": "Bearer " + accessToken}

	ctx := context.Background()

	rolesCreateEndpoint := newAPIEndpoint(http.MethodPost, "/roles")
	response, err := sendHTTPRequest(t, ctx, rolesCreateEndpoint, role, accessTokenHeader)
	assert.NoError(t, err, "Failed to create test role")
	defer response.Body.Close()

	assert.Equal(t, http.StatusCreated, response.StatusCode, "Failed to create test role, status code not 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))
	return roleID
}

func TestPolicyCreate(t *testing.T) {
	// Test policy creation
	t.Run("create_policy", func(t *testing.T) {
		t.Parallel()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		// 2. Create a new policy
		policyID := uuid.Must(uuid.NewV7())

		policy := map[string]any{
			"id":               policyID.String(),
			"name":             "test_policy_" + policyID.String(),
			"description":      "This is a test policy " + policyID.String(),
			"allowed_action":   "GET",
			"allowed_resource": "/users/" + policyID.String(),
		}

		// 2.1 Use access token from admin to have access to the endpoint
		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		ctx := context.Background()
		response, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
		assert.NoError(t, err, "Error sending request: %v", err)
		defer response.Body.Close()

		assert.Equal(t, http.StatusCreated, response.StatusCode, "Expected status code 201 Created. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		t.Cleanup(func() {
			deletePolicyByIDFromDB(t, policyID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})

		assert.Equal(t, model.PoliciesPolicyCreatedSuccessfully, apiResp.Message, "Unexpected response message")
		assert.Equal(t, policyCreateEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, policyCreateEndpoint.Path(), apiResp.Path, "Expected path to be set")
	})

	// Test creating a policy with invalid data format
	t.Run("create_policy_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Define test cases for various invalid inputs
		testCases := []struct {
			name          string
			invalidPolicy map[string]any
			expectedError string
		}{
			{
				name: "Invalid ID format",
				invalidPolicy: map[string]any{
					"id":               "not-a-valid-uuid",
					"name":             "Test Policy",
					"description":      "Test policy description",
					"allowed_action":   "GET",
					"allowed_resource": "/users/*",
				},
				expectedError: "invalid uuid",
			},
			{
				name: "Empty name",
				invalidPolicy: map[string]any{
					"id":               uuid.Must(uuid.NewV7()).String(),
					"name":             "",
					"description":      "Test policy description",
					"allowed_action":   "GET",
					"allowed_resource": "/users/*",
				},
				expectedError: "cannot be empty",
			},
			{
				name: "Invalid action",
				invalidPolicy: map[string]any{
					"id":               uuid.Must(uuid.NewV7()).String(),
					"name":             "Test Policy",
					"description":      "Test policy description",
					"allowed_action":   "INVALID_ACTION",
					"allowed_resource": "/users/*",
				},
				expectedError: "invalid action",
			},
			{
				name: "Empty resource",
				invalidPolicy: map[string]any{
					"id":               uuid.Must(uuid.NewV7()).String(),
					"name":             "Test Policy",
					"description":      "Test policy description",
					"allowed_action":   "GET",
					"allowed_resource": "",
				},
				expectedError: "cannot be empty",
			},
		}

		// 3. Run each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Send request with the invalid policy data
				response, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, tc.invalidPolicy, accessTokenHeader)
				assert.NoError(t, err, "Failed to send request for %s", tc.name)
				defer response.Body.Close()

				// Verify we get a 400 Bad Request response
				assert.Equal(t, http.StatusBadRequest, response.StatusCode,
					"Expected status code 400 Bad Request for %s. Got %d. Message: %s", tc.name, response.StatusCode, readResponseBody(t, response))

				// Parse and verify the error response
				errorResp, err := parserResponseBody[model.HTTPMessage](t, response)
				assert.NoError(t, err, "Failed to parse error response for %s", tc.name)

				// Verify error message contains information specific to this validation failure
				assert.Contains(t, strings.ToLower(errorResp.Message), strings.ToLower(tc.expectedError),
					"Error message should indicate %s validation failure", tc.name)
				assert.Equal(t, policyCreateEndpoint.method, errorResp.Method, "Expected method to be set")
				assert.Equal(t, policyCreateEndpoint.Path(), errorResp.Path, "Expected path to be set")
			})
		}

		// 4. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test creating policies with existing ID or name
	t.Run("create_policy_already_exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. First create a valid policy that will be our reference policy
		policyID := uuid.Must(uuid.NewV7())

		policyName := "test_" + policyID.String()
		allowedAction := "GET"
		allowedResource := "/users/*"

		existingPolicy := map[string]any{
			"id":               policyID.String(),
			"name":             policyName,
			"description":      "This is a test policy for duplicate checks",
			"allowed_action":   allowedAction,
			"allowed_resource": allowedResource,
		}

		// Create the first policy
		createResponse, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, existingPolicy, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to create initial policy")
		defer createResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 for initial policy creation. Got %d. Message: %s", createResponse.StatusCode, readResponseBody(t, createResponse))

		// 3. Define test cases for duplicate scenarios
		testCases := []struct {
			name            string
			duplicatePolicy map[string]any
			expectedStatus  int
			expectedError   string
		}{
			{
				name: "Policy with existing ID",
				duplicatePolicy: map[string]any{
					"id":               policyID.String(), // Same ID as existing policy
					"name":             "Different_" + policyName,
					"description":      "This is a different policy",
					"allowed_action":   "POST",
					"allowed_resource": "/roles",
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "already exists",
			},
			{
				name: "Policy with existing name",
				duplicatePolicy: map[string]any{
					"id":               uuid.Must(uuid.NewV7()).String(), // Different ID
					"name":             policyName,                       // Same name as existing policy has the same allowed action and allowed resource
					"description":      "This is another policy",
					"allowed_action":   allowedAction,
					"allowed_resource": allowedResource,
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "already exists",
			},
		}

		// 4. Run each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Try to create a policy with duplicate ID or name
				response, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, tc.duplicatePolicy, accessTokenHeader)
				assert.NoError(t, err, "Failed to send request for %s", tc.name)
				defer response.Body.Close()

				// Verify we get the expected conflict status
				assert.Equal(t, tc.expectedStatus, response.StatusCode,
					"Expected status code %d for %s. Got %d. Message: %s", tc.expectedStatus, tc.name, response.StatusCode, readResponseBody(t, response))

				// Parse and verify the error response
				errorResp, err := parserResponseBody[model.HTTPMessage](t, response)
				assert.NoError(t, err, "Failed to parse error response for %s", tc.name)

				// Verify error message contains information about the conflict
				assert.Contains(t, strings.ToLower(errorResp.Message), strings.ToLower(tc.expectedError),
					"Error message should indicate conflict for %s", tc.name)
				assert.Equal(t, policyCreateEndpoint.method, errorResp.Method, "Expected method to be set")
				assert.Equal(t, policyCreateEndpoint.Path(), errorResp.Path, "Expected path to be set")
			})
		}

		// 5. Cleanup
		t.Cleanup(func() {
			deletePolicyByIDFromDB(t, policyID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	t.Run("require_authentication", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		policy := map[string]any{
			"id":          "00000000-0000-0000-0000-000000000000",
			"name":        "Test Policy",
			"description": "Test policy description",
		}

		resp, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestPolicyGet(t *testing.T) {
	// Test policy retrieval
	t.Run("get_policy", func(t *testing.T) {
		t.Parallel()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		// 2. Create a new policy
		policyID := uuid.Must(uuid.NewV7())
		policyName := "test_policy_get_" + policyID.String()
		policyDesc := "This is a test policy for get " + policyID.String()
		policyAction := "GET"
		policyResource := "/roles/" + policyID.String()

		policy := map[string]any{
			"id":               policyID.String(),
			"name":             policyName,
			"description":      policyDesc,
			"allowed_action":   policyAction,
			"allowed_resource": policyResource,
		}

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		ctx := context.Background()
		createResponse, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
		assert.NoError(t, err, "Error sending create request: %v", err)
		defer createResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 Created for setup. Got %d. Message: %s", createResponse.StatusCode, readResponseBody(t, createResponse))

		// t.Logf("Create response: %v", createResponse)

		// 3. Get the policy
		getEndpoint := policyGetEndpoint.RewriteSlugs(policyID.String())
		// t.Logf("Get endpoint: %s", getEndpoint.Path())

		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get request: %v", err)
		defer getResponse.Body.Close()

		// t.Logf("Get response: %v", getResponse)

		// 4. Check the response
		assert.Equal(t, http.StatusOK, getResponse.StatusCode, "Expected status code 200 OK for get. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))
		getAPIResp, err := parserResponseBody[model.Policy](t, getResponse)
		assert.NoError(t, err, "Failed to parse get response body", err)

		assert.Equal(t, policyID, getAPIResp.ID, "Expected policy ID to match")
		assert.Equal(t, policyName, getAPIResp.Name, "Expected policy name to match")
		assert.Equal(t, policyDesc, getAPIResp.Description, "Expected policy description to match")
		assert.Equal(t, policyAction, getAPIResp.AllowedAction, "Expected policy action to match")
		assert.Equal(t, policyResource, getAPIResp.AllowedResource, "Expected policy resource to match")

		// 5. Cleanup
		t.Cleanup(func() {
			deletePolicyByIDFromDB(t, policyID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test retrieving a non-existent policy
	t.Run("get_policy_not_found", func(t *testing.T) {
		t.Parallel()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Generate a random UUID that doesn't exist in the database
		nonExistentPolicyID := uuid.Must(uuid.NewV7())

		// 3. Try to get the non-existent policy
		getEndpoint := policyGetEndpoint.RewriteSlugs(nonExistentPolicyID.String())
		ctx := context.Background()
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to get non-existent policy")
		defer getResponse.Body.Close()

		// 4. Check that we get a 404 Not Found response
		assert.Equal(t, http.StatusNotFound, getResponse.StatusCode, "Expected status code 404 for non-existent policy. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))

		// 5. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, getResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// Verify error message contains information about the policy not being found
		assert.Contains(t, errorResp.Message, "not found", "Error message should indicate that the policy was not found")
		assert.Equal(t, getEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, getEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test retrieving a policy with an invalid ID format
	t.Run("get_policy_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Try to get a policy with an invalid ID format (not a UUID)
		invalidPolicyID := "not-a-valid-uuid"
		getEndpoint := policyGetEndpoint.RewriteSlugs(invalidPolicyID)
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to get policy with invalid ID")
		defer getResponse.Body.Close()

		// 3. Check that we get a 400 Bad Request response
		assert.Equal(t, http.StatusBadRequest, getResponse.StatusCode, "Expected status code 400 for invalid policy ID format. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))

		// 4. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, getResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// Verify error message contains information about the invalid UUID format
		assert.Contains(t, errorResp.Message, "invalid", "Error message should indicate that the policy ID format is invalid")
		assert.Equal(t, getEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, getEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 5. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	t.Run("require_authentication", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		getEndpoint := policyGetEndpoint.RewriteSlugs("00000000-0000-0000-0000-000000000000")
		resp, err := sendHTTPRequest(t, ctx, getEndpoint, nil)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestPolicyDelete(t *testing.T) {
	// Test policy deletion
	t.Run("delete_policy", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a new policy
		policyID := uuid.Must(uuid.NewV7())
		policy := map[string]any{
			"id":               policyID.String(),
			"name":             "test_policy_delete_" + policyID.String(),
			"description":      "This is a test policy for delete " + policyID.String(),
			"allowed_action":   "DELETE",
			"allowed_resource": "/roles/" + policyID.String(),
		}

		createResponse, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
		assert.NoError(t, err, "Error sending create request: %v", err)
		defer createResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 Created for setup")

		// 3. Delete the policy
		deleteEndpoint := policyDeleteEndpoint.RewriteSlugs(policyID.String())
		deleteResponse, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending delete request: %v", err)
		defer deleteResponse.Body.Close()

		// 4. Check the delete response
		assert.Equal(t, http.StatusOK, deleteResponse.StatusCode, "Expected status code 200 OK for delete")
		deleteAPIResp, err := parserResponseBody[model.HTTPMessage](t, deleteResponse)
		assert.NoError(t, err, "Failed to parse delete response body")

		assert.Equal(t, model.PoliciesPolicyDeletedSuccessfully, deleteAPIResp.Message, "Unexpected delete response message")
		assert.Equal(t, deleteEndpoint.method, deleteAPIResp.Method, "Expected method to be set for delete")
		assert.Equal(t, deleteEndpoint.Path(), deleteAPIResp.Path, "Expected path to be set for delete")

		// 5. Verify policy is actually deleted (try to get it)
		getEndpoint := policyGetEndpoint.RewriteSlugs(policyID.String())
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get request after delete: %v", err)
		defer getResponse.Body.Close()
		assert.Equal(t, http.StatusNotFound, getResponse.StatusCode, "Expected status code 404 Not Found after deletion")

		// 6. Cleanup admin user
		t.Cleanup(func() {
			// Policy should already be deleted by the test, but try again just in case
			deletePolicyByIDFromDB(t, policyID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test deleting a policy with an invalid ID format
	t.Run("delete_policy_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Try to delete a policy with an invalid ID format (not a UUID)
		invalidPolicyID := "not-a-valid-uuid"
		deleteEndpoint := policyDeleteEndpoint.RewriteSlugs(invalidPolicyID)

		deleteResponse, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to delete policy with invalid ID")
		defer deleteResponse.Body.Close()

		// 3. Check that we get a 400 Bad Request response
		assert.Equal(t, http.StatusBadRequest, deleteResponse.StatusCode, "Expected status code 400 for invalid policy ID format. Got %d. Message: %s", deleteResponse.StatusCode, readResponseBody(t, deleteResponse))

		// 4. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, deleteResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// 5. Verify error message contains information about the invalid UUID format
		assert.Contains(t, errorResp.Message, "invalid", "Error message should indicate that the policy ID format is invalid")
		assert.Equal(t, deleteEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, deleteEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test deleting a non-existent policy
	t.Run("delete_policy_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Generate a UUID that doesn't exist in the database
		nonExistentPolicyID := uuid.Must(uuid.NewV7())
		deleteEndpoint := policyDeleteEndpoint.RewriteSlugs(nonExistentPolicyID.String())

		deleteResponse, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to delete non-existent policy")
		defer deleteResponse.Body.Close()

		// 3. Check the response - this should still return StatusOK even though the policy doesn't exist
		// This is because deleting a non-existent resource is considered idempotent in RESTful APIs
		assert.Equal(t, http.StatusOK, deleteResponse.StatusCode, "Expected status code 200 OK for deleting non-existent policy. Got %d. Message: %s", deleteResponse.StatusCode, readResponseBody(t, deleteResponse))

		// 4. Parse and verify the success response
		deleteAPIResp, err := parserResponseBody[model.HTTPMessage](t, deleteResponse)
		assert.NoError(t, err, "Failed to parse response body")

		// 5. Verify success message for deletion
		assert.Equal(t, model.PoliciesPolicyDeletedSuccessfully, deleteAPIResp.Message, "Expected success message")
		assert.Equal(t, deleteEndpoint.method, deleteAPIResp.Method, "Expected method to be set")
		assert.Equal(t, deleteEndpoint.Path(), deleteAPIResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	t.Run("require_authentication", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		deleteEndpoint := policyDeleteEndpoint.RewriteSlugs("00000000-0000-0000-0000-000000000000")
		resp, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestPolicyUpdate(t *testing.T) {
	// Test policy update
	t.Run("update_policy", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a new policy
		policyID := uuid.Must(uuid.NewV7())
		originalName := "test_policy_update_" + policyID.String()
		originalDesc := "Original description " + policyID.String()
		originalAction := "GET"
		originalResource := "/roles/*"

		policy := map[string]any{
			"id":               policyID.String(),
			"name":             originalName,
			"description":      originalDesc,
			"allowed_action":   originalAction,
			"allowed_resource": originalResource,
		}

		createResponse, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
		assert.NoError(t, err, "Error sending create request: %v", err)
		defer createResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 Created for setup")

		// 3. Update the policy
		updatedName := "updated_" + originalName
		updatedDesc := "Updated description " + policyID.String()
		updatedAction := "PUT"
		updatedResource := "/roles/*"

		updatedPolicy := map[string]any{
			"name":             updatedName,
			"description":      updatedDesc,
			"allowed_action":   updatedAction,
			"allowed_resource": updatedResource,
		}

		updateEndpoint := policyUpdateEndpoint.RewriteSlugs(policyID.String())
		updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, updatedPolicy, accessTokenHeader)
		assert.NoError(t, err, "Error sending update request: %v", err)
		defer updateResponse.Body.Close()

		// 4. Check the update response
		assert.Equal(t, http.StatusOK, updateResponse.StatusCode, "Expected status code 200 OK for update")
		updateAPIResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
		assert.NoError(t, err, "Failed to parse update response body")

		assert.Equal(t, model.PoliciesPolicyUpdatedSuccessfully, updateAPIResp.Message, "Unexpected update response message")
		assert.Equal(t, updateEndpoint.method, updateAPIResp.Method, "Expected method to be set for update")
		assert.Equal(t, updateEndpoint.Path(), updateAPIResp.Path, "Expected path to be set for update")

		// 5. Verify policy is actually updated (get it again)
		getEndpoint := policyGetEndpoint.RewriteSlugs(policyID.String())
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get request after update: %v", err)
		defer getResponse.Body.Close()

		assert.Equal(t, http.StatusOK, getResponse.StatusCode, "Expected status code 200 OK when getting updated policy")

		getAPIResp, err := parserResponseBody[model.Policy](t, getResponse)
		assert.NoError(t, err, "Failed to parse get response body for updated policy")

		assert.Equal(t, policyID, getAPIResp.ID, "Expected policy ID to remain the same")
		assert.Equal(t, updatedName, getAPIResp.Name, "Expected policy name to be updated")
		assert.Equal(t, updatedDesc, getAPIResp.Description, "Expected policy description to be updated")
		assert.Equal(t, updatedAction, getAPIResp.AllowedAction, "Expected policy action to be updated")
		assert.Equal(t, updatedResource, getAPIResp.AllowedResource, "Expected policy resource to be updated")

		// 6. Cleanup
		t.Cleanup(func() {
			deletePolicyByIDFromDB(t, policyID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test updating a policy with invalid data
	t.Run("update_policy_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a valid policy first that we'll try to update with invalid data
		policyID := uuid.Must(uuid.NewV7())
		policy := map[string]any{
			"id":               policyID.String(),
			"name":             "test_policy_update_invalid_" + policyID.String(),
			"description":      "This is a test policy for update with invalid data",
			"allowed_action":   "GET",
			"allowed_resource": "/users/*",
		}

		createResponse, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to create policy")
		defer createResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 for initial policy creation. Got %d. Message: %s", createResponse.StatusCode, readResponseBody(t, createResponse))

		// 3. Set up test cases for various invalid inputs
		testCases := []struct {
			name          string
			invalidUpdate map[string]any
			expectedError string
		}{
			{
				name: "Empty name",
				invalidUpdate: map[string]any{
					"name": "",
				},
				expectedError: "cannot be empty",
			},
			{
				name: "Invalid action",
				invalidUpdate: map[string]any{
					"allowed_action": "INVALID_ACTION",
				},
				expectedError: "invalid action",
			},
			{
				name: "Empty resource",
				invalidUpdate: map[string]any{
					"allowed_resource": "",
				},
				expectedError: "cannot be empty",
			},
		}

		// 4. Run each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Try to update the policy with invalid data
				updateEndpoint := policyUpdateEndpoint.RewriteSlugs(policyID.String())
				updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, tc.invalidUpdate, accessTokenHeader)
				assert.NoError(t, err, "Failed to send request for %s", tc.name)
				defer updateResponse.Body.Close()

				// Verify we get a 400 Bad Request response
				assert.Equal(t, http.StatusBadRequest, updateResponse.StatusCode,
					"Expected status code 400 Bad Request for %s. Got %d. Message: %s", tc.name, updateResponse.StatusCode, readResponseBody(t, updateResponse))

				// Parse and verify the error response
				errorResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
				assert.NoError(t, err, "Failed to parse error response for %s", tc.name)

				// Verify error message contains information specific to this validation failure
				assert.Contains(t, strings.ToLower(errorResp.Message), strings.ToLower(tc.expectedError),
					"Error message should indicate %s validation failure", tc.name)
				assert.Equal(t, updateEndpoint.method, errorResp.Method, "Expected method to be set")
				assert.Equal(t, updateEndpoint.Path(), errorResp.Path, "Expected path to be set")
			})
		}

		// 5. Cleanup
		t.Cleanup(func() {
			deletePolicyByIDFromDB(t, policyID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test updating a non-existent policy
	t.Run("update_policy_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Generate a random UUID that doesn't exist
		nonExistentPolicyID := uuid.Must(uuid.NewV7())
		updateEndpoint := policyUpdateEndpoint.RewriteSlugs(nonExistentPolicyID.String())

		// Create update data
		updatedPolicy := map[string]any{
			"name":             "Updated Policy Name",
			"description":      "Updated policy description",
			"allowed_action":   "PUT",
			"allowed_resource": "/roles/*",
		}

		// 3. Try to update the non-existent policy
		updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, updatedPolicy, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to update non-existent policy")
		defer updateResponse.Body.Close()

		// 4. Check that we get a 404 Not Found response
		assert.Equal(t, http.StatusNotFound, updateResponse.StatusCode, "Expected status code 404 for non-existent policy. Got %d. Message: %s", updateResponse.StatusCode, readResponseBody(t, updateResponse))

		// 5. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// Verify error message contains information about the policy not being found
		assert.Contains(t, errorResp.Message, "not found", "Error message should indicate that the policy was not found")
		assert.Equal(t, updateEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, updateEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test updating a policy with a name that already exists
	t.Run("update_policy_conflict_name", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create two policies with different names
		// First policy - this is the one we'll try to update
		policyID1 := uuid.Must(uuid.NewV7())
		policyName1 := "test_conflict_1_" + policyID1.String()
		policy1 := map[string]any{
			"id":               policyID1.String(),
			"name":             policyName1,
			"description":      "First test policy for conflict check",
			"allowed_action":   "GET",
			"allowed_resource": "/users/*",
		}

		createResponse1, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy1, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to create first policy")
		defer createResponse1.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse1.StatusCode, "Expected status code 201 for first policy creation. Got %d. Message: %s", createResponse1.StatusCode, readResponseBody(t, createResponse1))

		// Second policy - we'll try to use this policy's name when updating the first policy
		policyID2 := uuid.Must(uuid.NewV7())
		policyName2 := "test_conflict_2_" + policyID2.String()
		policy2 := map[string]any{
			"id":               policyID2.String(),
			"name":             policyName2,
			"description":      "Second test policy for conflict check",
			"allowed_action":   "GET",
			"allowed_resource": "/roles/*",
		}

		createResponse2, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy2, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to create second policy")
		defer createResponse2.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse2.StatusCode, "Expected status code 201 for second policy creation. Got %d. Message: %s", createResponse2.StatusCode, readResponseBody(t, createResponse2))

		// 3. Try to update the first policy with the second policy's name
		updatePolicy1WithConflict := map[string]any{
			"name":             policyName2, // This will cause a conflict because policyName2, allows the same action and resource as policyName1
			"allowed_action":   "GET",
			"allowed_resource": "/roles/*",
		}

		updateEndpoint := policyUpdateEndpoint.RewriteSlugs(policyID1.String())
		updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, updatePolicy1WithConflict, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to update policy with conflicting name")
		defer updateResponse.Body.Close()

		// 4. Check that we get a 409 Conflict response
		assert.Equal(t, http.StatusConflict, updateResponse.StatusCode, "Expected status code 409 Conflict for update with already used name. Got %d. Message: %s", updateResponse.StatusCode, readResponseBody(t, updateResponse))

		// 5. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// 6. Verify error message contains information about the name being already in use
		assert.Contains(t, strings.ToLower(errorResp.Message), "already exists", "Error message should indicate that the name already exists")
		assert.Equal(t, updateEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, updateEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 7. Cleanup
		t.Cleanup(func() {
			deletePolicyByIDFromDB(t, policyID1)
			deletePolicyByIDFromDB(t, policyID2)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})
}

func TestPolicyList(t *testing.T) {
	// Test policy listing
	t.Run("list_policies", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a couple of new policies
		policyIDs := []uuid.UUID{uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())}
		policiesToCreate := map[string]map[string]any{}

		// validActions := strings.ReplaceAll(model.GetValidActions(), " ", "")
		// actionsAllowed := strings.Split(validActions, ",")
		actionsAllowed := []string{"GET"}
		resourcesAllowed := []string{
			"/roles/*",
			"/users/*",
			"/policies/*",
		}

		for i, policyID := range policyIDs {
			randomAction := actionsAllowed[rand.Intn(len(actionsAllowed))]
			randomResource := resourcesAllowed[rand.Intn(len(resourcesAllowed))]

			policy := map[string]any{
				"id":               policyID.String(),
				"name":             "test_policy_list_" + policyID.String(),
				"description":      "This is a test policy for list " + policyID.String(),
				"allowed_action":   randomAction,
				"allowed_resource": randomResource,
			}
			policiesToCreate[policyID.String()] = policy

			createResponse, err := sendHTTPRequest(t, ctx, policyCreateEndpoint, policy, accessTokenHeader)
			assert.NoError(t, err, "Failed to send request to create policy %d", i+1)
			if createResponse != nil {
				defer createResponse.Body.Close()

				createResponseMessage, err := parserResponseBody[model.HTTPMessage](t, createResponse)
				assert.NoError(t, err, "Failed to parse create response body for policy %d", i+1)

				assert.Equal(t, model.PoliciesPolicyCreatedSuccessfully, createResponseMessage.Message, "Unexpected response message for policy %d", i+1)
				assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 for policy.")
			}
		}

		// 3. List the policies
		listResponse, err := sendHTTPRequest(t, ctx, policyListEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to list policies")
		defer listResponse.Body.Close()

		// 4. Check the list response
		assert.Equal(t, http.StatusOK, listResponse.StatusCode, "Expected status code 200 for list. Got %d. Message: %s", listResponse.StatusCode, readResponseBody(t, listResponse))
		// Assuming model.ListPoliciesOutput exists and has an Items field []model.Policy
		listAPIResp, err := parserResponseBody[model.ListPoliciesOutput](t, listResponse)
		assert.NoError(t, err, "Failed to parse list response body")

		// 5. Verify the created policies are in the list

		for _, listedPolicy := range listAPIResp.Items {
			if _, ok := policiesToCreate[listedPolicy.Name]; ok {
				// Optionally assert other fields match
				for _, createdPolicy := range policiesToCreate {
					if createdPolicy["name"] == listedPolicy.Name {
						assert.Equal(t, createdPolicy["description"], listedPolicy.Description)
						assert.Equal(t, createdPolicy["action"], listedPolicy.AllowedAction)
						assert.Equal(t, createdPolicy["resource"], listedPolicy.AllowedResource)
						break
					}
				}
			}
		}
		assert.GreaterOrEqual(t, len(listAPIResp.Items), len(policiesToCreate), "Expected to find at least the created policies in the list")

		// 6. Cleanup
		t.Cleanup(func() {
			for _, policyID := range policyIDs {
				deletePolicyByIDFromDB(t, policyID)
			}
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})
}

func TestPolicyLinkRoles(t *testing.T) {
	t.Run("link_roles_to_policy", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Setup: Create admin, policy, and roles
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		policyID := createTestPolicy(t, adminToken.AccessToken, "link_test_policy_")
		roleID1 := createTestRole(t, adminToken.AccessToken, "link_test_role_1_")
		roleID2 := createTestRole(t, adminToken.AccessToken, "link_test_role_2_")

		// 2. Link roles to the policy
		linkEndpoint := policyLinkRolesEndpoint.RewriteSlugs(policyID.String())
		linkPayload := map[string]any{
			"role_ids": []string{roleID1.String(), roleID2.String()},
		}

		linkResponse, err := sendHTTPRequest(t, ctx, linkEndpoint, linkPayload, accessTokenHeader) // Pass struct directly
		assert.NoError(t, err, "Error sending link roles request: %v", err)
		defer linkResponse.Body.Close()

		// 3. Check link response
		assert.Equal(t, http.StatusOK, linkResponse.StatusCode, "Expected status code 200 OK for linking roles. Got %d. Message: %s", linkResponse.StatusCode, readResponseBody(t, linkResponse))
		linkAPIResp, err := parserResponseBody[model.HTTPMessage](t, linkResponse)
		assert.NoError(t, err, "Failed to parse link roles response body")

		assert.Equal(t, model.PoliciesRolesLinkedSuccessfully, linkAPIResp.Message, "Unexpected link roles response message")

		// 4. Verify roles are linked (by getting one of the roles)
		getRoleEndpoint := rolesGetEndpoint.RewriteSlugs(roleID1.String())
		getRoleResponse, err := sendHTTPRequest(t, ctx, getRoleEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get role request after linking: %v", err)
		defer getRoleResponse.Body.Close()

		assert.Equal(t, http.StatusOK, getRoleResponse.StatusCode, "Expected status code 200 OK when getting role after linking. Got %d. Message: %s", getRoleResponse.StatusCode, readResponseBody(t, getRoleResponse))
		assert.NoError(t, err, "Failed to parse get role response body after linking")

		t.Cleanup(func() {
			// Attempt cleanup even if tests fail
			deletePolicyByIDFromDB(t, policyID)
			deleteRoleByIDFromDB(t, roleID1)
			deleteRoleByIDFromDB(t, roleID2)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})
}

func TestPolicyUnlinkRoles(t *testing.T) {
	t.Run("unlink_roles_from_policy", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Setup: Create admin, policy, roles, and link them
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		policyID := createTestPolicy(t, adminToken.AccessToken, "unlink_test_policy_")
		roleID1 := createTestRole(t, adminToken.AccessToken, "unlink_test_role_1_")
		roleID2 := createTestRole(t, adminToken.AccessToken, "unlink_test_role_2_")

		// Link roles first
		linkEndpoint := policyLinkRolesEndpoint.RewriteSlugs(policyID.String())
		linkPayload := map[string]any{
			"role_ids": []string{roleID1.String(), roleID2.String()},
		}

		linkResponse, err := sendHTTPRequest(t, ctx, linkEndpoint, linkPayload, accessTokenHeader)
		assert.NoError(t, err, "Error sending link roles request during setup: %v", err)
		defer linkResponse.Body.Close()

		assert.Equal(t, http.StatusOK, linkResponse.StatusCode, "Expected status code 200 OK for linking roles during setup. Got %d. Message: %s", linkResponse.StatusCode, readResponseBody(t, linkResponse))

		// 2. Unlink one role
		time.Sleep(1 * time.Second) // Ensure different timestamps for unlinking
		unlinkEndpoint := policyUnlinkRolesEndpoint.RewriteSlugs(policyID.String())
		unlinkPayload := map[string]any{
			"role_ids": []string{roleID1.String()},
		}

		unlinkResponse, err := sendHTTPRequest(t, ctx, unlinkEndpoint, unlinkPayload, accessTokenHeader)
		assert.NoError(t, err, "Error sending unlink role request: %v", err)
		defer unlinkResponse.Body.Close()

		// 3. Check unlink response
		assert.Equal(t, http.StatusOK, unlinkResponse.StatusCode, "Expected status code 200 OK for unlinking role. Got %d. Message: %s", unlinkResponse.StatusCode, readResponseBody(t, unlinkResponse))
		unlinkAPIResp, err := parserResponseBody[model.HTTPMessage](t, unlinkResponse)
		assert.NoError(t, err, "Failed to parse unlink role response body")

		assert.Equal(t, model.PoliciesRolesUnlinkedSuccessfully, unlinkAPIResp.Message, "Unexpected unlink role response message")

		// 4. Verify role is unlinked (by getting the roles)
		rolesGetEndpoint := newAPIEndpoint(http.MethodGet, "/roles/{role_id}") // Define locally or ensure global access

		// Check Role 1 (should be unlinked)
		getRole1Endpoint := rolesGetEndpoint.RewriteSlugs(roleID1.String())
		getRole1Response, err := sendHTTPRequest(t, ctx, getRole1Endpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get role 1 request after unlinking: %v", err)
		defer getRole1Response.Body.Close()
		assert.Equal(t, http.StatusOK, getRole1Response.StatusCode, "Expected status code 200 OK when getting role 1 after unlinking. Got %d. Message: %s", getRole1Response.StatusCode, readResponseBody(t, getRole1Response))

		// Check Role 2 (should still be linked)
		getRole2Endpoint := rolesGetEndpoint.RewriteSlugs(roleID2.String())
		getRole2Response, err := sendHTTPRequest(t, ctx, getRole2Endpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get role 2 request after unlinking: %v", err)
		defer getRole2Response.Body.Close()
		assert.Equal(t, http.StatusOK, getRole2Response.StatusCode, "Expected status code 200 OK when getting role 2 after unlinking")

		t.Cleanup(func() {
			// Attempt cleanup even if tests fail
			deletePolicyByIDFromDB(t, policyID)
			deleteRoleByIDFromDB(t, roleID1)
			deleteRoleByIDFromDB(t, roleID2)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})
}
