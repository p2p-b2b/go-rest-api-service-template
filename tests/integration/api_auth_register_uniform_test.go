//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/payload"
)

// register submits a registration and returns what the caller sees.
func register(t *testing.T, email string) (int, string) {
	t.Helper()

	firstName, lastName, _ := generateUserData(t)

	response, err := sendHTTPRequest(t, t.Context(), authRegisterEndpoint, map[string]any{
		"id":         mustUUIDString(t),
		"first_name": firstName,
		"last_name":  lastName,
		"email":      email,
		"password":   generatePassword(t),
	})
	require.NoError(t, err, "Error registering")

	body, err := parserResponseBody[payload.HTTPMessage](t, response)
	require.NoError(t, err, "Error parsing the response")
	require.NoError(t, response.Body.Close())

	return response.StatusCode, body.Message
}

// TestRegistrationDoesNotConfirmTheAddress covers the last of the three account
// oracles.
//
// Registration answered 409 "user: already exists: email=<address>" for a taken
// address and 201 for a free one, so one unauthenticated request said whether an
// address was registered. Login and password recovery were both closed for
// exactly this; registration was left because a person genuinely has to learn
// their address is taken.
//
// It is closed by telling them somewhere the prober cannot see: the response is
// the same either way, and the owner is told by email.
func TestRegistrationDoesNotConfirmTheAddress(t *testing.T) {
	t.Parallel()

	firstName, lastName, taken := generateUserData(t)
	createUserInDB(t, firstName, lastName, taken, generatePassword(t))
	t.Cleanup(func() { deleteUserByEmailFromDB(t, taken) })

	_, _, free := generateUserData(t)
	t.Cleanup(func() { deleteUserByEmailFromDB(t, free) })

	freeStatus, freeMessage := register(t, free)
	require.Equal(t, http.StatusCreated, freeStatus,
		"registering a free address must succeed, or this test proves nothing. Message: %s", freeMessage)

	takenStatus, takenMessage := register(t, taken)

	assert.Equal(t, freeStatus, takenStatus,
		"a taken address must answer with the same status as a free one")
	assert.Equal(t, freeMessage, takenMessage,
		"a taken address must answer with the same message as a free one")
	assert.NotContains(t, takenMessage, taken,
		"the response must not echo the address that was probed")

	// And it must not have quietly created a second account for that address.
	assert.Equal(t, 1, countUsersByEmailInDB(t, taken),
		"answering silently must not create a second account for an address that already has one")

	// The owner has to find out, or somebody who simply forgot they had an
	// account is told "registered" and then never hears anything.
	time.Sleep(700 * time.Millisecond)

	assert.True(t, emailWithSubjectExists(t, taken, "You already have an account"),
		"the address owner must be told that someone tried to register with it")
}

// TestRegistrationWithATakenIDIsNotATakenAddress: "registered, check your
// email" is the answer for an address that already has an account, and the
// owner of that address is the one who is mailed. The route also takes an
// optional id, and an id that is taken is refused by the same create. It was
// answered the same way: 201, no account, and "you already have an account"
// mailed to an address that had none. It is the conflict it is.
//
// With the address taken as well, the answer is the same 409: the id's
// conflict, whichever one the database reported, so that a taken id does
// not tell a taken address from a free one.
func TestRegistrationWithATakenIDIsNotATakenAddress(t *testing.T) {
	t.Parallel()

	existingEmail, existing := loginAs(t)
	_, _, freshEmail := generateUserData(t)

	for name, email := range map[string]string{"a_free_address": freshEmail, "a_taken_address": existingEmail} {
		t.Run(name, func(t *testing.T) {
			firstName, lastName, _ := generateUserData(t)

			response, err := sendHTTPRequest(t, t.Context(), authRegisterEndpoint, map[string]any{
				"id":         existing.UserID.String(),
				"first_name": firstName,
				"last_name":  lastName,
				"email":      email,
				"password":   generatePassword(t),
			})
			require.NoError(t, err)

			body := readResponseBody(t, response)
			require.NoError(t, response.Body.Close())

			assert.Equal(t, http.StatusConflict, response.StatusCode, "body: %s", body)
			assert.NotContains(t, body, email, "the refusal must not echo the address")
			assert.LessOrEqual(t, countUsersByEmailInDB(t, email), 1, "nothing may be created")

			time.Sleep(700 * time.Millisecond)
			assert.False(t, emailWithSubjectExists(t, email, "You already have an account"),
				"the address was told it has an account through a taken id")
		})
	}
}
