//go:build unit

package usecase

import (
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// An account with no password keeps a way in when one of its identities is
// unlinked only if it has another, at a provider that is enabled. An
// identity at a disabled provider is no way in: counted as one, it let an
// account unlink its only door, and when that account was the last
// administrator account that could sign in, nobody could any more.
func TestAnUnlinkKeepsAWayIn(t *testing.T) {
	t.Parallel()

	here, enabled, disabled := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

	at := func(idp uuid.UUID, on bool) domain.UserIdentity {
		return domain.UserIdentity{IDPID: idp, IDPEnabled: on}
	}

	for name, c := range map[string]struct {
		identities []domain.UserIdentity
		keeps      bool
	}{
		"the_only_identity":                     {identities: []domain.UserIdentity{at(here, true)}},
		"another_at_an_enabled_provider":        {identities: []domain.UserIdentity{at(here, true), at(enabled, true)}, keeps: true},
		"another_at_a_disabled_provider_only":   {identities: []domain.UserIdentity{at(here, true), at(disabled, false)}},
		"one_enabled_among_the_others":          {identities: []domain.UserIdentity{at(here, true), at(disabled, false), at(enabled, true)}, keeps: true},
		"the_unlinked_one_is_the_disabled_one":  {identities: []domain.UserIdentity{at(here, false), at(enabled, true)}, keeps: true},
		"no_identity_at_all":                    {},
		"an_identity_it_does_not_have_unlinked": {identities: []domain.UserIdentity{at(enabled, true)}, keeps: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := keepsAWayIn(c.identities, here); got != c.keeps {
				t.Errorf("keeps a way in = %v, want %v", got, c.keeps)
			}
		})
	}
}
