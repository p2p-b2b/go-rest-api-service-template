//go:build integration

package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"github.com/stretchr/testify/assert"
)

// apiEndpoint is a struct to hold endpoint information
// It contains the endpoint URL, path, and HTTP method
type apiEndpoint struct {
	apiURL     *url.URL
	requestURL *url.URL
	path       string
	method     string
}

func (e *apiEndpoint) String() string {
	return e.requestURL.Path
}

func (e *apiEndpoint) Path() string {
	return e.path
}

// Clone creates a new apiEndpoint with the same method and path
func (e *apiEndpoint) Clone() *apiEndpoint {
	return newAPIEndpoint(e.method, e.path)
}

// RewriteSlugs clones the apiEndpoint and replaces the slugs in the path
// with the provided slugs.
func (e *apiEndpoint) RewriteSlugs(slugs ...string) *apiEndpoint {
	if len(slugs) != 0 {
		// Clone the apiEndpoint
		eClone := newAPIEndpoint(e.method, e.path)

		eClone.path = replaceSlugs(e.path, slugs...)
		apiEndpointURL, err := url.Parse(apiEndpointURL)
		if err != nil {
			panic(fmt.Sprintf("❌ Failed to parse API endpoint URL: %v", err))
		}

		requestURL, err := url.Parse(apiEndpointURL.String() + eClone.path)
		if err != nil {
			panic(fmt.Sprintf("❌ Failed to parse request URL: %v", err))
		}

		eClone.requestURL = requestURL

		return eClone
	}

	// If no slugs are provided, return the original apiEndpoint
	return e
}

func (e *apiEndpoint) SetQueryParam(key, value string) {
	if e.requestURL == nil {
		panic("❌ requestURL is nil")
	}

	query := e.requestURL.Query()
	query.Set(key, value)
	e.requestURL.RawQuery = query.Encode()
}

func (e *apiEndpoint) SetQueryParams(params map[string]string) {
	if e.requestURL == nil {
		panic("❌ requestURL is nil")
	}

	query := e.requestURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	e.requestURL.RawQuery = query.Encode()
}

// newAPIEndpoint is a helper function to create a new API endpoint
// It takes the HTTP method and path as parameters
func newAPIEndpoint(method, path string) *apiEndpoint {
	baseURL, err := url.Parse(apiEndpointURL)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse API endpoint URL: %v", err))
	}

	requestURL, err := url.Parse(baseURL.String() + path)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse request URL: %v", err))
	}

	return &apiEndpoint{
		apiURL:     baseURL,
		requestURL: requestURL,
		path:       path,
		method:     method,
	}
}

func newMailAPIEndpoint(method, path string) *apiEndpoint {
	apiEndpointURL, err := url.Parse(mailServerEndpointURL)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse API endpoint URL: %v", err))
	}

	requestURL, err := url.Parse(apiEndpointURL.String() + path)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse request URL: %v", err))
	}

	return &apiEndpoint{
		apiURL:     apiEndpointURL,
		requestURL: requestURL,
		method:     method,
	}
}

// pointerTo is a helper function to return a pointer to the given value
// It takes a value of any type and returns a pointer to that value
func pointerTo[T any](value T) *T {
	return &value
}

// sendHTTPRequest is a helper function to send HTTP requests
// It takes a testing.T object, an apiEndpoint object, and a request body
// It returns the HTTP response and an error if any
// It uses the default HTTP client to send the request
// It marshals the request body to JSON if provided
func sendHTTPRequest(t *testing.T, ctx context.Context, endpoint *apiEndpoint, body map[string]any, headers ...map[string]string) (*http.Response, error) {
	if t == nil {
		t = &testing.T{}
	}
	t.Helper()

	client := http.DefaultClient

	var jsonBody io.ReadWriter
	var err error
	if body != nil {
		jsonBody = new(bytes.Buffer)
		enc := json.NewEncoder(jsonBody)
		enc.SetEscapeHTML(false)
		err = enc.Encode(body)
		if err != nil {
			t.Errorf("Failed to encode request body: %v", err)
			return nil, err
		}

	}

	// t.Logf("Sending %s request to %s with body: %v", endpoint.method, endpoint.requestURL.String(), body)
	req, err := http.NewRequestWithContext(ctx, endpoint.method, endpoint.requestURL.String(), jsonBody)
	if err != nil {
		return nil, err
	}

	// obligatory headers
	req.Header.Set("Accept", "application/json")

	// optional headers
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
		return nil, err
	}

	return resp, nil
}

