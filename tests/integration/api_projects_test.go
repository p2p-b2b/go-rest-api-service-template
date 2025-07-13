//go:build integration

package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	projectsCreateEndpoint = newAPIEndpoint(http.MethodPost, "/projects")
	projectsListEndpoint   = newAPIEndpoint(http.MethodGet, "/projects")
	projectsGetEndpoint    = newAPIEndpoint(http.MethodGet, "/projects/{project_id}")
	projectsUpdateEndpoint = newAPIEndpoint(http.MethodPut, "/projects/{project_id}")
	projectsDeleteEndpoint = newAPIEndpoint(http.MethodDelete, "/projects/{project_id}")
)

// Helper function to create a test project
func createTestProject(t *testing.T, ctx context.Context, accessToken, namePrefix string) (uuid.UUID, string, string) {
	t.Helper()

	projectID := uuid.Must(uuid.NewV7())
	projectName := namePrefix + projectID.String()
	projectDesc := "Test project " + projectID.String()
	project := map[string]any{
		"id":          projectID.String(),
		"name":        projectName,
		"description": projectDesc,
	}

	accessTokenHeader := map[string]string{"Authorization": "Bearer " + accessToken}

	response, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, project, accessTokenHeader)
	assert.NoError(t, err, "Failed to create test project")
	if response != nil {
		defer response.Body.Close()
		assert.Equal(t, http.StatusCreated, response.StatusCode, "Failed to create test project, status code not 201")
	}

	return projectID, projectName, projectDesc
}

