//go:build unit

package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// administratorRepository holds a few accounts, says how many accounts with
// the flag can still sign in and whether one of them can, and records every
// write it is asked for. Any other method panics: the embedded port is nil.
type administratorRepository struct {
	repository.Users
	accounts map[uuid.UUID]*domain.User
	// identities is the accounts with an identity at an enabled provider.
	identities map[uuid.UUID]bool
	// unreadable is an account whose read fails.
	unreadable uuid.UUID
	enabled    int
	writes     []string
	// written is the last update the statement was handed.
	written *domain.UpdateUserInput
	// holders are the accounts that hold each role; holdersErr is what
	// reading them answers instead, when set.
	holders    map[uuid.UUID][]uuid.UUID
	holdersErr error
}

func (r *administratorRepository) SelectByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if id == r.unreadable {
		return nil, errTheDatabaseIsDown
	}

	if account, ok := r.accounts[id]; ok {
		return account, nil
	}

	return nil, &domain.UserNotFoundError{ID: id}
}

func (r *administratorRepository) CountEnabledAdministrators(context.Context) (int, error) {
	return r.enabled, nil
}

// IsEnabledAdministrator is the repository's predicate: the flag, enabled,
// and a password login or an identity at an enabled provider.
func (r *administratorRepository) IsEnabledAdministrator(_ context.Context, id uuid.UUID) (bool, error) {
	account, ok := r.accounts[id]
	if !ok {
		return false, nil
	}

	local := account.LocalAccount != nil && *account.LocalAccount

	return hasAdminFlag(account) && account.Disabled != nil && !*account.Disabled && (local || r.identities[id]), nil
}

// errTheDatabaseIsDown is the failure of an account's read.
var errTheDatabaseIsDown = errors.New("the database is down")

func (r *administratorRepository) SelectAdministratorsHolding(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	if r.holdersErr != nil {
		return nil, r.holdersErr
	}

	return r.SelectAdministratorsAmong(ctx, r.holders[roleID])
}

func (r *administratorRepository) SelectAdministratorsAmong(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	var found []uuid.UUID

	for _, id := range ids {
		if hasAdminFlag(r.accounts[id]) {
			found = append(found, id)
		}
	}

	return found, nil
}

func (r *administratorRepository) UpdateByID(_ context.Context, in *domain.UpdateUserInput) error {
	r.writes = append(r.writes, "UpdateByID")
	r.written = in

	return nil
}

func (r *administratorRepository) DeleteByID(_ context.Context, in *domain.DeleteUserInput) error {
	if _, ok := r.accounts[in.ID]; !ok {
		return &domain.UserNotFoundError{ID: in.ID}
	}

	r.writes = append(r.writes, "DeleteByID")

	return nil
}

func (r *administratorRepository) UnlinkRoles(context.Context, *domain.UnlinkRolesFromUsersInput) error {
	r.writes = append(r.writes, "UnlinkRoles")

	return nil
}

// The accounts of these tests: six with the `admin` flag (two enabled, one
// disabled, one federated with no identity, one federated with one, one
// read without its `local_account`); a user manager; an ordinary account;
// one read without its flag (a nil Admin). The columns are NOT NULL; a nil
// field is a row read without it.
var (
	adminA, adminB, adminOff = uuid.NewV7(), uuid.NewV7(), uuid.NewV7()
	adminFederated           = uuid.NewV7()
	adminSSO, adminNeverSet  = uuid.NewV7(), uuid.NewV7()
	unreadableAcct           = uuid.NewV7()
	managerAcct, ordinary    = uuid.NewV7(), uuid.NewV7()
	ordinaryOff              = uuid.NewV7()
	unflaggedAcct            = uuid.NewV7()
)