// parserResponseBody is a helper function generic to parse the response body
// It takes an HTTP response and a pointer to a struct to unmarshal the response into
// It returns the unmarshaled struct and an error if any
func parserResponseBody[T any](t *testing.T, resp *http.Response) (T, error) {
	if t == nil {
		t = &testing.T{}
	}
	t.Helper()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// t.Logf("Response: %+v", result)

	return result, nil
}

// generatePassword generates a random password of the specified length
// It uses the crypto/rand package to generate a secure random password
// It returns the generated password as a string
func generatePassword(t *testing.T, length ...int) string {
	t.Helper()
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"

	if len(length) == 0 {
		length = append(length, 12) // Default length
	}

	password := make([]byte, length[0])
	_, err := io.ReadFull(rand.Reader, password)
	if err != nil {
		t.Fatalf("Failed to generate password: %v", err)
		return ""
	}

	for i := range password {
		password[i] = charset[int(password[i])%len(charset)]
	}

	return string(password)
}

// generateRandomName generates a random name
// It uses a charset = "abcdefghijklmnopqrstuvwxyz"
// and the name has a dynamic length between 3 and 10 characters
func generateRandomName(t *testing.T, prefix string) string {
	t.Helper()

	const charset = "abcdefghijklmnopqrstuvwxyz"

	// Generate random length between 3 and 10
	randomBytes := make([]byte, 1)
	_, err := rand.Read(randomBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
		return ""
	}
	nameLength := 3 + int(randomBytes[0]%8)
	name := make([]byte, nameLength)

	_, err = io.ReadFull(rand.Reader, name)
	if err != nil {
		t.Fatalf("Failed to generate name: %v", err)
		return ""
	}

	for i := range name {
		name[i] = charset[int(name[i])%len(charset)]
	}

	if prefix != "" {
		prefix += "_"
	}

	return fmt.Sprintf("%s%s", prefix, string(name))
}

// readResponseBody is a helper function to read the response body
// It takes a testing.T object and an HTTP response
// It returns the response body as a byte slice and an error if any
func readResponseBody(t *testing.T, resp *http.Response) string {
	if t == nil {
		t = &testing.T{}
	}
	t.Helper()

	if resp == nil {
		t.Errorf("Response is nil, cannot read body")
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
		return ""
	}

	// Close the original body
	resp.Body.Close()

	// Create a new ReadCloser and replace the response body
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body)
}

// getAdminUserTokens is a helper function to get the admin user tokens
// It takes a testing.T object and the admin user's email address
// It returns the access token and refresh token as strings
func getAdminUserTokens(t *testing.T) model.LoginUserResponse {
	t.Helper()

	ctx := context.Background()

	tx, txErr := testDBPool.Begin(ctx)
	if txErr != nil {
		t.Fatalf("Failed to begin transaction: %v", txErr)
	}

	// 1. Insert a user into the database
	query1 := `
        INSERT INTO users (id, first_name, last_name, email, password_hash, disabled, admin)
        VALUES ($1, $2, $3, $4, $5, $6, $7);
    `

	userID, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate user ID: %v", err)
	}

	firstName, lastName, email := generateUserData(t)
	password := generatePassword(t)
	hashPwd, err := service.HashAndSaltPassword(password)
	assert.NoError(t, err, "Failed to hash password")

	_, txErr = tx.Exec(context.Background(), query1,
		userID,
		firstName,
		lastName,
		email,
		hashPwd,
		false,
		true, // admin = true
	)
	if txErr != nil {
		t.Fatalf("Failed to insert user into database: %v", txErr)
	}

	// 2. Get the role_id for the admin role and assign it to the user
	query2 := `
    WITH
        role_id AS (
            SELECT id FROM roles WHERE name = 'Administrator' LIMIT 1
        )
        INSERT INTO users_roles (users_id, roles_id)
        SELECT $1, id FROM role_id;
    `

	_, txErr = tx.Exec(context.Background(), query2, userID)
	if txErr != nil {
		t.Fatalf("Failed to assign role to user in database directly on integration test: %v", txErr)
	}

	if txErr != nil {
		if err := tx.Rollback(ctx); err != nil {
			t.Errorf("Failed to rollback transaction: %v", err)
		}
	} else {
		if err := tx.Commit(ctx); err != nil {
			t.Errorf("Failed to commit transaction: %v", err)
		}
	}

	// 3. Login the user
	// wait for login verification in the database
	// time.Sleep(200 * time.Millisecond)

	loginUser := map[string]any{
		"email":    email,
		"password": password,
	}

	loginResponse, err := sendHTTPRequest(t, ctx, authLoginEndpoint, loginUser)
	assert.NoError(t, err)
	defer loginResponse.Body.Close()

	assert.Equal(t, loginResponse.StatusCode, http.StatusOK, "Expected status code 200 OK. Got %d. Message: %s", loginResponse.StatusCode, readResponseBody(t, loginResponse))
	loginAPIResp, err := parserResponseBody[model.LoginUserResponse](t, loginResponse)
	assert.NoError(t, err)

	assert.NotEmpty(t, loginAPIResp.AccessToken, "Expected access token to be present")
	assert.NotEmpty(t, loginAPIResp.RefreshToken, "Expected refresh token to be present")
	assert.NotEmpty(t, loginAPIResp.UserID, "Expected user ID to be present")

	assert.Equal(t, loginAPIResp.UserID.String(), userID.String(), "Expected user ID to match")
	assert.Equal(t, loginAPIResp.TokenType, "Bearer", "Expected token type to be Bearer")

	return loginAPIResp
}

