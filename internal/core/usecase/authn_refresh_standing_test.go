//go:build unit

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// refreshStore is the revoked-tokens store of a refresh whose token was
// never spent: nothing is on the denylist, and a rotation is recorded. Any
// other method panics: the embedded port is nil.
type refreshStore struct {
	repository.RevokedTokens
	rotations int
}

func (s *refreshStore) Get(context.Context, uuid.UUID) (*domain.TokenRevocation, error) {
	return nil, nil
}

func (s *refreshStore) Rotate(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) error {
	s.rotations++

	return nil
}

// standingThatFails is the users port of a refresh whose question about the
// account cannot be answered.
type standingThatFails struct {
	*fakeUserStore
	err error
}

func (s standingThatFails) Standing(context.Context, uuid.UUID) (domain.AccountStanding, error) {
	return domain.AccountUnknown, s.err
}

// A refresh reads the account by its id, and that read is cached for twelve
// hours. Whether the account may renew its session is not decided from it:
// the cached copy of an account that has been disabled or deleted says
// "enabled" until its entry is dropped, and dropping it is best effort. The
// repository is asked, before anything is issued or rotated.
func TestARefreshAsksTheRepositoryWhetherTheAccountMayAct(t *testing.T) {
	t.Parallel()

	fault := errors.New("the database is away")

	for name, c := range map[string]struct {
		// cachedDisabled is what the cached copy says; standing what the
		// repository does.
		cachedDisabled bool
		standing       domain.AccountStanding
		fault          error
		renewed        bool
	}{
		"enabled_in_both":                       {standing: domain.AccountActive, renewed: true},
		"disabled_and_the_cache_does_not_know":  {standing: domain.AccountDisabled},
		"deleted_and_the_cache_does_not_know":   {standing: domain.AccountMissing},
		"a_standing_nobody_gave_is_not_a_yes":   {standing: domain.AccountUnknown},
		"the_cache_says_disabled_and_it_is_not": {cachedDisabled: true, standing: domain.AccountActive},
		"the_question_cannot_be_answered":       {fault: fault},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newVerifiedFixture(t)
			store := &refreshStore{}
			f.svc.revokedTokens = store

			u := account(t, name+"@example.com", true, c.cachedDisabled)
			f.users.add(u)
			f.users.standings = map[uuid.UUID]domain.AccountStanding{u.ID: c.standing}

			if c.fault != nil {
				f.svc.userService = standingThatFails{fakeUserStore: f.users, err: c.fault}
			}

			token, err := f.signer.Sign(t.Context(), domain.JWTClaims{
				Email: u.Email, Subject: u.ID.String(), Issuer: "https://api.example",
				TokenType: domain.TokenTypeRefresh, TokenDuration: time.Hour,
			})
			if err != nil {
				t.Fatalf("Sign: %v", err)
			}

			out, err := f.svc.RefreshAccessToken(t.Context(), &domain.RefreshAccessTokenInput{RefreshToken: token})

			if c.renewed {
				if err != nil || out == nil || out.AccessToken == "" {
					t.Fatalf("an account that may act did not renew its session: %+v, %v", out, err)
				}

				return
			}

			if out != nil {
				t.Errorf("a session was renewed: %+v", out)
			}

			if store.rotations != 0 {
				t.Errorf("the refresh token was rotated %d times for an account that may not act", store.rotations)
			}

			if c.fault != nil {
				if !errors.Is(err, c.fault) {
					t.Fatalf("got %v, want the fault: it is not a refusal, and not a yes", err)
				}

				return
			}

			if _, ok := errors.AsType[*domain.InvalidRefreshTokenError](err); !ok {
				t.Fatalf("got %T (%v), want the refusal a refresh token that no longer works gets", err, err)
			}
		})
	}
}
