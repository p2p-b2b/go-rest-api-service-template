//go:build unit

package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// administratorsCounted answers how many administrator accounts can sign in,
// and how many could without one provider, and records what it was asked.
type administratorsCounted struct {
	err error
	// errNow is what the count of all of them answers, when the count
	// without a provider answered.
	errNow       error
	askedWithout []uuid.UUID
	now, without int
	askedNow     int
}

func (a *administratorsCounted) CountEnabledAdministrators(context.Context) (int, error) {
	a.askedNow++

	if a.errNow != nil {
		return 0, a.errNow
	}

	return a.now, a.err
}

func (a *administratorsCounted) CountEnabledAdministratorsWithout(_ context.Context, idpID uuid.UUID) (int, error) {
	a.askedWithout = append(a.askedWithout, idpID)

	return a.without, a.err
}

// providersStored is the identity providers' repository for an update and a
// delete: one provider, enabled or not, and what was written. Any other
// method panics: the embedded port is nil.
type providersStored struct {
	repository.IDPs
	deleteAnswers error
	writes        []string
	enabled       bool
}

func (p *providersStored) SelectByID(_ context.Context, id uuid.UUID) (*domain.IDP, error) {
	return &domain.IDP{ID: id, Enabled: p.enabled}, nil
}

func (p *providersStored) UpdateByID(context.Context, *domain.UpdateIDPInput) error {
	p.writes = append(p.writes, "UpdateByID")

	return nil
}

func (p *providersStored) DeleteByID(context.Context, *domain.DeleteIDPInput) error {
	if p.deleteAnswers != nil {
		return p.deleteAnswers
	}

	p.writes = append(p.writes, "DeleteByID")

	return nil
}

// slotsCounted is the usage limits a delete gives its slot back to.
type slotsCounted struct {
	ResourcesLimitsServiceConsumer
	givenBack int
}

func (s *slotsCounted) DecrementUsage(context.Context, domain.ResourcesLimitsScope, domain.ResourcesLimitsResourceType) error {
	s.givenBack++

	return nil
}

func providersService(t *testing.T, repo *providersStored, administrators *administratorsCounted, slots *slotsCounted) *IDPsService {
	t.Helper()

	svc, err := NewIDPsService(IDPsServiceConf{
		Repository:      repo,
		IDPTypes:        typesNeverAsked{},
		Administrators:  administrators,
		Cipher:          fakeCipher{},
		ResourcesLimits: slots,
		OT:              testOT(t),
	})
	if err != nil {
		t.Fatalf("NewIDPsService: %v", err)
	}

	return svc
}

// typesNeverAsked is the provider types' repository of a test that changes
// neither a provider's kind nor its issuer.
type typesNeverAsked struct{ repository.IDPTypes }

const lastDoor = "the last enabled administrator account signs in through it"

func refusedAsTheLastDoor(t *testing.T, err error, action string) {
	t.Helper()

	refusal, ok := errors.AsType[*domain.IDPProtectedError](err)
	if !ok {
		t.Fatalf("got %T (%v), want the refusal the status table answers 403", err, err)
	}

	if want := "identity provider cannot be " + action + ": system/protected resource: " + lastDoor; refusal.Error() != want {
		t.Errorf("refused as %q, want %q", refusal.Error(), want)
	}
}

// Disabling an identity provider takes away every sign-in that goes through
// it, until it is enabled again. It is allowed, unless nobody would be left
// to enable it: the `admin` flag may be held by an account that signs in
// through a provider and has no password, and with the seeded account
// disabled that provider is the last door.
//
// It is refused for everybody: the route is the Administrator role's, and
// the role is not the flag.
func TestAProviderIsNotDisabledUnderTheLastAdministrator(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		enabled      *bool
		stored       bool
		now, without int
		refused      bool
		asked        bool
	}{
		"others_can_still_sign_in":          {enabled: new(false), stored: true, now: 2, without: 1, asked: true},
		"it_is_the_last_door":               {enabled: new(false), stored: true, now: 1, without: 0, asked: true, refused: true},
		"the_last_two_both_come_through_it": {enabled: new(false), stored: true, now: 2, without: 0, asked: true, refused: true},
		"nobody_can_sign_in_as_it_is":       {enabled: new(false), stored: true, now: 0, without: 0, asked: true},
		"it_is_disabled_already":            {enabled: new(false), stored: false, now: 1, without: 0},
		"it_is_enabled":                     {enabled: new(true), stored: false, now: 1, without: 0},
		"enabled_sent_back_as_it_is":        {enabled: new(true), stored: true, now: 1, without: 0},
		"the_body_does_not_say":             {stored: true, now: 1, without: 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo, administrators := &providersStored{enabled: c.stored}, &administratorsCounted{now: c.now, without: c.without}
			id := uuid.NewV7()

			err := providersService(t, repo, administrators, &slotsCounted{}).UpdateByID(t.Context(), &domain.UpdateIDPInput{ID: id, Name: new("renamed"), Enabled: c.enabled})

			if asked := len(administrators.askedWithout) == 1; asked != c.asked {
				t.Errorf("the administrators were counted without the provider: %v, want %v", asked, c.asked)
			}

			if c.asked && administrators.askedWithout[0] != id {
				t.Errorf("counted without %s, want the provider %s", administrators.askedWithout[0], id)
			}

			if c.refused {
				refusedAsTheLastDoor(t, err, "disabled")

				if len(repo.writes) != 0 {
					t.Errorf("the provider was written all the same: %v", repo.writes)
				}

				return
			}

			if err != nil {
				t.Fatalf("refused: %v", err)
			}

			if len(repo.writes) != 1 {
				t.Errorf("writes = %v, want the update", repo.writes)
			}
		})
	}
}

