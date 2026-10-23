package repositorypg

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// TestEveryFlagOfAnAccountIsNotNull: a boolean of `users` is set or it is
// not. `disabled`, `admin` and `local_account` were nullable, and a NULL was
// read more than one way: with `disabled` NULL and the flag, the sign-in and
// the account's standing took the account for enabled while the count of
// enabled administrator accounts (`disabled IS FALSE`) left it out, and the
// projects' view, which splits on `admin = TRUE` and `admin = FALSE`, gave an
// account with `admin` NULL no project at all. Every reader would have to
// spell the third value the same way; the column has none.
//
// The integration suite shows the table refusing the NULL; this is what the
// pull request gate runs.
func TestEveryFlagOfAnAccountIsNotNull(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../../../database/migrations/00002_users_tables.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	table := regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS users \((.*?)\n\);`).FindSubmatch(b)
	if table == nil {
		t.Fatal("00002_users_tables.sql has no CREATE TABLE for users: the migration or this test's pattern changed")
	}

	column := regexp.MustCompile(`(?m)^\s*([a-z_]+)\s+BOOLEAN\b(.*)$`)

	var flags []string

	for _, m := range column.FindAllStringSubmatch(string(table[1]), -1) {
		name, definition := m[1], m[2]
		flags = append(flags, name)

		if !strings.Contains(definition, "NOT NULL") {
			t.Errorf("users.%s is a BOOLEAN that may be NULL: give it NOT NULL and a DEFAULT", name)
		}
	}

	// So that a pattern that stopped matching does not pass on nothing.
	for _, want := range []string{"disabled", "verified", "disabled_by_admin", "admin", "local_account"} {
		if !slices.Contains(flags, want) {
			t.Errorf("users.%s was not found among the table's BOOLEAN columns %v", want, flags)
		}
	}

	// Nor is it taken off later.
	migrations, err := filepath.Glob("../../../../database/migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("no migrations found: %v", err)
	}

	dropped := regexp.MustCompile(`(?is)ALTER\s+TABLE\s+(IF\s+EXISTS\s+)?(ONLY\s+)?users\b[^;]*DROP\s+NOT\s+NULL`)

	for _, path := range migrations {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if dropped.Match(b) {
			t.Errorf("%s drops a NOT NULL of users", filepath.Base(path))
		}
	}
}
