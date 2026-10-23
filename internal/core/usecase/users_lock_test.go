//go:build unit

package usecase

import (
	"errors"
	"strconv"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// `disabled` alone cannot say "an administrator shut this account" for a
// registration whose link is still out: it is disabled already, the value is
// no change, and a form sends it back with every rename. So the
// administrator's lock is a field of its own, `disabled_by_admin`: true locks
// whatever `disabled` is, false unlocks.
//
// A client may send both fields, and the values it read are no change. What
// the statement is handed is what the client changed. And when the body
// carries the lock, a change of `disabled` beside it agrees with that lock,
// whether or not the lock is itself a change: a body that says "enable" and
// "locked" is refused whole. Not refused, with the lock dropped as no
// change, the `disabled` that was left enabled a locked account, and the
// same body sent again locked it.
func TestTheLockIsAFieldOfItsOwn(t *testing.T) {
	t.Parallel()

	row := func(disabled, locked bool) *domain.User {
		return &domain.User{ID: uuid.NewV7(), Disabled: new(disabled), DisabledByAdmin: new(locked)}
	}

	show := func(v *bool) string {
		if v == nil {
			return "-"
		}

		return strconv.FormatBool(*v)
	}

	for name, c := range map[string]struct {
		stored   *domain.User
		disabled *bool
		lock     *bool
		// handed is what is left for the statement, as "disabled=<v>
		// lock=<v>" with "-" for none; "refused" is the 400.
		handed string
	}{
		// The case the field exists for: a registration waiting for its
		// link is disabled and not locked.
		"a_registration_is_locked":                     {stored: row(true, false), lock: new(true), handed: "disabled=- lock=true"},
		"an_enabled_account_is_locked":                 {stored: row(false, false), lock: new(true), handed: "disabled=- lock=true"},
		"a_locked_account_is_unlocked":                 {stored: row(true, true), lock: new(false), handed: "disabled=- lock=false"},
		"the_lock_sent_back_as_read_is_no_change":      {stored: row(true, true), lock: new(true), handed: "disabled=- lock=-"},
		"no_lock_sent_back_as_read_is_no_change":       {stored: row(true, false), lock: new(false), handed: "disabled=- lock=-"},
		"no_lock_in_the_body_leaves_disabled_as_it_is": {stored: row(true, false), disabled: new(true), handed: "disabled=true lock=-"},
		// A row read without its `disabled` is not disabled.
		"a_row_whose_disabled_was_never_set_is_locked": {stored: &domain.User{ID: uuid.NewV7(), DisabledByAdmin: new(false)}, disabled: new(false), lock: new(true), handed: "disabled=- lock=true"},
		"and_enabled_with_the_lock_sent_back_as_read":  {stored: &domain.User{ID: uuid.NewV7(), DisabledByAdmin: new(false)}, disabled: new(false), lock: new(false), handed: "disabled=- lock=-"},
		// Both fields, as a form that knows both sends them.
		"locked_with_disabled_sent_back":          {stored: row(true, false), disabled: new(true), lock: new(true), handed: "disabled=- lock=true"},
		"unlocked_with_disabled_sent_back":        {stored: row(true, true), disabled: new(true), lock: new(false), handed: "disabled=- lock=false"},
		"locked_and_disabled_together":            {stored: row(false, false), disabled: new(true), lock: new(true), handed: "disabled=true lock=true"},
		"unlocked_and_enabled_together":           {stored: row(true, true), disabled: new(false), lock: new(false), handed: "disabled=false lock=false"},
		"the_whole_form_as_read_changes_nothing":  {stored: row(true, true), disabled: new(true), lock: new(true), handed: "disabled=- lock=-"},
		"enabled_with_the_lock_sent_back_as_read": {stored: row(true, false), disabled: new(false), lock: new(false), handed: "disabled=false lock=-"},
		// Two directions in one body: a change of `disabled` that disagrees
		// with the lock the body states.
		"locked_and_enabled_is_refused":                     {stored: row(true, false), disabled: new(false), lock: new(true), handed: "refused"},
		"enabled_under_a_lock_sent_back_as_read_is_refused": {stored: row(true, true), disabled: new(false), lock: new(true), handed: "refused"},
		"disabled_and_not_by_an_administrator_is_refused":   {stored: row(false, false), disabled: new(true), lock: new(false), handed: "refused"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			input := &domain.UpdateUserInput{ID: c.stored.ID, Disabled: c.disabled, DisabledByAdmin: c.lock}

			err := lockMayChange(input, c.stored)

			if (c.handed == "refused") != (err != nil) {
				t.Fatalf("refused = %v (%v), want %v", err != nil, err, c.handed == "refused")
			}

			if c.handed == "refused" {
				refused, ok := errors.AsType[*domain.InvalidInputError](err)
				if !ok || refused.Field != domain.FieldDisabledByAdmin {
					t.Fatalf("got %T (%v), want the body refused on the lock", err, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("refused: %v", err)
			}

			if got := "disabled=" + show(input.Disabled) + " lock=" + show(input.DisabledByAdmin); got != c.handed {
				t.Errorf("the statement is handed %q, want %q", got, c.handed)
			}
		})
	}
}

// The lock is the account's sign-in, as `disabled` is. On an account with the
// `admin` flag it is set, and taken off, only by an account that has the
// flag, and the last enabled one is locked by nobody: the rule `disabled`
// has, reached through the other field.
func TestTheLockOnAnAdministratorAccountIsAnAdministrators(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		target  uuid.UUID
		caller  uuid.UUID
		lock    bool
		enabled int
		refused string
		handed  string
	}{
		"a_user_manager_locks_it":            {target: adminA, caller: managerAcct, lock: true, enabled: 2, refused: theAdministratorAccount},
		"a_user_manager_unlocks_it":          {target: adminLocked, caller: managerAcct, lock: false, enabled: 2, refused: theAdministratorAccount},
		"an_administrator_locks_another":     {target: adminA, caller: adminB, lock: true, enabled: 2, handed: "true"},
		"an_administrator_unlocks_another":   {target: adminLocked, caller: adminB, lock: false, enabled: 2, handed: "false"},
		"the_last_enabled_one_is_not_locked": {target: adminA, caller: adminA, lock: true, enabled: 1, refused: theLastOne},
		"a_user_manager_locks_an_ordinary_one": {
			target: ordinary, caller: managerAcct, lock: true, enabled: 2, handed: "true",
		},
		// The value it already has is no change, and is nobody's to refuse.
		"a_user_manager_sends_the_lock_as_it_is": {target: adminLocked, caller: managerAcct, lock: true, enabled: 2, handed: "-"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(c.enabled)
			repo.accounts[adminLocked] = &domain.User{ID: adminLocked, Email: "locked@example.com", Admin: new(true), Disabled: new(true), DisabledByAdmin: new(true), LocalAccount: new(true)}

			input := domain.UpdateUserInput{ID: c.target, CallerID: c.caller, DisabledByAdmin: new(c.lock)}

			err := usersServiceOver(t, repo).UpdateByID(t.Context(), &input)

			if c.refused != "" {
				refusedAs(t, err, c.refused)

				if len(repo.writes) != 0 {
					t.Errorf("the account was written all the same: %v", repo.writes)
				}

				return
			}

			if err != nil {
				t.Fatalf("refused: %v", err)
			}

			got := "-"
			if repo.written != nil && repo.written.DisabledByAdmin != nil {
				got = strconv.FormatBool(*repo.written.DisabledByAdmin)
			}

			if got != c.handed {
				t.Errorf("the statement is handed the lock %q, want %q", got, c.handed)
			}
		})
	}
}