func administratorAccounts(enabled int) *administratorRepository {
	return &administratorRepository{
		enabled:    enabled,
		identities: map[uuid.UUID]bool{adminSSO: true},
		unreadable: unreadableAcct,
		accounts: map[uuid.UUID]*domain.User{
			adminA:         {ID: adminA, Email: "a@example.com", Admin: new(true), Disabled: new(false), LocalAccount: new(true)},
			adminB:         {ID: adminB, Email: "b@example.com", Admin: new(true), Disabled: new(false), LocalAccount: new(true)},
			adminOff:       {ID: adminOff, Email: "off@example.com", Admin: new(true), Disabled: new(true), LocalAccount: new(true)},
			adminFederated: {ID: adminFederated, Email: "fed@example.com", Admin: new(true), Disabled: new(false), LocalAccount: new(false)},
			adminSSO:       {ID: adminSSO, Email: "sso@example.com", Admin: new(true), Disabled: new(false), LocalAccount: new(false)},
			adminNeverSet:  {ID: adminNeverSet, Email: "ns@example.com", Admin: new(true), Disabled: new(false)},
			managerAcct:    {ID: managerAcct, Email: "m@example.com", Admin: new(false), Disabled: new(false), LocalAccount: new(true)},
			ordinary:       {ID: ordinary, Email: "o@example.com", Admin: new(false), Disabled: new(false), LocalAccount: new(true)},
			ordinaryOff:    {ID: ordinaryOff, Email: "oo@example.com", Admin: new(false), Disabled: new(true), LocalAccount: new(false)},
			unflaggedAcct:  {ID: unflaggedAcct, Email: "u@example.com", Disabled: new(false), LocalAccount: new(true)},
		},
	}
}

// slotsGivenBack is the usage limits a delete gives its slot back to.
type slotsGivenBack struct{ ResourcesLimitsServiceConsumer }

func (slotsGivenBack) DecrementUsage(context.Context, domain.ResourcesLimitsScope, domain.ResourcesLimitsResourceType) error {
	return nil
}

func usersServiceOver(t *testing.T, repo *administratorRepository) *UsersService {
	t.Helper()

	svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: repo, ResourcesLimits: slotsGivenBack{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewUsersService: %v", err)
	}

	return svc
}

// refusedAs reports whether err is the 403 a protected administrator account
// answers, and with what words.
func refusedAs(t *testing.T, err error, wantMessage string) {
	t.Helper()

	refusal, ok := errors.AsType[*domain.AdministratorProtectedError](err)
	if !ok {
		t.Fatalf("got %T (%v), want the refusal the handlers answer 403", err, err)
	}

	if refusal.Message != wantMessage {
		t.Errorf("refused as %q, want a user refused as %q", refusal.Error(), wantMessage)
	}
}

const (
	theLastOne = theLastAdministratorAccount
)

// An account with the `admin` flag is what a deployment is operated from, and
// the seeded UserManager role manages accounts. With only that role an
// account disabled the seeded administrator.
//
// An account with the flag is disabled, made federated, stripped of a role
// or deleted only by an account with the flag, and the last that can still
// sign in by nobody. Every other account stays the
// role's to manage.
func TestAnAdministratorAccountIsLockedOnlyByOne(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		target  uuid.UUID
		caller  uuid.UUID
		locks   bool
		enabled int
		// refused is the words of the refusal, "" for none; fault is a
		// route wired without its caller.
		refused string
		fault   bool
	}{
		"an_ordinary_account_by_a_user_manager":           {target: ordinary, caller: managerAcct, locks: true, enabled: 2},
		"an_ordinary_account_by_nobody_in_particular":     {target: ordinary, locks: true, enabled: 2},
		"an_account_whose_flag_was_never_set":             {target: unflaggedAcct, caller: managerAcct, locks: true, enabled: 2},
		"an_administrator_account_by_a_user_manager":      {target: adminA, caller: managerAcct, locks: true, enabled: 2, refused: theAdministratorAccount},
		"an_administrator_account_by_a_missing_caller":    {target: adminA, caller: uuid.NewV7(), locks: true, enabled: 2, refused: theAdministratorAccount},
		"an_administrator_account_by_an_unflagged_caller": {target: adminA, caller: unflaggedAcct, locks: true, enabled: 2, refused: theAdministratorAccount},
		"an_administrator_account_with_no_caller":         {target: adminA, locks: true, enabled: 2, fault: true},
		"an_administrator_account_by_another_one":         {target: adminA, caller: adminB, locks: true, enabled: 2},
		"an_administrator_account_by_itself":              {target: adminA, caller: adminA, locks: true, enabled: 2},
		// It reaches the guard with an access token that has not expired yet:
		// locked, it is not one that may ask.
		"an_administrator_account_by_a_disabled_one": {target: adminA, caller: adminOff, locks: true, enabled: 2, refused: theAdministratorAccount},
		// The last that can still sign in, by anybody. A caller without
		// the flag is not told that it was the last.
		"the_last_enabled_one_by_another_administrator": {target: adminA, caller: adminB, locks: true, enabled: 1, refused: theLastOne},
		"the_last_enabled_one_by_itself":                {target: adminA, caller: adminA, locks: true, enabled: 1, refused: theLastOne},
		"the_last_enabled_one_by_a_user_manager":        {target: adminA, caller: managerAcct, locks: true, enabled: 1, refused: theAdministratorAccount},
		// What does not take a sign-in away is not counted.
		"a_role_taken_from_the_last_enabled_one": {target: adminA, caller: adminB, locks: false, enabled: 1},
		// One that cannot sign in as it is has no sign-in to take, whatever
		// the count: disabled already, or federated with no identity.
		"a_disabled_one_when_one_is_left":                   {target: adminOff, caller: adminA, locks: true, enabled: 1},
		"a_federated_one_with_no_identity_when_one_is_left": {target: adminFederated, caller: adminA, locks: true, enabled: 1},
		// One that signs in through a provider is one that can.
		"the_last_enabled_one_signs_in_through_a_provider": {target: adminSSO, caller: adminA, locks: true, enabled: 1, refused: theLastOne},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(c.enabled)

			err := guardAdministrator(t.Context(), repo, repo.accounts[c.target], c.caller, administratorDisabled, c.locks)

			switch {
			case c.fault:
				if !errors.Is(err, errAdministratorChangeWithoutCaller) {
					t.Fatalf("got %T (%v), want the fault of a route wired without its caller", err, err)
				}
			case c.refused != "":
				refusedAs(t, err, c.refused)
			case err != nil:
				t.Fatalf("refused: %v", err)
			}
		})
	}
}

