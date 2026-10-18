package repositorypg

import (
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// The provider's delete decides, in its own statement, whether an account
// would be left with no way in: an account that is not a local one (a
// `local_account` never set is no password login either), with an identity
// at this provider and at no other. The pull request gate has no
// database, so the statement is held whole, spaces apart: a term dropped
// from it (which account, which provider, "no other") still compiles and
// deletes more than it should, or nothing.
func TestTheProvidersDeleteKeepsItForAnAccountWithNoOtherWayIn(t *testing.T) {
	t.Parallel()

	const want = "DELETE FROM idps WHERE id = $1 AND NOT EXISTS ( " +
		"SELECT 1 FROM users_identities AS only_way JOIN users ON users.id = only_way.users_id " +
		"WHERE only_way.idps_id = $1 AND users.local_account IS NOT TRUE " +
		"AND NOT EXISTS ( SELECT 1 FROM users_identities AS other " +
		"WHERE other.users_id = only_way.users_id AND other.idps_id <> $1 ) );"

	if got := strings.Join(strings.Fields(queryDeleteIDP), " "); got != want {
		t.Errorf("the statement changed:\n got: %s\nwant: %s", got, want)
	}

	// And the read that tells a missing provider from a kept one asks about
	// the same provider.
	if got, want := strings.Join(strings.Fields(queryIDPExists), " "), "SELECT EXISTS (SELECT 1 FROM idps WHERE id = $1);"; got != want {
		t.Errorf("the read changed:\n got: %s\nwant: %s", got, want)
	}
}

// A delete that deleted nothing is a 404 for a provider that is not there
// and the 409 for one the statement kept, and the 409 says what keeps it.
func TestADeleteThatDeletedNothingSaysWhy(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()

	kept, ok := errors.AsType[*domain.IDPInUseError](idpNotDeleted(true, id))
	if !ok {
		t.Fatalf("a kept provider is answered %T, want the 409", idpNotDeleted(true, id))
	}

	if want := "identity provider cannot be deleted: still used by at least one account that has no other way to sign in"; kept.Error() != want {
		t.Errorf("the 409 says %q, want %q", kept.Error(), want)
	}

	missing, ok := errors.AsType[*domain.IDPNotFoundError](idpNotDeleted(false, id))
	if !ok || missing.ID != id {
		t.Fatalf("a provider that is not there is answered %T (%v), want its 404", idpNotDeleted(false, id), idpNotDeleted(false, id))
	}
}

// "Can sign in", for an account with the `admin` flag: enabled, and with a
// password login or an identity at a provider that is enabled, and that is
// not the provider being asked about. One predicate for the three questions
// (how many, how many without a provider, is this account one of them), held
// whole: the guard of the administrator account and the guard of a provider
// both stand on it, and two spellings of it would be two meanings of "the
// last one".
func TestWhoCountsAsAnAdministratorThatCanSignIn(t *testing.T) {
	t.Parallel()

	const predicate = "admin IS TRUE AND disabled IS FALSE " +
		"AND (local_account OR EXISTS ( SELECT 1 FROM users_identities " +
		"JOIN idps ON idps.id = users_identities.idps_id " +
		"WHERE users_identities.users_id = users.id AND idps.enabled " +
		"AND ($1::uuid IS NULL OR idps.id <> $1::uuid) ))"

	spaced := func(statement string) string { return strings.Join(strings.Fields(statement), " ") }

	if got := spaced(enabledAdministrator); got != predicate {
		t.Errorf("the predicate changed:\n got: %s\nwant: %s", got, predicate)
	}

	if got, want := spaced(queryEnabledAdministrators), "SELECT count(*) FROM users WHERE "+predicate+";"; got != want {
		t.Errorf("the count changed:\n got: %s\nwant: %s", got, want)
	}

	if got, want := spaced(queryIsEnabledAdministrator), "SELECT EXISTS ( SELECT 1 FROM users WHERE id = $2::uuid AND "+predicate+" );"; got != want {
		t.Errorf("the question about one account changed:\n got: %s\nwant: %s", got, want)
	}
}

// The two questions that name something refuse a nil id before anything is
// asked: "without no provider" is the other count, and "is nobody one of
// them" has no answer.
func TestTheCountsRefuseWhatWasNotNamed(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ot := &o11y.OpenTelemetry{
		Traces:  o11y.NewOpenTelemetryTracer(ctx, &o11y.OpenTelemetryTracerConfig{Name: "test"}),
		Metrics: o11y.NewOpenTelemetryMeter(ctx, &o11y.OpenTelemetryMeterConfig{Name: "test"}),
	}

	// The pool is never dialled: both refusals come before a statement.
	repo, err := NewUsersRepository(UsersRepositoryConfig{DB: &pgxpool.Pool{}, OT: ot, MaxPingTimeout: 5 * time.Second, MaxQueryTimeout: 10 * time.Second})
	if err != nil {
		t.Fatalf("NewUsersRepository: %v", err)
	}

	if _, err := repo.CountEnabledAdministratorsWithout(t.Context(), uuid.Nil()); err == nil {
		t.Error("the administrators were counted without a provider that was not named")
	}

	if _, err := repo.IsEnabledAdministrator(t.Context(), uuid.Nil()); err == nil {
		t.Error("an account that was not named was asked about")
	}
}
