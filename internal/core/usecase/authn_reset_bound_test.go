//go:build unit

package usecase

import (
	"errors"
	"slices"
	"testing"
	"time"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// A reset link is bound to the account and to the address it was mailed to,
// and it is held to the account as it is when the link is followed. A link
// mailed to an address the account no longer has set the password until it
// expired; so did one mailed before an administrator disabled the account.
// Every refusal is the answer of a link that is not valid, in one sentence:
// which of them it was is not the holder's to learn.
func TestAResetLinkIsBoundToTheAccountAsItIsNow(t *testing.T) {
	t.Parallel()

	const mailedTo = "holder@example.com"

	for name, c := range map[string]struct {
		// stored is what the repository holds now; the cached copy is the
		// account as it was when the link was mailed.
		stored  func(u domain.User) *domain.User
		claim   string
		written bool
		missing bool // the claim set has no address
	}{
		"the_account_the_link_was_mailed_to": {stored: func(u domain.User) *domain.User { return &u }, claim: mailedTo, written: true},
		"its_address_changed_since": {stored: func(u domain.User) *domain.User {
			u.Email = "corrected@example.com"
			return &u
		}, claim: mailedTo},
		"an_administrator_disabled_it_since": {stored: func(u domain.User) *domain.User {
			u.Disabled = new(true)
			return &u
		}, claim: mailedTo},
		"it_is_no_longer_a_local_account": {stored: func(u domain.User) *domain.User {
			u.LocalAccount = new(false)
			return &u
		}, claim: mailedTo},
		"it_is_gone":             {stored: func(domain.User) *domain.User { return nil }, claim: mailedTo},
		"a_link_with_no_address": {stored: func(u domain.User) *domain.User { return &u }, missing: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newVerifiedFixture(t)
			revoked := &fakeRevoked{}
			f.svc.revokedTokens = revoked

			u := account(t, mailedTo, true, false)
			f.users.add(u)

			f.users.stored = map[uuid.UUID]*domain.User{}
			if stored := c.stored(*u); stored != nil {
				f.users.stored[u.ID] = stored
			} else {
				delete(f.users.byID, u.ID)
			}

			jti := uuid.NewV7()

			err := f.svc.ResetPassword(t.Context(), &domain.ResetPasswordInput{
				UserID: u.ID, Password: "Quartz9Falcon!", Email: c.claim,
				TokenID: jti, TokenExpiresAt: time.Now().Add(time.Hour),
			})

			switch {
			case c.written:
				if err != nil {
					t.Fatalf("refused: %v", err)
				}

				if want := []string{u.ID.String() + " " + mailedTo}; !slices.Equal(f.users.resets, want) {
					t.Errorf("written %v, want %v", f.users.resets, want)
				}
			case c.missing:
				invalid, ok := errors.AsType[*domain.InvalidJWTError](err)
				if !ok || invalid.Message != "email claim is missing" {
					t.Fatalf("got %T (%v), want the 401 of a link with no address", err, err)
				}

				// Refused before it is spent: the link was never valid.
				if revoked.spent[jti] {
					t.Error("a link with no address was spent")
				}
			default:
				invalid, ok := errors.AsType[*domain.InvalidJWTError](err)
				if !ok || invalid.Message != resetLinkNotForTheAccount {
					t.Fatalf("got %T (%v), want the one 401 of a link that is not the account's", err, err)
				}

				if len(f.users.resets) != 0 {
					t.Errorf("a password was written: %v", f.users.resets)
				}

				// Spent: each of these wants a new link.
				if !revoked.spent[jti] {
					t.Error("the refused link can be followed again")
				}
			}
		})
	}
}

