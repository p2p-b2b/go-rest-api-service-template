//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAProviderThatCannotBeUsedIsTheServicesSide: a provider row the
// service cannot sign in with (an OpenID Connect provider with no issuer,
// written by hand: the API refuses it) answered the public route that
// starts a sign-in with a 400, "idp <its name> has no issuer URL". The
// caller sent nothing wrong, and the route is anonymous. It is the
// service's side now: a 503 in the service's words, naming nothing.
func TestAProviderThatCannotBeUsedIsTheServicesSide(t *testing.T) {
	t.Parallel()

	admin := getAdminUserTokens(t)
	t.Cleanup(func() { deleteUserByIDFromDB(t, admin.UserID) })

	provider := newFakeProvider(t)
	idp := createOIDCIDP(t, admin.AccessToken, provider, false)

	var name string
	require.NoError(t, testDBPool.QueryRow(context.Background(),
		`UPDATE idps SET issuer_url = NULL WHERE id = $1 RETURNING name`, idp).Scan(&name))

	response, err := sendHTTPRequest(t, context.WithoutCancel(t.Context()), idpLoginEndpoint.Clone().RewriteSlugs(idp), nil, nil)
	require.NoError(t, err)

	body := readResponseBody(t, response)
	response.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, response.StatusCode, "a provider the service cannot use is not the caller's fault: %s", body)
	assert.NotContains(t, body, name, "the public answer names the provider")
	assert.Contains(t, body, "sign-in with this identity provider is not available right now", "not the callback's sentence")
}
