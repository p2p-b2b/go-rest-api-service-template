package repositorypg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
	"github.com/slashdevops/go-rest-api-service-template/pkg/cslog"
)

// Constraint names this repository matches on. A CONSTRAINT NAME REFERENCED
// FROM GO IS A CONTRACT: renaming one in the migration without changing the
// constant here turns a documented 409 into a 500.
const (
	constraintUsersIdentitiesPK      = "pk_users_identities"
	constraintUsersIdentitiesUserIDP = "unique_users_identities_user_idp"
)

// queryLinkIdentity stores an account's identity at a provider: the link of
// an existing account (Link), and the identity of an account provisioned
// from it, in the account's own transaction (UsersRepository.Insert).
//
// Only under the provider the sign-in used: the row is read with the issuer
// and the kind the sign-in's token came from ($5, $6), FOR KEY SHARE. The
// issuer is compared as the reads answer it, the empty string for a github
// provider, whether a writer stored NULL or an empty one. A provider may be
// pointed at another issuer or type while nobody is linked to it
// (IDPsRepository.UpdateByID);
// a sign-in already in flight, or one that read the provider through a
// cached copy, would have stored an identity under the provider as it is
// now, for a subject another issuer asserted. The repoint takes the row FOR
// UPDATE first, which conflicts with KEY SHARE: an insert that comes second
// waits for it and then reads the row as it was left, and inserts nothing
// when it changed; a repoint that comes second waits for the insert and
// sees its identity (409). No row inserted is the refusal
// (linkIdentityRefused).
const queryLinkIdentity = `
        INSERT INTO users_identities (users_id, idps_id, subject, email)
        SELECT $1, idps.id, $3, $4
        FROM idps
        WHERE idps.id = $2
          AND idps.idp_types = $6
          AND COALESCE(idps.issuer_url, '') = $5
        FOR KEY SHARE OF idps;
    `

// linkIdentityRefused is what an identity's insert that stored nothing
// answers: the provider is not as the sign-in used it, or not there.
func linkIdentityRefused() error {
	return &domain.IDPChangedDuringSignInError{}
}

type UsersIdentitiesRepositoryConfig struct {
	DB              *pgxpool.Pool
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
	MaxPingTimeout  time.Duration
	MaxQueryTimeout time.Duration
}

// UsersIdentitiesRepository is the PostgreSQL store for the link between an
// account and a provider identity.
type UsersIdentitiesRepository struct {
	db              *pgxpool.Pool
	ot              *o11y.OpenTelemetry
	metrics         *o11y.LayerMetrics
	metricsMetadata o11y.Metadata
	metricsPrefix   string
	maxPingTimeout  time.Duration
	maxQueryTimeout time.Duration
}

// NewUsersIdentitiesRepository creates a new UsersIdentitiesRepository.
func NewUsersIdentitiesRepository(conf UsersIdentitiesRepositoryConfig) (*UsersIdentitiesRepository, error) {
	if conf.DB == nil {
		return nil, &domain.InvalidDBConfigurationError{Message: "invalid database configuration. It is nil"}
	}

	if conf.MaxPingTimeout < domain.ValidDatabaseMinPingTimeout || conf.MaxPingTimeout > domain.ValidDatabaseMaxPingTimeout {
		return nil, &domain.InvalidDBMaxPingTimeoutError{Message: fmt.Sprintf("invalid max ping timeout. It must be between %d and %d", domain.ValidDatabaseMinPingTimeout, domain.ValidDatabaseMaxPingTimeout)}
	}

	if conf.MaxQueryTimeout < domain.ValidDatabaseMinQueryTimeout || conf.MaxQueryTimeout > domain.ValidDatabaseMaxQueryTimeout {
		return nil, &domain.InvalidDBMaxQueryTimeoutError{Message: fmt.Sprintf("invalid max query timeout. It must be between %d and %d", domain.ValidDatabaseMinQueryTimeout, domain.ValidDatabaseMaxQueryTimeout)}
	}

	if conf.OT == nil {
		return nil, &domain.InvalidOTConfigurationError{Message: "invalid OpenTelemetry configuration. It is nil"}
	}

	ref := &UsersIdentitiesRepository{
		db:              conf.DB,
		maxPingTimeout:  conf.MaxPingTimeout,
		maxQueryTimeout: conf.MaxQueryTimeout,
		ot:              conf.OT,
		metricsMetadata: o11y.Metadata{Layer: AppLayer, Domain: "UsersIdentities", Action: "NewUsersIdentitiesRepository"},
	}

	if conf.MetricsPrefix != "" {
		ref.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_") + "_"
	}

	metrics, err := o11y.NewLayerMetrics(ref.ot.Metrics.Meter, ref.metricsPrefix)
	if err != nil {
		return nil, err
	}

	ref.metrics = metrics

	return ref, nil
}

