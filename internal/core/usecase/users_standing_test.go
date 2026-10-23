//go:build unit

package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/cache"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// standingUsers holds a few accounts, or fails its one read. Any other method
// panics: the embedded port is nil. The standing reads one column
// (SelectDisabled) and not the row with its password hash, so a SelectByID
// here would be the regression, and it panics.
type standingUsers struct {
	repository.Users
	accounts map[uuid.UUID]*domain.User
	err      error
}

func (s *standingUsers) SelectDisabled(_ context.Context, id uuid.UUID) (bool, error) {
	if s.err != nil {
		return false, s.err
	}

	if account, ok := s.accounts[id]; ok {
		return account.Disabled != nil && *account.Disabled, nil
	}

	return false, &domain.UserNotFoundError{ID: id}
}

// wrongCache answers what it was told, whatever the repository holds.
type wrongCache struct {
	entries map[string]any
}

func (c *wrongCache) Get(ctx context.Context, id cache.Identifier, dest any, fetcher cache.Fetcher) error {
	if v, ok := c.entries[id.String()]; ok {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(v))

		return nil
	}

	v, _, err := fetcher(ctx)
	if err != nil {
		return err
	}

	reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(v))

	return nil
}

func (c *wrongCache) Invalidate(context.Context, cache.Identifier) error { return nil }

// noLimits is the usage limits of a service whose read asks for none. Any
// method panics: the embedded port is nil.
type noLimits struct{ ResourcesLimitsServiceConsumer }

// An account's standing decides whether a token that verifies may act. It is
// read from the repository on every request and never from the cache: the
// cached account lives twelve hours, an invalidation is best effort, and a
// failed one would leave a disabled account working for half a day.
func TestAnAccountsStandingIsReadFromTheRepository(t *testing.T) {
	t.Parallel()

	enabled, disabled, neverSet, staleOff := uuid.NewV7(), uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

	accounts := func() map[uuid.UUID]*domain.User {
		return map[uuid.UUID]*domain.User{
			enabled:  {ID: enabled, Disabled: new(false)},
			disabled: {ID: disabled, Disabled: new(true)},
			// A row read without its `disabled` is not disabled.
			neverSet: {ID: neverSet},
			staleOff: {ID: staleOff, Disabled: new(true)},
		}
	}

	for name, c := range map[string]struct {
		id   uuid.UUID
		want domain.AccountStanding
	}{
		"an_enabled_account":                         {id: enabled, want: domain.AccountActive},
		"an_account_whose_disabled_was_never_set":    {id: neverSet, want: domain.AccountActive},
		"a_disabled_account":                         {id: disabled, want: domain.AccountDisabled},
		"an_account_that_is_not_there":               {id: uuid.NewV7(), want: domain.AccountMissing},
		"a_disabled_account_the_cache_calls_enabled": {id: staleOff, want: domain.AccountDisabled},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// What the cache holds must not be believed: here it says the
			// disabled account is enabled, as it would after an
			// invalidation that failed.
			stale := &wrongCache{entries: map[string]any{
				cache.Identifier{Type: "user", ID: staleOff.String()}.String(): &domain.User{ID: staleOff, Disabled: new(false)},
				cache.Identifier{Type: "user", ID: disabled.String()}.String(): &domain.User{ID: disabled, Disabled: new(false)},
			}}

			svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: &standingUsers{accounts: accounts()}, CacheService: stale, ResourcesLimits: noLimits{}, OT: testOT(t)})
			if err != nil {
				t.Fatalf("NewUsersService: %v", err)
			}

			// As in a request: what a layer records as a failure is held for
			// the access line.
			ctx := o11y.WithFailure(t.Context())

			got, err := svc.Standing(ctx, c.id)
			if err != nil {
				t.Fatalf("Standing: %v", err)
			}

			if got != c.want {
				t.Errorf("standing = %v, want %v", got, c.want)
			}

			// A disabled or missing account is an answer, not a failure of
			// the read: recorded as one, every request of such an account
			// would be an error of the users use-case.
			if held, ok := o11y.FailureFrom(ctx); ok {
				t.Errorf("an answer was recorded as a failure: %v", held.Err)
			}
		})
	}

	t.Run("a_fault_is_a_fault_not_a_standing", func(t *testing.T) {
		t.Parallel()

		fault := errors.New("the database is away")

		svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: &standingUsers{err: fault}, ResourcesLimits: noLimits{}, OT: testOT(t)})
		if err != nil {
			t.Fatalf("NewUsersService: %v", err)
		}

		standing, err := svc.Standing(t.Context(), uuid.NewV7())
		if !errors.Is(err, fault) {
			t.Fatalf("got %v, want the fault: a caller that is told \"active\" on a failed read lets the request through", err)
		}

		// And beside the error, nothing a caller that forgot to look at it
		// could take for a yes.
		if standing == domain.AccountActive {
			t.Error("a failed read answers \"may act\" beside its error")
		}
	})

	// The zero value is no answer.
	t.Run("the_zero_value_is_not_active", func(t *testing.T) {
		t.Parallel()

		var zero domain.AccountStanding

		if zero == domain.AccountActive || zero != domain.AccountUnknown {
			t.Errorf("the zero value is %v", zero)
		}
	})
}
