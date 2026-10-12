package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"go.opentelemetry.io/otel/attribute"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/cache"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/repository"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

type UsersServiceConf struct {
	// Guard refuses a grant the caller does not hold. Required: a service
	// that can widen permissions without it is the escalation path.
	Guard           *GrantGuard
	Repository      repository.Users
	CacheService    cache.Cache
	ResourcesLimits ResourcesLimitsServiceConsumer
	OT              *o11y.OpenTelemetry
	MetricsPrefix   string
}

type UsersService struct {
	guard           *GrantGuard
	repository      repository.Users
	cacheService    cache.Cache
	resourcesLimits ResourcesLimitsServiceConsumer
	ot              *o11y.OpenTelemetry
	metrics         *o11y.LayerMetrics
	metricsMetadata o11y.Metadata
	metricsPrefix   string
}

func NewUsersService(conf UsersServiceConf) (*UsersService, error) {
	if conf.Guard == nil {
		return nil, &domain.InvalidInputError{Message: "Guard is nil, but it is required for UsersService"}
	}

	if conf.Repository == nil {
		return nil, &domain.InvalidRepositoryError{Message: "Repository is nil, but it is required for UsersService"}
	}

	if conf.ResourcesLimits == nil {
		return nil, &domain.InvalidResourcesLimitsError{Message: "ResourcesLimits is nil, but it is required for UsersService"}
	}

	if conf.OT == nil {
		return nil, &domain.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for UsersService"}
	}

	ref := &UsersService{
		guard:           conf.Guard,
		repository:      conf.Repository,
		cacheService:    conf.CacheService,
		resourcesLimits: conf.ResourcesLimits,
		ot:              conf.OT,
		metricsMetadata: o11y.Metadata{
			Layer:  AppLayer,
			Domain: "Users",
			Action: "NewUsersService",
		},
	}

	if conf.MetricsPrefix != "" {
		ref.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		ref.metricsPrefix += "_"
	}

	metrics, err := o11y.NewLayerMetrics(ref.ot.Metrics.Meter, ref.metricsPrefix)
	if err != nil {
		return nil, err
	}

	ref.metrics = metrics

	return ref, nil
}

