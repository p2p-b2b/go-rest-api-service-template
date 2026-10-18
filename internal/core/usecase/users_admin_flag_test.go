//go:build unit

package usecase

import (
	"errors"
	"strconv"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// The `admin` flag is the administrator account: every project, an account's
// address, another administrator account's sign-in. It is granted and taken
// back through PUT /users/{user_id}, by an account that has it and by nobody
// else: the route's grant is the seeded UserManager's too, and were the
// field open to the grant, that role would make itself the administrator
// account in one request.
//
// What is refused is a change. A form sends back what it read, so a value
// the account already has is nobody's to refuse, and it is taken out of the
// update: the statement is handed only what an administrator account
// changed. And the last account with the flag that can still sign in keeps
// it, whoever asks.
func TestTheAdminFlagIsGrantedAndTakenByAnAdministratorAccount(t *testing.T) {
	t.Parallel()

	const (
		onlyTheAdministrator = "only the administrator account"
		noCaller             = "no caller"
	)

	for name, c := range map[string]struct {
		target  uuid.UUID
		caller  uuid.UUID
		admin   bool
		enabled int
		// refused is how: one of the two constants above, or the words of
		// the guard's 403. Empty is a write, and handed is the flag the
		// statement was given ("-" for none).
		refused string
		handed  string
	}{
		// Granting.
		"an_administrator_grants_it": {target: ordinary, caller: adminA, admin: true, enabled: 2, handed: "true"},
		"an_administrator_grants_it_to_one_whose_flag_was_never_set": {
			target: unflaggedAcct, caller: adminA, admin: true, enabled: 2, handed: "true",
		},
		"an_administrator_grants_it_to_a_disabled_account": {target: ordinaryOff, caller: adminA, admin: true, enabled: 2, handed: "true"},
		"a_user_manager_grants_it":                         {target: ordinary, caller: managerAcct, admin: true, enabled: 2, refused: onlyTheAdministrator},
		"a_user_manager_grants_it_to_itself":               {target: managerAcct, caller: managerAcct, admin: true, enabled: 2, refused: onlyTheAdministrator},
		"a_disabled_administrator_grants_it":               {target: ordinary, caller: adminOff, admin: true, enabled: 2, refused: onlyTheAdministrator},
		"an_account_that_is_gone_grants_it":                {target: ordinary, caller: uuid.NewV7(), admin: true, enabled: 2, refused: onlyTheAdministrator},
		"an_account_whose_flag_was_never_set_grants_it":    {target: ordinary, caller: unflaggedAcct, admin: true, enabled: 2, refused: onlyTheAdministrator},
		"nobody_grants_it":                                 {target: ordinary, caller: uuid.Nil(), admin: true, enabled: 2, refused: noCaller},

		// Taking it back.
		"an_administrator_takes_it_from_another":         {target: adminB, caller: adminA, admin: false, enabled: 2, handed: "false"},
		"an_administrator_gives_its_own_up":              {target: adminA, caller: adminA, admin: false, enabled: 2, handed: "false"},
		"a_user_manager_takes_it":                        {target: adminA, caller: managerAcct, admin: false, enabled: 2, refused: onlyTheAdministrator},
		"a_disabled_administrator_takes_it":              {target: adminA, caller: adminOff, admin: false, enabled: 2, refused: onlyTheAdministrator},
		"the_last_enabled_one_keeps_it_from_itself":      {target: adminA, caller: adminA, admin: false, enabled: 1, refused: theLastOne},
		"the_last_enabled_one_keeps_it_from_another":     {target: adminA, caller: adminB, admin: false, enabled: 1, refused: theLastOne},
		"a_user_manager_is_not_told_it_was_the_last_one": {target: adminA, caller: managerAcct, admin: false, enabled: 1, refused: onlyTheAdministrator},
		"a_disabled_ones_flag_is_taken_when_one_is_left": {target: adminOff, caller: adminA, admin: false, enabled: 1, handed: "false"},

		// The value the account already has: no change, whoever sends it.
		"a_user_manager_sends_back_no_flag": {target: ordinary, caller: managerAcct, admin: false, enabled: 2, handed: "-"},
		"a_user_manager_sends_back_no_flag_for_one_never_set": {
			target: unflaggedAcct, caller: managerAcct, admin: false, enabled: 2, handed: "-",
		},
		"a_user_manager_sends_back_the_flag":     {target: adminA, caller: managerAcct, admin: true, enabled: 2, handed: "-"},
		"the_last_one_sends_back_its_flag":       {target: adminA, caller: adminA, admin: true, enabled: 1, handed: "-"},
		"nobody_sends_back_what_the_account_has": {target: ordinary, caller: uuid.Nil(), admin: false, enabled: 2, handed: "-"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(c.enabled)
			input := domain.UpdateUserInput{ID: c.target, CallerID: c.caller, FirstName: new("Jane"), Admin: new(c.admin)}

			err := usersServiceOver(t, repo).UpdateByID(t.Context(), &input)

			if c.refused != "" {
				switch c.refused {
				case onlyTheAdministrator:
					refusal, ok := errors.AsType[*domain.AdministratorOnlyError](err)
					if !ok || refusal.Entity != "user" || refusal.Field != domain.FieldAdmin {
						t.Fatalf("got %T (%v), want the 403 of a field that is the administrator account's", err, err)
					}
				case noCaller:
					// A route wired without its caller: no domain error, so
					// the fixed 500, never a grant.
					if !errors.Is(err, errAdministratorChangeWithoutCaller) {
						t.Fatalf("got %T (%v), want the fault of a change with no caller", err, err)
					}
				default:
					refusedAs(t, err, c.refused)

					// In the words of what was asked: the flag, not the
					// account's sign-in.
					if want := "user cannot be stripped of the admin flag: system/protected resource: " + c.refused; err.Error() != want {
						t.Errorf("refused as %q, want %q", err.Error(), want)
					}
				}

				if len(repo.writes) != 0 {
					t.Errorf("the account was written all the same: %v", repo.writes)
				}

				return
			}

			if err != nil {
				t.Fatalf("refused: %v", err)
			}

			// The rest of the body is written: the rename is the caller's.
			if repo.written == nil || repo.written.FirstName == nil {
				t.Fatalf("the update did not reach the statement: %v", repo.writes)
			}

			got := "-"
			if repo.written.Admin != nil {
				got = strconv.FormatBool(*repo.written.Admin)
			}

			if got != c.handed {
				t.Errorf("the statement is handed the flag %q, want %q", got, c.handed)
			}
		})
	}
}

