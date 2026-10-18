package repositorypg

import (
	"errors"
	"os"
	"strings"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// The seeded administrator keeps the Administrator role, and is deleted by
// nobody: a delete takes its link to the role with it. The trigger that
// guards that link refuses the delete and both unlinks, and the repositories
// passed the database's error on as it came, a 500; each answers the refusal
// a protected administrator account has, a 403.
//
// The mapping is by the trigger's own words, since P0001 is every RAISE
// EXCEPTION, so the words are held against the migration that raises them.
func TestTheSeededAdministratorsLinkIsA403(t *testing.T) {
	t.Parallel()

	migration, err := os.ReadFile("../../../../database/migrations/00007_roles_policies_tables.sql")
	if err != nil {
		t.Fatalf("read the migration: %v", err)
	}

	if !strings.Contains(string(migration), "RAISE EXCEPTION '"+bootstrapUserLinkMessage+"'") {
		t.Errorf("migration 00007 no longer raises %q", bootstrapUserLinkMessage)
	}

	raised := &pgconn.PgError{Code: "P0001", Message: bootstrapUserLinkMessage}

	for _, tc := range []struct {
		name   string
		mapped error
		want   string
	}{
		{
			name:   "the_delete",
			mapped: (&UsersRepository{}).handlePgError(raised, &domain.DeleteUserInput{ID: uuid.NewV7(), CallerID: uuid.NewV7()}),
			want:   "user cannot be deleted: system/protected resource: the bootstrap administrator",
		},
		{
			name:   "the_users_unlink",
			mapped: (&UsersRepository{}).handlePgError(raised, &domain.UnlinkRolesFromUsersInput{UserID: uuid.NewV7()}),
			want:   "user cannot be stripped of a role: system/protected resource: the bootstrap administrator",
		},
		{
			name:   "the_roles_unlink",
			mapped: (&RolesRepository{}).handlePgError(raised, &domain.UnlinkUsersFromRoleInput{RoleID: uuid.NewV7()}),
			want:   "user cannot be stripped of a role: system/protected resource: the bootstrap administrator",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			refusal, ok := errors.AsType[*domain.AdministratorProtectedError](tc.mapped)
			if !ok {
				t.Fatalf("the trigger's refusal is %T (%v), not the 403 of a protected administrator account", tc.mapped, tc.mapped)
			}

			if refusal.Error() != tc.want {
				t.Errorf("refused as %q, want %q", refusal.Error(), tc.want)
			}
		})
	}

	// Only that refusal, and only for those statements: another exception,
	// and the same one raised under another statement, are not answered as
	// it.
	for name, other := range map[string]error{
		"another_exception": (&UsersRepository{}).handlePgError(&pgconn.PgError{Code: "P0001", Message: "something else was raised"}, &domain.DeleteUserInput{ID: uuid.NewV7()}),
		"another_statement": (&UsersRepository{}).handlePgError(raised, &domain.UpdateUserInput{ID: uuid.NewV7()}),
		"a_roles_exception": (&RolesRepository{}).handlePgError(&pgconn.PgError{Code: "P0001", Message: "something else was raised"}, &domain.UnlinkUsersFromRoleInput{RoleID: uuid.NewV7()}),
	} {
		if _, protected := errors.AsType[*domain.AdministratorProtectedError](other); protected {
			t.Errorf("%s is answered as the seeded administrator's refusal: %v", name, other)
		}
	}
}