// adminLocked is an account with the flag that an administrator locked.
var adminLocked = uuid.NewV7()

// The refusal reaches the caller of the update, and nothing of the body is
// written. And an account that may not change the target at all hears that
// first: it is not told, by a 400, what the row holds.
func TestABodyInTwoDirectionsIsRefusedWhole(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		target  uuid.UUID
		caller  uuid.UUID
		input   domain.UpdateUserInput
		twoWays bool
		refused string
	}{
		// A locked ordinary account, told to be enabled and to stay locked.
		"an_ordinary_account": {
			target: ordinaryLocked, caller: managerAcct, twoWays: true,
			input: domain.UpdateUserInput{FirstName: new("Jane"), Disabled: new(false), DisabledByAdmin: new(true)},
		},
		"an_administrator_account_by_an_administrator": {
			target: adminLocked, caller: adminB, twoWays: true,
			input: domain.UpdateUserInput{Disabled: new(false), DisabledByAdmin: new(true)},
		},
		// The same body from an account that may not change it: the 403.
		"an_administrator_account_by_a_user_manager": {
			target: adminLocked, caller: managerAcct, refused: theAdministratorAccount,
			input: domain.UpdateUserInput{Disabled: new(false), DisabledByAdmin: new(true)},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(2)
			repo.accounts[adminLocked] = &domain.User{ID: adminLocked, Email: "locked@example.com", Admin: new(true), Disabled: new(true), DisabledByAdmin: new(true), LocalAccount: new(true)}
			repo.accounts[ordinaryLocked] = &domain.User{ID: ordinaryLocked, Email: "ol@example.com", Admin: new(false), Disabled: new(true), DisabledByAdmin: new(true), LocalAccount: new(true)}

			input := c.input
			input.ID, input.CallerID = c.target, c.caller

			err := usersServiceOver(t, repo).UpdateByID(t.Context(), &input)

			if len(repo.writes) != 0 {
				t.Errorf("the account was written all the same: %v", repo.writes)
			}

			if c.refused != "" {
				refusedAs(t, err, c.refused)

				return
			}

			refused, ok := errors.AsType[*domain.InvalidInputError](err)
			if !ok || refused.Field != domain.FieldDisabledByAdmin {
				t.Fatalf("got %T (%v), want the body refused on the lock", err, err)
			}
		})
	}
}

// ordinaryLocked is an account without the flag that an administrator locked.
var ordinaryLocked = uuid.NewV7()
