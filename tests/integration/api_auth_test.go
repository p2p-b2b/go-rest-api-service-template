//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	authLoginEndpoint    = newAPIEndpoint(http.MethodPost, "/auth/login")
	authLogoutEndpoint   = newAPIEndpoint(http.MethodDelete, "/auth/logout")
	authRefreshEndpoint  = newAPIEndpoint(http.MethodPost, "/auth/refresh")
	authRegisterEndpoint = newAPIEndpoint(http.MethodPost, "/auth/register")
	authReVerifyEndpoint = newAPIEndpoint(http.MethodPost, "/auth/verify")
	// authVerifyEndpoint    = newAPIEndpoint(http.MethodGet, "/auth/verify/{jwt}")
	// authUserAuthzEndpoint = newAPIEndpoint(http.MethodGet, "/users/{user_id}/authz")
)

func TestAuthRegisterUser(t *testing.T) {
	t.Run("test_register_single_user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")
		defer response.Body.Close()

		assert.Equal(t, http.StatusCreated, response.StatusCode, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})

		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, authRegisterEndpoint.Path(), apiResp.Path, "Expected path to be set")
	})

	t.Run("test_register_user_twice_get_409_error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		// 1. Register the user
		firstResponse, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")
		defer firstResponse.Body.Close()

		assert.Equal(t, firstResponse.StatusCode, http.StatusCreated, "Expected status code 201. Got %d. Message: %s", firstResponse.StatusCode, readResponseBody(t, firstResponse))

		apiFirstResp, err := parserResponseBody[model.HTTPMessage](t, firstResponse)
		if err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}

		assert.Equal(t, http.StatusCreated, firstResponse.StatusCode, "Expected status code 201. Got %d. Message: %s", firstResponse.StatusCode, readResponseBody(t, firstResponse))
		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiFirstResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiFirstResp.Method, "Expected method to be set")
		assert.Equal(t, authRegisterEndpoint.Path(), apiFirstResp.Path, "Expected path to be set")

		// 2. Try to register the same user again
		secondResponse, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")
		defer secondResponse.Body.Close()

		assert.Equal(t, http.StatusConflict, secondResponse.StatusCode, "Expected status code 409 for duplicate registration. Got %d. Message: %s", secondResponse.StatusCode, readResponseBody(t, secondResponse))

		apiSecondResp, err := parserResponseBody[model.HTTPMessage](t, secondResponse)
		assert.NoError(t, err, "Failed to parse response body")
		assert.Equal(t, http.StatusConflict, secondResponse.StatusCode, "Expected status code 409. Got %d. Message: %s", secondResponse.StatusCode, readResponseBody(t, secondResponse))

		emailAlreadyExistsError := &model.UserEmailAlreadyExistsError{Email: email}
		assert.Equal(t, emailAlreadyExistsError.Error(), apiSecondResp.Message, "Expected email already exists message")

		assert.Equal(t, authRegisterEndpoint.method, apiSecondResp.Method, "Expected method to be set")
		assert.Equal(t, authRegisterEndpoint.Path(), apiSecondResp.Path, "Expected path to be set")

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})
	})

	t.Run("test_create_and_verify_user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Register the user
		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		// 1. Register the user
		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")

		assert.Equal(t, response.StatusCode, http.StatusCreated, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))
		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, apiResp.Path, authRegisterEndpoint.Path(), "Expected path to be set")
		assert.Equal(t, http.StatusCreated, apiResp.StatusCode, "Expected status code 201. Got %d. Message: %s", apiResp.StatusCode, readResponseBody(t, response))

		// wait for the verification email to be sent
		time.Sleep(500 * time.Millisecond)

		// 2. Verify the user
		verifyLink := getVerifyLinkFromEmail(t, verifyEmailAddress, email)
		assert.NotEmpty(t, verifyLink, "Expected verify link to be generated")

		verificationRawResponse, err := http.Get(verifyLink)
		assert.NoError(t, err, "Failed to send request")

		assert.Equal(t, http.StatusOK, verificationRawResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", verificationRawResponse.StatusCode, readResponseBody(t, verificationRawResponse))

		verificationResponse, err := parserResponseBody[model.HTTPMessage](t, verificationRawResponse)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, verificationResponse.StatusCode, http.StatusOK, "Expected status code 200.")
		assert.Equal(t, model.AuthnUserVerifiedSuccessfully, verificationResponse.Message, "Expected verification success message")
		assert.Equal(t, http.MethodGet, verificationResponse.Method, "Expected method to be set")
		assert.Equal(t, removeAPIEndpointFromURL(verifyLink), verificationResponse.Path, "Expected path to be set")

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})
	})

	t.Run("test_create_and_verify_user_and_then_reverify", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Register the user
		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		// 1. Register the user
		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")

		assert.Equal(t, response.StatusCode, http.StatusCreated, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))
		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, apiResp.Path, authRegisterEndpoint.Path(), "Expected path to be set")
		assert.Equal(t, http.StatusCreated, apiResp.StatusCode, "Expected status code 201. Got %d. Message: %s", apiResp.StatusCode, readResponseBody(t, response))

		// wait for the verification email to be sent
		time.Sleep(500 * time.Millisecond)

		// 2. Verify the user
		verifyLink := getVerifyLinkFromEmail(t, verifyEmailAddress, email)
		assert.NotEmpty(t, verifyLink, "Expected verify link to be generated")

		verificationRawResponse, err := http.Get(verifyLink)
		assert.NoError(t, err, "Failed to send request")

		assert.Equal(t, http.StatusOK, verificationRawResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", verificationRawResponse.StatusCode, readResponseBody(t, verificationRawResponse))

		verificationResponse, err := parserResponseBody[model.HTTPMessage](t, verificationRawResponse)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, verificationResponse.StatusCode, http.StatusOK, "Expected status code 200.")
		assert.Equal(t, model.AuthnUserVerifiedSuccessfully, verificationResponse.Message, "Expected verification success message")
		assert.Equal(t, http.MethodGet, verificationResponse.Method, "Expected method to be set")
		assert.Equal(t, removeAPIEndpointFromURL(verifyLink), verificationResponse.Path, "Expected path to be set")

		// 3. Re-Verify the user
		reVerifyPayload := map[string]any{
			"email": email,
		}

		reVerifyResponse, err := sendHTTPRequest(t, ctx, authReVerifyEndpoint, reVerifyPayload)
		assert.NoError(t, err, "Failed to send request")
		assert.Equal(t, reVerifyResponse.StatusCode, http.StatusOK, "Expected status code 200. Got %d. Message: %s", reVerifyResponse.StatusCode, readResponseBody(t, reVerifyResponse))
		defer reVerifyResponse.Body.Close()

		reVerifyAPIResp, err := parserResponseBody[model.HTTPMessage](t, reVerifyResponse)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, reVerifyResponse.StatusCode, http.StatusOK, "Expected status code 200.")
		assert.Equal(t, model.AuthnUserVerificationEmailSent, reVerifyAPIResp.Message, "Expected verification email sent message")
		assert.Equal(t, authReVerifyEndpoint.method, reVerifyAPIResp.Method)
		assert.Equal(t, authReVerifyEndpoint.Path(), reVerifyAPIResp.Path)

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})
	})
}