// PingContext verifies the connection is alive.
func (ref *UsersIdentitiesRepository) PingContext(ctx context.Context) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxPingTimeout, ref.metricsMetadata, "PingContext")
	defer cancel()
	defer span.End()

	if err := ref.db.Ping(ctx); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	return nil
}

// Link implements [repository.UsersIdentities].
func (ref *UsersIdentitiesRepository) Link(ctx context.Context, input *domain.LinkUserIdentityInput) error {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "Link")
	defer cancel()
	defer span.End()

	if input == nil {
		return o11y.RecordError(ctx, span, start, &domain.InvalidInputError{Message: "input is nil"}, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()), attribute.String("idp.id", input.IDPID.String()))

	cslog.Trace(ctx, "sql", "query", prettyPrint(queryLinkIdentity))

	result, err := ref.db.Exec(ctx, queryLinkIdentity, input.UserID, input.IDPID, input.Subject, input.Email, input.IssuerURL, input.IDPTypeID)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)
	}

	if result.RowsAffected() == 0 {
		return o11y.RecordError(ctx, span, start, linkIdentityRefused(), ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "identity linked")

	return nil
}

// queryLockAccountForUnlink takes the account's row before the unlink's
// statement, as a provider's delete takes the rows of the accounts linked to
// it (queryLockAccountsOfIDP): two writes that each remove one of an
// account's ways in are one at a time, and the second reads what the first
// left. FOR NO KEY UPDATE does not hold a sign-in's identity insert.
const queryLockAccountForUnlink = `SELECT TRUE FROM users WHERE id = $1 FOR NO KEY UPDATE;`

// queryUnlinkIdentity removes an identity unless it is the account's only way
// in: an account with no password and no identity at another enabled
// provider. The use-case asks the same first (keepsAWayIn); this statement,
// under the account's lock, is what holds when a provider's delete or
// another unlink took the other way in meanwhile, which the use-case's two
// reads did not see.
const queryUnlinkIdentity = `
        DELETE FROM users_identities
        WHERE users_id = $1 AND idps_id = $2
          AND EXISTS (
                SELECT 1 FROM users
                WHERE users.id = $1
                  AND (
                        users.local_account IS TRUE
                        OR EXISTS (
                              SELECT 1 FROM users_identities AS other
                              JOIN idps ON idps.id = other.idps_id
                              WHERE other.users_id = $1 AND other.idps_id <> $2 AND idps.enabled
                        )
                  )
          );
    `

// queryIdentityExists tells an identity the unlink kept from one that is not
// there.
const queryIdentityExists = `SELECT EXISTS (SELECT 1 FROM users_identities WHERE users_id = $1 AND idps_id = $2);`

// Unlink implements [repository.UsersIdentities]: one transaction, the
// account's row taken first, and the rule in the statement that deletes.
func (ref *UsersIdentitiesRepository) Unlink(ctx context.Context, input *domain.UnlinkUserIdentityInput) (err error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "Unlink")
	defer cancel()
	defer span.End()

	if input == nil {
		return o11y.RecordError(ctx, span, start, &domain.InvalidInputError{Message: "input is nil"}, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	tx, err := ref.db.Begin(ctx)
	if err != nil {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)
	}
	defer endTx(ctx, tx, &err, func(e error) error {
		return o11y.RecordError(ctx, span, start, ref.handlePgError(e), ref.metrics, attrs)
	})

	cslog.Trace(ctx, "sql", "query", prettyPrint(queryLockAccountForUnlink, input.UserID.String()))

	// Run to its end, its row dropped: an account that is not there has no
	// identity either, which the delete below answers.
	if _, err = tx.Exec(ctx, queryLockAccountForUnlink, input.UserID); err != nil {
		err = o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)

		return err
	}

	cslog.Trace(ctx, "sql", "query", prettyPrint(queryUnlinkIdentity, input.UserID.String(), input.IDPID.String()))

	result, err := tx.Exec(ctx, queryUnlinkIdentity, input.UserID, input.IDPID)
	if err != nil {
		err = o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)

		return err
	}

	if result.RowsAffected() == 0 {
		// Nothing was removed: no such identity, or the statement kept the
		// account's only way in. Which of the two is a read under the lock.
		var there bool

		cslog.Trace(ctx, "sql", "query", prettyPrint(queryIdentityExists, input.UserID.String(), input.IDPID.String()))

		if err = tx.QueryRow(ctx, queryIdentityExists, input.UserID, input.IDPID).Scan(&there); err != nil {
			err = o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)

			return err
		}

		if there {
			err = o11y.RecordError(ctx, span, start, &domain.UserIdentityAlreadyLinkedError{Message: domain.UnlinkOnlyWayInMessage}, ref.metrics, attrs)

			return err
		}

		err = o11y.RecordError(ctx, span, start, &domain.UserIdentityNotFoundError{Message: "this account has no identity at that provider"}, ref.metrics, attrs)

		return err
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "identity unlinked")

	return nil
}

