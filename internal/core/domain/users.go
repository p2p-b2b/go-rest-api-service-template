package domain

import (
	"fmt"
	"reflect"
	"time"

	"uuid"
)

const (
	UsersUserCreatedSuccessfully          = "User created successfully"
	UsersUserUpdatedSuccessfully          = "User updated successfully"
	UsersUserDeletedSuccessfully          = "User deleted successfully"
	UsersRoleLinkedToUserSuccessfully     = "User role linked successfully"
	UsersRoleUnlinkedFromUserSuccessfully = "User role unlinked successfully"
	UsersUserFound                        = "User found"

	UsersProjectsLinkedToUserSuccessfully     = "Projects linked to user successfully"
	UsersProjectsUnlinkedFromUserSuccessfully = "Projects unlinked from user successfully"
)

var (
	UsersFilterFields  = []string{FieldID, FieldFirstName, FieldLastName, FieldEmail, FieldDisabled, FieldDisabledByAdmin, FieldVerified, FieldLocalAccount, FieldCreatedAt, FieldUpdatedAt}
	UsersSortFields    = []string{FieldID, FieldFirstName, FieldLastName, FieldEmail, FieldDisabled, FieldDisabledByAdmin, FieldVerified, FieldLocalAccount, FieldCreatedAt, FieldUpdatedAt}
	UsersPartialFields = []string{FieldID, FieldFirstName, FieldLastName, FieldEmail, FieldDisabled, FieldDisabledByAdmin, FieldVerified, FieldLocalAccount, FieldCreatedAt, FieldUpdatedAt}
)

// User is the user-account entity.
//
// Disabled and Verified are two facts, not one. Disabled says whether the
// account may sign in and is what an administrator switches. Verified says
// whether the address has been proven by following the verification link (or
// by an identity provider, which proves its subject, or by the administrator
// who created the account and vouches for it). A registration is unverified
// and disabled; verification flips both, unless an administrator disabled
// the account meanwhile (DisabledByAdmin). Those three are the only writers
// of Verified, and no update sets it. They used to be one flag, and password
// recovery could not tell an account that had never followed its link from
// one that had been switched off, so it refused both.
//
// DisabledByAdmin is the administrator's lock: an account that is disabled
// because an administrator said so, and not because its registration is
// waiting for its link. A verification link does not open such an account.
// It is read and written through the API as `disabled_by_admin`.
type User struct {
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Disabled        *bool
	DisabledByAdmin *bool
	Verified        *bool
	Admin           *bool
	LocalAccount    *bool
	FirstName       string
	LastName        string
	Email           string
	Password        string
	PasswordHash    string
	SerialID        int64
	ID              uuid.UUID
}

// AccountStanding says whether an account may act on a request it is
// authenticated for: it is there, and it is not disabled.
type AccountStanding int

const (
	// AccountUnknown is the zero value, and it is no answer: whatever
	// returns it without having asked (a failed read, a double that was
	// told nothing) must not be read as "may act".
	AccountUnknown AccountStanding = iota
	// AccountActive: the account exists and may act.
	AccountActive
	// AccountDisabled: the account exists and an administrator, or a
	// registration not yet verified, keeps it from acting.
	AccountDisabled
	// AccountMissing: there is no such account any more.
	AccountMissing
)

type InsertUserInput struct {
	Disabled     *bool
	Verified     *bool
	LocalAccount *bool
	// Identity is the provider identity an account is provisioned from, on
	// a first sign-in: stored with the account, in the same transaction, so
	// that neither is ever there without the other. Nil for every other
	// account. Its UserID is the account's ID.
	Identity     *LinkUserIdentityInput
	FirstName    string
	LastName     string
	Email        string
	Password     string
	PasswordHash string
	ID           uuid.UUID
	// DisabledByAdmin says that Disabled is an administrator's doing and
	// not a registration waiting for its link: POST /users, which stores
	// the account disabled for an administrator to enable. A verification
	// link does not open such a row.
	DisabledByAdmin bool
}