func TestProjectCreate(t *testing.T) {
	t.Run("create_project", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a new project
		projectID := uuid.Must(uuid.NewV7())

		project := map[string]any{
			"id":          projectID.String(),
			"name":        "test_project_" + projectID.String(),
			"description": "This is a test project " + projectID.String(),
		}

		response, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, project, accessTokenHeader)
		assert.NoError(t, err, "Error sending request: %v", err)
		defer response.Body.Close()

		assert.Equal(t, http.StatusCreated, response.StatusCode, "Expected status code 201 Created. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		t.Cleanup(func() {
			deleteProjectByIDFromDB(t, projectID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})

		assert.Equal(t, model.ProjectsProjectCreatedSuccessfully, apiResp.Message, "Unexpected response message")
		assert.Equal(t, projectsCreateEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, projectsCreateEndpoint.Path(), apiResp.Path, "Expected path to be set")
	})

	// Test creating a project with invalid data format
	t.Run("create_project_bad_request", func(t *testing.T) {
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
			name           string
			invalidProject map[string]any
			expectedError  string
		}{
			{
				name: "Empty project name",
				invalidProject: map[string]any{
					"id":          uuid.Must(uuid.NewV7()).String(),
					"name":        "",
					"description": "Project with empty name",
				},
				expectedError: "cannot be empty",
			},
			{
				name: "Name too long",
				invalidProject: map[string]any{
					"id":          uuid.Must(uuid.NewV7()).String(),
					"name":        string(make([]byte, 256)), // Very long name
					"description": "Project with too long name",
				},
				expectedError: "must be between",
			},
			{
				name: "Invalid ID format",
				invalidProject: map[string]any{
					"id":          "not-a-valid-uuid",
					"name":        "Valid Project Name",
					"description": "Project with invalid ID",
				},
				expectedError: "invalid uuid",
			},
			{
				name: "Description too long",
				invalidProject: map[string]any{
					"id":          uuid.Must(uuid.NewV7()).String(),
					"name":        "Valid Name",
					"description": string(make([]byte, 5000)), // Very long description
				},
				expectedError: "must be between",
			},
		}

		// 3. Run each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Send request with the invalid project data
				response, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, tc.invalidProject, accessTokenHeader)
				assert.NoError(t, err, "Failed to send request for %s", tc.name)
				defer response.Body.Close()

				// Verify we get a 400 Bad Request response
				assert.Equal(t, http.StatusBadRequest, response.StatusCode,
					"Expected status code 400 Bad Request for %s, got %d", tc.name, response.StatusCode)

				// Parse and verify the error response
				errorResp, err := parserResponseBody[model.HTTPMessage](t, response)
				assert.NoError(t, err, "Failed to parse error response for %s", tc.name)

				// Verify error message contains information specific to this validation failure
				assert.Contains(t, strings.ToLower(errorResp.Message), strings.ToLower(tc.expectedError),
					"Error message should indicate %s validation failure", tc.name)
				assert.Equal(t, projectsCreateEndpoint.method, errorResp.Method, "Expected method to be set")
				assert.Equal(t, projectsCreateEndpoint.Path(), errorResp.Path, "Expected path to be set")
			})
		}

		// 4. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test creating projects with existing ID or name
	t.Run("create_project_already_exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. First create a valid project that will be our reference project
		projectID := uuid.Must(uuid.NewV7())
		projectName := "project_" + projectID.String()
		existingProject := map[string]any{
			"id":          projectID.String(),
			"name":        projectName,
			"description": "This is an existing project",
		}

		// Create the first project
		createResponse, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, existingProject, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to create initial project")
		defer createResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 for initial project creation. Got %d. Message: %s", createResponse.StatusCode, readResponseBody(t, createResponse))

		// 3. Define test cases for duplicate scenarios
		testCases := []struct {
			name             string
			duplicateProject map[string]any
			expectedStatus   int
			expectedError    string
		}{
			{
				name: "Project with existing ID",
				duplicateProject: map[string]any{
					"id":          projectID.String(), // Same ID as existing project
					"name":        "Different Project Name",
					"description": "This is a different project",
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "already exists",
			},
			{
				name: "Project with existing name",
				duplicateProject: map[string]any{
					"id":          uuid.Must(uuid.NewV7()).String(), // Different ID
					"name":        projectName,                      // Same name as existing project
					"description": "This is another project",
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "already exists",
			},
		}

		// 4. Run each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Try to create a project with duplicate ID or name
				response, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, tc.duplicateProject, accessTokenHeader)
				assert.NoError(t, err, "Failed to send request for %s", tc.name)
				defer response.Body.Close()

				// Verify we get the expected conflict status
				assert.Equal(t, tc.expectedStatus, response.StatusCode,
					"Expected status code %d for %s, got %d", tc.expectedStatus, tc.name, response.StatusCode)

				// Parse and verify the error response
				errorResp, err := parserResponseBody[model.HTTPMessage](t, response)
				assert.NoError(t, err, "Failed to parse error response for %s", tc.name)

				// Verify error message contains information about the conflict
				assert.Contains(t, strings.ToLower(errorResp.Message), strings.ToLower(tc.expectedError),
					"Error message should indicate conflict for %s", tc.name)
				assert.Equal(t, projectsCreateEndpoint.method, errorResp.Method, "Expected method to be set")
				assert.Equal(t, projectsCreateEndpoint.Path(), errorResp.Path, "Expected path to be set")
			})
		}

		// 5. Cleanup
		t.Cleanup(func() {
			deleteProjectByIDFromDB(t, projectID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	t.Run("require_authentication", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		project := map[string]any{
			"id":          "00000000-0000-0000-0000-000000000000",
			"name":        "Test Project",
			"description": "Test project description",
		}

		resp, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, project)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestProjectGet(t *testing.T) {
	t.Run("get_project", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a new project
		projectID, projectName, projectDesc := createTestProject(t, ctx, adminToken.AccessToken, "test_")
		assert.NotEmpty(t, projectID, "Project ID should not be empty")
		assert.NotEmpty(t, projectName, "Project name should not be empty")
		assert.NotEmpty(t, projectDesc, "Project description should not be empty")

		// 3. Get the project
		getEndpoint := projectsGetEndpoint.RewriteSlugs(projectID.String())
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get request: %v", err)
		defer getResponse.Body.Close()

		// 4. Check the response
		assert.Equal(t, http.StatusOK, getResponse.StatusCode, "Expected status code 200 OK for get. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))
		getAPIResp, err := parserResponseBody[model.Project](t, getResponse)
		assert.NoError(t, err, "Failed to parse get response body", err)

		assert.Equal(t, projectID, getAPIResp.ID, "Expected project ID to match")
		assert.Equal(t, projectName, getAPIResp.Name, "Expected project name to match")
		assert.Equal(t, projectDesc, getAPIResp.Description, "Expected project description to match")
		assert.Equal(t, pointerTo(false), getAPIResp.Disabled, "Expected disabled to be false")
		assert.Equal(t, pointerTo(false), getAPIResp.System, "Expected system to be false")

		// 5. Cleanup
		t.Cleanup(func() {
			deleteProjectByIDFromDB(t, projectID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test retrieving a non-existent project
	t.Run("get_project_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Generate a random UUID that doesn't exist in the database
		nonExistentProjectID := uuid.Must(uuid.NewV7())

		// 3. Try to get the non-existent project
		getEndpoint := projectsGetEndpoint.RewriteSlugs(nonExistentProjectID.String())
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to get non-existent project")
		defer getResponse.Body.Close()

		// 4. Check that we get a 404 Not Found response
		assert.Equal(t, http.StatusNotFound, getResponse.StatusCode, "Expected status code 404 for non-existent project. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))

		// 5. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, getResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// Verify error message contains information about the project not being found
		assert.Contains(t, errorResp.Message, "not found", "Error message should indicate that the project was not found")
		assert.Equal(t, getEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, getEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test retrieving a project with an invalid ID format
	t.Run("get_project_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Try to get a project with an invalid ID format (not a UUID)
		invalidProjectID := "not-a-valid-uuid"
		getEndpoint := projectsGetEndpoint.RewriteSlugs(invalidProjectID)
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to get project with invalid ID")
		defer getResponse.Body.Close()

		// 3. Check that we get a 400 Bad Request response
		assert.Equal(t, http.StatusBadRequest, getResponse.StatusCode, "Expected status code 400 for invalid project ID format. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))

		// 4. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, getResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// Verify error message contains information about the invalid UUID format
		assert.Contains(t, errorResp.Message, "invalid", "Error message should indicate that the project ID format is invalid")
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

		getEndpoint := projectsGetEndpoint.RewriteSlugs("00000000-0000-0000-0000-000000000000")
		resp, err := sendHTTPRequest(t, ctx, getEndpoint, nil)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestProjectDelete(t *testing.T) {
	t.Run("delete_project", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a new project
		projectID, _, _ := createTestProject(t, ctx, adminToken.AccessToken, "test_")
		assert.NotEmpty(t, projectID, "Project ID should not be empty")

		// 3. Delete the project
		deleteEndpoint := projectsDeleteEndpoint.RewriteSlugs(projectID.String())
		deleteResponse, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending delete request: %v", err)
		defer deleteResponse.Body.Close()

		// 4. Check the delete response
		assert.Equal(t, http.StatusOK, deleteResponse.StatusCode, "Expected status code 200 OK for delete. Got %d. Message: %s", deleteResponse.StatusCode, readResponseBody(t, deleteResponse))
		deleteAPIResp, err := parserResponseBody[model.HTTPMessage](t, deleteResponse)
		assert.NoError(t, err, "Failed to parse delete response body")

		assert.Equal(t, model.ProjectsProjectDeletedSuccessfully, deleteAPIResp.Message, "Unexpected delete response message")
		assert.Equal(t, deleteEndpoint.method, deleteAPIResp.Method, "Expected method to be set for delete")
		assert.Equal(t, deleteEndpoint.Path(), deleteAPIResp.Path, "Expected path to be set for delete")

		// 5. Verify project is actually deleted (try to get it)
		getEndpoint := projectsGetEndpoint.RewriteSlugs(projectID.String())
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get request after delete: %v", err)
		defer getResponse.Body.Close()
		assert.Equal(t, http.StatusNotFound, getResponse.StatusCode, "Expected status code 404 Not Found after deletion. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))

		// 6. Cleanup admin user
		t.Cleanup(func() {
			// Project should already be deleted by the test, but try again just in case
			deleteProjectByIDFromDB(t, projectID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test deleting a project with an invalid ID format
	t.Run("delete_project_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Try to delete a project with an invalid ID format (not a UUID)
		invalidProjectID := "not-a-valid-uuid"
		deleteEndpoint := projectsDeleteEndpoint.RewriteSlugs(invalidProjectID)

		deleteResponse, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to delete project with invalid ID")
		defer deleteResponse.Body.Close()

		// 3. Check that we get a 400 Bad Request response
		assert.Equal(t, http.StatusBadRequest, deleteResponse.StatusCode, "Expected status code 400 for invalid project ID format. Got %d. Message: %s", deleteResponse.StatusCode, readResponseBody(t, deleteResponse))

		// 4. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, deleteResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// 5. Verify error message contains information about the invalid UUID format
		assert.Contains(t, errorResp.Message, "invalid", "Error message should indicate that the project ID format is invalid")
		assert.Equal(t, deleteEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, deleteEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test deleting a non-existent project
	t.Run("delete_project_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Generate a UUID that doesn't exist in the database
		nonExistentProjectID := uuid.Must(uuid.NewV7())
		deleteEndpoint := projectsDeleteEndpoint.RewriteSlugs(nonExistentProjectID.String())

		deleteResponse, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to delete non-existent project")
		defer deleteResponse.Body.Close()

		// 3. Check the response - this should still return StatusOK even though the project doesn't exist
		// This is because deleting a non-existent resource is considered idempotent in RESTful APIs
		assert.Equal(t, http.StatusOK, deleteResponse.StatusCode, "Expected status code 200 OK for deleting non-existent project. Got %d. Message: %s", deleteResponse.StatusCode, readResponseBody(t, deleteResponse))

		// 4. Parse and verify the success response
		deleteAPIResp, err := parserResponseBody[model.HTTPMessage](t, deleteResponse)
		assert.NoError(t, err, "Failed to parse response body")

		// 5. Verify success message for deletion
		assert.Equal(t, model.ProjectsProjectDeletedSuccessfully, deleteAPIResp.Message, "Expected success message")
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

		deleteEndpoint := projectsDeleteEndpoint.RewriteSlugs("00000000-0000-0000-0000-000000000000")
		resp, err := sendHTTPRequest(t, ctx, deleteEndpoint, nil)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestProjectUpdate(t *testing.T) {
	t.Run("update_project", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a new project
		projectID, originalName, _ := createTestProject(t, ctx, adminToken.AccessToken, "t_")

		// 3. Update the project
		updatedName := "updated_" + originalName
		updatedDesc := "Updated description " + projectID.String()
		disabled := true
		updatedProject := map[string]any{
			"name":        updatedName,
			"description": updatedDesc,
			"disabled":    disabled,
		}

		updateEndpoint := projectsUpdateEndpoint.RewriteSlugs(projectID.String())
		updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, updatedProject, accessTokenHeader)
		assert.NoError(t, err, "Error sending update request: %v", err)
		defer updateResponse.Body.Close()

		// 4. Check the update response
		assert.Equal(t, http.StatusOK, updateResponse.StatusCode, "Expected status code 200 OK for update. Got %d. Message: %s", updateResponse.StatusCode, readResponseBody(t, updateResponse))
		updateAPIResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
		assert.NoError(t, err, "Failed to parse update response body")

		assert.Equal(t, model.ProjectsProjectUpdatedSuccessfully, updateAPIResp.Message, "Unexpected update response message")
		assert.Equal(t, updateEndpoint.method, updateAPIResp.Method, "Expected method to be set for update")
		assert.Equal(t, updateEndpoint.Path(), updateAPIResp.Path, "Expected path to be set for update")

		// 5. Verify project is actually updated (get it again)
		getEndpoint := projectsGetEndpoint.RewriteSlugs(projectID.String())
		getResponse, err := sendHTTPRequest(t, ctx, getEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Error sending get request after update: %v", err)
		defer getResponse.Body.Close()
		assert.Equal(t, http.StatusOK, getResponse.StatusCode, "Expected status code 200 OK when getting updated project. Got %d. Message: %s", getResponse.StatusCode, readResponseBody(t, getResponse))

		getAPIResp, err := parserResponseBody[model.Project](t, getResponse)
		assert.NoError(t, err, "Failed to parse get response body for updated project")

		assert.Equal(t, projectID, getAPIResp.ID, "Expected project ID to remain the same")
		assert.Equal(t, updatedName, getAPIResp.Name, "Expected project name to be updated")
		assert.Equal(t, updatedDesc, getAPIResp.Description, "Expected project description to be updated")
		assert.Equal(t, pointerTo(disabled), getAPIResp.Disabled, "Expected disabled to be updated")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteProjectByIDFromDB(t, projectID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test updating a project with invalid data format
	t.Run("update_project_bad_request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a valid project first
		projectID, _, _ := createTestProject(t, ctx, adminToken.AccessToken, "bad_req_test_")

		// 3. Set up test cases for different bad request scenarios
		testCases := []struct {
			name          string
			updateData    map[string]any
			expectedCode  int
			expectedError string
		}{
			{
				name:          "Empty project name",
				updateData:    map[string]any{"name": ""},
				expectedCode:  http.StatusBadRequest,
				expectedError: "cannot be empty",
			},
			{
				name:          "Name too long",
				updateData:    map[string]any{"name": string(make([]byte, 256))}, // Very long name
				expectedCode:  http.StatusBadRequest,
				expectedError: "must be between",
			},
			{
				name:          "Description too long",
				updateData:    map[string]any{"description": string(make([]byte, 5000))}, // Very long description
				expectedCode:  http.StatusBadRequest,
				expectedError: "must be between",
			},
		}

		// 4. Run each test case
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				updateEndpoint := projectsUpdateEndpoint.RewriteSlugs(projectID.String())
				updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, tc.updateData, accessTokenHeader)
				assert.NoError(t, err, "Failed to send request for %s", tc.name)
				defer updateResponse.Body.Close()

				// Verify we get the expected status code
				assert.Equal(t, tc.expectedCode, updateResponse.StatusCode,
					"Expected status code %d for %s, got %d", tc.expectedCode, tc.name, updateResponse.StatusCode)

				// Parse and verify the error response
				errorResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
				assert.NoError(t, err, "Failed to parse error response for %s", tc.name)

				// Verify error message contains relevant information
				assert.Contains(t, strings.ToLower(errorResp.Message), strings.ToLower(tc.expectedError),
					"Error message should indicate %s validation failure", tc.name)
				assert.Equal(t, updateEndpoint.method, errorResp.Method, "Expected method to be set")
				assert.Equal(t, updateEndpoint.Path(), errorResp.Path, "Expected path to be set")
			})
		}

		// 5. Cleanup
		t.Cleanup(func() {
			deleteProjectByIDFromDB(t, projectID)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test updating a non-existent project
	t.Run("update_project_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Generate a UUID that doesn't exist in the database
		nonExistentProjectID := uuid.Must(uuid.NewV7())
		updateEndpoint := projectsUpdateEndpoint.RewriteSlugs(nonExistentProjectID.String())

		updatedProject := map[string]any{
			"name":        "UpdatedProjectName",
			"description": "Updated project description",
			"disabled":    true,
		}

		updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, updatedProject, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to update non-existent project")
		defer updateResponse.Body.Close()

		// 3. Check that we get a 404 Not Found response
		assert.Equal(t, http.StatusNotFound, updateResponse.StatusCode, "Expected status code 404 for non-existent project. Got %d. Message: %s", updateResponse.StatusCode, readResponseBody(t, updateResponse))

		// 4. Parse and verify the error response
		errorResp, err := parserResponseBody[model.HTTPMessage](t, updateResponse)
		assert.NoError(t, err, "Failed to parse error response")

		// 5. Verify error message contains information about the project not being found
		assert.Contains(t, errorResp.Message, "not found", "Error message should indicate that the project was not found")
		assert.Equal(t, updateEndpoint.method, errorResp.Method, "Expected method to be set")
		assert.Equal(t, updateEndpoint.Path(), errorResp.Path, "Expected path to be set")

		// 6. Cleanup
		t.Cleanup(func() {
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	// Test updating a project with a name that already exists (conflict case)
	t.Run("update_project_conflict_name", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create two projects with different names
		// First project - this is the one we'll try to update
		projectID1, _, _ := createTestProject(t, ctx, adminToken.AccessToken, "conflict_test1_")

		// Second project - we'll try to use this project's name when updating the first project
		projectID2, projectName2, _ := createTestProject(t, ctx, adminToken.AccessToken, "conflict_test2_")

		// 3. Try to update the first project with the second project's name
		updateProject1WithConflict := map[string]any{
			"name": projectName2, // This will cause a conflict because projectName2 is already being used
		}

		updateEndpoint := projectsUpdateEndpoint.RewriteSlugs(projectID1.String())
		updateResponse, err := sendHTTPRequest(t, ctx, updateEndpoint, updateProject1WithConflict, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to update project with conflicting name")
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
			deleteProjectByIDFromDB(t, projectID1)
			deleteProjectByIDFromDB(t, projectID2)
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	t.Run("require_authentication", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		updateEndpoint := projectsUpdateEndpoint.RewriteSlugs("00000000-0000-0000-0000-000000000000")
		updateData := map[string]any{
			"name":        "Updated Project Name",
			"description": "Updated project description",
		}

		resp, err := sendHTTPRequest(t, ctx, updateEndpoint, updateData)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestProjectList(t *testing.T) {
	t.Run("list_projects", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Create an administrator user and get the token
		adminToken := getAdminUserTokens(t)
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		accessTokenHeader := map[string]string{
			"Authorization": "Bearer " + adminToken.AccessToken,
		}

		// 2. Create a couple of new projects
		projectIDs := []uuid.UUID{uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())}
		projectsToCreate := []map[string]any{}

		for i, projectID := range projectIDs {
			project := map[string]any{
				"id":          projectID.String(),
				"name":        "test_" + projectID.String(),
				"description": "This is a test project for list " + projectID.String(),
			}
			projectsToCreate = append(projectsToCreate, project)

			createResponse, err := sendHTTPRequest(t, ctx, projectsCreateEndpoint, project, accessTokenHeader)
			assert.NoError(t, err, "Failed to send request to create project %d", i+1)

			if createResponse != nil {
				defer createResponse.Body.Close()
				assert.Equal(t, http.StatusCreated, createResponse.StatusCode, "Expected status code 201 for project %d. Got %d. Message: %s", i+1, createResponse.StatusCode, readResponseBody(t, createResponse))
			}
		}

		// Wait briefly to ensure projects are created
		time.Sleep(1 * time.Second)

		// 3. List the projects
		listResponse, err := sendHTTPRequest(t, ctx, projectsListEndpoint, nil, accessTokenHeader)
		assert.NoError(t, err, "Failed to send request to list projects")
		defer listResponse.Body.Close()

		// 4. Check the list response
		assert.Equal(t, http.StatusOK, listResponse.StatusCode, "Expected status code 200 for list. Got %d. Message: %s", listResponse.StatusCode, readResponseBody(t, listResponse))
		// Assuming model.ListProjectsOutput exists and has an Items field []model.Project
		listAPIResp, err := parserResponseBody[model.ListProjectsOutput](t, listResponse)
		assert.NoError(t, err, "Failed to parse list response body")

		// 5. Verify the created projects are in the list
		foundCount := 0
		projectMap := make(map[string]bool) // Use name for checking presence
		for _, createdProject := range projectsToCreate {
			projectMap[createdProject["name"].(string)] = true
		}

		for _, listedProject := range listAPIResp.Items {
			if _, ok := projectMap[listedProject.Name]; ok {
				foundCount++
				// Optionally assert other fields match
				for _, createdProject := range projectsToCreate {
					if createdProject["name"] == listedProject.Name {
						assert.Equal(t, createdProject["description"], listedProject.Description)
						break
					}
				}
			}
		}
		assert.GreaterOrEqual(t, foundCount, len(projectsToCreate), "Expected to find at least the created projects in the list")

		// 6. Cleanup
		t.Cleanup(func() {
			for _, projectID := range projectIDs {
				deleteProjectByIDFromDB(t, projectID)
			}
			deleteUserByIDFromDB(t, adminToken.UserID)
		})
	})

	t.Run("require_authentication", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		resp, err := sendHTTPRequest(t, ctx, projectsListEndpoint, nil)
		assert.NoError(t, err, "Failed to send request without authentication")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
