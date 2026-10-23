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
)

// usersAddressRepository holds a few accounts and records what was read and
// whether anything was written. Any other method panics: the embedded port
// is nil.
type usersAddressRepository struct {
	repository.Users
	accounts map[uuid.UUID]*domain.User
	read     []uuid.UUID
	written  *domain.UpdateUserInput
	// unreadable is an account whose read fails.
	unreadable uuid.UUID
}

var errUnreadable = errors.New("the database is away")

func (r *usersAddressRepository) SelectByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	r.read = append(r.read, id)

	if id == r.unreadable && id != uuid.Nil() {
		return nil, errUnreadable
	}

	if account, ok := r.accounts[id]; ok {
		return account, nil
	}

	return nil, &domain.UserNotFoundError{ID: id}
}

func (r *usersAddressRepository) UpdateByID(_ context.Context, in *domain.UpdateUserInput) error {
	r.written = in

	return nil
}

// staleUsers is a cache that answers what it was told, whatever the
// repository holds, and forgets nothing.
type staleUsers struct {
	entries map[string]any
}

func (c *staleUsers) Get(ctx context.Context, id cache.Identifier, dest any, fetcher cache.Fetcher) error {
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

func (c *staleUsers) Invalidate(context.Context, cache.Identifier) error { return nil }

// limitsNeverAsked is the usage limits of a service whose update asks for
// none. Any method panics: the embedded port is nil.
type limitsNeverAsked struct{ ResourcesLimitsServiceConsumer }

// An account's e-mail address is where its password reset is sent, so it is
// part of how the account is held. Only the administrator account, the one
// with the `admin` flag, changes an account's address.
func TestOnlyTheAdministratorAccountChangesAnAddress(t *testing.T) {
	t.Parallel()

	target, manager, administrator := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()
	older, unflagged, gone, locked := uuid.NewV7(), uuid.NewV7(), uuid.NewV7(), uuid.NewV7()
	neverLocked, unreadable := uuid.NewV7(), uuid.NewV7()

	accounts := func() map[uuid.UUID]*domain.User {
		return map[uuid.UUID]*domain.User{
			target:        {ID: target, Email: "holder.smith@example.com", Admin: new(false)},
			manager:       {ID: manager, Email: "manager@example.com", Admin: new(false)},
			administrator: {ID: administrator, Email: "admin@example.com", Admin: new(true), Disabled: new(false)},
			// A row older than the spelling every address is given now.
			older: {ID: older, Email: "Older.Holder@Example.com", Admin: new(false)},
			// The column is nullable: a flag that was never set.
			unflagged: {ID: unflagged, Email: "unflagged@example.com"},
			// An account with the flag that was disabled.
			locked: {ID: locked, Email: "locked@example.com", Admin: new(true), Disabled: new(true)},
			// The column is nullable here too: never set is not disabled.
			neverLocked: {ID: neverLocked, Email: "never@example.com", Admin: new(true)},
		}
	}

	// What the cache holds is not what decides: here it says the manager is
	// the administrator, and that the target's address is already the
	// manager's. Neither may be believed.
	stale := func() *staleUsers {
		return &staleUsers{entries: map[string]any{
			cache.Identifier{Type: "user", ID: manager.String()}.String(): &domain.User{ID: manager, Email: "manager@example.com", Admin: new(true)},
			cache.Identifier{Type: "user", ID: target.String()}.String():  &domain.User{ID: target, Email: "mine@example.com", Admin: new(false)},
		}}
	}

	type want int

	const (
		written want = iota
		refused
		fault
		missing
		unread
	)

	for name, c := range map[string]struct {
		input domain.UpdateUserInput
		want  want
		// reads is how many accounts are looked up: none when the body
		// carries no address.
		reads int
		// address is what the statement is handed, when something is
		// written: nothing, unless the administrator account changed it.
		address string
	}{
		"another_accounts_address_by_an_account_without_the_flag": {
			input: domain.UpdateUserInput{ID: target, CallerID: manager, Email: new("mine@example.com")},
			want:  refused, reads: 2,
		},
		"its_own_address_by_an_account_without_the_flag": {
			input: domain.UpdateUserInput{ID: manager, CallerID: manager, Email: new("other@example.com")},
			want:  refused, reads: 2,
		},
		"another_accounts_address_by_the_administrator_account": {
			input: domain.UpdateUserInput{ID: target, CallerID: administrator, Email: new("corrected@example.com")},
			want:  written, reads: 2, address: "corrected@example.com",
		},
		// A client sends the whole form back: the address it read is no
		// change, the rest of the body is written, and the statement is
		// handed no address at all.
		"the_stored_address_sent_back": {
			input: domain.UpdateUserInput{ID: target, CallerID: manager, Email: new("holder.smith@example.com"), FirstName: new("Jane")},
			want:  written, reads: 1,
		},
		"the_stored_address_typed_in_another_case": {
			input: domain.UpdateUserInput{ID: target, CallerID: manager, Email: new("Holder.Smith@Example.com")},
			want:  written, reads: 1,
		},
		// The row itself is in mixed case, and the form sends back what the
		// service made of it: no change, and the row is not rewritten.
		"an_older_rows_address_sent_back": {
			input: domain.UpdateUserInput{ID: older, CallerID: manager, Email: new("older.holder@example.com")},
			want:  written, reads: 2,
		},
		// The administrator account may write it in the service's spelling:
		// such a row cannot sign in, since a sign-in looks the address up
		// as the service spells it. For that account only the same bytes
		// are no change.
		"an_older_rows_address_respelled_by_the_administrator_account": {
			input: domain.UpdateUserInput{ID: older, CallerID: administrator, Email: new("older.holder@example.com")},
			want:  written, reads: 2, address: "older.holder@example.com",
		},
		"an_older_rows_address_sent_back_with_no_caller": {
			input: domain.UpdateUserInput{ID: older, Email: new("older.holder@example.com")},
			want:  written, reads: 1,
		},
		// Whether the caller may respell it is not known, and the rest of
		// the body is not written without it.
		"the_caller_of_a_respelling_cannot_be_read": {
			input: domain.UpdateUserInput{ID: older, CallerID: unreadable, Email: new("older.holder@example.com"), FirstName: new("Jane")},
			want:  unread, reads: 2,
		},
		// A long s for an s. A comparison that folds case calls the two
		// equal; no lowering turns one into the other; the statement
		// compares bytes. It is another address.
		"an_address_that_only_folds_to_the_stored_one": {
			input: domain.UpdateUserInput{ID: target, CallerID: manager, Email: new("holder.ſmith@example.com")},
			want:  refused, reads: 2,
		},
		"an_address_the_stored_one_begins": {
			input: domain.UpdateUserInput{ID: target, CallerID: manager, Email: new("holder.smith@example.com.other.example")},
			want:  refused, reads: 2,
		},
		"a_caller_whose_flag_was_never_set": {
			input: domain.UpdateUserInput{ID: target, CallerID: unflagged, Email: new("mine@example.com")},
			want:  refused, reads: 2,
		},
		"a_caller_that_is_not_there": {
			input: domain.UpdateUserInput{ID: target, CallerID: gone, Email: new("mine@example.com")},
			want:  refused, reads: 2,
		},
		"a_disabled_account_with_the_flag": {
			input: domain.UpdateUserInput{ID: target, CallerID: locked, Email: new("mine@example.com")},
			want:  refused, reads: 2,
		},
		"an_account_with_the_flag_whose_disabled_was_never_set": {
			input: domain.UpdateUserInput{ID: target, CallerID: neverLocked, Email: new("corrected@example.com")},
			want:  written, reads: 2, address: "corrected@example.com",
		},
		// A read that fails is not an answer: nothing is written, whichever
		// of the two accounts could not be read.
		"the_caller_cannot_be_read": {
			input: domain.UpdateUserInput{ID: target, CallerID: unreadable, Email: new("mine@example.com")},
			want:  unread, reads: 2,
		},
		"the_account_cannot_be_read": {
			input: domain.UpdateUserInput{ID: unreadable, CallerID: administrator, Email: new("mine@example.com")},
			want:  unread, reads: 1,
		},
		"no_address_in_the_body": {
			input: domain.UpdateUserInput{ID: target, CallerID: manager, FirstName: new("Jane"), LastName: new("Doe")},
			want:  written, reads: 0,
		},
		// The flows that update an account for itself carry no address and
		// no caller, and are not asked who they are.
		"no_address_and_no_caller": {
			input: domain.UpdateUserInput{ID: target, FirstName: new("Jane")},
			want:  written, reads: 0,
		},
		"a_change_with_no_caller_is_a_fault_not_a_refusal": {
			input: domain.UpdateUserInput{ID: target, Email: new("mine@example.com")},
			want:  fault, reads: 1,
		},
		"an_account_that_does_not_exist": {
			input: domain.UpdateUserInput{ID: uuid.NewV7(), CallerID: administrator, Email: new("mine@example.com")},
			want:  missing, reads: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &usersAddressRepository{accounts: accounts(), unreadable: unreadable}

			svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: repo, CacheService: stale(), ResourcesLimits: limitsNeverAsked{}, OT: testOT(t)})
			if err != nil {
				t.Fatalf("NewUsersService: %v", err)
			}

			input := c.input
			err = svc.UpdateByID(t.Context(), &input)

			if len(repo.read) != c.reads {
				t.Errorf("%d accounts were looked up, want %d: %v", len(repo.read), c.reads, repo.read)
			}

			switch c.want {
			case written:
				if err != nil {
					t.Fatalf("refused: %v", err)
				}

				if repo.written == nil {
					t.Fatal("nothing was written")
				}

				switch handed := repo.written.Email; {
				case c.address == "" && handed != nil:
					t.Errorf("the statement was handed the address %q for a body that changes none", *handed)
				case c.address != "" && (handed == nil || *handed != c.address):
					t.Errorf("the statement was handed %v, want the address %q", handed, c.address)
				}

			case refused:
				refusal, ok := errors.AsType[*domain.AdministratorOnlyError](err)
				if !ok {
					t.Fatalf("got %T (%v), want the refusal the handler answers 403", err, err)
				}

				if refusal.Field != domain.FieldEmail {
					t.Errorf("the refusal names %q, want the address", refusal.Field)
				}

				if repo.written != nil {
					t.Errorf("the account was written all the same: %+v", repo.written)
				}

			case fault:
				if !errors.Is(err, errAddressChangeWithoutCaller) {
					t.Fatalf("got %T (%v), want the fault of a route wired without its caller", err, err)
				}

				if repo.written != nil {
					t.Error("the account was written all the same")
				}

			case unread:
				if !errors.Is(err, errUnreadable) {
					t.Fatalf("got %T (%v), want the read's own failure", err, err)
				}

				if repo.written != nil {
					t.Errorf("an account was written on a read that failed: %+v", repo.written)
				}

			case missing:
				if _, ok := errors.AsType[*domain.UserNotFoundError](err); !ok {
					t.Fatalf("got %T (%v), want the account's own not-found", err, err)
				}

				if repo.written != nil {
					t.Error("something was written for an account that does not exist")
				}
			}
		})
	}
}

// The refusal says what is refused and names no address: not the one that
// was sent, and not the one that is stored.
func TestTheAddressRefusalNamesNoAddress(t *testing.T) {
	t.Parallel()

	got := (&domain.AdministratorOnlyError{Entity: "user", Field: domain.FieldEmail}).Error()

	if got != "user email can be changed only by the administrator account" {
		t.Errorf("the refusal says %q", got)
	}
}