func TestAuthLoginUser(t *testing.T) {
	t.Run("test_login_user_with_verification", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		userID, err := uuid.NewV7()
		assert.NoError(t, err, "Failed to generate user ID")
		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"id":         userID.String(),
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		// t.Logf("User Email: %s, Password: %s", user["email"], user["password"])

		// 1. Register the user
		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")
		defer response.Body.Close()

		assert.Equal(t, response.StatusCode, http.StatusCreated, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, authRegisterEndpoint.Path(), apiResp.Path, "Expected path to be set")

		// 2. Verify the user
		// wait for the verification email to be sent
		time.Sleep(500 * time.Millisecond)

		verifyLink := getVerifyLinkFromEmail(t, verifyEmailAddress, email)
		assert.NotEmpty(t, verifyLink, "Expected verify link to be generated")

		verificationRawResponse, err := http.Get(verifyLink)
		assert.NoError(t, err, "Failed to send request")
		assert.Equal(t, http.StatusOK, verificationRawResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", verificationRawResponse.StatusCode, readResponseBody(t, verificationRawResponse))

		verificationResponse, err := parserResponseBody[model.HTTPMessage](t, verificationRawResponse)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, http.StatusOK, verificationResponse.StatusCode, "Expected status code 200.")
		assert.Equal(t, model.AuthnUserVerifiedSuccessfully, verificationResponse.Message, "Expected verification success message")
		assert.Equal(t, http.MethodGet, verificationResponse.Method, "Expected method to be set")
		assert.Equal(t, removeAPIEndpointFromURL(verifyLink), verificationResponse.Path, "Expected path to be set")

		// 3. Login the user
		// wait for login verification in the database
		time.Sleep(500 * time.Millisecond)

		loginUser := map[string]any{
			"email":    user["email"],
			"password": user["password"],
		}

		loginResponse, err := sendHTTPRequest(t, ctx, authLoginEndpoint, loginUser)
		assert.NoError(t, err)

		defer loginResponse.Body.Close()

		assert.Equal(t, http.StatusOK, loginResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", loginResponse.StatusCode, readResponseBody(t, loginResponse))

		loginAPIResp, err := parserResponseBody[model.LoginUserResponse](t, loginResponse)
		assert.NoError(t, err)

		assert.Equal(t, userID, loginAPIResp.UserID, "Expected user ID to be set")
		assert.Equal(t, "Bearer", loginAPIResp.TokenType, "Expected token type to be Bearer")

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})
	})

	t.Run("test_login_user_without_verification", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		userID, err := uuid.NewV7()
		assert.NoError(t, err, "Failed to generate user ID")

		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"id":         userID.String(),
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		// t.Logf("User Email: %s, Password: %s", user["email"], user["password"])

		// 1. Register the user
		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")
		defer response.Body.Close()

		assert.Equal(t, response.StatusCode, http.StatusCreated, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, apiResp.Method, authRegisterEndpoint.method, "Expected method to be set")
		assert.Equal(t, apiResp.Path, authRegisterEndpoint.Path(), "Expected path to be set")

		// 2. Login the user
		// wait for login verification in the database
		time.Sleep(500 * time.Millisecond)

		loginUser := map[string]any{
			"email":    user["email"],
			"password": user["password"],
		}

		loginResponse, err := sendHTTPRequest(t, ctx, authLoginEndpoint, loginUser)
		assert.NoError(t, err)

		defer loginResponse.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, loginResponse.StatusCode, "Expected status code 401. Got %d. Message: %s", loginResponse.StatusCode, readResponseBody(t, loginResponse))

		loginAPIResp, err := parserResponseBody[model.HTTPMessage](t, loginResponse)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, loginAPIResp.StatusCode, "Expected status code 401. Got %d. Message: %s", loginAPIResp.StatusCode, readResponseBody(t, loginResponse))
		assert.Equal(t, loginAPIResp.Path, authLoginEndpoint.Path(), "Expected path to be set")
		assert.Equal(t, loginAPIResp.Method, authLoginEndpoint.method, "Expected method to be set")

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})
	})
}

