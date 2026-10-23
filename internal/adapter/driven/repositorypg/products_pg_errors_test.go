package repositorypg

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// A product's refusals are in this service's words. Postgres's sentence --
// "duplicate key value violates unique constraint ...", the foreign key's --
// was the answer's text, so a driver or server upgrade would have rewritten
// the API's contract, and the constraint's name was published with it.
func TestAProductsRefusalIsInOurWords(t *testing.T) {
	t.Parallel()

	ref := &ProductsRepository{}

	for name, tt := range map[string]struct {
		pg   *pgconn.PgError
		want string
	}{
		"a_name_that_is_taken": {
			pg:   &pgconn.PgError{Code: "23505", ConstraintName: constraintProductsProjectName, Message: `duplicate key value violates unique constraint "unique_project_product_name"`},
			want: (&domain.ProductAlreadyExistsError{Message: "the name is taken"}).Error(),
		},
		"an_id_that_is_taken": {
			pg:   &pgconn.PgError{Code: "23505", ConstraintName: "products_pkey", Message: `duplicate key value violates unique constraint "products_pkey"`},
			want: (&domain.ProductAlreadyExistsError{Message: "the id is taken"}).Error(),
		},
		"a_project_that_is_gone": {
			pg:   &pgconn.PgError{Code: "23503", ConstraintName: "products_projects_id_fkey", Message: `insert or update on table "products" violates foreign key constraint "products_projects_id_fkey"`},
			want: (&domain.ProjectNotFoundError{}).Error(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ref.handlePgError(tt.pg, nil)
			if err.Error() != tt.want {
				t.Errorf("answered %q, want %q", err, tt.want)
			}

			if strings.Contains(err.Error(), "violates") || strings.Contains(err.Error(), tt.pg.ConstraintName) {
				t.Errorf("the database's words reached the answer: %q", err)
			}
		})
	}

	if _, ok := errors.AsType[*domain.ProductAlreadyExistsError](ref.handlePgError(&pgconn.PgError{Code: "23505"}, nil)); !ok {
		t.Error("a unique violation is no longer the product's conflict")
	}
}

// The constraint the name's refusal is read from is the migration's: a
// rename there without one here would answer a taken name as a taken id.
func TestTheProductsNameConstraintIsTheMigrations(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../../../database/migrations/00009_products_tables.sql")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "CONSTRAINT "+constraintProductsProjectName+" UNIQUE") {
		t.Errorf("the migration does not declare %s", constraintProductsProjectName)
	}
}
