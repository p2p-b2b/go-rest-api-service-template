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
	UsersFilterFields  = []string{FieldID, FieldFirstName, FieldLastName, FieldEmail, FieldDisabled, FieldVerified, FieldLocalAccount, FieldCreatedAt, FieldUpdatedAt}
	UsersSortFields    = []string{FieldID, FieldFirstName, FieldLastName, FieldEmail, FieldDisabled, FieldVerified, FieldLocalAccount, FieldCreatedAt, FieldUpdatedAt}
	UsersPartialFields = []string{FieldID, FieldFirstName, FieldLastName, FieldEmail, FieldDisabled, FieldVerified, FieldLocalAccount, FieldCreatedAt, FieldUpdatedAt}
)

// User is the user-account entity.
//
// Disabled and Verified are two facts, not one. Disabled says whether the
// account may sign in and is what an administrator switches. Verified says
// whether the address has been proven by following the verification link (or
// by an identity provider, which proves its subject, or by the administrator
// who created the account and vouches for it). A registration is unverified
// and disabled; verification flips both, unless an administrator disabled
// the account meanwhile, which the row remembers in a column this type does
// not carry. Those three are the only writers of Verified, and no update
// sets it. They used to be one flag, and
// password recovery could not tell an account that had never followed its
// link from one that had been switched off, so it refused both.
type User struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Disabled     *bool
	Verified     *bool
	Admin        *bool
	LocalAccount *bool
	FirstName    string
	LastName     string
	Email        string
	Password     string
	PasswordHash string
	SerialID     int64
	ID           uuid.UUID
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
	LocalAccount *bool
	ID           uuid.UUID
	// CallerID is who asks, on PUT /users/{user_id}: a change of the
	// account's e-mail address is held to it (UsersService.UpdateByID). The
	// flows that update an account for itself (the profile, the
	// verification, a password reset) carry no address and name no caller.
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

	parseSortExpression(&errs, UsersSortParser, ref.Sort, "INVALID_SORT_FIELD")
	parseFilterExpression(&errs, UsersFilterParser, ref.Filter, "INVALID_FILTER_FIELD")
	parseFieldsExpression(&errs, UsersFieldsParser, ref.Fields, "INVALID_FIELD")

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