// A delete takes the provider's identities with it. The statement keeps the
// provider for an account whose only way in it is (the 409); an
// administrator account that also has an identity at a disabled provider is
// not kept by that rule and still signs in through this one alone.
func TestAProviderIsNotDeletedUnderTheLastAdministrator(t *testing.T) {
	t.Parallel()

	inUse := &domain.IDPInUseError{UsedBy: "account that has no other way to sign in"}

	for name, c := range map[string]struct {
		keeps        error
		now, without int
		refused      bool
	}{
		"others_can_still_sign_in":    {now: 2, without: 1},
		"it_is_the_last_door":         {now: 1, without: 0, refused: true},
		"nobody_can_sign_in_as_it_is": {now: 0, without: 0},
		// The statement's refusal reaches the caller as what it is, and the
		// provider's slot is not given back for a delete that did not happen.
		"an_account_has_no_other_way_in": {now: 2, without: 2, keeps: inUse},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo, administrators, slots := &providersStored{enabled: true, deleteAnswers: c.keeps}, &administratorsCounted{now: c.now, without: c.without}, &slotsCounted{}
			id := uuid.NewV7()

			err := providersService(t, repo, administrators, slots).DeleteByID(t.Context(), &domain.DeleteIDPInput{ID: id})

			if len(administrators.askedWithout) != 1 || administrators.askedWithout[0] != id {
				t.Errorf("counted without %v, want the provider %s, once", administrators.askedWithout, id)
			}

			switch {
			case c.refused:
				refusedAsTheLastDoor(t, err, "deleted")
			case c.keeps != nil:
				if _, ok := errors.AsType[*domain.IDPInUseError](err); !ok {
					t.Fatalf("got %T (%v), want the statement's refusal, the 409", err, err)
				}
			default:
				if err != nil {
					t.Fatalf("refused: %v", err)
				}

				if len(repo.writes) != 1 || slots.givenBack != 1 {
					t.Errorf("writes = %v, slots given back = %d; want the delete and its slot", repo.writes, slots.givenBack)
				}

				return
			}

			if len(repo.writes) != 0 || slots.givenBack != 0 {
				t.Errorf("writes = %v, slots given back = %d; want neither", repo.writes, slots.givenBack)
			}
		})
	}
}

// A count that fails is not "others can still sign in".
func TestAProvidersGuardDoesNotGuessTheCount(t *testing.T) {
	t.Parallel()

	fault := errors.New("the database is away")
	repo := &providersStored{enabled: true}
	svc := providersService(t, repo, &administratorsCounted{err: fault, now: 2, without: 2}, &slotsCounted{})

	if err := svc.UpdateByID(t.Context(), &domain.UpdateIDPInput{ID: uuid.NewV7(), Enabled: new(false)}); !errors.Is(err, fault) {
		t.Errorf("a disable with the count away answered %v, want the fault", err)
	}

	if err := svc.DeleteByID(t.Context(), &domain.DeleteIDPInput{ID: uuid.NewV7()}); !errors.Is(err, fault) {
		t.Errorf("a delete with the count away answered %v, want the fault", err)
	}

	if len(repo.writes) != 0 {
		t.Errorf("the provider was written all the same: %v", repo.writes)
	}

	// Nor is the second count's failure "nobody can sign in as it is": with
	// none left without the provider, a count that cannot be read is not the
	// zero that lets everything through.
	second := providersService(t, repo, &administratorsCounted{errNow: fault, without: 0}, &slotsCounted{})

	if err := second.UpdateByID(t.Context(), &domain.UpdateIDPInput{ID: uuid.NewV7(), Enabled: new(false)}); !errors.Is(err, fault) {
		t.Errorf("a disable with the second count away answered %v, want the fault", err)
	}

	if err := second.DeleteByID(t.Context(), &domain.DeleteIDPInput{ID: uuid.NewV7()}); !errors.Is(err, fault) {
		t.Errorf("a delete with the second count away answered %v, want the fault", err)
	}

	if len(repo.writes) != 0 {
		t.Errorf("the provider was written all the same: %v", repo.writes)
	}
}

// The guard is a dependency the service does not start without: a provider
// service with no count would disable and delete under anybody.
func TestTheProvidersServiceNeedsItsCount(t *testing.T) {
	t.Parallel()

	_, err := NewIDPsService(IDPsServiceConf{
		Repository:      &providersStored{},
		IDPTypes:        typesNeverAsked{},
		Cipher:          fakeCipher{},
		ResourcesLimits: &slotsCounted{},
		OT:              testOT(t),
	})

	missing, ok := errors.AsType[*domain.InvalidRepositoryError](err)
	if !ok || !strings.Contains(missing.Message, "Administrators") {
		t.Fatalf("got %T (%v), want the service to refuse to start without its Administrators", err, err)
	}
}
