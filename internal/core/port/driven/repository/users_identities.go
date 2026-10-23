package repository

import (
	"context"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

//go:generate go tool mockgen -package=mocks -destination=../../../../../mocks/service/users_identities.go -source=users_identities.go UsersIdentities

// UsersIdentities is the driven persistence port for the link between an
// account and a provider identity.
//
// The identity is (idp, subject). Absence means "unknown identity", which is
// the ordinary answer for a first sign-in and is returned as
// *domain.UserIdentityNotFoundError so the caller can branch on it. An error
// from the store is a fault, never "unknown": treating an unreachable table as
// an empty one would turn every sign-in into a refused one -- or, worse, into a
// provisioned duplicate.
type UsersIdentities interface {
	// Link records that subject at idp belongs to user. A subject already
	// linked to another account, or a user already linked at that idp, is a
	// *domain.UserIdentityAlreadyLinkedError.
	Link(ctx context.Context, input *domain.LinkUserIdentityInput) error

	// Unlink removes the account's identity at idp. Not found is an error:
	// the caller asked to remove something that was not there.
	Unlink(ctx context.Context, input *domain.UnlinkUserIdentityInput) error

	// SelectBySubject answers "whose identity is this", under the provider
	// as the sign-in used it: its kind and its issuer ("" for a kind with
	// none). An identity under a provider that was pointed at another
	// issuer or type since the sign-in read it is not found: the subject was
	// proven by an issuer that no longer vouches for it.
	SelectBySubject(ctx context.Context, idpID, idpTypeID uuid.UUID, issuerURL, subject string) (*domain.UserIdentity, error)

	// SelectByUserID lists an account's identities, for the profile page.
	SelectByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserIdentity, error)
}
