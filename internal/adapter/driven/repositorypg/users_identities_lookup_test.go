package repositorypg

import (
	"strings"
	"testing"
)

// A sign-in finds an existing identity only under the provider as it used
// it, as an identity is stored only under it (queryLinkIdentity): read by
// the provider and the subject alone, a sign-in that read the provider
// before it was pointed at another issuer signed in as the account linked
// under the new one. The pull request gate has no database: the statement's
// condition is held whole, and its arguments in its order.
func TestAnIdentityIsFoundOnlyUnderTheProviderItSignedInWith(t *testing.T) {
	t.Parallel()

	source := strings.Join(strings.Fields(functionSource(t, "users_identities.go", "SelectBySubject")), " ")

	for _, want := range []string{
		"WHERE ui.idps_id = $1 AND ui.subject = $2 AND idp.idp_types = $3 AND COALESCE(idp.issuer_url, '') = $4;",
		"ref.db.QueryRow(ctx, query, idpID, subject, idpTypeID, issuerURL)",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("SelectBySubject has no %q", want)
		}
	}
}
