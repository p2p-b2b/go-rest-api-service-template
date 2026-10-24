//go:build unit

package usecase

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/cache"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
)

// dependentCache is a cache that keeps what each entry was read from and,
// as the real cache does, drops an entry with every entry that named it.
type dependentCache struct {
	entries map[string]dependentEntry
}

type dependentEntry struct {
	value any
	deps  []string
}

func newDependentCache() *dependentCache {
	return &dependentCache{entries: map[string]dependentEntry{}}
}

func (c *dependentCache) Get(ctx context.Context, id cache.Identifier, dest any, fetcher cache.Fetcher) error {
	if e, ok := c.entries[id.String()]; ok {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(e.value))

		return nil
	}

	v, deps, err := fetcher(ctx)
	if err != nil {
		return err
	}

	e := dependentEntry{value: v}
	for _, d := range deps {
		e.deps = append(e.deps, d.String())
	}

	c.entries[id.String()] = e
	reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(v))

	return nil
}

func (c *dependentCache) Invalidate(_ context.Context, id cache.Identifier) error {
	c.drop(id.String())

	return nil
}

// drop removes the entry under key and everything that named it as a
// dependency, and what named those, to the end.
func (c *dependentCache) drop(key string) {
	delete(c.entries, key)

	for k, e := range c.entries {
		if slices.Contains(e.deps, key) {
			c.drop(k)
		}
	}
}

func (c *dependentCache) has(key string) bool {
	_, ok := c.entries[key]

	return ok
}

// memberProjects is a projects repository that answers one project to its
// members alone, counts its reads, and deletes it. Any other method panics:
// the embedded port is nil.
type memberProjects struct {
	repository.Projects
	project *domain.Project
	members map[uuid.UUID]bool
	reads   []uuid.UUID
}

func (r *memberProjects) SelectByIDByUserID(_ context.Context, id, userID uuid.UUID) (*domain.Project, error) {
	r.reads = append(r.reads, userID)

	if id != r.project.ID || !r.members[userID] {
		return nil, &domain.ProjectNotFoundError{ID: id}
	}

	copied := *r.project

	return &copied, nil
}

func (r *memberProjects) DeleteByID(context.Context, *domain.DeleteProjectInput) error {
	return nil
}

func (r *memberProjects) UpdateByID(context.Context, *domain.UpdateProjectInput) error {
	return nil
}

func (r *memberProjects) LinkUsers(context.Context, *domain.LinkUsersToProjectInput) error {
	return nil
}

func (r *memberProjects) UnlinkUsers(context.Context, *domain.UnlinkUsersFromProjectInput) error {
	return nil
}