// The update refuses a CHANGE of `disabled` or `local_account` on an
// administrator account, in either direction, and not the field: a client
// sends the whole form back with every edit, `disabled: false` and the rest,
// and a user manager who renames the administrator account sends what it
// read.
func TestAUsersUpdateDoesNotLockAnAdministratorAccount(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		input   domain.UpdateUserInput
		refused string
		// fault: a change that names no caller, from a route wired wrongly.
		fault bool
		// handed is what the statement must be handed of the two fields, as
		// "disabled=<v> local=<v>" with "-" for none; "" does not look.
		handed string
	}{
		"a_user_manager_disables_it":              {input: domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, Disabled: new(true)}, refused: theAdministratorAccount},
		"a_user_manager_makes_it_federated":       {input: domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, LocalAccount: new(false)}, refused: theAdministratorAccount},
		"a_user_manager_renames_it":               {input: domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, FirstName: new("Jane"), Disabled: new(false), LocalAccount: new(true)}},
		"a_user_manager_sends_disabled_as_it_is":  {input: domain.UpdateUserInput{ID: adminOff, CallerID: managerAcct, FirstName: new("Jane"), Disabled: new(true)}},
		"a_user_manager_disables_an_ordinary_one": {input: domain.UpdateUserInput{ID: ordinary, CallerID: managerAcct, Disabled: new(true)}},
		"a_user_manager_federates_an_ordinary_one": {
			input: domain.UpdateUserInput{ID: ordinary, CallerID: managerAcct, LocalAccount: new(false)},
		},
		"an_administrator_disables_another": {input: domain.UpdateUserInput{ID: adminA, CallerID: adminB, Disabled: new(true)}, handed: "disabled=true local=-"},
		// What the frontend's form sends: the names, the address it read
		// and `disabled`, together. The address is no change; the lock is.
		"a_user_manager_sends_the_whole_form_with_disabled": {
			input:   domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, FirstName: new("Jane"), LastName: new("Doe"), Email: new("a@example.com"), Disabled: new(true)},
			refused: theAdministratorAccount,
		},
		"a_user_manager_sends_the_whole_form_as_it_is": {
			input: domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, FirstName: new("Jane"), LastName: new("Doe"), Email: new("a@example.com"), Disabled: new(false)},
		},
		// A federated one, told again to be federated: no change.
		"a_user_manager_sends_local_account_as_it_is": {
			input: domain.UpdateUserInput{ID: adminFederated, CallerID: managerAcct, FirstName: new("Jane"), LocalAccount: new(false)},
		},
		// The other direction: what an administrator locked is
		// unlocked by an administrator.
		"a_user_manager_enables_one_an_administrator_disabled": {
			input: domain.UpdateUserInput{ID: adminOff, CallerID: managerAcct, Disabled: new(false)}, refused: theAdministratorAccount,
		},
		"a_user_manager_makes_a_federated_one_local_again": {
			input: domain.UpdateUserInput{ID: adminFederated, CallerID: managerAcct, LocalAccount: new(true)}, refused: theAdministratorAccount,
		},
		"a_user_manager_does_it_with_the_whole_form": {
			input:   domain.UpdateUserInput{ID: adminOff, CallerID: managerAcct, FirstName: new("Jane"), LastName: new("Doe"), Email: new("off@example.com"), Disabled: new(false)},
			refused: theAdministratorAccount,
		},
		"an_administrator_enables_another":             {input: domain.UpdateUserInput{ID: adminOff, CallerID: adminB, Disabled: new(false)}, handed: "disabled=false local=-"},
		"an_administrator_makes_a_federated_one_local": {input: domain.UpdateUserInput{ID: adminFederated, CallerID: adminB, LocalAccount: new(true)}, handed: "disabled=- local=true"},
		"a_user_manager_enables_an_ordinary_one":       {input: domain.UpdateUserInput{ID: ordinaryOff, CallerID: managerAcct, Disabled: new(false), LocalAccount: new(true)}, handed: "disabled=false local=true"},
		"an_unlock_that_names_no_caller_is_a_fault":    {input: domain.UpdateUserInput{ID: adminOff, Disabled: new(false)}, fault: true},
		// An account read without its `local_account` has no password
		// login: false is no change, true makes it a local account.
		"a_user_manager_sends_local_account_false_where_it_was_never_set": {
			input: domain.UpdateUserInput{ID: adminNeverSet, CallerID: managerAcct, FirstName: new("Jane"), LocalAccount: new(false)}, handed: "disabled=- local=-",
		},
		"a_user_manager_makes_one_never_set_a_local_account": {
			input: domain.UpdateUserInput{ID: adminNeverSet, CallerID: managerAcct, LocalAccount: new(true)}, refused: theAdministratorAccount,
		},
		// A locked account with the flag is not one that may ask: not for
		// itself, and not against another.
		"a_disabled_administrator_does_not_enable_itself": {
			input: domain.UpdateUserInput{ID: adminOff, CallerID: adminOff, Disabled: new(false)}, refused: theAdministratorAccount,
		},
		"a_disabled_administrator_does_not_disable_another": {
			input: domain.UpdateUserInput{ID: adminA, CallerID: adminOff, Disabled: new(true)}, refused: theAdministratorAccount,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(2)
			input := c.input

			err := usersServiceOver(t, repo).UpdateByID(t.Context(), &input)

			if c.fault {
				if !errors.Is(err, errAdministratorChangeWithoutCaller) {
					t.Fatalf("got %T (%v), want the fault of a route wired without its caller", err, err)
				}

				if len(repo.writes) != 0 {
					t.Errorf("the account was written all the same: %v", repo.writes)
				}

				return
			}

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

			if len(repo.writes) != 1 {
				t.Fatalf("writes = %v, want the update", repo.writes)
			}

			if c.handed != "" {
				show := func(v *bool) string {
					if v == nil {
						return "-"
					}

					return strconv.FormatBool(*v)
				}

				if got := "disabled=" + show(repo.written.Disabled) + " local=" + show(repo.written.LocalAccount); got != c.handed {
					t.Errorf("the statement was handed %s, want %s", got, c.handed)
				}
			}
		})
	}

	for name, input := range map[string]domain.UpdateUserInput{
		"the_last_enabled_one_is_disabled_by_nobody":  {ID: adminA, CallerID: adminA, Disabled: new(true)},
		"the_last_enabled_one_is_federated_by_nobody": {ID: adminA, CallerID: adminB, LocalAccount: new(false)},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(1)

			refusedAs(t, usersServiceOver(t, repo).UpdateByID(t.Context(), &input), theLastOne)

			if len(repo.writes) != 0 {
				t.Errorf("the account was written all the same: %v", repo.writes)
			}
		})
	}

	// A locking value that is no change is not handed to the statement, on
	// an account with the flag: between the check and the write an
	// administrator may have enabled it, and "disabled, as it is" would
	// disable it again.
	t.Run("a_locking_value_that_is_no_change_is_not_written", func(t *testing.T) {
		t.Parallel()

		repo := administratorAccounts(2)

		input := domain.UpdateUserInput{ID: adminOff, CallerID: managerAcct, FirstName: new("Jane"), Disabled: new(true)}
		if err := usersServiceOver(t, repo).UpdateByID(t.Context(), &input); err != nil {
			t.Fatalf("refused: %v", err)
		}

		if repo.written == nil || repo.written.Disabled != nil {
			t.Errorf("the statement was handed disabled = %v for a value that is no change", repo.written)
		}

		federated := domain.UpdateUserInput{ID: adminFederated, CallerID: managerAcct, FirstName: new("Jane"), LocalAccount: new(false)}
		if err := usersServiceOver(t, repo).UpdateByID(t.Context(), &federated); err != nil {
			t.Fatalf("refused: %v", err)
		}

		if repo.written == nil || repo.written.LocalAccount != nil {
			t.Errorf("the statement was handed local_account = %v for a value that is no change", repo.written)
		}

		// And the other way: "enabled and local, as it is" would undo an
		// administrator who disabled or federated it in between.
		asItIs := domain.UpdateUserInput{ID: adminA, CallerID: managerAcct, FirstName: new("Jane"), Disabled: new(false), LocalAccount: new(true)}
		if err := usersServiceOver(t, repo).UpdateByID(t.Context(), &asItIs); err != nil {
			t.Fatalf("refused: %v", err)
		}

		if repo.written == nil || repo.written.Disabled != nil || repo.written.LocalAccount != nil {
			t.Errorf("the statement was handed %+v for values that are no change", repo.written)
		}
	})

	// What gives a sign-in back is not counted: the last enabled account
	// with the flag enables another.
	t.Run("the_last_enabled_one_enables_another", func(t *testing.T) {
		t.Parallel()

		repo := administratorAccounts(1)

		if err := usersServiceOver(t, repo).UpdateByID(t.Context(), &domain.UpdateUserInput{ID: adminOff, CallerID: adminA, Disabled: new(false)}); err != nil {
			t.Fatalf("refused: %v", err)
		}

		if err := usersServiceOver(t, repo).UpdateByID(t.Context(), &domain.UpdateUserInput{ID: adminFederated, CallerID: adminA, LocalAccount: new(true)}); err != nil {
			t.Fatalf("making a federated one local again was refused as if it took a sign-in away: %v", err)
		}
	})
}

