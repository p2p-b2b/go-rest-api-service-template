package domain

import (
	"time"

	"uuid"
)

// UserIdentity links one provider identity to one account.
//
// The identity is the provider's stable subject -- OIDC `sub`, GitHub's
// numeric user id -- never the email. An email is a mutable attribute that an
// Entra admin, a GitHub user or the account holder can change, and matching
// accounts on it let anyone who controlled an email at ONE provider sign in as
// the account behind it. The email is kept only as the value seen at link
// time, for the profile page.
type UserIdentity struct {
	LinkedAt time.Time
	Subject  string
	Email    string

	// Provider name and type, filled by the listing for the profile page.
	IDPName     string
	IDPTypeName string

	// IDPEnabled says whether the provider is enabled: an identity at a
	// provider that is disabled is not a way into the account.
	IDPEnabled bool

	UserID uuid.UUID
	IDPID  uuid.UUID
}

// LinkUserIdentityInput is what a link writes.
type LinkUserIdentityInput struct {
	Subject string
	Email   string
	// IssuerURL and IDPTypeID are the provider as the sign-in used it: the
	// issuer whose token proved the subject, and its kind. The identity is
	// stored only while the provider's row still says the same, so a sign-in
	// through a provider that was pointed elsewhere meanwhile, or read from
	// a stale cached copy, stores nothing. IssuerURL is "" for a kind with
	// no issuer (github).
	IssuerURL string
	UserID    uuid.UUID
	IDPID     uuid.UUID
	IDPTypeID uuid.UUID
}

func (ref *LinkUserIdentityInput) Validate() error {
	var errs ValidationErrors

	errs.Add(ValidateUUID(ref.UserID, 7, FieldUserID))
	errs.Add(ValidateUUID(ref.IDPID, 7, FieldIDPID))
	errs.Add(ValidateUUID(ref.IDPTypeID, 7, FieldIDPTypeID))

	if ref.Subject == "" {
		errs.AddError(FieldSubject, "the provider's subject is required", "REQUIRED")
	}

	if len(ref.Subject) > ValidIDPClientIDMaxLength {
		errs.AddError(FieldSubject, "the provider's subject is too long", "TOO_LONG")
	}

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

// UnlinkUserIdentityInput removes one provider from one account.
type UnlinkUserIdentityInput struct {
	UserID uuid.UUID
	IDPID  uuid.UUID
}

func (ref *UnlinkUserIdentityInput) Validate() error {
	var errs ValidationErrors

	errs.Add(ValidateUUID(ref.UserID, 7, FieldUserID))
	errs.Add(ValidateUUID(ref.IDPID, 7, FieldIDPID))

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

// UserIdentityNotFoundError: the (idp, subject) or (user, idp) pair has no row.
// The ordinary case for a first sign-in; an error only because the caller has
// to branch on it.
type UserIdentityNotFoundError struct {
	Message string
}

func (e *UserIdentityNotFoundError) Error() string {
	if e.Message == "" {
		return "no account is linked to this provider identity"
	}

	return e.Message
}

// UnlinkOnlyWayInMessage is the refusal of an unlink that would leave an
// account with no way in: no password, and no identity at another enabled
// provider. The use-case says it first; the unlink's statement, which reads
// the account's identities under its row's lock, says it when an unlink or
// a provider's delete took the other way in meanwhile.
const UnlinkOnlyWayInMessage = "this is the only way into the account; set a password or link another provider first"

// UserIdentityAlreadyLinkedError: the provider identity belongs to another
// account, or this account already has an identity at that provider.
type UserIdentityAlreadyLinkedError struct {
	Message string
}

func (e *UserIdentityAlreadyLinkedError) Error() string {
	if e.Message == "" {
		return "this provider identity is already linked to an account"
	}

	return e.Message
}

// IDPIdentityNotLinkedError is what an IdP sign-in answers when the identity
// is unknown and cannot be provisioned. One wording whatever the reason -- an
// existing account with that email, a provider that does not vouch for the
// email, auto-provisioning switched off -- because the differences would tell
// whoever controls the provider account which addresses have accounts here.
type IDPIdentityNotLinkedError struct {
	Detail string // the operator's: which of the reasons it was, see FailureDetail
}

func (e *IDPIdentityNotLinkedError) Error() string {
	return "this identity is not linked to an account here; sign in with your password and link the provider from your profile, or ask an administrator"
}

// IDPChangedDuringSignInError refuses an identity whose provider was pointed
// at another issuer or type, or deleted, while the sign-in ran: the subject
// was proven by the provider as it was, and is not stored under the
// provider as it is. Nothing of the sign-in is stored; signing in again
// reads the provider as it is now.
type IDPChangedDuringSignInError struct{}

func (e *IDPChangedDuringSignInError) Error() string {
	return "the identity provider changed while signing in; sign in again"
}
