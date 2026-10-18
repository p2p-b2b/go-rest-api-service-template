package domain

import (
	"uuid"
)

// UserNotFoundError represents an error when a user cannot be found
type UserNotFoundError struct {
	Email   string // optional: search by email
	Message string // optional: additional context
	ID      uuid.UUID
}

func (e *UserNotFoundError) Error() string {
	base := &BaseNotFoundError{
		Entity:  "user",
		ID:      e.ID,
		Message: e.Message,
	}

	if e.Email != "" {
		base.Field = "email"
		base.Value = e.Email
	}

	return base.Error()
}

// UserAlreadyExistsError represents an error when a user already exists
// Consolidates UserExistsError, UserAlreadyExistsError, EmailExistsError, UserEmailAlreadyExistsError
type UserAlreadyExistsError struct {
	Email   string // which field conflicts
	Message string // optional: additional context
	ID      uuid.UUID
}

func (e *UserAlreadyExistsError) Error() string {
	base := &BaseAlreadyExistsError{
		Entity:  "user",
		ID:      e.ID,
		Message: e.Message,
	}

	if e.Email != "" {
		base.Field = "email"
		base.Value = e.Email
	}

	return base.Error()
}

// InvalidUserIDError represents an invalid user ID error
type InvalidUserIDError struct {
	Message string
	ID      uuid.UUID
}

func (e *InvalidUserIDError) Error() string {
	return (&BaseInvalidFieldError{
		Entity: "user",
		Field:  "ID",
		Value:  e.ID.String(),
		Reason: e.Message,
	}).Error()
}

// UserDisabledError represents an error when trying to use a disabled user account
type UserDisabledError struct {
	Username string
	UserID   uuid.UUID
}

func (e *UserDisabledError) Error() string {
	base := &BaseInvalidFieldError{
		Entity: "user",
		Field:  "status",
		Reason: "user is disabled",
	}

	if e.Username != "" {
		base.Value = e.Username
	}

	if e.UserID != uuid.Nil() {
		base.Value = e.UserID.String()
	}

	return base.Error()
}

// InvalidEmailError represents an invalid email error
type InvalidEmailError struct {
	Email   string
	Reason  string // optional: why it's invalid
	Message string // optional: additional context
}

func (e *InvalidEmailError) Error() string {
	return (&BaseInvalidFieldError{
		Field:  "email",
		Value:  e.Email,
		Reason: e.Reason,
	}).Error()
}

// InvalidPasswordError represents an invalid password error
type InvalidPasswordError struct {
	Message   string
	MinLength int
	MaxLength int
}

func (e *InvalidPasswordError) Error() string {
	return (&BaseInvalidFieldError{
		Field:     "password",
		Reason:    e.Message,
		MinLength: e.MinLength,
		MaxLength: e.MaxLength,
	}).Error()
}

// InvalidUserPasswordError represents an invalid user password error (for updates/changes)
type InvalidUserPasswordError struct {
	Reason  string // optional: why password is invalid
	Message string // optional: additional context
}

func (e *InvalidUserPasswordError) Error() string {
	return (&BaseInvalidFieldError{
		Entity: "user",
		Field:  "password",
		Reason: e.Reason,
	}).Error()
}

// InvalidProjectLinkError represents an invalid project link error
type InvalidProjectLinkError struct {
	Message   string
	ProjectID uuid.UUID
	UserID    uuid.UUID
}

func (e *InvalidProjectLinkError) Error() string {
	base := &BaseInvalidFieldError{
		Field:  "projectLink",
		Reason: e.Message,
	}

	if e.ProjectID != uuid.Nil() {
		base.Value = e.ProjectID.String()
	}

	if e.UserID != uuid.Nil() && e.ProjectID == uuid.Nil() {
		base.Value = e.UserID.String()
	}

	return base.Error()
}

// ProjectNotOfCallerError refuses a change to a user's project links that
// names a project the caller is not linked to. A project of another account
// and one that does not exist are refused in the same words, with the id the
// caller sent: the caller is not told which of the two it is.
type ProjectNotOfCallerError struct {
	// ProjectID is the id that is not the caller's; nil when the write found
	// it gone and cannot say which of the ids it was.
	ProjectID uuid.UUID
}

func (e *ProjectNotOfCallerError) Error() string {
	if e.ProjectID == uuid.Nil() {
		return "invalid project_ids: one of the ids is not one of your projects"
	}

	return "invalid project_ids: " + e.ProjectID.String() + " is not one of your projects"
}

// AdministratorProtectedError refuses what would take an administrator
// account out of its holder's hands: disabling or enabling it, making it a
// federated or a local account, taking a role from it, deleting it. An
// administrator account is the one with the `admin` flag, which is not the
// Administrator role.
//
// It is a 403: the caller is who they say they are and holds the route's
// grant, and this account is still not theirs to lock or unlock. Message
// says which rule it was ("the administrator account", for a caller that is
// not one; "the last enabled administrator account", for everybody).
type AdministratorProtectedError struct {
	Action  string // what was attempted: "disabled", "deleted", ...
	Message string // which rule refused it
}

func (e *AdministratorProtectedError) Error() string {
	return (&BaseProtectedError{Entity: "user", Action: e.Action, Message: e.Message}).Error()
}
