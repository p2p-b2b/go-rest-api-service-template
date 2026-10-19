package repositorypg

import (
	"errors"
	"strings"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// A provider's update decides, in its own statement, whether it points the
// provider at another issuer or type under its identities: an identity is
// (provider, subject), and whoever pointed a provider at an issuer of their
// own asserted an existing subject and was signed in as that account. The
// pull request gate has no database, so the condition is held whole, spaces
// apart: a term dropped from it still compiles and writes the repoint.
func TestAProvidersUpdateKeepsItForItsIdentities(t *testing.T) {
	t.Parallel()

	const condition = "WHERE id = $1 AND ( ( ($2::uuid IS NULL OR $2::uuid = idp_types) " +
		"AND ($6::varchar IS NULL OR NULLIF($6::varchar, '') IS NOT DISTINCT FROM issuer_url) ) " +
		"OR NOT EXISTS (SELECT 1 FROM users_identities WHERE idps_id = $1) );"

	if got := strings.Join(strings.Fields(queryUpdateIDP), " "); !strings.HasSuffix(got, condition) {
		t.Errorf("the update's condition changed:\n got: %s\nwant it to end: %s", got, condition)
	}

	// The two it compares are the two it writes.
	for _, write := range []string{
		"idp_types = COALESCE(NULLIF($2::uuid, idp_types), idp_types),",
		"issuer_url = CASE WHEN $6::varchar IS NULL THEN issuer_url ELSE NULLIF($6::varchar, '') END,",
	} {
		if !strings.Contains(strings.Join(strings.Fields(queryUpdateIDP), " "), write) {
			t.Errorf("the update no longer writes %q as its condition reads it", write)
		}
	}
}

// An update that wrote nothing is a 404 for a provider that is not there
// and a 409 for one the statement kept, and the 409 says why.
func TestAnUpdateThatWroteNothingSaysWhy(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()

	kept, ok := errors.AsType[*domain.IDPInUseError](idpNotRepointed(true, id))
	if !ok {
		t.Fatalf("a kept provider is answered %T, want the 409", idpNotRepointed(true, id))
	}

	if want := "identity provider cannot be pointed at another issuer or type: still used by at least one linked account"; kept.Error() != want {
		t.Errorf("the 409 says %q, want %q", kept.Error(), want)
	}

	if _, missing := errors.AsType[*domain.IDPNotFoundError](idpNotRepointed(false, id)); !missing {
		t.Fatalf("a provider that is not there is answered %T, want its 404", idpNotRepointed(false, id))
	}
}