// getVerifyLinkFromEmail is a helper function to get the verification link from the email
// It takes a testing.T object, the sender's email address, and the recipient's email address
func getVerifyLinkFromEmail(t *testing.T, from, to string) string {
	mailSearchEndpoint := newMailAPIEndpoint(http.MethodGet, "/search")
	apiQueryParam := "query="
	apiQuery := fmt.Sprintf("From:%s To:%s", from, to)
	mailSearchEndpoint.requestURL.RawQuery = apiQueryParam + url.QueryEscape(apiQuery)

	resp, err := sendHTTPRequest(t, context.Background(), mailSearchEndpoint, nil)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	type searchResponse struct {
		Messages []struct {
			ID string `json:"ID"`
		} `json:"Messages"`
	}

	mailResponse, err := parserResponseBody[searchResponse](t, resp)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if len(mailResponse.Messages) == 0 {
		t.Fatalf("No emails found for From: %s, To: %s", from, to)
	}

	mailID := mailResponse.Messages[0].ID
	mailGetMessageSourceEndpoint := newMailAPIEndpoint(http.MethodGet, fmt.Sprintf("/message/%s/raw", mailID))

	rawContentResp, err := sendHTTPRequest(t, context.Background(), mailGetMessageSourceEndpoint, nil)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer rawContentResp.Body.Close()

	if rawContentResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", rawContentResp.StatusCode)
	}

	rawContent, err := io.ReadAll(rawContentResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// regex to extract the verification link from the email content
	// https://regex101.com/r/2uC89J/2
	re := regexp.MustCompile(`http:\/\/[^\s]+\/verify\/[a-zA-Z0-9._-]+`)
	matches := re.FindStringSubmatch(string(rawContent))
	if len(matches) < 1 {
		t.Fatalf("No verification link found in the email content")
	}
	verifyLink := matches[0]

	// Check if the verification link is valid
	_, err = url.ParseRequestURI(verifyLink)
	if err != nil {
		t.Fatalf("Invalid verification link: %v", err)
	}

	return verifyLink
}

// replaceSlugs is a helper function to build a path with multiple slugs.
// NOTE: This is used for testing purposes only.
// replace the slugs in the path with the slugs provided in the order they are provided.
// if no slugs are provided, the first slug found in the path will be replaced.
func replaceSlugs(path string, slugs ...string) string {
	re := regexp.MustCompile(`\{([^}]+)\}`)

	var val string
	if len(slugs) == 0 {
		val = ""
	} else {
		for _, slug := range slugs {
			found := re.FindStringSubmatchIndex(path)
			if found == nil {
				break
			}

			start, end := found[0], found[1]
			path = strings.Replace(path, path[start:end], slug, 1)
		}

		val = path
	}

	return val
}

// removeAPIEndpointFromURL is a helper function to remove the API endpoint from the URL
// It takes a URL string and returns the URL string without the API endpoint
func removeAPIEndpointFromURL(urlStr string) string {
	return strings.TrimPrefix(urlStr, apiEndpointURL)
}