// The delete, and the unlink of a user's roles.
func TestAnAdministratorAccountIsDeletedOrStrippedOnlyByOne(t *testing.T) {
	t.Parallel()

	role := uuid.NewV7()

	for name, c := range map[string]struct {
		call    func(*UsersService) error
		enabled int
		refused string
		// missing: the account is not there, and the repository says so.
		missing bool
		// failed: the account could not be read, and nothing is done.
		failed bool
	}{
		"a_user_manager_deletes_it": {
			call: func(s *UsersService) error {
				return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: adminA, CallerID: managerAcct})
			},
			enabled: 2, refused: theAdministratorAccount,
		},
		"an_administrator_deletes_another": {
			call: func(s *UsersService) error {
				return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: adminA, CallerID: adminB})
			},
			enabled: 2,
		},
		"nobody_deletes_the_last_enabled_one": {
			call: func(s *UsersService) error {
				return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: adminA, CallerID: adminB})
			},
			enabled: 1, refused: theLastOne,
		},
		"a_user_manager_deletes_an_ordinary_account": {
			call: func(s *UsersService) error {
				return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: ordinary, CallerID: managerAcct})
			},
			enabled: 2,
		},
		// A read of the account that fails refuses: whether it has the flag
		// is not known.
		"an_account_that_cannot_be_read_is_not_deleted": {
			call: func(s *UsersService) error {
				return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: unreadableAcct, CallerID: adminA})
			},
			enabled: 2, failed: true,
		},
		"an_account_that_cannot_be_read_is_not_stripped": {
			call: func(s *UsersService) error {
				return s.UnlinkRoles(t.Context(), &domain.UnlinkRolesFromUsersInput{UserID: unreadableAcct, CallerID: adminA, RoleIDs: []uuid.UUID{role}})
			},
			enabled: 2, failed: true,
		},
		// A delete that names no caller, of an account without the flag.
		"an_ordinary_account_with_no_caller": {
			call:    func(s *UsersService) error { return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: ordinary}) },
			enabled: 2,
		},
		"an_account_that_is_not_there": {
			call: func(s *UsersService) error {
				return s.DeleteByID(t.Context(), &domain.DeleteUserInput{ID: uuid.NewV7(), CallerID: managerAcct})
			},
			enabled: 2, missing: true,
		},
		"a_user_manager_takes_a_role_from_it": {
			call: func(s *UsersService) error {
				return s.UnlinkRoles(t.Context(), &domain.UnlinkRolesFromUsersInput{UserID: adminA, CallerID: managerAcct, RoleIDs: []uuid.UUID{role}})
			},
			enabled: 2, refused: theAdministratorAccount,
		},
		"an_administrator_takes_a_role_from_the_last_one": {
			call: func(s *UsersService) error {
				return s.UnlinkRoles(t.Context(), &domain.UnlinkRolesFromUsersInput{UserID: adminA, CallerID: adminB, RoleIDs: []uuid.UUID{role}})
			},
			enabled: 1,
		},
		"a_user_manager_takes_a_role_from_an_ordinary_account": {
			call: func(s *UsersService) error {
				return s.UnlinkRoles(t.Context(), &domain.UnlinkRolesFromUsersInput{UserID: ordinary, CallerID: managerAcct, RoleIDs: []uuid.UUID{role}})
			},
			enabled: 2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := administratorAccounts(c.enabled)

			err := c.call(usersServiceOver(t, repo))

			switch {
			case c.refused != "":
				refusedAs(t, err, c.refused)

				if len(repo.writes) != 0 {
					t.Errorf("it was done all the same: %v", repo.writes)
				}
			case c.failed:
				if !errors.Is(err, errTheDatabaseIsDown) {
					t.Fatalf("got %T (%v), want the read's own failure", err, err)
				}

				if len(repo.writes) != 0 {
					t.Errorf("it was done all the same: %v", repo.writes)
				}
			case c.missing:
				if _, missing := errors.AsType[*domain.UserNotFoundError](err); !missing {
					t.Fatalf("got %T (%v), want the account's own not-found", err, err)
				}
			default:
				if err != nil {
					t.Fatalf("refused: %v", err)
				}

				if len(repo.writes) != 1 {
					t.Errorf("writes = %v, want one", repo.writes)
				}
			}
		})
	}
}

