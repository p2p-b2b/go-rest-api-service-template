package respond

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// StatusFor is the HTTP status a domain error is answered with, or 500 for an
// error it does not recognise.
//
// It is the one place that says which side of 500 each error is on: what
// the caller got wrong, or what broke. The handlers still carry their own
// arms (an `errors.AsType` per type and per method) and answer through them;
// this table was read off those arms, and TestEveryArmAgreesWithTheTable
// holds every arm to the side the table says: an arm that answers a type
// below 500 finds it below 500 here, and an arm that answers it 500 finds it
// at 500 or not here at all. What reads the table is what must agree with
// the answers without being one of them: IsRefusal, which tells a refused
// call from a failed one.
//
// Where the arms answer one type with two statuses on the same side (a 404,
// and a 200 where a policy's delete is idempotent or a verification resend
// will not say whether an account exists), the table holds the one most of
// them answer; its side is what is relied on. An error the handlers build
// themselves and answer without an arm (a path id that is not a uuid) is
// in it too, so that it is not taken for a failure.
func StatusFor(err error) int {
	for _, row := range statusTable {
		if row.match(err) {
			return row.status
		}
	}

	return http.StatusInternalServerError
}

// WriteError answers err with [StatusFor]: its message with that status, or
// the fixed 500 body with the cause on the request's log line.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status := StatusFor(err)
	if status == http.StatusInternalServerError {
		WriteInternalError(w, r, err)

		return
	}

	WriteJSONMessage(w, r, status, err.Error())
}

// statusRow is one type of the table: its name as a handler's arm writes
// it, how an error in a chain is matched to it, and its status.
type statusRow struct {
	match  func(error) bool
	name   string
	status int
}

// row is the row of the error type T.
func row[T error](status int) statusRow {
	var zero T

	return statusRow{
		name:   fmt.Sprintf("%T", zero),
		match:  func(err error) bool { _, ok := errors.AsType[T](err); return ok },
		status: status,
	}
}

// statusTable is checked in order; a wrapped error matches the first row any
// error in its chain matches.
var statusTable = []statusRow{
	// 400: the request itself is wrong: a field, an id, a value Postgres
	// refused.
	row[*domain.ValidationError](http.StatusBadRequest),
	row[*domain.ValidationErrors](http.StatusBadRequest),
	row[*domain.InvalidInputError](http.StatusBadRequest),
	row[*domain.InvalidEmailError](http.StatusBadRequest),
	row[*domain.InvalidPasswordError](http.StatusBadRequest),
	row[*domain.InvalidActionError](http.StatusBadRequest),
	row[*domain.InvalidResourceError](http.StatusBadRequest),
	row[*domain.InvalidIdentityProvidersError](http.StatusBadRequest),
	row[*domain.ProjectNotOfCallerError](http.StatusBadRequest),
	row[*domain.InvalidRateLimitTargetError](http.StatusBadRequest),
	row[*domain.InvalidRateLimitStrategyError](http.StatusBadRequest),
	// Built by the handlers' own parsers and answered 400: a path id that
	// is not a uuid, a page token or a limit that does not parse, and a
	// body the authentication routes could not read.
	row[*domain.InvalidUUIDError](http.StatusBadRequest),
	row[*domain.InvalidTokenError](http.StatusBadRequest),
	row[*domain.InvalidLimitError](http.StatusBadRequest),
	row[*domain.InvalidRequestError](http.StatusBadRequest),
	// Built by the payloads' Validate and answered 400 by the handler that
	// asked: an update that sets nothing, a link body that names nothing.
	row[*domain.InvalidUserUpdateError](http.StatusBadRequest),
	row[*domain.InvalidIDPUpdateError](http.StatusBadRequest),
	row[*domain.InvalidProjectLinkError](http.StatusBadRequest),
	row[*domain.DatatypeMismatchError](http.StatusBadRequest),
	row[*domain.InvalidByteSequenceError](http.StatusBadRequest),
	row[*domain.InvalidMessageFormatError](http.StatusBadRequest),
	row[*domain.UndefinedColumnError](http.StatusBadRequest),
	// A provider type named in a provider's body (four routes); the type's
	// own route answers it 404.
	row[*domain.IDPTypesNotFoundError](http.StatusBadRequest),

	// 401: the caller is not who they need to be.
	row[*domain.InvalidJWTError](http.StatusUnauthorized),
	row[*domain.InvalidRefreshTokenError](http.StatusUnauthorized),
	row[*domain.InvalidCredentialsError](http.StatusUnauthorized),
	row[*domain.InvalidAuthnServiceError](http.StatusUnauthorized),
	row[*domain.IDPIdentityNotLinkedError](http.StatusUnauthorized),
	row[*domain.IDPChangedDuringSignInError](http.StatusUnauthorized),

	// 403: the caller may not do this to this row.
	row[*domain.AdministratorOnlyError](http.StatusForbidden),
	row[*domain.AdministratorProtectedError](http.StatusForbidden),
	row[*domain.GrantNotHeldError](http.StatusForbidden),
	row[*domain.IDPProtectedError](http.StatusForbidden),
	row[*domain.SystemPolicyError](http.StatusForbidden),
	row[*domain.SystemProjectError](http.StatusForbidden),
	row[*domain.SystemRateLimitError](http.StatusForbidden),
	row[*domain.SystemRoleError](http.StatusForbidden),

	// 404: what the request names is not there.
	row[*domain.UserNotFoundError](http.StatusNotFound),
	row[*domain.ProjectNotFoundError](http.StatusNotFound),
	row[*domain.ProductNotFoundError](http.StatusNotFound),
	row[*domain.RoleNotFoundError](http.StatusNotFound),
	row[*domain.PolicyNotFoundError](http.StatusNotFound),
	row[*domain.ResourceNotFoundError](http.StatusNotFound),
	row[*domain.RateLimitNotFoundError](http.StatusNotFound),
	row[*domain.IDPNotFoundError](http.StatusNotFound),
	row[*domain.UserIdentityNotFoundError](http.StatusNotFound),

	// 409: the request conflicts with what is stored.
	row[*domain.UserAlreadyExistsError](http.StatusConflict),
	row[*domain.UserAlreadyVerifiedError](http.StatusConflict),
	row[*domain.ProjectAlreadyExistsError](http.StatusConflict),
	row[*domain.ProductAlreadyExistsError](http.StatusConflict),
	row[*domain.RoleAlreadyExistsError](http.StatusConflict),
	row[*domain.PolicyAlreadyExistsError](http.StatusConflict),
	row[*domain.RateLimitAlreadyExistsError](http.StatusConflict),
	row[*domain.IDPAlreadyExistsError](http.StatusConflict),
	row[*domain.IDPInUseError](http.StatusConflict),
	row[*domain.UserIdentityAlreadyLinkedError](http.StatusConflict),
	row[*domain.ResourcesLimitsHardLimitReachedError](http.StatusConflict),

	// 429: the caller is asking too often.
	row[*domain.TooManyLoginAttemptsError](http.StatusTooManyRequests),
	row[*domain.TooManyRecoveryRequestsError](http.StatusTooManyRequests),

	// 503: a dependency of the service's is not there. A failure, not a
	// refusal: it is in the table so that its answer is not the 500.
	row[*domain.IDPUnreachableError](http.StatusServiceUnavailable),
}