// A project's read is scoped to the user who asks, by the statement: a
// member reads it, anyone else is answered that there is none. Its cached
// copy was stored under the project's key alone, so whoever asked after a
// member was answered the member's copy and the statement was never asked.
// No route reaches that, because the membership check in front of the read
// is not cached. The key does not rely on that.
func TestAProjectsCachedReadIsItsReadersAlone(t *testing.T) {
	t.Parallel()

	project, member, stranger := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

	repo := &memberProjects{project: &domain.Project{ID: project, Name: "theirs"}, members: map[uuid.UUID]bool{member: true}}

	svc, err := NewProjectsService(ProjectsServiceConf{Repository: repo, CacheService: newDependentCache(), ResourcesLimits: slotsGivenBack{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewProjectsService: %v", err)
	}

	for range 2 {
		got, err := svc.GetByIDByUserID(t.Context(), project, member)
		if err != nil || got.Name != "theirs" {
			t.Fatalf("a member: %+v, %v", got, err)
		}
	}

	if len(repo.reads) != 1 {
		t.Fatalf("the member's second read was not answered from the cache: %v", repo.reads)
	}

	got, err := svc.GetByIDByUserID(t.Context(), project, stranger)
	if _, ok := errors.AsType[*domain.ProjectNotFoundError](err); !ok {
		t.Fatalf("somebody who is not a member was answered %+v (%v), want what a project that does not exist is answered", got, err)
	}

	if len(repo.reads) != 2 || repo.reads[1] != stranger {
		t.Errorf("the statement that scopes the read was not asked for the other user: %v", repo.reads)
	}
}

// The project's key drops every user's read of it: its update, delete and
// membership changes invalidate that key, and nothing is stored under it.
// One user going drops that user's read and nothing else.
func TestTheProjectsKeyDropsEveryReadOfIt(t *testing.T) {
	t.Parallel()

	project, alice, bob := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

	fixture := func(t *testing.T) *dependentCache {
		t.Helper()

		repo := &memberProjects{project: &domain.Project{ID: project}, members: map[uuid.UUID]bool{alice: true, bob: true}}
		c := newDependentCache()

		svc, err := NewProjectsService(ProjectsServiceConf{Repository: repo, CacheService: c, ResourcesLimits: slotsGivenBack{}, OT: testOT(t)})
		if err != nil {
			t.Fatalf("NewProjectsService: %v", err)
		}

		for _, who := range []uuid.UUID{alice, bob} {
			if _, err := svc.GetByIDByUserID(t.Context(), project, who); err != nil {
				t.Fatalf("the read that warms the cache: %v", err)
			}
		}

		return c
	}

	t.Run("a_user_going_drops_that_users_read_and_nothing_else", func(t *testing.T) {
		t.Parallel()

		c := fixture(t)

		if err := c.Invalidate(t.Context(), cache.Identifier{Type: "user", ID: alice.String()}); err != nil {
			t.Fatalf("Invalidate: %v", err)
		}

		if c.has(projectReadCacheKey(alice, project).String()) {
			t.Error("the read of a user who is gone is still cached")
		}

		if !c.has(projectReadCacheKey(bob, project).String()) {
			t.Error("one user going dropped another user's read of the project")
		}
	})

	t.Run("the_projects_key_drops_every_read_of_it", func(t *testing.T) {
		t.Parallel()

		c := fixture(t)

		if err := c.Invalidate(t.Context(), projectCacheKey(project)); err != nil {
			t.Fatalf("Invalidate: %v", err)
		}

		for _, who := range []uuid.UUID{alice, bob} {
			if c.has(projectReadCacheKey(who, project).String()) {
				t.Errorf("a member's read outlived the project's key")
			}
		}
	})
}

// Each change of a project drops every user's read of it, through the service:
// the update and both membership changes invalidate the project's key, and
// the reads depend on it.
func TestEveryChangeOfAProjectDropsItsReads(t *testing.T) {
	t.Parallel()

	project, alice, bob := uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

	for name, change := range map[string]func(*ProjectsService) error{
		"update": func(svc *ProjectsService) error {
			return svc.UpdateByID(t.Context(), &domain.UpdateProjectInput{ID: project, UserID: alice, Description: new("changed")})
		},
		"link": func(svc *ProjectsService) error {
			return svc.LinkUsers(t.Context(), &domain.LinkUsersToProjectInput{ProjectID: project, UserIDs: []uuid.UUID{uuid.NewV7()}})
		},
		"unlink": func(svc *ProjectsService) error {
			return svc.UnlinkUsers(t.Context(), &domain.UnlinkUsersFromProjectInput{ProjectID: project, UserIDs: []uuid.UUID{bob}})
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &memberProjects{project: &domain.Project{ID: project, Name: "theirs"}, members: map[uuid.UUID]bool{alice: true, bob: true}}
			c := newDependentCache()

			svc, err := NewProjectsService(ProjectsServiceConf{Repository: repo, CacheService: c, ResourcesLimits: slotsGivenBack{}, OT: testOT(t)})
			if err != nil {
				t.Fatalf("NewProjectsService: %v", err)
			}

			for _, who := range []uuid.UUID{alice, bob} {
				if _, err := svc.GetByIDByUserID(t.Context(), project, who); err != nil {
					t.Fatalf("the read that warms the cache: %v", err)
				}
			}

			if err := change(svc); err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			for _, who := range []uuid.UUID{alice, bob} {
				if c.has(projectReadCacheKey(who, project).String()) {
					t.Errorf("a member's read outlived the project's %s", name)
				}
			}
		})
	}
}

// decrementFails is the usage limits of a delete whose decrement fails after
// the row is gone: a fault of the database, or of the counter's signature.
type decrementFails struct{ ResourcesLimitsServiceConsumer }

func (decrementFails) DecrementUsage(context.Context, domain.ResourcesLimitsScope, domain.ResourcesLimitsResourceType) error {
	return errors.New("the counter's signature does not verify")
}

// A delete whose decrement fails returns that error, and the row is gone:
// what is cached of it was dropped first. It was dropped after the
// decrement, so the deleted project was answered from the cache for the
// cache's lifetime.
func TestADeletedProjectIsNotCachedWhenItsDecrementFails(t *testing.T) {
	t.Parallel()

	project, member := uuid.NewV7(), uuid.NewV7()

	repo := &memberProjects{project: &domain.Project{ID: project}, members: map[uuid.UUID]bool{member: true}}
	c := newDependentCache()

	svc, err := NewProjectsService(ProjectsServiceConf{Repository: repo, CacheService: c, ResourcesLimits: decrementFails{}, OT: testOT(t)})
	if err != nil {
		t.Fatalf("NewProjectsService: %v", err)
	}

	if _, err := svc.GetByIDByUserID(t.Context(), project, member); err != nil {
		t.Fatalf("the read that warms the cache: %v", err)
	}

	if err := svc.DeleteByID(t.Context(), &domain.DeleteProjectInput{ID: project, UserID: member}); err == nil {
		t.Fatal("the decrement's failure was not the call's")
	}

	if c.has(projectReadCacheKey(member, project).String()) {
		t.Error("the deleted project is still cached")
	}
}
