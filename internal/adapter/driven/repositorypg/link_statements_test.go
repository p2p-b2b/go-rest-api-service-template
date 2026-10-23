package repositorypg

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every bulk link statement is INSERT ... SELECT FROM UNNEST(...) ON CONFLICT
// ... DO UPDATE, and Postgres refuses to update one row twice in one
// statement (SQLSTATE 21000): a body that named an id twice was a 500. Each
// selects DISTINCT. The pull request gate has no database, so the source is
// read: a new link statement without it fails here, not at the first body
// with a repeated id.
func TestEveryLinkStatementSelectsEachPairOnce(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	unnest := regexp.MustCompile(`SELECT\s+(DISTINCT\s+)?\*\s+FROM\s+UNNEST\(`)
	seen := 0

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		for _, m := range unnest.FindAllStringSubmatch(string(raw), -1) {
			seen++

			if m[1] == "" {
				t.Errorf("%s: a link statement selects FROM UNNEST without DISTINCT: %q", file, m[0])
			}
		}
	}

	// Five today: a project's users, a user's projects, a user's roles, a
	// role's users, a role's policies.
	if seen != 5 {
		t.Errorf("found %d link statements, want 5: one was added or removed; check it, and this count", seen)
	}
}
