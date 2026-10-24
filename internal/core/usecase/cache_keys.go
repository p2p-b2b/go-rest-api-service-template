package usecase

import (
	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/cache"
)

// authzCacheKey is the one key a user's effective permissions live under.
// Eight call sites used to spell it out by hand, which is exactly the drift
// a single helper exists to prevent; the AuthzServiceCache interface that was
// meant to own it was declared and implemented by nothing.
func authzCacheKey(userID uuid.UUID) cache.Identifier {
	return cache.Identifier{Type: "authz", ID: userID.String()}
}

// projectCacheKey is the key a project's update, delete and membership
// changes invalidate: the dependency of everything read from the project.
// Nothing is stored under it.
func projectCacheKey(projectID uuid.UUID) cache.Identifier {
	return cache.Identifier{Type: "project", ID: projectID.String()}
}

// projectReadCacheKey is the key a project's read for one user is stored
// under. The read is scoped to the user by its statement (a member reads the
// project, anyone else is answered that there is none), and a cache hit
// bypasses the statement: under the project's key alone, whoever asked after
// a member was answered the member's copy. No route reaches that, because
// the membership check in front of it is not cached; the key does not rely
// on that.
func projectReadCacheKey(userID, projectID uuid.UUID) cache.Identifier {
	return cache.Identifier{Type: "project", ID: userID.String() + ":" + projectID.String()}
}
