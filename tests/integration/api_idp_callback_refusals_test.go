//go:build integration

package integration

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callbackWith calls the callback with the query a test names.
func callbackWith(t *testing.T, idpID string, query url.Values) (int, string) {
	t.Helper()

	ep := idpCallbackEndpoint.Clone().RewriteSlugs(idpID)
	ep.requestURL.RawQuery = query.Encode()

	resp, err := sendHTTPRequest(t, t.Context(), ep, nil)
	require.NoError(t, err)

	body := readResponseBody(t, resp)
	resp.Body.Close()

	return resp.StatusCode, body
}

// TestIDPCallbackARefusedSignInIsNotAnOutage: a callback fails at the
// provider for one of two kinds of reason, and they are answered apart.
//
// The person's side is a 401: they cancelled, or the code is not one the
// provider accepts. The provider's side is a 503: it rejects this service's
// client, or it does not answer. They were one error answered 401, so a
// provider that was down told every visitor their sign-in was refused, and
// every cancelled sign-in was counted as a failure of the service.
//
// No answer carries the provider's own words or the client library's.
func TestIDPCallbackARefusedSignInIsNotAnOutage(t *testing.T) {
	admin := getAdminUserTokens(t)

	// One sentence for every 503 of the callback: which provider-side
	// failure it was is in the log and on the metric, not in a public body.
	const unavailable = "sign-in with this identity provider is not available right now"

	noForeignWords := func(t *testing.T, body string) {
		t.Helper()

		for _, leak := range []string{"oauth2:", "invalid_grant", "invalid_client", "access_denied", "the provider's own words"} {
			assert.NotContains(t, body, leak, "the provider's or the library's words reached the caller: %s", body)
		}
	}

	t.Run("a_code_the_provider_does_not_accept_is_a_401", func(t *testing.T) {
		f := newFakeProvider(t)
		idpID := createOIDCIDP(t, admin.AccessToken, f, true)

		state, _ := f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person("sub-x", "x@example.com", true))

		status, body := callbackWith(t, idpID, url.Values{"state": {state}, "code": {"a-code-the-provider-never-issued"}})
		assert.Equal(t, http.StatusUnauthorized, status, body)
		noForeignWords(t, body)

		f.mu.Lock()
		defer f.mu.Unlock()
		assert.Equal(t, 1, f.exchanges, "the code was presented more than once")
	})

	t.Run("a_provider_that_rejects_this_service_is_a_503", func(t *testing.T) {
		f := newFakeProvider(t)
		idpID := createOIDCIDP(t, admin.AccessToken, f, true)

		state, code := f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person("sub-y", "y@example.com", true))

		f.mu.Lock()
		f.tokenError = "invalid_client"
		f.mu.Unlock()

		status, body := callbackWith(t, idpID, url.Values{"state": {state}, "code": {code}})
		assert.Equal(t, http.StatusServiceUnavailable, status, body)
		assert.Contains(t, body, unavailable, body)
		assert.NotContains(t, body, "this service", "that the provider rejects this service is the operator's to read: %s", body)
		noForeignWords(t, body)

		f.mu.Lock()
		defer f.mu.Unlock()
		assert.Equal(t, 1, f.exchanges, "the code was presented more than once")
	})

	// The token endpoint fails behind its proxy. That is a 503, and the code
	// was presented once: it is good for one exchange, and the retries of
	// the service's shared client sent it again and again (with the OAuth
	// library presenting the exchange a second time on top), so a provider
	// that had spent it answered the next attempt `invalid_grant`.
	t.Run("a_provider_that_fails_is_a_503_and_is_asked_once", func(t *testing.T) {
		f := newFakeProvider(t)
		idpID := createOIDCIDP(t, admin.AccessToken, f, true)

		state, code := f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person("sub-w", "w@example.com", true))

		f.mu.Lock()
		f.tokenStatus = http.StatusBadGateway
		f.mu.Unlock()

		status, body := callbackWith(t, idpID, url.Values{"state": {state}, "code": {code}})
		assert.Equal(t, http.StatusServiceUnavailable, status, body)
		assert.Contains(t, body, unavailable, body)
		noForeignWords(t, body)

		f.mu.Lock()
		defer f.mu.Unlock()
		assert.Equal(t, 1, f.exchanges, "the code was presented more than once")
	})

	// The provider is gone by the time the code comes back. That is a 503,
	// and it is answered at once: the identity providers' client does not
	// retry. On the shared outbound client this one callback took about two
	// and a half minutes (the retries of a refused connection, and the OAuth
	// library presenting the exchange a second time), with the person
	// looking at a spinner and the request holding a connection.
	t.Run("a_provider_that_does_not_answer_is_a_503_at_once", func(t *testing.T) {
		f := newFakeProvider(t)
		idpID := createOIDCIDP(t, admin.AccessToken, f, true)

		// Discovery succeeds while the flow starts.
		state, code := f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person("sub-z", "z@example.com", true))
		f.srv.Close()

		started := time.Now()
		status, body := callbackWith(t, idpID, url.Values{"state": {state}, "code": {code}})
		waited := time.Since(started)

		assert.Equal(t, http.StatusServiceUnavailable, status, body)
		assert.Contains(t, body, unavailable, body)
		assert.NotContains(t, body, f.srv.URL, "the provider's address reached the caller: %s", body)
		noForeignWords(t, body)
		assert.Less(t, waited, 5*time.Second, "a provider that refuses the connection was waited for: something retries")
	})

	// The provider sends the person back with `error=` and no code. That is
	// their one callback: a 401, and the state is spent. It used to stay
	// good for a callback with a code for the rest of its fifteen minutes.
	t.Run("a_sign_in_cancelled_at_the_provider_is_a_401_and_spends_the_state", func(t *testing.T) {
		f := newFakeProvider(t)
		idpID := createOIDCIDP(t, admin.AccessToken, f, true)

		_, _, email := generateUserData(t)
		t.Cleanup(func() { deleteUserByEmailFromDB(t, email) })

		state, code := f.authorize(t, startFlow(t, idpLoginEndpoint, idpID, nil), person("sub-"+email, email, true))
		jti := jwtClaims(t, state)["jti"].(string)

		status, body := callbackWith(t, idpID, url.Values{"state": {state}, "error": {"access_denied"}, "error_description": {"the provider's own words"}})
		assert.Equal(t, http.StatusUnauthorized, status, body)
		noForeignWords(t, body)
		assert.True(t, isTokenRevokedInDB(t, jti), "the state of a cancelled sign-in is spent")

		// The code is one the provider would honour. The state is not.
		status, body = callbackWith(t, idpID, url.Values{"state": {state}, "code": {code}})
		assert.Equal(t, http.StatusBadRequest, status, body)
		assert.Contains(t, body, "state is not valid", body)
		assert.Zero(t, countUsersByEmailInDB(t, email), "the spent state provisioned an account")
	})
}