// deleteUserByIDFromDB is a helper function to delete a user from the database
// It takes a testing.T object and the user's ID
// It returns an error if any
func deleteUserByIDFromDB(t *testing.T, userID uuid.UUID) {
	t.Helper()

	query := `DELETE FROM users WHERE id = $1;`
	_, err := testDBPool.Exec(context.Background(), query, userID)
	if err != nil {
		t.Errorf("Failed to delete user from database: %v", err)
	}
}

// deleteRoleByIDFromDB is a helper function to delete a role from the database
// It takes a testing.T object and the role's ID
// It returns an error if any
func deleteRoleByIDFromDB(t *testing.T, roleID uuid.UUID) {
	t.Helper()

	query := `DELETE FROM roles WHERE id = $1;`
	_, err := testDBPool.Exec(context.Background(), query, roleID)
	if err != nil {
		t.Errorf("Failed to delete role from database: %v", err)
	}
}

// deletePolicyByIDFromDB is a helper function to delete a policy from the database
// It takes a testing.T object and the policy's ID
// It returns an error if any
func deletePolicyByIDFromDB(t *testing.T, policyID uuid.UUID) {
	t.Helper()

	query := `DELETE FROM policies WHERE id = $1;`
	_, err := testDBPool.Exec(context.Background(), query, policyID)
	if err != nil {
		t.Errorf("Failed to delete policy from database: %v", err)
	}
}

// enableUserByEmailFromDB is a helper function to enable a user in the database
// It takes a testing.T object and the user's email address
// It returns an error if any
func enableUserByEmailFromDB(t *testing.T, email string) {
	t.Helper()

	query := `UPDATE users SET disabled = false WHERE email = $1;`
	_, err := testDBPool.Exec(context.Background(), query, email)
	if err != nil {
		t.Errorf("Failed to enable user in database: %v", err)
	}
}

// deleteUserByEmailFromDB is a helper function to delete a user from the database
// It takes a testing.T object and the user's email address
// It returns an error if any
func deleteUserByEmailFromDB(t *testing.T, email string) {
	t.Helper()

	query := `DELETE FROM users WHERE email = $1;`
	_, err := testDBPool.Exec(context.Background(), query, email)
	if err != nil {
		t.Errorf("Failed to delete user from database: %v", err)
	}
}

