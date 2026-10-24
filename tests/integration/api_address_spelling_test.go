//go:build integration

package integration

import (
	"io"
	"net/http"
	"testing"

	"uuid"

	"github.com/stretchr/testify/require"
)

// An address is checked twice on its way in, by the handler and by the
// use-case on what the first check returned. "a b"@example.com was spelled
// a b@example.com, and "Ada <a@b.c>" a@b.c, both of which the second check
// refused: a 500 on each of these routes, an ERROR line for a caller's typo.
// Each is the first check's 400 now.
func TestAnAddressTheServiceCannotSpellIsA400(t *testing.T) {
	ctx := t.Context()
	admin := getAdminUserTokens(t)
	header := map[string]string{"Authorization": "Bearer " + admin.AccessToken}

	for _, a := range []struct{ label, address string }{
		{"a_quoted_local_part", `"a b"@example.com`},
		{"a_display_name_around_a_short_address", "Ada <a@b.c>"},
	} {
		address := a.address

		for _, tc := range []struct {
			name     string
			endpoint *apiEndpoint
			body     map[string]any
			headers  []map[string]string
		}{
			{name: "register", endpoint: newAPIEndpoint(http.MethodPost, "/auth/register"), body: map[string]any{
				"email": address, "first_name": "Ada", "last_name": "Lovelace", "password": generatePassword(t),
			}},
			{name: "recover", endpoint: newAPIEndpoint(http.MethodPost, "/auth/password/recover"), body: map[string]any{"email": address}},
			{name: "re_verify", endpoint: newAPIEndpoint(http.MethodPost, "/auth/verify"), body: map[string]any{"email": address}},
			{name: "create_user", endpoint: newAPIEndpoint(http.MethodPost, "/users"), body: map[string]any{
				"id": uuid.NewV7().String(), "email": address, "first_name": "Ada", "last_name": "Lovelace", "password": generatePassword(t),
			}, headers: []map[string]string{header}},
		} {
			t.Run(tc.name+"_"+a.label, func(t *testing.T) {
				resp, err := sendHTTPRequest(t, ctx, tc.endpoint, tc.body, tc.headers...)
				require.NoError(t, err)
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "body: %s", body)
			})
		}
	}
}