func TestAuthReVerifyUser(t *testing.T) {
	t.Run("test_register_user_then_delete_it_and_reverify_user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Register the user
		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")
		defer response.Body.Close()

		assert.Equal(t, http.StatusCreated, response.StatusCode, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))

		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, authRegisterEndpoint.Path(), apiResp.Path, "Expected path to be set")

		// 2. Delete the user
		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})

		// 3. Re-verify the user
		reVerifyPayload := map[string]any{
			"email": user["email"],
		}

		reVerifyResponse, err := sendHTTPRequest(t, ctx, authReVerifyEndpoint, reVerifyPayload)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, reVerifyResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", reVerifyResponse.StatusCode, readResponseBody(t, reVerifyResponse))
		defer reVerifyResponse.Body.Close()

		reVerifyAPIResp, err := parserResponseBody[model.HTTPMessage](t, reVerifyResponse)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, reVerifyResponse.StatusCode, "Expected status code 200.")

		assert.Equal(t, model.AuthnUserVerificationEmailSent, reVerifyAPIResp.Message)
		assert.Equal(t, authReVerifyEndpoint.method, reVerifyAPIResp.Method)
		assert.Equal(t, authReVerifyEndpoint.Path(), reVerifyAPIResp.Path)
	})

	t.Run("test_verify_user_does_not_exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		reVerifyPayload := map[string]any{
			"email": "does.notexist@mail.com",
		}

		reVerifyResponse, err := sendHTTPRequest(t, ctx, authReVerifyEndpoint, reVerifyPayload)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, reVerifyResponse.StatusCode, "Expected status code 404. Got %d. Message: %s", reVerifyResponse.StatusCode, readResponseBody(t, reVerifyResponse))

		defer reVerifyResponse.Body.Close()

		reVerifyAPIResp, err := parserResponseBody[model.HTTPMessage](t, reVerifyResponse)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, reVerifyResponse.StatusCode, "Expected status code 200.")

		assert.Equal(t, model.AuthnUserVerificationEmailSent, reVerifyAPIResp.Message)
		assert.Equal(t, authReVerifyEndpoint.method, reVerifyAPIResp.Method)
		assert.Equal(t, authReVerifyEndpoint.Path(), reVerifyAPIResp.Path)
	})
}

