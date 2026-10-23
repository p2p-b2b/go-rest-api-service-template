package usecase

import (
	"context"
	"errors"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// AdministratorAccounts is what the guard of the administrator account reads:
// an account, how many accounts with the flag may still sign in and whether
// one of them is among those, and which of a set of accounts have the flag. The users repository provides it.
type AdministratorAccounts interface {
	SelectByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	CountEnabledAdministrators(ctx context.Context) (int, error)
	IsEnabledAdministrator(ctx context.Context, id uuid.UUID) (bool, error)
	SelectAdministratorsAmong(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)
	SelectAdministratorsHolding(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
}

// errAdministratorChangeWithoutCaller is what a change to an administrator
// account answers when it does not know who asks: a route wired without its
// caller. It is no domain error, so the answer is the fixed 500.
var errAdministratorChangeWithoutCaller = errors.New("a change to an administrator account was asked for with no caller")

// What is done to an account, in the words of the refusal.
const (
	administratorDisabled  = "disabled"
	administratorEnabled   = "enabled"
	administratorFederated = "made a federated account"
	administratorMadeLocal = "made a local account"
	administratorDeleted   = "deleted"
	administratorStripped  = "stripped of a role"
	// Taking the flag away takes the account out of the set the guard
	// counts, as disabling it does.
	administratorDemoted = "stripped of the admin flag"
)

// The two rules, in the words of the refusal.
const (
	theAdministratorAccount     = "the administrator account"
	theLastAdministratorAccount = "the last enabled administrator account"
)

// guardAdministrator refuses what would take an administrator account out of
// its holder's hands, unless an administrator account asks.
//
// "Administrator account" is the `admin` flag, not the Administrator role.
// The flag is every project and the things only it may do. The seeded
// UserManager role manages accounts, and that included this one: with that
// role's grants an account with the flag could be disabled, made a federated
// account (which silences its password and its recovery), stripped of its
// roles or deleted, and the account a deployment is operated from was out of
// its holder's hands until somebody wrote to the database.
//
// An account with the flag is disabled, made federated, stripped of a role
// or deleted only by an account with the flag, and the last one that can
// still sign in by nobody. Every other account stays the role's to manage.
//
// The other direction is the flag's too: enabling such an account that an
// administrator disabled, and making one that an administrator federated a
// local account again, which lets its stored password and its recovery work
// again. An account with the flag has its sign-in changed, either way, only
// by one that has the flag.
//
// locks says whether the change takes the account's sign-in away: those
// that do are also refused, for everybody, on the last account with the flag
// that can sign in. What gives a sign-in back is not counted.
//
// An account that cannot sign in as it is has no sign-in to take: one that
// is disabled, one that is no longer local and has no identity, one whose
// only identity is at a provider that is disabled. Locking it takes nothing
// away, whatever the count is. Which accounts those are is the count's own
// predicate (IsEnabledAdministrator), not a second opinion held here, so an
// account the count leaves out is never refused as "the last one".
//
// A caller that does not have the flag is not told whether the account was
// the last one.
func guardAdministrator(ctx context.Context, accounts AdministratorAccounts, target *domain.User, callerID uuid.UUID, action string, locks bool) error {
	if !hasAdminFlag(target) {
		return nil
	}

	asked, err := isAdministratorAccount(ctx, accounts, callerID)
	if err != nil {
		return err
	}

	if !asked {
		return &domain.AdministratorProtectedError{Action: action, Message: theAdministratorAccount}
	}

	if !locks {
		return nil
	}

	signsIn, err := accounts.IsEnabledAdministrator(ctx, target.ID)
	if err != nil {
		return err
	}

	if !signsIn {
		return nil
	}

	enabled, err := accounts.CountEnabledAdministrators(ctx)
	if err != nil {
		return err
	}

	if enabled <= 1 {
		return &domain.AdministratorProtectedError{Action: action, Message: theLastAdministratorAccount}
	}

	return nil
}

// isAdministratorAccount reports whether id is an account with the `admin`
// flag that may act: one that is not disabled. It is read from the
// repository, not the cache: the answer decides who may have, or lock, an
// account. An account that is not there, and a row read without its flag,
// are not the administrator's.
//
// A disabled account does not sign in, and is asked about all the same: it
// is the account an administrator locked, and the one thing it must not be
// able to do, with a token issued before, is enable itself.
func isAdministratorAccount(ctx context.Context, accounts AdministratorAccounts, id uuid.UUID) (bool, error) {
	if id == uuid.Nil() {
		return false, errAdministratorChangeWithoutCaller
	}

	caller, err := accounts.SelectByID(ctx, id)
	if _, missing := errors.AsType[*domain.UserNotFoundError](err); missing {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return hasAdminFlag(caller) && !isDisabled(caller), nil
}