// rolesUnlinkRepository records the one write of the role's side. Any other
// method panics: the embedded port is nil.
type rolesUnlinkRepository struct {
	repository.Roles
	unlinked bool
}

func (r *rolesUnlinkRepository) UnlinkUsers(context.Context, *domain.UnlinkUsersFromRoleInput) error {
	r.unlinked = true

	return nil
}

// The same strip from the other side: DELETE /roles/{role_id}/users names the
// users in its body. One administrator account among them refuses the whole
// request, unless an administrator account asks.
func TestARoleIsNotTakenFromAnAdministratorAccountThroughTheRole(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		users   []uuid.UUID
		caller  uuid.UUID
		refused bool
		fault   bool
	}{
		"ordinary_accounts_by_a_user_manager":     {users: []uuid.UUID{ordinary, unflaggedAcct}, caller: managerAcct},
		"an_administrator_account_among_them":     {users: []uuid.UUID{ordinary, adminA}, caller: managerAcct, refused: true},
		"an_administrator_account_by_another":     {users: []uuid.UUID{ordinary, adminA}, caller: adminB},
		"an_administrator_account_with_no_caller": {users: []uuid.UUID{adminA}, fault: true},
		// A locked account with the flag is not one that may ask.
		"an_administrator_account_by_a_disabled_one":     {users: []uuid.UUID{adminA}, caller: adminOff, refused: true},
		"ordinary_accounts_with_no_caller_are_as_before": {users: []uuid.UUID{ordinary}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			roles := &rolesUnlinkRepository{}

			svc, err := NewRolesService(RolesServiceConf{Guard: &GrantGuard{}, Repository: roles, Administrators: administratorAccounts(2), OT: testOT(t)})
			if err != nil {
				t.Fatalf("NewRolesService: %v", err)
			}

			err = svc.UnlinkUsers(t.Context(), &domain.UnlinkUsersFromRoleInput{RoleID: uuid.NewV7(), CallerID: c.caller, UserIDs: c.users})

			switch {
			case c.fault:
				if !errors.Is(err, errAdministratorChangeWithoutCaller) {
					t.Fatalf("got %T (%v), want the fault of a route wired without its caller", err, err)
				}

				if roles.unlinked {
					t.Error("the role was taken all the same")
				}
			case c.refused:
				refusedAs(t, err, theAdministratorAccount)

				if roles.unlinked {
					t.Error("the role was taken all the same")
				}
			default:
				if err != nil {
					t.Fatalf("refused: %v", err)
				}

				if !roles.unlinked {
					t.Error("nothing was unlinked")
				}
			}
		})
	}
}