// A body carries the flag beside the account's sign-in, and each is judged
// for itself against the row as it is stored: the flag a body takes away
// does not make the account an ordinary one for the rest of that body, and
// the flag a body grants does not make its other changes an administrator
// account's.
func TestTheAdminFlagBesideTheRestOfTheBody(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		input   domain.UpdateUserInput
		enabled int
		refused string
	}{
		// Taken and disabled in one body: two changes that each take the
		// account out of the ones that can sign in, both the guard's.
		"an_administrator_takes_it_and_disables_the_account": {
			input: domain.UpdateUserInput{ID: adminB, CallerID: adminA, Admin: new(false), Disabled: new(true)}, enabled: 2,
		},
		"the_last_one_is_not_demoted_and_disabled": {
			input: domain.UpdateUserInput{ID: adminA, CallerID: adminA, Admin: new(false), Disabled: new(true)}, enabled: 1, refused: theLastOne,
		},
		// The flag sent back as it is, with a change that is not the
		// caller's: the 403 of that change.
		"a_user_manager_sends_the_flag_back_and_disables_it": {
			input: domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, Admin: new(true), Disabled: new(true)}, enabled: 2, refused: theAdministratorAccount,
		},
		// Granted and disabled in one body, by an administrator account.
		"an_administrator_grants_it_to_an_account_it_disables": {
			input: domain.UpdateUserInput{ID: ordinary, CallerID: adminA, Admin: new(true), Disabled: new(true)}, enabled: 2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(c.enabled)

			err := usersServiceOver(t, repo).UpdateByID(t.Context(), &c.input)

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

			if repo.written == nil || repo.written.Admin == nil || repo.written.Disabled == nil {
				t.Errorf("the statement was not handed both changes: %+v", repo.written)
			}
		})
	}
}

// A grant from an account that may not make it is the 403 of the flag, and
// it comes before the 400 of a body that pulls the account in two
// directions: an account without the flag is not told, by a 400, what the
// row holds.
func TestAGrantThatIsNotTheCallersIsRefusedBeforeTheBodyIsJudged(t *testing.T) {
	t.Parallel()

	repo := administratorAccounts(2)
	repo.accounts[ordinaryLocked] = &domain.User{ID: ordinaryLocked, Email: "ol@example.com", Admin: new(false), Disabled: new(true), DisabledByAdmin: new(true), LocalAccount: new(true)}

	input := domain.UpdateUserInput{ID: ordinaryLocked, CallerID: managerAcct, Admin: new(true), Disabled: new(false), DisabledByAdmin: new(true)}

	err := usersServiceOver(t, repo).UpdateByID(t.Context(), &input)

	if _, ok := errors.AsType[*domain.AdministratorOnlyError](err); !ok {
		t.Fatalf("got %T (%v), want the 403 of the flag", err, err)
	}

	if len(repo.writes) != 0 {
		t.Errorf("the account was written all the same: %v", repo.writes)
	}
}
