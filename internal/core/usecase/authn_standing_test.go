//go:build unit

package usecase

import (
	"errors"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// A sign-in that is not by password (an identity provider's callback) reads
// the account by its id, and that read is cached for twelve hours. Whether
// the account may sign in is not decided from it: the cached copy of an
// account that has been disabled says "enabled" until its entry is dropped,
// and dropping it is best effort. The repository is asked.
func TestASignInByIDAsksTheRepositoryWhetherTheAccountMayAct(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		// cachedDisabled is what the cached copy says; standing what the
		// repository does.
		cachedDisabled bool
		standing       domain.AccountStanding
		refused        bool
	}{
		"enabled_in_both":                       {standing: domain.AccountActive},
		"disabled_and_the_cache_does_not_know":  {standing: domain.AccountDisabled, refused: true},
		"deleted_and_the_cache_does_not_know":   {standing: domain.AccountMissing, refused: true},
		"disabled_in_both":                      {cachedDisabled: true, standing: domain.AccountDisabled, refused: true},
		"a_standing_nobody_gave_is_not_a_yes":   {standing: domain.AccountUnknown, refused: true},
		"the_cache_says_disabled_and_it_is_not": {cachedDisabled: true, standing: domain.AccountActive, refused: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newVerifiedFixture(t)

			u := account(t, name+"@example.com", true, c.cachedDisabled)
			f.users.add(u)
			f.users.standings = map[uuid.UUID]domain.AccountStanding{u.ID: c.standing}

			out, err := f.svc.LoginUserByID(t.Context(), u.ID, domain.LoginMethodOAuth)

			if c.refused {
				if _, ok := errors.AsType[*domain.InvalidCredentialsError](err); !ok {
					t.Fatalf("got %T (%v) and %+v, want the refusal every failed sign-in gets", err, err, out)
				}

				return
			}

			if err != nil {
				t.Fatalf("refused: %v", err)
			}

			if out == nil || out.AccessToken == "" {
				t.Fatalf("no session was issued: %+v", out)
			}
		})
	}
}