func (ref *UsersService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "GetByID")
	defer span.End()

	if !domain.IsUUIDV7(id) {
		errorType := &domain.InvalidUserIDError{Message: "user ID cannot be empty"}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", id.String()))

	var out *domain.User
	var err error

	userFetcher := func(ctx context.Context) (*domain.User, []cache.Identifier, error) {
		out, err := ref.repository.SelectByID(ctx, id)
		if err != nil {
			return nil, nil, err
		}

		return withoutCredentials(out), nil, nil
	}

	if ref.cacheService == nil {
		slog.DebugContext(ctx, "cache lookup", "cache", "disabled")

		// Through the same fetcher, so the value a caller receives does not
		// depend on whether caching happens to be enabled.
		out, _, err = userFetcher(ctx)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	} else {

		cacheKey := cache.Identifier{
			Type: "user",
			ID:   id.String(),
		}

		slog.DebugContext(ctx, "cache lookup", "cache", "enabled", "user_id", id)
		out, err = cache.GetTyped[*domain.User](ctx, ref.cacheService, cacheKey, userFetcher)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user found", attribute.String("user.email", out.Email))

	return out, nil
}

// withoutCredentials returns a copy of u with the credential fields cleared.
//
// [UsersService.GetByEmail] and [UsersService.GetByID] are cached, and their
// results are written to Valkey — a store outside the database, with a twelve
// hour hard TTL by default and, unless cache.tls.enabled is set, reached over a
// cleartext connection. A bcrypt hash does not belong there.
//
// It returns a copy rather than clearing in place: the value belongs to the
// caller of the fetcher, and mutating it would surprise anyone who still holds
// a reference to what the repository returned.
//
// The one caller that genuinely needs a hash uses
// [UsersService.GetByEmailForAuth], which does not cache. Everything else gets
// a user with no credentials on it, so no future caller can start depending on
// one being there.
func withoutCredentials(u *domain.User) *domain.User {
	if u == nil {
		return nil
	}

	stripped := *u
	stripped.Password = ""
	stripped.PasswordHash = ""

	return &stripped
}

// GetByEmailForAuth returns the user for email with the password hash intact,
// reading straight from the repository.
//
// This is the only path in the service that returns credentials, and it is
// deliberately uncached. Caching it would put bcrypt hashes in Valkey, and
// caching it *without* the hash would be worse still: the compare in
// [AuthnService.LoginUser] would run against an empty string.
//
// The cost is one indexed lookup per login. Logins are rare next to
// authenticated requests — the per-request authorization data is cached
// separately under authz:<userID> and is unaffected — so there is nothing to
// save here worth the exposure.
func (ref *UsersService) GetByEmailForAuth(ctx context.Context, email string) (*domain.User, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "GetByEmailForAuth")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	if email == "" {
		errorType := &domain.InvalidEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	out, err := ref.repository.SelectByEmail(ctx, email)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user found", attribute.String("user.email", out.Email))

	return out, nil
}

func (ref *UsersService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "GetByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	if email == "" {
		errorType := &domain.InvalidEmailError{Email: email}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	var out *domain.User
	var err error

	userFetcher := func(ctx context.Context) (*domain.User, []cache.Identifier, error) {
		out, err := ref.repository.SelectByEmail(ctx, email)
		if err != nil {
			return nil, nil, err
		}

		// since this is fetching by email, we can also add the user ID cache dependency
		dependencies := []cache.Identifier{
			{
				Type: "user",
				ID:   out.ID.String(),
			},
		}

		return withoutCredentials(out), dependencies, nil
	}

	if ref.cacheService == nil {
		slog.DebugContext(ctx, "cache lookup", "cache", "disabled")

		// Through the same fetcher, so the value a caller receives does not
		// depend on whether caching happens to be enabled.
		out, _, err = userFetcher(ctx)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	} else {
		cacheKey := cache.Identifier{
			Type: "user",
			ID:   email,
		}

		slog.DebugContext(ctx, "cache lookup", "cache", "enabled")
		out, err = cache.GetTyped[*domain.User](ctx, ref.cacheService, cacheKey, userFetcher)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user found", attribute.String("user.email", out.Email))

	return out, nil
}

func (ref *UsersService) Create(ctx context.Context, input *domain.CreateUserInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "Create")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.email", input.Email))

	var err error
	input.ID, err = domain.EnsureUUIDV7(input.ID)
	if err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	// Check resource limits
	rlScope := domain.ResourcesLimitsScope{
		Type: domain.ResourcesLimitsScopeTypeSystem,
	}

	resourceLimitParams := ResourceLimitCheckParams{
		Ctx:             ctx,
		ResourcesLimits: ref.resourcesLimits,
		Scope:           rlScope,
		ResourceType:    domain.ResourcesLimitsResourceTypeUsers,
	}

	// Reserve the slot before creating, and give it back if anything after this
	// point fails.
	if err := ReserveResourceSlot(resourceLimitParams); err != nil {
		return err
	}

	hashPwd, err := HashAndSaltPassword(input.Password)
	if err != nil {
		ReleaseResourceSlot(resourceLimitParams)
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	input.PasswordHash = hashPwd

	if err := ref.repository.Insert(ctx, input); err != nil {
		ReleaseResourceSlot(resourceLimitParams)
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	WarnOnSoftLimit(resourceLimitParams)

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user created successfully",
		attribute.String("user.email", input.Email),
		attribute.String("user.id", input.ID.String()))

	return nil
}

// Standing reports whether the account may act on a request it is
// authenticated for.
//
// It reads the repository and not the cache. The cached account lives twelve
// hours and an invalidation is best effort: a failed one would leave an
// account that was disabled working for half a day, on the one read that
// decides whether "disabled" means anything. It is one read by primary key
// for every authenticated request.
func (ref *UsersService) Standing(ctx context.Context, id uuid.UUID) (domain.AccountStanding, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "Standing")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id.String()))

	account, err := ref.repository.SelectByID(ctx, id)
	if _, gone := errors.AsType[*domain.UserNotFoundError](err); gone {
		o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "the account is not there")

		return domain.AccountMissing, nil
	}

	if err != nil {
		return domain.AccountUnknown, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if account.Disabled != nil && *account.Disabled {
		o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "the account is disabled")

		return domain.AccountDisabled, nil
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "the account may act")

	return domain.AccountActive, nil
}

