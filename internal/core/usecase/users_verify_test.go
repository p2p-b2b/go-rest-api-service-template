//go:build unit

package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/cache"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// verifyingUsers records what the verification asked the repository for, and
// holds the row the use-case reads when the statement wrote nothing. Any
// other method panics: the embedded port is nil.
type verifyingUsers struct {
	repository.Users
	verified []string
	err      error
	// stored is the row as it is, for the question "why was nothing
	// written"; nil is an account that is gone. reads counts the question.
	stored *domain.User
	reads  int
	// readErr is what that read answers instead, when it fails.
	readErr error
}

func (v *verifyingUsers) Verify(_ context.Context, id uuid.UUID, address string) error {
	v.verified = append(v.verified, id.String()+" "+address)

	return v.err
}

func (v *verifyingUsers) SelectByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	v.reads++

	if v.readErr != nil {
		return nil, v.readErr
	}

	if v.stored == nil {
		return nil, &domain.UserNotFoundError{ID: id}
	}

	return v.stored, nil
}

// droppingCache records what was invalidated.
type droppingCache struct {
	cache.Cache
	dropped []string
}

func (d *droppingCache) Invalidate(_ context.Context, id cache.Identifier) error {
	d.dropped = append(d.dropped, id.String())

	return nil
}

func verifyingService(t *testing.T, repo *verifyingUsers, dropped *droppingCache) *UsersService {
	t.Helper()

	svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: repo, CacheService: dropped, ResourcesLimits: noLimits{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewUsersService: %v", err)
	}

	return svc
}

// Following a verification link proves an address, and opens a registration
// whose link was the only thing keeping it shut. Whether it does is not for
// the use-case to decide from a copy of the account.
//
// The verification is one statement of the repository, which decides
// from the row it writes. The use-case hands it the account and the address
// the link was mailed to, and nothing else: it reads no row to decide from,
// and sends no update that could enable.
func TestTheVerificationIsTheRepositorysOneStatement(t *testing.T) {
	t.Parallel()

	repo := &verifyingUsers{}
	dropped := &droppingCache{}
	id := uuid.NewV7()

	if err := verifyingService(t, repo, dropped).Verify(t.Context(), id, "fresh@example.com"); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if want := id.String() + " fresh@example.com"; len(repo.verified) != 1 || repo.verified[0] != want {
		t.Fatalf("verified = %v, want the account with the address, once", repo.verified)
	}

	if repo.reads != 0 {
		t.Errorf("the account was read %d times: the statement decides, not a copy", repo.reads)
	}

	// The cached account says "unverified, disabled" for twelve hours
	// otherwise, and the sign-in reads it.
	want := cache.Identifier{Type: "user", ID: id.String()}.String()
	if len(dropped.dropped) != 1 || dropped.dropped[0] != want {
		t.Errorf("dropped from the cache = %v, want %s", dropped.dropped, want)
	}
}

// The statement writes only a row with that id and that address that is not
// verified yet, and answers "no such row" for the three ways there is none.
// Which one it is, is read from the row, so that whoever followed the link
// hears what they would have heard had the flow's own copy been current.
func TestAVerificationThatWritesNothingSaysWhy(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	noRow := &domain.UserNotFoundError{ID: id}

	for name, c := range map[string]struct {
		answered error
		stored   *domain.User
		readErr  error
		reads    int
		check    func(err error) bool
	}{
		"the_account_is_gone": {
			answered: noRow, reads: 1,
			check: func(err error) bool { return errors.Is(err, noRow) },
		},
		"it_is_verified_already": {
			answered: noRow, reads: 1,
			stored: &domain.User{ID: id, Email: "fresh@example.com", Verified: new(true), Disabled: new(true)},
			check: func(err error) bool {
				already, ok := errors.AsType[*domain.UserAlreadyVerifiedError](err)

				return ok && already.UserID == id.String()
			},
		},
		"its_address_is_another_now": {
			answered: noRow, reads: 1,
			stored: &domain.User{ID: id, Email: "corrected@example.com", Verified: new(false), Disabled: new(true)},
			check: func(err error) bool {
				return errors.Is(err, ErrVerifiedAddressIsNotTheAccounts) && !strings.Contains(err.Error(), "@")
			},
		},
		// The row cannot be read to say why. That is a fault, and not "the
		// account is gone": a link is answered 200 for an account that is
		// gone, and a read that failed must not be that 200.
		"the_row_cannot_be_read": {
			answered: noRow, reads: 1, readErr: errFaultOfTheTest,
			check: func(err error) bool {
				_, gone := errors.AsType[*domain.UserNotFoundError](err)

				return errors.Is(err, errFaultOfTheTest) && !gone
			},
		},
		// Anything else the statement answers is not a question about the
		// row, and is passed on as it is.
		"a_fault": {
			answered: errFaultOfTheTest, reads: 0,
			check: func(err error) bool { return errors.Is(err, errFaultOfTheTest) },
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &verifyingUsers{err: c.answered, stored: c.stored, readErr: c.readErr}
			dropped := &droppingCache{}

			err := verifyingService(t, repo, dropped).Verify(t.Context(), id, "fresh@example.com")
			if !c.check(err) {
				t.Fatalf("got %T (%v)", err, err)
			}

			if repo.reads != c.reads {
				t.Errorf("the account was read %d times, want %d", repo.reads, c.reads)
			}

			if len(dropped.dropped) != 0 {
				t.Errorf("nothing was written and the cache was dropped: %v", dropped.dropped)
			}
		})
	}
}

var errFaultOfTheTest = errors.New("the database is away")