// SelectBySubject implements [repository.UsersIdentities].
func (ref *UsersIdentitiesRepository) SelectBySubject(ctx context.Context, idpID, idpTypeID uuid.UUID, issuerURL, subject string) (*domain.UserIdentity, error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "SelectBySubject")
	defer cancel()
	defer span.End()

	if idpID == uuid.Nil() || subject == "" {
		return nil, o11y.RecordError(ctx, span, start, &domain.InvalidInputError{Message: "idp and subject are required"}, ref.metrics, attrs)
	}

	// Under the provider as the sign-in used it, as queryLinkIdentity
	// stores one: a sign-in that read the provider before it was pointed at
	// another issuer, held at the old issuer or through a cached copy,
	// proved its subject with an issuer the provider no longer trusts. Read
	// by the provider and the subject alone, it signed in as the account
	// linked to that subject under the new issuer. Not found here, it is an
	// identity this provider does not know, and the insert that would
	// provision it refuses the provider as changed.
	const query = `
        SELECT ui.users_id, ui.idps_id, ui.subject, ui.email, ui.linked_at, idp.name, idpt.name, idp.enabled
        FROM users_identities AS ui
            JOIN idps AS idp ON idp.id = ui.idps_id
            JOIN idp_types AS idpt ON idpt.id = idp.idp_types
        WHERE ui.idps_id = $1 AND ui.subject = $2
          AND idp.idp_types = $3 AND COALESCE(idp.issuer_url, '') = $4;
    `

	cslog.Trace(ctx, "sql", "query", prettyPrint(query, idpID, subject, idpTypeID, issuerURL))

	var item domain.UserIdentity

	err := ref.db.QueryRow(ctx, query, idpID, subject, idpTypeID, issuerURL).Scan(&item.UserID, &item.IDPID, &item.Subject, &item.Email, &item.LinkedAt, &item.IDPName, &item.IDPTypeName, &item.IDPEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// The ordinary answer for a first sign-in; an error so the caller
			// branches on it rather than on a nil.
			return nil, o11y.RecordError(ctx, span, start, &domain.UserIdentityNotFoundError{}, ref.metrics, attrs)
		}

		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "identity found")

	return &item, nil
}

// SelectByUserID implements [repository.UsersIdentities].
func (ref *UsersIdentitiesRepository) SelectByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserIdentity, error) {
	start := time.Now()
	ctx, span, attrs, cancel := o11y.SetupTraceWithTimeout(ctx, ref.ot.Traces.Tracer, ref.maxQueryTimeout, ref.metricsMetadata, "SelectByUserID")
	defer cancel()
	defer span.End()

	if userID == uuid.Nil() {
		return nil, o11y.RecordError(ctx, span, start, &domain.InvalidInputError{Message: "user id is required"}, ref.metrics, attrs)
	}

	const query = `
        SELECT ui.users_id, ui.idps_id, ui.subject, ui.email, ui.linked_at, idp.name, idpt.name, idp.enabled
        FROM users_identities AS ui
            JOIN idps AS idp ON idp.id = ui.idps_id
            JOIN idp_types AS idpt ON idpt.id = idp.idp_types
        WHERE ui.users_id = $1
        ORDER BY ui.linked_at ASC;
    `

	cslog.Trace(ctx, "sql", "query", prettyPrint(query, userID))

	rows, err := ref.db.Query(ctx, query, userID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, ref.handlePgError(err), ref.metrics, attrs)
	}
	defer rows.Close()

	items := make([]domain.UserIdentity, 0, 4)

	for rows.Next() {
		var item domain.UserIdentity
		if err := rows.Scan(&item.UserID, &item.IDPID, &item.Subject, &item.Email, &item.LinkedAt, &item.IDPName, &item.IDPTypeName, &item.IDPEnabled); err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "identities listed", attribute.Int("count", len(items)))

	return items, nil
}

func (ref *UsersIdentitiesRepository) handlePgError(err error) error {
	return identityPgError(err)
}

// identityPgError maps the errors an identity's insert or delete raises: the
// identity's two unique keys, and a foreign key whose account or provider is
// gone. UsersRepository.Insert, which stores a provisioned account's
// identity in its own transaction, answers through it too.
func identityPgError(err error) error {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return err
	}

	switch pgErr.Code {
	case "23505": // unique violation
		switch {
		case strings.Contains(pgErr.ConstraintName, constraintUsersIdentitiesPK):
			return &domain.UserIdentityAlreadyLinkedError{Message: "this provider identity is already linked to an account"}
		case strings.Contains(pgErr.ConstraintName, constraintUsersIdentitiesUserIDP):
			return &domain.UserIdentityAlreadyLinkedError{Message: "this account already has an identity at that provider"}
		}
	case "23503": // foreign key: the account is gone
		// The account: the provider's row is read and held by the insert's
		// own statement (queryLinkIdentity), so a provider that is gone is
		// no row inserted, not this.
		return &domain.UserIdentityNotFoundError{Message: "the account does not exist"}
	}

	return err
}

var _ repository.UsersIdentities = (*UsersIdentitiesRepository)(nil)
