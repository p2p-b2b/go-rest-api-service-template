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
	// CountEnabledAdministrators is how many accounts with the `admin` flag
	// may still sign in: enabled, and with a password login or an identity at
	// a provider that is enabled.
	CountEnabledAdministrators(ctx context.Context) (int, error)
	// IsEnabledAdministrator is whether the account is one of those
	// CountEnabledAdministrators counts.
	IsEnabledAdministrator(ctx context.Context, id uuid.UUID) (bool, error)
	// SelectAdministratorsAmong answers which of ids are accounts with the
	// `admin` flag.
	SelectAdministratorsAmong(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)
	SelectByRoleID(ctx context.Context, roleID uuid.UUID, input *domain.ListUsersInput) (*domain.ListUsersOutput, error)
	SelectByProjectID(ctx context.Context, projectID uuid.UUID, input *domain.SelectUsersInput) (*domain.SelectUsersOutput, error)
	Select(ctx context.Context, input *domain.SelectUsersInput) (*domain.SelectUsersOutput, error)

	SelectAuthz(ctx context.Context, userID uuid.UUID) (*domain.SelectAuthzOutput, error)

	LinkRoles(ctx context.Context, input *domain.LinkRolesToUserInput) error
	UnlinkRoles(ctx context.Context, input *domain.UnlinkRolesFromUsersInput) error

	LinkProjects(ctx context.Context, input *domain.LinkProjectsToUserInput) error
	UnlinkProjects(ctx context.Context, input *domain.UnlinkProjectsFromUserInput) error
}