func TestAuthRefreshTokens(t *testing.T) {
	t.Run("test_refresh_token", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Register the user
		userID, err := uuid.NewV7()
		assert.NoError(t, err, "Failed to generate user ID")

		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"id":         userID.String(),
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		// 1. Register the user
		response, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send request")

		assert.Equal(t, response.StatusCode, http.StatusCreated, "Expected status code 201. Got %d. Message: %s", response.StatusCode, readResponseBody(t, response))
		apiResp, err := parserResponseBody[model.HTTPMessage](t, response)
		assert.NoError(t, err, "Failed to parser response body")

		assert.Equal(t, model.AuthnUserRegisteredSuccessfully, apiResp.Message, "Expected success message")
		assert.Equal(t, authRegisterEndpoint.method, apiResp.Method, "Expected method to be set")
		assert.Equal(t, apiResp.Path, authRegisterEndpoint.Path(), "Expected path to be set")
		assert.Equal(t, http.StatusCreated, apiResp.StatusCode, "Expected status code 201. Got %d. Message: %s", apiResp.StatusCode, readResponseBody(t, response))

		// wait for the verification email to be sent
		time.Sleep(500 * time.Millisecond)

		// 2. Verify the user
		verifyLink := getVerifyLinkFromEmail(t, verifyEmailAddress, email)
		assert.NotEmpty(t, verifyLink, "Expected verify link to be generated")

		verificationRawResponse, err := http.Get(verifyLink)
		assert.NoError(t, err, "Failed to send request")

		assert.Equal(t, http.StatusOK, verificationRawResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", verificationRawResponse.StatusCode, readResponseBody(t, verificationRawResponse))

		verificationResponse, err := parserResponseBody[model.HTTPMessage](t, verificationRawResponse)
		assert.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, verificationResponse.StatusCode, http.StatusOK, "Expected status code 200.")
		assert.Equal(t, model.AuthnUserVerifiedSuccessfully, verificationResponse.Message, "Expected verification success message")
		assert.Equal(t, http.MethodGet, verificationResponse.Method, "Expected method to be set")
		assert.Equal(t, removeAPIEndpointFromURL(verifyLink), verificationResponse.Path, "Expected path to be set")

		// 3. Login the user
		// wait for login verification in the database
		time.Sleep(500 * time.Millisecond)

		loginUser := map[string]any{
			"email":    user["email"],
			"password": user["password"],
		}

		loginResponse, err := sendHTTPRequest(t, ctx, authLoginEndpoint, loginUser)
		assert.NoError(t, err)
		defer loginResponse.Body.Close()

		assert.Equal(t, http.StatusOK, loginResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", loginResponse.StatusCode, readResponseBody(t, loginResponse))
		loginAPIResp, err := parserResponseBody[model.LoginUserResponse](t, loginResponse)
		assert.NoError(t, err)

		assert.Equal(t, user["id"], loginAPIResp.UserID.String(), "Expected user ID to be set")
		assert.Equal(t, "Bearer", loginAPIResp.TokenType, "Expected token type to be Bearer")
		assert.NotEmpty(t, loginAPIResp.AccessToken, "Expected access token to be set")
		assert.NotEmpty(t, loginAPIResp.RefreshToken, "Expected refresh token to be set")

		// 4. Assign permissions to the user for POST /auth/refresh

		// 5. Refresh the token
		refreshTokenPayload := map[string]any{
			"refresh_token": loginAPIResp.RefreshToken,
		}

		// user the refresh token to get a new access token
		refreshTokenHeader := map[string]string{
			"Authorization": "Bearer " + loginAPIResp.RefreshToken,
		}

		refreshResponse, err := sendHTTPRequest(t, ctx, authRefreshEndpoint, refreshTokenPayload, refreshTokenHeader)
		assert.NoError(t, err)
		defer refreshResponse.Body.Close()

		assert.Equal(t, http.StatusOK, refreshResponse.StatusCode, "Expected status code 200. Got %d. Message: %s", refreshResponse.StatusCode, readResponseBody(t, refreshResponse))

		refreshAPIResp, err := parserResponseBody[model.RefreshTokenResponse](t, refreshResponse)
		assert.NoError(t, err)

		assert.NotEmpty(t, refreshAPIResp.AccessToken, "Expected access token to be set")
		assert.NotEmpty(t, refreshAPIResp.TokenType, "Expected token type to be set")
	})
}