func (ref *InsertUserInput) Validate() error {
	var errs ValidationErrors

	errs.Add(ValidateUUID(ref.ID, 7, FieldID))

	if normalizedFirstName, err := ValidateString(ref.FirstName, StringValidationOptions{
		MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength,
		TrimWhitespace: true, AllowEmpty: false, NoControlChars: true, NoHTMLTags: true,
		NoNullBytes: true, NormalizeUnicode: true, FieldName: FieldFirstName,
	}); err != nil {
		errs.Add(err)
	} else {
		ref.FirstName = normalizedFirstName
	}

	if normalizedLastName, err := ValidateString(ref.LastName, StringValidationOptions{
		MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength,
		TrimWhitespace: true, AllowEmpty: false, NoControlChars: true, NoHTMLTags: true,
		NoNullBytes: true, NormalizeUnicode: true, FieldName: FieldLastName,
	}); err != nil {
		errs.Add(err)
	} else {
		ref.LastName = normalizedLastName
	}

	if normalizedEmail, err := ValidateEmail(ref.Email, FieldEmail); err != nil {
		errs.Add(err)
	} else {
		ref.Email = normalizedEmail
	}

	if ref.PasswordHash != "" {
		if _, err := ValidateString(ref.PasswordHash, StringValidationOptions{
			MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength,
			TrimWhitespace: false, AllowEmpty: false, NoControlChars: true, NoNullBytes: true,
			FieldName: FieldPasswordHash,
		}); err != nil {
			errs.Add(err)
		}
	}

	if ref.Password != "" {
		errs.Add(ValidatePassword(ref.Password, FieldPassword))
	}

	if ref.LocalAccount == nil {
		errs.AddError("local_account", "local_account cannot be nil", "REQUIRED")
	}

	if ref.Identity != nil {
		if err := ref.Identity.Validate(); err != nil {
			errs.Add(err)
		} else if ref.Identity.UserID != ref.ID {
			errs.AddError(FieldUserID, "the identity is not the account's", "INVALID_VALUE")
		}

		// An identity is how a federated account signs in: a local account
		// is never provisioned with one, so that no route that creates a
		// local account can link a subject on the way.
		if ref.LocalAccount != nil && *ref.LocalAccount {
			errs.AddError("local_account", "a local account is not provisioned with a provider identity", "INVALID_COMBINATION")
		}
	}

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

type CreateUserInput = InsertUserInput

type UpdateUserInput struct {
	FirstName    *string
	LastName     *string
	Email        *string
	Password     *string
	PasswordHash *string
	Disabled     *bool
	// DisabledByAdmin is the administrator's lock, set by itself: true
	// locks the account whatever Disabled already is (a registration whose
	// link is still out is disabled already, and "disabled: true" for it is
	// no change and is not remembered); false takes the lock off, and the
	// account is enabled if its address is proven and waits for its link
	// if not. Nil leaves the lock to Disabled, which sets it when it is a
	// change (UsersRepository.UpdateByID).
	DisabledByAdmin *bool
	LocalAccount    *bool
	// Admin is the `admin` flag, the administrator account: every project,
	// and the few things only that account does. True grants it and false
	// takes it back; a value the account already has is no change. A change
	// is made by an account that has the flag and by nobody else, and the
	// last such account that can sign in keeps it (UsersService.flagMayChange).
	Admin *bool
	ID    uuid.UUID
	// CallerID is who asks, on PUT /users/{user_id}. Three things in that
	// body are held to it (UsersService.UpdateByID): a change of address, a
	// change of the `admin` flag, and a change of `disabled`, of the lock or
	// of `local_account` on an account with the flag. The flows that update
	// an account for itself (the profile, a password reset) carry none of
	// those and name no caller.
	CallerID uuid.UUID
}

func (ref *UpdateUserInput) Validate() error {
	if reflect.DeepEqual(ref, &UpdateUserInput{}) {
		return &InvalidUserUpdateError{Message: "at least one field must be updated"}
	}

	var errs ValidationErrors

	if err := ValidateUUID(ref.ID, 7, FieldID); err != nil {
		errs.Add(err)
	}

	if ref.FirstName != nil {
		if normalizedFirstName, err := ValidateString(*ref.FirstName, StringValidationOptions{
			MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength,
			TrimWhitespace: true, AllowEmpty: false, NoControlChars: true, NoHTMLTags: true,
			NoNullBytes: true, NormalizeUnicode: true, FieldName: FieldFirstName,
		}); err != nil {
			errs.Add(err)
		} else {
			ref.FirstName = &normalizedFirstName
		}
	}

	if ref.LastName != nil {
		if normalizedLastName, err := ValidateString(*ref.LastName, StringValidationOptions{
			MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength,
			TrimWhitespace: true, AllowEmpty: false, NoControlChars: true, NoHTMLTags: true,
			NoNullBytes: true, NormalizeUnicode: true, FieldName: FieldLastName,
		}); err != nil {
			errs.Add(err)
		} else {
			ref.LastName = &normalizedLastName
		}
	}

	if ref.Email != nil {
		if normalizedEmail, err := ValidateEmail(*ref.Email, FieldEmail); err != nil {
			errs.Add(err)
		} else {
			ref.Email = &normalizedEmail
		}
	}

	if ref.PasswordHash != nil && *ref.PasswordHash != "" {
		if _, err := ValidateString(*ref.PasswordHash, StringValidationOptions{
			MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength,
			TrimWhitespace: false, AllowEmpty: false, NoControlChars: true, NoNullBytes: true,
			FieldName: FieldPasswordHash,
		}); err != nil {
			errs.Add(err)
		}
	}

	if ref.Password != nil && *ref.Password != "" {
		errs.Add(ValidatePassword(*ref.Password, FieldPassword))
	}

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

type DeleteUserInput struct {
	ID uuid.UUID
	// CallerID is who asks, on DELETE /users/{user_id}: an account with the
	// `admin` flag is deleted only by one that has it. A delete that names no
	// caller deletes an account without the flag, and is a fault on one with
	// it.
	CallerID uuid.UUID
}

func (ref *DeleteUserInput) Validate() error {
	var errs ValidationErrors
	errs.Add(ValidateUUID(ref.ID, 7, FieldID))
	if errs.HasErrors() {
		return &errs
	}
	return nil
}

type SelectUsersInput struct {
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectUsersInput) Validate() error {
	var errs ValidationErrors

	errs.Add(ref.Paginator.Validate())

	validateListExpressions(&errs, listExpressions{
		sort: ref.Sort, sortParser: UsersSortParser, sortFields: UsersSortFields,
		filter: ref.Filter, filterParser: UsersFilterParser, filterFields: UsersFilterFields,
		fields: ref.Fields, fieldsParser: UsersFieldsParser, partialFields: UsersPartialFields,
	})

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

type ListUsersInput = SelectUsersInput

type SelectUsersOutput struct {
	Items     []User
	Paginator Paginator
}

type ListUsersOutput = SelectUsersOutput

type LinkRolesToUserInput struct {
	RoleIDs []uuid.UUID
	// CallerID is who is granting; the guard checks they hold what they hand
	// out, and refuses a zero value.
	CallerID uuid.UUID
	UserID   uuid.UUID
}

func (ref *LinkRolesToUserInput) Validate() error {
	var errs ValidationErrors

	errs.Add(ValidateUUID(ref.UserID, 7, FieldUserID))

	if len(ref.RoleIDs) < 1 {
		errs.AddError("role_ids", "must be at least one role ID", "REQUIRED")
	}

	for i, roleID := range ref.RoleIDs {
		if err := ValidateUUID(roleID, 7, fmt.Sprintf("role_ids[%d]", i)); err != nil {
			errs.Add(err)
		}
	}

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

type UnlinkRolesFromUsersInput = LinkRolesToUserInput

type LinkProjectsToUserInput struct {
	ProjectIDs []uuid.UUID
	UserID     uuid.UUID
	// CallerID is who asks, on POST and DELETE /users/{user_id}/projects:
	// a caller links or unlinks a user only on a project the caller is
	// linked to (projectsOfCaller, in the repository).
	CallerID uuid.UUID
}

func (ref *LinkProjectsToUserInput) Validate() error {
	var errs ValidationErrors

	errs.Add(ValidateUUID(ref.UserID, 7, FieldUserID))

	for i, projectID := range ref.ProjectIDs {
		if err := ValidateUUID(projectID, 7, fmt.Sprintf("project_ids[%d]", i)); err != nil {
			errs.Add(err)
		}
	}

	if errs.HasErrors() {
		return &errs
	}

	return nil
}

type UnlinkProjectsFromUserInput = LinkProjectsToUserInput

type SelectAuthzOutput struct {
	Permissions map[string]any `json:"permissions"`
	Roles       []string       `json:"roles"`
	Policies    []string       `json:"policies"`
}