// deleteAllEmails is a helper function to delete all emails from the mail server
// It takes a testing.T object and the sender's email address
// It returns an error if any
func deleteAllEmails() error {
	mailListEndpoint := newMailAPIEndpoint(http.MethodGet, "/messages")

	listResp, err := sendHTTPRequest(nil, context.Background(), mailListEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status code 200, got %d", listResp.StatusCode)
	}

	type listResponse struct {
		Messages []struct {
			ID string `json:"ID"`
		} `json:"Messages"`
	}

	mailListResponse, err := parserResponseBody[listResponse](nil, listResp)
	if err != nil {
		return fmt.Errorf("failed to parse response body: %v", err)
	}

	mailDeletePayload := make(map[string]any)
	mailIDsToDelete := make([]string, len(mailListResponse.Messages))

	for i, message := range mailListResponse.Messages {
		mailIDsToDelete[i] = message.ID
	}

	mailDeletePayload["IDs"] = strings.Join(mailIDsToDelete, ",")

	mailDeleteEndpoint := newMailAPIEndpoint(http.MethodDelete, "/messages")

	deleteResp, err := sendHTTPRequest(nil, context.Background(), mailDeleteEndpoint, mailDeletePayload)
	if err != nil {
		return fmt.Errorf("failed to delete emails: %v", err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status code 200, got %d", deleteResp.StatusCode)
	}

	return nil
}

// generateUserData generates a random email address
// It uses a charset = "abcdefghijklmnopqrstuvwxyz"
// and the email has the patter word1.word2@<mailDomain>
// the mailDomain is an optional parameter and the size of the word1 and word2 is
// dynamic between 3 and 10 characters
func generateUserData(t *testing.T, mailDomain ...string) (firstName, lastName, email string) {
	t.Helper()

	const charset = "abcdefghijklmnopqrstuvwxyz"

	// if mailDomain is not provided, use a default domain
	if len(mailDomain) == 0 {
		mailDomain = append(mailDomain, "mail.com")
	}

	// Generate random length between 3 and 10
	randomBytes := make([]byte, 2)
	_, err := rand.Read(randomBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
		return "", "", ""
	}

	word1Length := 3 + int(randomBytes[0]%8)
	word2Length := 3 + int(randomBytes[1]%8)

	word1 := make([]byte, word1Length)
	word2 := make([]byte, word2Length)

	_, err = io.ReadFull(rand.Reader, word1)
	if err != nil {
		t.Fatalf("Failed to generate email: %v", err)
	}

	_, err = io.ReadFull(rand.Reader, word2)
	if err != nil {
		t.Fatalf("Failed to generate email: %v", err)
		return "", "", ""
	}

	for i := range word1 {
		word1[i] = charset[int(word1[i])%len(charset)]
	}
	for i := range word2 {
		word2[i] = charset[int(word2[i])%len(charset)]
	}

	firstName = string(word1)
	lastName = string(word2)
	email = fmt.Sprintf("%s.%s@%s", string(word1), string(word2), mailDomain[0])

	// Check if the email is valid
	_, err = mail.ParseAddress(email)
	if err != nil {
		t.Fatalf("Invalid email address: %v", err)
		return "", "", ""
	}

	return firstName, lastName, email
}

// deleteProjectByIDFromDB is a helper function to delete a project from the database
// It takes a testing.T object and the project's ID
// It returns an error if any
func deleteProjectByIDFromDB(t *testing.T, projectID uuid.UUID) {
	t.Helper()

	query := `DELETE FROM projects WHERE id = $1;`
	_, err := testDBPool.Exec(context.Background(), query, projectID)
	if err != nil {
		t.Errorf("Failed to delete project from database: %v", err)
	}
}

// createProjectInDB is a helper function to create a project in the database
// It takes a testing.T object and the project name
// It returns the created project and an error if any
// If the project ID is not provided, it generates a new UUID
// If the project name is not provided, it generates a random name
// If the project description is not provided, it uses an empty string
func createProjectInDB(t *testing.T, id uuid.UUID, name, description string) (*model.Project, error) {
	t.Helper()

	if id == uuid.Nil {
		var err error
		id, err = uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate project ID: %v", err)
			return nil, err
		}
	}

	if name == "" {
		name = fmt.Sprintf("Project-%d", time.Now().UnixNano())
	}

	if description == "" {
		description = fmt.Sprintf("Description for project %s", name)
	}

	query := `
        INSERT INTO projects (id, name, description)
        VALUES ($1, $2, $3)
        RETURNING id, name, description;
    `

	var project model.Project
	err := testDBPool.QueryRow(context.Background(), query, id, name, description).Scan(&project.ID, &project.Name, &project.Description)
	if err != nil {
		t.Fatalf("Failed to insert project into database: %v", err)
		return nil, err
	}

	return &project, nil
}

// assignProjectToUserInDB is a helper function to assign a project to a user in the database
// It creates an entry in the projects_users table to establish the relationship
func assignProjectToUserInDB(t *testing.T, projectID, userID uuid.UUID) error {
	t.Helper()

	query := `
        INSERT INTO projects_users (projects_id, users_id)
        VALUES ($1, $2)
        ON CONFLICT (projects_id, users_id) DO NOTHING;
    `

	_, err := testDBPool.Exec(context.Background(), query, projectID, userID)
	if err != nil {
		t.Fatalf("Failed to assign project to user in database: %v", err)
		return err
	}

	return nil
}