func TestAuthLogoutUser(t *testing.T) {
	t.Run("test_logout_user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// 1. Register the user
		userID, err := uuid.NewV7()
		assert.NoError(t, err, "Failed to generate user ID")

		firstName, lastName, email := generateUserData(t)
		user := map[string]any{
			"id":         userID.String(),
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"password":   generatePassword(t),
		}

		registerResponse, err := sendHTTPRequest(t, ctx, authRegisterEndpoint, user)
		assert.NoError(t, err, "Failed to send register request")
		defer registerResponse.Body.Close()
		assert.Equal(t, http.StatusCreated, registerResponse.StatusCode, "Expected status code 201 for registration. Got %d. Message: %s", registerResponse.StatusCode, readResponseBody(t, registerResponse))

		// 2. Verify the user
		time.Sleep(1 * time.Second) // Allow time for email processing
		verifyLink := getVerifyLinkFromEmail(t, verifyEmailAddress, email)
		assert.NotEmpty(t, verifyLink, "Expected verify link to be generated")
		verificationRawResponse, err := http.Get(verifyLink)
		assert.NoError(t, err, "Failed to send verification request")
		defer verificationRawResponse.Body.Close()
		assert.Equal(t, http.StatusOK, verificationRawResponse.StatusCode, "Expected status code 200 for verification. Got %d. Message: %s", verificationRawResponse.StatusCode, readResponseBody(t, verificationRawResponse))

		// 3. Login the user
		time.Sleep(1 * time.Second) // Allow time for verification update
		loginUser := map[string]any{
			"email":    user["email"],
			"password": user["password"],
		}
		loginResponse, err := sendHTTPRequest(t, ctx, authLoginEndpoint, loginUser)
		assert.NoError(t, err, "Failed to send login request")
		defer loginResponse.Body.Close()
		assert.Equal(t, http.StatusOK, loginResponse.StatusCode, "Expected status code 200 for login. Got %d. Message: %s", loginResponse.StatusCode, readResponseBody(t, loginResponse))

		loginAPIResp, err := parserResponseBody[model.LoginUserResponse](t, loginResponse)
		assert.NoError(t, err, "Failed to parse login response body")
		assert.NotEmpty(t, loginAPIResp.AccessToken, "Expected access token to be set")
		assert.NotEmpty(t, loginAPIResp.RefreshToken, "Expected refresh token to be set")

		// 4. Logout the user
		logoutHeader := map[string]string{
			"Authorization": "Bearer " + loginAPIResp.AccessToken,
		}

		logoutResponse, err := sendHTTPRequest(t, ctx, authLogoutEndpoint, nil, logoutHeader) // Logout usually doesn't need a body
		assert.NoError(t, err, "Failed to send logout request")
		defer logoutResponse.Body.Close()

		assert.Equal(t, http.StatusOK, logoutResponse.StatusCode, "Expected status code 200 for logout. Got %d. Message: %s", logoutResponse.StatusCode, readResponseBody(t, logoutResponse))
		logoutAPIResp, err := parserResponseBody[model.HTTPMessage](t, logoutResponse)
		assert.NoError(t, err, "Failed to parse logout response body")

		assert.Equal(t, model.AuthnUserLoggedOutSuccessfully, logoutAPIResp.Message, "Expected logout success message")
		assert.Equal(t, authLogoutEndpoint.method, logoutAPIResp.Method, "Expected method to be DELETE")
		assert.Equal(t, authLogoutEndpoint.Path(), logoutAPIResp.Path, "Expected path to be /auth/logout")

		// 5. (Optional but recommended) Verify tokens are invalidated
		// Try using the old access token - this should fail
		// Example: Try to access a protected endpoint like getting user details
		// getUserEndpoint := newAPIEndpoint(http.MethodGet, "/users/"+userID.String()) // Assuming such endpoint exists
		// protectedResponse, err := sendHTTPRequest(t, ctx, getUserEndpoint, nil, logoutHeader)
		// assert.NoError(t, err)
		// defer protectedResponse.Body.Close()
		// assert.Equal(t, http.StatusUnauthorized, protectedResponse.StatusCode, "Access token should be invalid after logout")

		// Try using the old refresh token - this should fail
		refreshTokenPayload := map[string]any{
			"refresh_token": loginAPIResp.RefreshToken,
		}
		refreshTokenHeader := map[string]string{
			"Authorization": "Bearer " + loginAPIResp.RefreshToken,
		}
		refreshResponse, err := sendHTTPRequest(t, ctx, authRefreshEndpoint, refreshTokenPayload, refreshTokenHeader)
		assert.NoError(t, err)
		defer refreshResponse.Body.Close()
		assert.Equal(t, http.StatusOK, refreshResponse.StatusCode, "Refresh token should be invalid after logout. Got %d. Message: %s", refreshResponse.StatusCode, readResponseBody(t, refreshResponse))

		t.Cleanup(func() {
			deleteUserByEmailFromDB(t, email)
		})
	})
}
