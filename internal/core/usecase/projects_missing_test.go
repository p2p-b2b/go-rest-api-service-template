//go:build unit

package usecase

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// TestAMembershipOfTheNilIDIsAnInvalidInput: the Nil uuid parses, so the
// membership middleware asks about it, and the answer goes through the status
// table. It was an InvalidProjectIDError, which has no row there, so
// `/projects/00000000-0000-0000-0000-000000000000/...` was a 500. An invalid
// input is the table's 400. The repository is not asked.
func TestAMembershipOfTheNilIDIsAnInvalidInput(t *testing.T) {
	t.Parallel()

	svc, err := NewProjectsService(ProjectsServiceConf{Repository: &listingProjects{}, ResourcesLimits: noLimits{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewProjectsService: %v", err)
	}

	for name, ids := range map[string][2]uuid.UUID{
		"the_project": {uuid.Nil(), uuid.NewV7()},
		"the_caller":  {uuid.NewV7(), uuid.Nil()},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			membership, err := svc.Membership(t.Context(), ids[0], ids[1])

			if _, ok := errors.AsType[*domain.InvalidInputError](err); !ok {
				t.Fatalf("got %T (%v), want an invalid input", err, err)
			}

			if membership != domain.ProjectMembershipNone {
				t.Errorf("membership %v, want none", membership)
			}
		})
	}
}

// deletingProjects is the projects repository for a delete that matched no
// row. Any other method panics: the embedded port is nil.
type deletingProjects struct{ repository.Projects }

func (deletingProjects) DeleteByID(_ context.Context, in *domain.DeleteProjectInput) error {
	return &domain.ProjectNotFoundError{ID: in.ID}
}

// countingLimits counts the usage given back.
type countingLimits struct {
	ResourcesLimitsServiceConsumer
	decrements int
}

func (c *countingLimits) DecrementUsage(context.Context, domain.ResourcesLimitsScope, domain.ResourcesLimitsResourceType) error {
	c.decrements++
	return nil
}

// TestDeletingAMissingProjectIsANotFoundAndGivesNothingBack: nothing was
// removed, so nothing is given back, or a caller could lower their own usage
// by deleting ids that never existed. The delete answered success for it, as
// if it had deleted; it is the 404 the handler answers, as for every delete
// by id. The membership middleware answers a missing project first, so this
// is the project deleted between that check and the statement.
func TestDeletingAMissingProjectIsANotFoundAndGivesNothingBack(t *testing.T) {
	t.Parallel()

	limits := &countingLimits{}

	svc, err := NewProjectsService(ProjectsServiceConf{Repository: deletingProjects{}, ResourcesLimits: limits, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewProjectsService: %v", err)
	}

	err = svc.DeleteByID(t.Context(), &domain.DeleteProjectInput{ID: uuid.NewV7(), UserID: uuid.NewV7()})

	if _, ok := errors.AsType[*domain.ProjectNotFoundError](err); !ok {
		t.Fatalf("got %T (%v), want the project's not found", err, err)
	}

	if limits.decrements != 0 {
		t.Errorf("usage given back %d times for a project that was not deleted", limits.decrements)
	}
}