// createProductInDB is a helper function to create a product in the database
// It takes a testing.T object, the product name, and the project ID
// It returns the created product and an error if any
func createProductInDB(t *testing.T, id, projectID uuid.UUID, name string, paymentProcessors []model.ProductPaymentProcessorRequest) (*model.Product, error) {
	ctx := context.Background()

	if id == uuid.Nil {
		var err error
		id, err = uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate product ID: %v", err)
			return nil, err
		}
	}

	if projectID == uuid.Nil {
		t.Fatalf("Project ID cannot be nil")
		return nil, fmt.Errorf("project ID cannot be nil")
	}

	if name == "" {
		name = fmt.Sprintf("Product-%d", time.Now().UnixNano())
	}

	description := fmt.Sprintf("Description for product %s", name)

	tx, txErr := testDBPool.Begin(ctx)
	if txErr != nil {
		t.Fatalf("Failed to begin transaction: %v", txErr)
		return nil, txErr
	}

	defer func() {
		if txErr != nil {
			if err := tx.Rollback(ctx); err != nil {
				t.Errorf("Failed to rollback transaction: %v", err)
			}
		} else {
			if err := tx.Commit(ctx); err != nil {
				t.Errorf("Failed to commit transaction: %v", err)
			}
		}
	}()

	query1 := `
        INSERT INTO products (id, projects_id, name, description)
        VALUES ($1, $2, $3, $4);
    `

	_, txErr = tx.Exec(ctx, query1, id, projectID, name, description)
	if txErr != nil {
		t.Fatalf("Failed to insert product into database: %v", txErr)
		return nil, txErr
	}

	if len(paymentProcessors) > 0 {
		query2 := `
        INSERT INTO products_payment_processors (product_id, payment_processor_id, payment_processor_product_id)
        VALUES ($1, $2, $3);
        `

		for _, pp := range paymentProcessors {
			_, txErr = tx.Exec(ctx, query2,
				id,
				pp.PaymentProcessorID,
				pp.PaymentProcessorProductID,
			)
			if txErr != nil {
				t.Errorf("Failed to insert product payment processor into database: %v", txErr)
				return nil, txErr
			}
		}
	}

	query3 := `
        SELECT
            p.id,
            p.name,
            p.description,
            p.created_at,
            p.updated_at,
            array_agg(DISTINCT(ARRAY[prj.id::varchar, prj.name])) AS projects
        FROM products AS p
            LEFT JOIN projects prj ON prj.id = p.projects_id
        WHERE p.id = $1 AND p.projects_id = $2
        GROUP BY p.id;
    `

	row := tx.QueryRow(ctx, query3, id, projectID)

	var product model.Product
	var projects []string

	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.CreatedAt, &product.UpdatedAt, &projects)
	if err != nil {
		t.Fatalf("Failed to scan product row: %v", err)
		return nil, err
	}

	// PostgreSQL -> {{f282315d-1e65-43fd-8f12-a9c27be60c9e, "Project Name"}}
	// Go -> [f282315d-1e65-43fd-8f12-a9c27be60c9e, Project Name]
	for i := 0; i < len(projects); i += 2 {
		id, err := uuid.Parse(projects[i])
		if err != nil {
			t.Errorf("invalid project ID: %v", projects[i])
			return nil, fmt.Errorf("invalid project ID: %v", projects[i])
		}

		product.Projects = &model.Project{
			ID:   id,
			Name: projects[i+1],
		}
	}

	return &product, nil
}

// deleteProductByIDFromDB is a helper function to delete a product from the database
// It takes a testing.T object and the product's ID
// It returns an error if any
func deleteProductByIDFromDB(t *testing.T, id uuid.UUID) {
	t.Helper()

	if id == uuid.Nil {
		t.Log("Skipping deletion of product with nil ID")
		return
	}

	_, err := testDBPool.Exec(context.Background(), "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		// It's a test helper, so we can be a bit more aggressive with logging
		t.Logf("Failed to delete product with ID %s: %v", id, err)
	}
}

// deletePaymentProcessorByIDFromDB is a helper function to delete a payment processor from the database
// It takes a testing.T object and the payment processor's ID
// It returns an error if any
func deletePaymentProcessorByIDFromDB(t *testing.T, id uuid.UUID) {
	t.Helper()

	if id == uuid.Nil {
		t.Log("Skipping deletion of payment processor with nil ID")
		return
	}

	_, err := testDBPool.Exec(context.Background(), "DELETE FROM payment_processors WHERE id = $1", id)
	if err != nil {
		// It's a test helper, so we can be a bit more aggressive with logging
		t.Logf("Failed to delete payment processor with ID %s: %v", id, err)
	}
}

// deletePaymentProcessorByIDAndProjectIDFromDB is a helper function to delete a payment processor by ID and project ID from the database
// It takes a testing.T object, the payment processor's ID, and the project ID
// It returns an error if any
func deletePaymentProcessorByIDAndProjectIDFromDB(t *testing.T, id, projectID uuid.UUID) {
	t.Helper()

	if id == uuid.Nil || projectID == uuid.Nil {
		t.Log("Skipping deletion of payment processor with nil ID or project ID")
		return
	}

	_, err := testDBPool.Exec(context.Background(), "DELETE FROM payment_processors WHERE id = $1 AND projects_id = $2", id, projectID)
	if err != nil {
		// It's a test helper, so we can be a bit more aggressive with logging
		t.Logf("Failed to delete payment processor with ID %s and project ID %s: %v", id, projectID, err)
	}
}
