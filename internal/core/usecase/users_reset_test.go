//go:build unit

package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/cache"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// resetRepository holds one account and answers the reset's statement with
// resetErr. Any other method panics: the embedded port is nil.
type resetRepository struct {
	repository.Users
	account  *domain.User
	resetErr error
	reset    bool
}

func (r *resetRepository) SelectByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if r.account == nil || r.account.ID != id {
		return nil, &domain.UserNotFoundError{ID: id}
	}

	return r.account, nil
}

func (r *resetRepository) ResetPassword(context.Context, uuid.UUID, string, string) error {
	if r.resetErr != nil {
		return r.resetErr
	}

	r.reset = true

	return nil
}

// droppedKeys records what is invalidated. Get is never asked here.
type droppedKeys struct {
	cache.Cache
	dropped []string
}

func (d *droppedKeys) Invalidate(_ context.Context, id cache.Identifier) error {
	d.dropped = append(d.dropped, id.String())

	return nil
}

func resetServiceOver(t *testing.T, repo *resetRepository, c cache.Cache) *UsersService {
	t.Helper()

	svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: repo, CacheService: c, ResourcesLimits: slotsGivenBack{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewUsersService: %v", err)
	}

	return svc
}

// A reset written drops the account's cached copy, as every write of the
// account does; one refused drops nothing and answers the statement's
// NotFoundError, which the authentication use-case turns into its 401.
func TestAResetDropsTheCachedAccountOnlyWhenItWrote(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	key := cache.Identifier{Type: "user", ID: id.String()}.String()

	t.Run("written", func(t *testing.T) {
		t.Parallel()

		repo := &resetRepository{}
		dropped := &droppedKeys{}

		if err := resetServiceOver(t, repo, dropped).ResetPassword(t.Context(), id, "holder@example.com", "hash"); err != nil {
			t.Fatalf("ResetPassword: %v", err)
		}

		if !repo.reset || !slices.Contains(dropped.dropped, key) {
			t.Errorf("written %v, dropped %v; want written and %s dropped", repo.reset, dropped.dropped, key)
		}
	})

	t.Run("refused", func(t *testing.T) {
		t.Parallel()

		repo := &resetRepository{resetErr: &domain.UserNotFoundError{ID: id}}
		dropped := &droppedKeys{}

		err := resetServiceOver(t, repo, dropped).ResetPassword(t.Context(), id, "holder@example.com", "hash")
		if !isUserGone(err) {
			t.Fatalf("got %T (%v), want the statement's NotFoundError", err, err)
		}

		if len(dropped.dropped) != 0 {
			t.Errorf("a refused reset dropped %v", dropped.dropped)
		}
	})

	t.Run("a_fault", func(t *testing.T) {
		t.Parallel()

		fault := errors.New("the database is gone")
		repo := &resetRepository{resetErr: fault}

		if err := resetServiceOver(t, repo, &droppedKeys{}).ResetPassword(t.Context(), id, "holder@example.com", "hash"); !errors.Is(err, fault) {
			t.Errorf("got %v, want the fault", err)
		}
	})
}

// Where a reset link is mailed is read from the repository, and the account
// it answers carries no credentials: it is the account every other read
// answers, without the cache.
func TestTheStoredAccountIsReadWithoutItsCredentials(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	repo := &resetRepository{account: &domain.User{ID: id, Email: "holder@example.com", PasswordHash: "$2a$10$secret", Password: "plain"}}

	got, err := resetServiceOver(t, repo, nil).StoredAccount(t.Context(), id)
	if err != nil {
		t.Fatalf("StoredAccount: %v", err)
	}

	if got.Email != "holder@example.com" || got.PasswordHash != "" || got.Password != "" {
		t.Errorf("answered %+v, want the address and no credentials", got)
	}

	if repo.account.PasswordHash == "" {
		t.Error("the repository's own copy was stripped in place")
	}

	if _, err := resetServiceOver(t, repo, nil).StoredAccount(t.Context(), uuid.NewV7()); !isUserGone(err) {
		t.Errorf("an account that is not there: got %v, want its NotFoundError", err)
	}
}

// isUserGone reports an account that is not there, as the twin's
// repositories answer it.
func isUserGone(err error) bool {
	_, gone := errors.AsType[*domain.UserNotFoundError](err)

	return gone
}