// What an administrator asks for (POST /users/{user_id}/password/reset) is
// mailed to the address the row has, read from the repository. The route
// read the account through the cache and then the address through the cache
// again: after an address was corrected, a failed invalidation mailed the
// previous one until the entry expired.
func TestAnAdministratorsResetIsMailedToTheStoredAddress(t *testing.T) {
	t.Parallel()

	t.Run("the_cache_has_the_previous_address", func(t *testing.T) {
		t.Parallel()

		f := newVerifiedFixture(t)

		u := account(t, "previous@example.com", true, false)
		f.users.add(u)

		corrected := *u
		corrected.Email = "corrected@example.com"
		f.users.stored = map[uuid.UUID]*domain.User{u.ID: &corrected}

		if err := f.svc.RecoverPasswordOf(t.Context(), u.ID); err != nil {
			t.Fatalf("RecoverPasswordOf: %v", err)
		}

		if want := []string{"corrected@example.com"}; !slices.Equal(f.notifier.resets, want) {
			t.Errorf("mailed %v, want %v", f.notifier.resets, want)
		}
	})

	t.Run("an_account_that_is_not_there_is_the_404", func(t *testing.T) {
		t.Parallel()

		f := newVerifiedFixture(t)

		err := f.svc.RecoverPasswordOf(t.Context(), uuid.NewV7())
		if !isUserGone(err) {
			t.Fatalf("got %T (%v), want the account's NotFoundError", err, err)
		}

		if len(f.notifier.resets) != 0 {
			t.Errorf("mailed %v", f.notifier.resets)
		}
	})

	t.Run("the_stored_address_spends_the_recovery_budget", func(t *testing.T) {
		t.Parallel()

		f := newVerifiedFixture(t)
		spent := &refusingThrottle{retryAfter: time.Minute}
		f.svc.loginThrottle = spent

		u := account(t, "previous@example.com", true, false)
		f.users.add(u)

		corrected := *u
		corrected.Email = "corrected@example.com"
		f.users.stored = map[uuid.UUID]*domain.User{u.ID: &corrected}

		err := f.svc.RecoverPasswordOf(t.Context(), u.ID)
		if _, ok := errors.AsType[*domain.TooManyRecoveryRequestsError](err); !ok {
			t.Fatalf("got %T (%v), want TooManyRecoveryRequestsError", err, err)
		}

		if want := []string{throttleKey(throttlePurposeRecovery, "corrected@example.com")}; !slices.Equal(spent.keys, want) {
			t.Errorf("spent %v, want %v", spent.keys, want)
		}

		if len(f.notifier.resets) != 0 {
			t.Errorf("mailed %v", f.notifier.resets)
		}
	})

	t.Run("a_stored_address_the_rule_refuses_is_the_400", func(t *testing.T) {
		t.Parallel()

		f := newVerifiedFixture(t)

		u := account(t, "not-an-address", true, false)
		f.users.add(u)

		// ValidateEmail's ValidationError, singular: the type the route
		// answers 400 for.
		err := f.svc.RecoverPasswordOf(t.Context(), u.ID)
		if _, ok := errors.AsType[*domain.ValidationError](err); !ok {
			t.Fatalf("got %T (%v), want the address's ValidationError", err, err)
		}

		if len(f.notifier.resets) != 0 {
			t.Errorf("mailed %v", f.notifier.resets)
		}
	})

	t.Run("a_disabled_account_is_mailed_nothing_and_told_nothing", func(t *testing.T) {
		t.Parallel()

		f := newVerifiedFixture(t)

		u := account(t, "locked@example.com", true, true)
		f.users.add(u)

		if err := f.svc.RecoverPasswordOf(t.Context(), u.ID); err != nil {
			t.Fatalf("RecoverPasswordOf: %v", err)
		}

		if len(f.notifier.resets) != 0 {
			t.Errorf("mailed %v", f.notifier.resets)
		}
	})
}

// refusingThrottle refuses every attempt and records the key it was spent on.
type refusingThrottle struct {
	keys       []string
	retryAfter time.Duration
}

func (ref *refusingThrottle) Attempt(key string) (time.Duration, bool) {
	ref.keys = append(ref.keys, key)

	return ref.retryAfter, false
}

func (ref *refusingThrottle) Succeed(string) {}
