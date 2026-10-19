package domain

import (
	"uuid"
)

// InvalidIdentityProvidersError represents an error with identity providers configuration
type InvalidIdentityProvidersError struct {
	Provider string // optional: which provider
	Message  string
}

func (e *InvalidIdentityProvidersError) Error() string {
	return (&BaseInvalidFieldError{
		Field:  "identity provider",
		Value:  e.Provider,
		Reason: e.Message,
	}).Error()
}

// IDPNotFoundError represents an error when an identity provider cannot be found
type IDPNotFoundError struct {
	Name    string // optional: search by name
	Message string // optional: additional context
	ID      uuid.UUID
}

func (e *IDPNotFoundError) Error() string {
	base := &BaseNotFoundError{
		Entity:  "identity provider",
		ID:      e.ID,
		Message: e.Message,
	}

	if e.Name != "" {
		base.Field = "name"
		base.Value = e.Name
	}

	return base.Error()
}

// IDPAlreadyExistsError represents an error when an identity provider already exists
// Consolidates IDPNameAlreadyExistsError
type IDPAlreadyExistsError struct {
	Name    string // which field conflicts
	Message string // optional: additional context
	ID      uuid.UUID
}

func (e *IDPAlreadyExistsError) Error() string {
	base := &BaseAlreadyExistsError{
		Entity:  "identity provider",
		ID:      e.ID,
		Message: e.Message,
	}

	if e.Name != "" {
		base.Field = "name"
		base.Value = e.Name
	}

	return base.Error()
}

// InvalidIDPIDError represents an invalid identity provider ID error
type InvalidIDPIDError struct {
	Message string
	ID      uuid.UUID
}

func (e *InvalidIDPIDError) Error() string {
	return (&BaseInvalidFieldError{
		Entity: "identity provider",
		Field:  "ID",
		Value:  e.ID.String(),
		Reason: e.Message,
	}).Error()
}

// InvalidIDPUpdateError represents an error during identity provider update
type InvalidIDPUpdateError struct {
	Field   string // optional: which field
	Reason  string // optional: why update failed
	Message string // optional: additional context
}

func (e *InvalidIDPUpdateError) Error() string {
	reason := e.Reason
	if reason == "" {
		reason = e.Message
	}

	if e.Field != "" {
		reason = "field '" + e.Field + "': " + reason
	}

	return (&BaseInvalidUpdateError{
		Entity: "identity provider",
		Reason: reason,
	}).Error()
}

// IDPInUseError refuses the delete of an identity provider that something
// depends on to work: an account whose only way to sign in is an identity at
// it. The delete took those identities with it, by foreign key, and left the
// accounts with no way in at all.
type IDPInUseError struct {
	UsedBy string // what depends on it, in the singular
	// Action is what was refused, when it is not the delete: the provider
	// pointed at another issuer or type under its identities.
	Action string
	ID     uuid.UUID
}

func (e *IDPInUseError) Error() string {
	action := e.Action
	if action == "" {
		action = "deleted"
	}

	return "identity provider cannot be " + action + ": still used by at least one " + e.UsedBy
}

// IDPProtectedError refuses to disable or delete an identity provider when
// that would leave no administrator account that can sign in.
type IDPProtectedError struct {
	Action  string // "disabled" or "deleted"
	Message string // which rule refused it
	ID      uuid.UUID
}

func (e *IDPProtectedError) Error() string {
	return (&BaseProtectedError{Entity: "identity provider", Action: e.Action, Message: e.Message}).Error()
}
