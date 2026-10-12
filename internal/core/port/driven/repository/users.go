package repository

import (
	"context"
	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

//go:generate go tool mockgen -package=mocks -destination=../../../../../mocks/service/users.go -source=users.go Users

// Users is the driven persistence port for user entities.
type Users interface {
	Insert(ctx context.Context, input *domain.InsertUserInput) error
	UpdateByID(ctx context.Context, input *domain.UpdateUserInput) error

	// Verify records that address is proven to be the account's and opens
	// the account, unless an administrator disabled it or it has the `admin`
	// flag, in one statement. No unverified account with that id and that
	// address is a NotFoundError.
	Verify(ctx context.Context, id uuid.UUID, address string) error
	DeleteByID(ctx context.Context, input *domain.DeleteUserInput) error

	SelectByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	SelectByEmail(ctx context.Context, email string) (*domain.User, error)
	SelectByRoleID(ctx context.Context, roleID uuid.UUID, input *domain.ListUsersInput) (*domain.ListUsersOutput, error)
	SelectByProjectID(ctx context.Context, projectID uuid.UUID, input *domain.SelectUsersInput) (*domain.SelectUsersOutput, error)
	Select(ctx context.Context, input *domain.SelectUsersInput) (*domain.SelectUsersOutput, error)

	SelectAuthz(ctx context.Context, userID uuid.UUID) (*domain.SelectAuthzOutput, error)

	LinkRoles(ctx context.Context, input *domain.LinkRolesToUserInput) error
	UnlinkRoles(ctx context.Context, input *domain.UnlinkRolesFromUsersInput) error

	LinkProjects(ctx context.Context, input *domain.LinkProjectsToUserInput) error
	UnlinkProjects(ctx context.Context, input *domain.UnlinkProjectsFromUserInput) error
}