func (ref *UsersService) UpdateByID(ctx context.Context, input *domain.UpdateUserInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "UpdateByID")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	// An address is not every caller's to change. It is judged against the
	// stored row, read only when the body carries one: the profile, the
	// verification and a password reset carry none.
	if input.Email != nil {
		if err := ref.addressMayChange(ctx, input); err != nil {
			return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	}

	if input.Password != nil {
		if len(*input.Password) < domain.ValidUserPasswordMinLength || len(*input.Password) > domain.ValidUserPasswordMaxLength {
			errorValue := &domain.InvalidUserPasswordError{Message: fmt.Sprintf("password must be between %d and %d characters", domain.ValidUserPasswordMinLength, domain.ValidUserPasswordMaxLength)}
			return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
		}

		hashPwd, err := HashAndSaltPassword(*input.Password)
		if err != nil {
			return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}

		input.PasswordHash = &hashPwd
	}

	if err := ref.repository.UpdateByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if ref.cacheService != nil {
		slog.DebugContext(ctx, "invalidate cache", "user_id", input.ID)

		cacheKey := cache.Identifier{
			Type: "user",
			ID:   input.ID.String(),
		}

		if err := ref.cacheService.Invalidate(ctx, cacheKey); err != nil {
			slog.WarnContext(ctx, "failed to invalidate cache", slog.Any("error", err), "user_id", input.ID)
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user updated successfully",
		attribute.String("user.id", input.ID.String()))

	return nil
}

// errAddressChangeWithoutCaller is what a change of address answers when it
// does not know who asks: a route wired without its caller. It is no domain
// error, so the answer is the fixed 500.
var errAddressChangeWithoutCaller = errors.New("an account's e-mail address was asked to change with no caller")

// addressMayChange refuses a change of an account's e-mail address by anyone
// but the administrator account.
//
// An address is where a password reset is sent, so it is part of how the
// account is held. Only the administrator account, the one with the `admin`
// flag, changes an account's address. The
// flag, not the Administrator role: the role is a set of grants and can be
// given; no route sets the flag.
//
// An address that is the stored one is no change, and is let through: a
// client sends the whole form back. "The stored one" is the same bytes, or
// the stored address in the spelling the service gives every address
// (ValidateEmail, which the request has been through), for a row older than
// that spelling. Nothing looser: a comparison that folds case folds more
// than case, and the statement compares bytes.
//
// And an address that is no change is taken out of the input: the statement
// is never handed one by anybody but the administrator account. The check
// and the write are two statements, and what was "the stored address" at
// the check would be written back over a correction made in between.
//
// Both rows are read from the repository and not through the cache: the
// answer decides who may have the account.
func (ref *UsersService) addressMayChange(ctx context.Context, input *domain.UpdateUserInput) error {
	stored, err := ref.repository.SelectByID(ctx, input.ID)
	if err != nil {
		return err
	}

	if sameAddress(stored.Email, *input.Email) {
		input.Email = nil

		return nil
	}

	if input.CallerID == uuid.Nil() {
		return errAddressChangeWithoutCaller
	}

	// An account that is not there, a flag that was never set, and an
	// account that is disabled are not the administrator's: a disabled
	// account does not sign in, and a token it still holds is not a reason
	// to let it write an address.
	caller, err := ref.repository.SelectByID(ctx, input.CallerID)
	if _, gone := errors.AsType[*domain.UserNotFoundError](err); gone {
		return &domain.AdministratorOnlyError{Entity: "user", Field: domain.FieldEmail}
	}

	if err != nil {
		return err
	}

	if hasAdminFlag(caller) && !isDisabled(caller) {
		return nil
	}

	return &domain.AdministratorOnlyError{Entity: "user", Field: domain.FieldEmail}
}

// sameAddress reports whether sent is the stored address: byte for byte, or
// once the stored one is spelled the way the service spells every address.
func sameAddress(stored, sent string) bool {
	if stored == sent {
		return true
	}

	spelled, err := domain.ValidateEmail(stored, domain.FieldEmail)

	return err == nil && spelled == sent
}

func hasAdminFlag(account *domain.User) bool {
	return account != nil && account.Admin != nil && *account.Admin
}

func isDisabled(account *domain.User) bool {
	return account != nil && account.Disabled != nil && *account.Disabled
}

func (ref *UsersService) DeleteByID(ctx context.Context, input *domain.DeleteUserInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "DeleteByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", input.ID.String()))

	if input.ID == uuid.Nil() {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := ref.repository.DeleteByID(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	// decrement resource usage
	systemScope := domain.ResourcesLimitsScope{
		Type: domain.ResourcesLimitsScopeTypeSystem,
	}

	if err := ref.resourcesLimits.DecrementUsage(ctx, systemScope, domain.ResourcesLimitsResourceTypeUsers); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if ref.cacheService != nil {
		slog.DebugContext(ctx, "invalidate cache", "user_id", input.ID)

		cacheKey := cache.Identifier{
			Type: "user",
			ID:   input.ID.String(),
		}

		if err := ref.cacheService.Invalidate(ctx, cacheKey); err != nil {
			slog.WarnContext(ctx, "failed to invalidate cache", slog.Any("error", err), "user_id", input.ID)
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user deleted successfully", attribute.String("user.id", input.ID.String()))

	return nil
}

func (ref *UsersService) List(ctx context.Context, input *domain.ListUsersInput) (*domain.ListUsersOutput, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "List")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is required"}
		_ = o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
		return nil, errorValue
	}

	span.SetAttributes(
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	out, err := ref.repository.Select(ctx, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "users listed successfully",
		attribute.Int("count", len(out.Items)))

	return out, nil
}

func (ref *UsersService) LinkRoles(ctx context.Context, input *domain.LinkRolesToUserInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "LinkRoles")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if err := ref.guard.CheckRoles(ctx, input.CallerID, input.RoleIDs); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if err := ref.repository.LinkRoles(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if ref.cacheService != nil {
		slog.DebugContext(ctx, "invalidate cache", "user_id", input.UserID.String())

		cacheKeys := []cache.Identifier{
			{
				Type: "user",
				ID:   input.UserID.String(),
			},
			authzCacheKey(input.UserID),
		}

		for _, roleID := range input.RoleIDs {
			cacheKey := cache.Identifier{
				Type: "role",
				ID:   roleID.String(),
			}

			cacheKeys = append(cacheKeys, cacheKey)
		}

		for _, cacheKey := range cacheKeys {
			slog.DebugContext(ctx, "invalidate cache", "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)

			if err := ref.cacheService.Invalidate(ctx, cacheKey); err != nil {
				slog.WarnContext(ctx, "failed to invalidate cache", slog.Any("error", err), "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)
			}
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "roles linked successfully",
		attribute.String("user.id", input.UserID.String()))

	return nil
}

func (ref *UsersService) UnlinkRoles(ctx context.Context, input *domain.UnlinkRolesFromUsersInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "UnlinkRoles")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if err := ref.repository.UnlinkRoles(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if ref.cacheService != nil {
		slog.DebugContext(ctx, "invalidate cache", "user_id", input.UserID.String())

		cacheKeys := []cache.Identifier{
			{
				Type: "user",
				ID:   input.UserID.String(),
			},
			authzCacheKey(input.UserID),
		}

		for _, roleID := range input.RoleIDs {
			cacheKey := cache.Identifier{
				Type: "role",
				ID:   roleID.String(),
			}

			cacheKeys = append(cacheKeys, cacheKey)
		}

		for _, cacheKey := range cacheKeys {
			slog.DebugContext(ctx, "invalidate cache", "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)

			if err := ref.cacheService.Invalidate(ctx, cacheKey); err != nil {
				slog.WarnContext(ctx, "failed to invalidate cache", slog.Any("error", err), "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)
			}
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "roles unlinked successfully",
		attribute.String("user.id", input.UserID.String()))

	return nil
}

func (ref *UsersService) LinkProjects(ctx context.Context, input *domain.LinkProjectsToUserInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "LinkProjects")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if err := ref.repository.LinkProjects(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if ref.cacheService != nil {
		slog.DebugContext(ctx, "invalidate cache", "user_id", input.UserID.String())

		cacheKeys := []cache.Identifier{
			{
				Type: "user",
				ID:   input.UserID.String(),
			},
			authzCacheKey(input.UserID),
		}

		for _, cacheKey := range cacheKeys {
			slog.DebugContext(ctx, "invalidate cache", "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)

			if err := ref.cacheService.Invalidate(ctx, cacheKey); err != nil {
				slog.WarnContext(ctx, "failed to invalidate cache", slog.Any("error", err), "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)
			}
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "projects linked successfully",
		attribute.String("user.id", input.UserID.String()))

	return nil
}

func (ref *UsersService) UnlinkProjects(ctx context.Context, input *domain.UnlinkProjectsFromUserInput) error {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "UnlinkProjects")
	defer span.End()

	if input == nil {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	span.SetAttributes(attribute.String("user.id", input.UserID.String()))

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if err := ref.repository.UnlinkProjects(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	if ref.cacheService != nil {
		slog.DebugContext(ctx, "invalidate cache", "user_id", input.UserID.String())

		cacheKeys := []cache.Identifier{
			{
				Type: "user",
				ID:   input.UserID.String(),
			},
			authzCacheKey(input.UserID),
		}

		for _, cacheKey := range cacheKeys {
			slog.DebugContext(ctx, "invalidate cache", "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)

			if err := ref.cacheService.Invalidate(ctx, cacheKey); err != nil {
				slog.WarnContext(ctx, "failed to invalidate cache", slog.Any("error", err), "cache_type", cacheKey.Type, "cache_key", cacheKey.ID)
			}
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "projects unlinked successfully",
		attribute.String("user.id", input.UserID.String()))

	return nil
}

// ListByRoleID returns a list of users by role ID
func (ref *UsersService) ListByRoleID(ctx context.Context, roleID uuid.UUID, input *domain.ListUsersInput) (*domain.ListUsersOutput, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "ListByRoleID")
	defer span.End()

	span.SetAttributes(attribute.String("role.id", roleID.String()))

	if roleID == uuid.Nil() {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	out, err := ref.repository.SelectByRoleID(ctx, roleID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "users found by role ID",
		attribute.String("role.id", roleID.String()),
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// ListByProjectID returns a list of users by project ID
func (ref *UsersService) ListByProjectID(ctx context.Context, projectID uuid.UUID, input *domain.ListUsersInput) (*domain.ListUsersOutput, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "ListByProjectID")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectID.String()))

	if projectID == uuid.Nil() {
		errorValue := &domain.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, start, errorValue, ref.metrics, attrs)
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	out, err := ref.repository.SelectByProjectID(ctx, projectID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "users found by project ID",
		attribute.String("project.id", projectID.String()),
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// SelectAuthz retrieves the authorization data for a user.
func (ref *UsersService) SelectAuthz(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	start := time.Now()
	ctx, span, attrs := o11y.SetupTrace(ctx, ref.ot.Traces.Tracer, ref.metricsMetadata, "SelectAuthz")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	if userID == uuid.Nil() {
		errorType := &domain.InvalidUserIDError{Message: "user ID cannot be empty"}
		return nil, o11y.RecordError(ctx, span, start, errorType, ref.metrics, attrs)
	}

	var userAuthPermissions map[string]any
	var err error

	// userAuthFetcher is a function that fetches the user authorization data from the repository.
	userAuthFetcher := func(ctx context.Context) (map[string]any, []cache.Identifier, error) {
		authz, err := ref.repository.SelectAuthz(ctx, userID)
		if err != nil {
			return nil, nil, err
		}

		dependencies := make([]cache.Identifier, 0, len(authz.Roles)+len(authz.Policies)+1)

		// add user dependency
		dependencies = append(dependencies, cache.Identifier{Type: "user", ID: userID.String()})

		// add role dependencies
		for _, roleID := range authz.Roles {
			dependencies = append(dependencies, cache.Identifier{Type: "role", ID: roleID})
		}

		// add policy dependencies
		for _, policyID := range authz.Policies {
			dependencies = append(dependencies, cache.Identifier{Type: "policy", ID: policyID})
		}

		return authz.Permissions, dependencies, nil
	}

	// Get the user Authz from the database, cache is disabled
	if ref.cacheService == nil {
		slog.DebugContext(ctx, "cache lookup", "cache", "disabled")

		userAuthPermissions, _, err = userAuthFetcher(ctx)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	} else {

		cacheKey := authzCacheKey(userID)

		slog.DebugContext(ctx, "cache lookup", "cache", "enabled", "key", cacheKey)

		userAuthPermissions, err = cache.GetTyped[map[string]any](ctx, ref.cacheService, cacheKey, userAuthFetcher)
		if err != nil {
			return nil, o11y.RecordError(ctx, span, start, err, ref.metrics, attrs)
		}
	}

	o11y.RecordSuccess(ctx, span, start, ref.metrics, attrs, "user authz found successfully",
		attribute.String("user.id", userID.String()))

	return userAuthPermissions, nil
}
