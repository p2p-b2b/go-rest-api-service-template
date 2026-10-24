package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/respond"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// ProjectMembershipChecker answers whether a user may act on a project.
type ProjectMembershipChecker interface {
	Membership(ctx context.Context, projectID, userID uuid.UUID) (domain.ProjectMembership, error)
}

// noteFailure is [o11y.NoteFailure] for a middleware that answers a request
// itself. NoteFailure records its caller's caller as where the failure arose,
// because it is called through a helper of respond; called straight from a
// middleware's closure, that is net/http's HandlerFunc.ServeHTTP. Through
// this one it is the middleware.
func noteFailure(ctx context.Context, err error) bool {
	return o11y.NoteFailure(ctx, err)
}

// ProjectIDPathValue is the path wildcard every project-scoped route uses.
const ProjectIDPathValue = "project_id"

// RequireProjectMembership refuses a project-scoped request from a caller who
// is neither a member of that project nor an administrator.
//
// The authorization check before this one answers "may this user call GET
// /projects/{any}/embedding_configs", and that is all it can answer: OPA sees
// (user, method, path) and expands the "*" in a grant to any uuid. So one
// project-scoped grant used to open every project -- measured 2026-09-06,
// user A with a grant on /projects/*/embedding_configs listed user B's
// configs by id. Membership is data the database owns; it is asked here,
// once per request, and the handler never has to know.
//
// It keys on the route's project_id path value and does nothing on a route
// that has none, so it sits in the shared authenticated chain. It runs AFTER
// CheckAuthz on purpose: a caller with no grant at all still gets 403; only a
// caller who is authorised for the route but not for this project gets 404 --
// the same answer a missing project gets, so the refusal does not confirm the
// project exists, which is what the projects handler already does.
//
// An administrator is admitted to every project that is there. That is a
// bypass of membership, and it is logged as one, with the project and the
// user. A project that is not there is a 404 for everybody.
func RequireProjectMembership(checker ProjectMembershipChecker) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.PathValue(ProjectIDPathValue)
			if raw == "" {
				next.ServeHTTP(w, r)
				return
			}

			projectID, err := uuid.Parse(raw)
			if err != nil {
				// The handler validates the id and answers 400 with its own
				// wording; an unparsable id cannot name a project to refuse.
				next.ServeHTTP(w, r)
				return
			}

			claims, _ := r.Context().Value(JwtClaims).(map[string]any)
			subStr, _ := claims["sub"].(string)

			userID, err := uuid.Parse(subStr)
			if err != nil {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "invalid sub claim")
				return
			}

			// The tenant boundary, for the access log and the span. It is the
			// one dimension every RAG metric and log line was aggregated
			// across, so a slow or failing project was invisible.
			SetSubjectProject(r.Context(), projectID.String(), false)

			membership, err := checker.Membership(r.Context(), projectID, userID)
			if err != nil {
				// Fail closed. An unreachable membership store must not admit
				// anyone; the same rule as the token denylist and CheckAuthz.
				// Through the status table: the Nil uuid parses and is
				// refused by the use-case as an invalid input, a 400, and it
				// was a 500 here.
				respond.WriteError(w, r, err)
				return
			}

			switch membership {
			case domain.ProjectMembershipAdmin:
				SetSubjectProject(r.Context(), projectID.String(), true)

				slog.InfoContext(r.Context(), "project membership bypassed by an administrator",
					"request_id", respond.RequestIDFrom(r.Context()),
					"project_id", projectID.String(),
					"user_id", userID.String(),
					"method", r.Method,
					"path", r.URL.Path,
				)
			case domain.ProjectMembershipMember:
			default:
				// Not a member, or not a project: one answer, so that the
				// refusal does not say whether the project is there. A
				// refusal is no error of the service: the access line carries
				// it at INFO with its error_type (NoteFailure), and it was a
				// WARN line of its own besides. Logged here only when there
				// is no access line to carry it.
				refusal := &domain.ProjectNotFoundError{ID: projectID}

				if !noteFailure(r.Context(), refusal) {
					slog.InfoContext(r.Context(), "project-scoped request refused",
						"project_id", projectID.String(),
						"user_id", userID.String(),
						"method", r.Method,
						"path", r.URL.Path,
					)
				}

				respond.WriteError(w, r, refusal)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