// rolesDeleteRepository holds a few roles and records the delete of one;
// another is the repository's 404. Any other method panics: the embedded
// port is nil.
type rolesDeleteRepository struct {
	repository.Roles
	roles   map[uuid.UUID]bool
	deleted bool
}

func (r *rolesDeleteRepository) DeleteByID(_ context.Context, in *domain.DeleteRoleInput) error {
	if !r.roles[in.ID] {
		return &domain.RoleNotFoundError{RoleID: in.ID.String()}
	}

	r.deleted = true

	return nil
}

// The same strip, of every holder at once: DELETE /roles/{role_id} takes the
// role from every account that holds it, by the links' cascade. While an
// account with the flag holds it, the delete is the flag's, as the unlink is.
func TestARoleIsNotTakenFromAnAdministratorAccountByItsDelete(t *testing.T) {
	t.Parallel()

	held, unheld := uuid.NewV7(), uuid.NewV7()

	for name, c := range map[string]struct {
		role    uuid.UUID
		caller  uuid.UUID
		refused bool
		fault   bool
		missing bool
	}{
		"a_role_only_ordinary_accounts_hold":       {role: unheld, caller: managerAcct},
		"a_role_an_administrator_account_holds":    {role: held, caller: managerAcct, refused: true},
		"by_an_administrator_account":              {role: held, caller: adminB},
		"by_a_disabled_administrator_account":      {role: held, caller: adminOff, refused: true},
		"held_by_an_administrator_with_no_caller":  {role: held, fault: true},
		"held_by_nobody_with_no_caller_as_before":  {role: unheld},
		"one_that_is_not_there_is_the_repositorys": {role: uuid.NewV7(), caller: managerAcct, missing: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			administrators := administratorAccounts(2)
			administrators.holders = map[uuid.UUID][]uuid.UUID{
				held:   {ordinary, adminA},
				unheld: {ordinary, unflaggedAcct},
			}

			roles := &rolesDeleteRepository{roles: map[uuid.UUID]bool{held: true, unheld: true}}

			svc, err := NewRolesService(RolesServiceConf{Guard: &GrantGuard{}, Repository: roles, Administrators: administrators, OT: testOT(t)})
			if err != nil {
				t.Fatalf("NewRolesService: %v", err)
			}

			err = svc.DeleteByID(t.Context(), &domain.DeleteRoleInput{ID: c.role, CallerID: c.caller})

			switch {
			case c.fault:
				if !errors.Is(err, errAdministratorChangeWithoutCaller) {
					t.Fatalf("got %T (%v), want the fault of a route wired without its caller", err, err)
				}

				if roles.deleted {
					t.Error("the role was deleted all the same")
				}
			case c.refused:
				refusedAs(t, err, theAdministratorAccount)

				if roles.deleted {
					t.Error("the role was deleted all the same")
				}
			case c.missing:
				if _, ok := errors.AsType[*domain.RoleNotFoundError](err); !ok {
					t.Fatalf("got %T (%v), want the repository's 404", err, err)
				}
			default:
				if err != nil {
					t.Fatalf("refused: %v", err)
				}

				if !roles.deleted {
					t.Error("nothing was deleted")
				}
			}
		})
	}

	// A read of the holders that fails is no answer: the delete is not done.
	t.Run("a_failed_read_of_its_holders_deletes_nothing", func(t *testing.T) {
		t.Parallel()

		fault := errors.New("the database is gone")

		administrators := administratorAccounts(2)
		administrators.holdersErr = fault

		roles := &rolesDeleteRepository{roles: map[uuid.UUID]bool{unheld: true}}

		svc, err := NewRolesService(RolesServiceConf{Guard: &GrantGuard{}, Repository: roles, Administrators: administrators, OT: testOT(t)})
		if err != nil {
			t.Fatalf("NewRolesService: %v", err)
		}

		if err := svc.DeleteByID(t.Context(), &domain.DeleteRoleInput{ID: unheld, CallerID: managerAcct}); !errors.Is(err, fault) {
			t.Errorf("got %v, want the read's fault", err)
		}

		if roles.deleted {
			t.Error("the role was deleted with its holders unread")
		}
	})
}
