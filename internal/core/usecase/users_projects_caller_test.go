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

// linkingUsers is the users repository for the two changes of a user's
// project links. It records the two accounts as they are when the repository
// is called, and answers the refusal it holds. Any other method panics: the
// embedded port is nil.
type linkingUsers struct {
	repository.Users
	refusal      error
	user, caller uuid.UUID
	asked        int
}

func (l *linkingUsers) LinkProjects(_ context.Context, in *domain.LinkProjectsToUserInput) error {
	l.asked++
	l.user, l.caller = in.UserID, in.CallerID

	return l.refusal
}

func (l *linkingUsers) UnlinkProjects(_ context.Context, in *domain.UnlinkProjectsFromUserInput) error {
	l.asked++
	l.user, l.caller = in.UserID, in.CallerID

	return l.refusal
}

// The repository holds a change of a user's project links to the projects of
// whoever asks, so the caller has to reach it as the handler gave it, beside
// the path's user. And its refusal has to reach the handler as what it is:
// the handler answers by the error's type, and a refusal passed on as a
// plain error is a 500.
func TestAUsersProjectLinksReachTheRepositoryWithTheirCaller(t *testing.T) {
	t.Parallel()

	for _, direction := range []string{"link", "unlink"} {
		t.Run(direction, func(t *testing.T) {
			t.Parallel()

			user, caller, project := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()
			foreign := uuid.NewV7()

			change := func(repo *linkingUsers) error {
				svc, err := NewUsersService(UsersServiceConf{Guard: &GrantGuard{}, Repository: repo, CacheService: &droppingCache{}, ResourcesLimits: noLimits{}, OT: testOT(t)})
				if err != nil {
					t.Fatalf("NewUsersService: %v", err)
				}

				input := &domain.LinkProjectsToUserInput{UserID: user, CallerID: caller, ProjectIDs: []uuid.UUID{project}}

				if direction == "link" {
					return svc.LinkProjects(t.Context(), input)
				}

				return svc.UnlinkProjects(t.Context(), input)
			}

			repo := &linkingUsers{}
			if err := change(repo); err != nil {
				t.Fatalf("refused: %v", err)
			}

			if repo.asked != 1 || repo.user != user {
				t.Fatalf("the repository was asked %d times, for user %s; want once, for %s", repo.asked, repo.user, user)
			}

			if repo.caller != caller {
				t.Errorf("the repository heard caller %s, want %s: the projects are held to who asks", repo.caller, caller)
			}

			refusing := &linkingUsers{refusal: &domain.ProjectNotOfCallerError{ProjectID: foreign}}

			refusal, ok := errors.AsType[*domain.ProjectNotOfCallerError](change(refusing))
			if !ok || refusal.ProjectID != foreign {
				t.Errorf("the refusal reached the handler as %v, want the repository's, with its project", refusal)
			}
		})
	}
}

// listingProjects is the projects repository for the list of a user's
// projects, and records the two accounts as they are when it is called.
type listingProjects struct {
	repository.Projects
	user, caller uuid.UUID
	asked        int
}

func (l *listingProjects) SelectByUserID(_ context.Context, in *domain.SelectProjectsInput) (*domain.SelectProjectsOutput, error) {
	l.asked++
	l.user, l.caller = in.UserID, in.CallerID

	return &domain.SelectProjectsOutput{}, nil
}

// The list of a user's projects names two accounts, and the repository reads
// the first one's projects that are the second one's too: each has to reach
// it as what it is.
func TestAUsersProjectsReachTheRepositoryWithTheirCaller(t *testing.T) {
	t.Parallel()

	repo := &listingProjects{}

	svc, err := NewProjectsService(ProjectsServiceConf{Repository: repo, ResourcesLimits: noLimits{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewProjectsService: %v", err)
	}

	user, caller := uuid.NewV7(), uuid.NewV7()

	input := &domain.ListProjectsInput{UserID: user, CallerID: caller, Paginator: domain.Paginator{Limit: 10}}

	if _, err := svc.ListByUserID(t.Context(), input); err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}

	if repo.asked != 1 || repo.user != user {
		t.Fatalf("the repository was asked %d times, for user %s; want once, for %s", repo.asked, repo.user, user)
	}

	if repo.caller != caller {
		t.Errorf("the repository heard caller %s, want %s", repo.caller, caller)
	}
}
