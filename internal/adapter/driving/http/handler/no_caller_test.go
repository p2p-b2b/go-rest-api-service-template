//go:build unit

package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/respond"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// TestARequestWithoutACallerIsAFaultAtEverySite: thirty-three handler
// functions ask the request for the account it was authenticated as, and two
// for the verified token itself. The authentication middleware puts both on
// every request it lets through, so a request without them reached the
// handler past a missing or broken middleware: a fault of the service.
//
// Each site is driven with nothing on the request and must answer the fixed
// 500, say nothing of the cause, and record a failure. Eleven answered 500
// for an error the status table called a 401, so the metric counted a
// refusal; eighteen answered 400 or 401, most of them with the error's
// text, which told a caller their token was at fault; and six read the
// claims by hand, one of them writing its reason into a 500's body.
func TestARequestWithoutACallerIsAFaultAtEverySite(t *testing.T) {
	t.Parallel()

	ot, metrics := testTelemetry(t)
	id := func() string { return uuid.NewV7().String() }
	ids := func(key string) string { return `{"` + key + `":["` + id() + `"]}` }

	authn := &AuthnHandler{ot: ot, metrics: metrics}
	idp := &AuthnIDPsHandler{ot: ot, metrics: metrics}
	idps := &IDPsHandler{ot: ot, metrics: metrics}
	me := &MeHandler{ot: ot, metrics: metrics}
	policies := &PoliciesHandler{ot: ot, metrics: metrics}
	products := &ProductsHandler{ot: ot, metrics: metrics}
	projects := &ProjectsHandler{ot: ot, metrics: metrics}
	limits := &ResourcesLimitsHandler{ot: ot, metrics: metrics}
	roles := &RolesHandler{ot: ot, metrics: metrics}
	lifetimes := &TokenLifetimesHandler{ot: ot, metrics: metrics}
	users := &UsersHandler{ot: ot, metrics: metrics}

	sites := []struct {
		name string
		call http.HandlerFunc
		// what the handler reads before it asks for the caller
		path map[string]string
		body string
	}{
		// The two that act on the verified token itself, not on whose it is.
		{name: "AuthnHandler.confirmVerification", call: authn.confirmVerification},
		{name: "AuthnHandler.refreshAccessToken", call: authn.refreshAccessToken},
		{name: "AuthnHandler.logout", call: authn.logout},
		{name: "AuthnHandler.resetPassword", call: authn.resetPassword, body: `{"password":"NewSecureP@ssw0rd1"}`},
		{name: "AuthnIDPsHandler.start", call: idp.link, path: map[string]string{"idp_id": id()}},
		{name: "AuthnIDPsHandler.listIdentities", call: idp.listIdentities},
		{name: "AuthnIDPsHandler.unlinkIdentity", call: idp.unlinkIdentity, path: map[string]string{"idp_id": id()}},
		{name: "IDPsHandler.updateByID", call: idps.updateByID, path: map[string]string{"idp_id": id()}, body: `{"enabled":false}`},
		{name: "MeHandler.authz", call: me.authz},
		{name: "MeHandler.update", call: me.update},
		{name: "MeHandler.getByID", call: me.getByID},
		{name: "PoliciesHandler.create", call: policies.create, body: `{"name":"Project Products Manager","description":"Allows managing the products of a project","allowed_action":"POST","allowed_resource":"/projects/*/products"}`},
		{name: "PoliciesHandler.updateByID", call: policies.updateByID, path: map[string]string{"policy_id": id()}, body: `{"name":"Updated Policy Name"}`},
		{name: "PoliciesHandler.linkRoles", call: policies.linkRoles, path: map[string]string{"policy_id": id()}, body: ids("role_ids")},
		{name: "ProductsHandler.productPathParams", call: products.createByProjectID, path: map[string]string{"project_id": id()}},
		{name: "ProductsHandler.list", call: products.list, path: map[string]string{"project_id": id()}},
		{name: "ProjectsHandler.create", call: projects.create},
		{name: "ProjectsHandler.getByIDByUserID", call: projects.getByIDByUserID, path: map[string]string{"project_id": id()}},
		{name: "ProjectsHandler.updateByID", call: projects.updateByID, path: map[string]string{"project_id": id()}},
		{name: "ProjectsHandler.deleteByID", call: projects.deleteByID, path: map[string]string{"project_id": id()}},
		{name: "ProjectsHandler.listByUserID", call: projects.listByUserID, path: map[string]string{"user_id": id()}},
		{name: "ProjectsHandler.list", call: projects.list},
		{name: "ResourcesLimitsHandler.statusForMe", call: limits.statusForMe},
		{name: "ResourcesLimitsHandler.list", call: limits.list},
		{name: "RolesHandler.deleteByID", call: roles.deleteByID, path: map[string]string{"role_id": id()}},
		{name: "RolesHandler.linkUsers", call: roles.linkUsers, path: map[string]string{"role_id": id()}, body: ids("user_ids")},
		{name: "RolesHandler.unlinkUsers", call: roles.unlinkUsers, path: map[string]string{"role_id": id()}, body: ids("user_ids")},
		{name: "RolesHandler.linkPolicies", call: roles.linkPolicies, path: map[string]string{"role_id": id()}, body: ids("policy_ids")},
		{name: "TokenLifetimesHandler.update", call: lifetimes.update},
		{name: "UsersHandler.updateByID", call: users.updateByID, path: map[string]string{"user_id": id()}},
		{name: "UsersHandler.deleteByID", call: users.deleteByID, path: map[string]string{"user_id": id()}},
		{name: "UsersHandler.linkRoles", call: users.linkRoles, path: map[string]string{"user_id": id()}, body: ids("role_ids")},
		{name: "UsersHandler.unlinkRoles", call: users.unlinkRoles, path: map[string]string{"user_id": id()}, body: ids("role_ids")},
		{name: "UsersHandler.linkProjects", call: users.linkProjects, path: map[string]string{"user_id": id()}, body: ids("project_ids")},
		{name: "UsersHandler.unlinkProjects", call: users.unlinkProjects, path: map[string]string{"user_id": id()}, body: ids("project_ids")},
	}

	for _, site := range sites {
		t.Run(site.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(site.body))
			req.Header.Set("Content-Type", "application/json")

			for key, value := range site.path {
				req.SetPathValue(key, value)
			}

			req = req.WithContext(o11y.WithFailure(req.Context()))

			rec := httptest.NewRecorder()
			site.call(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500. Body: %s", rec.Code, rec.Body.String())
			}

			body := rec.Body.String()
			if !strings.Contains(body, respond.InternalServerErrorMessage) {
				t.Errorf("the body is not the fixed one: %s", body)
			}

			for _, leak := range []string{"caller", "claim", "JWT", "token", "sub", "context"} {
				if strings.Contains(body, leak) {
					t.Errorf("the body says what was missing (%q): %s", leak, body)
				}
			}

			recorded, ok := o11y.TakeFailure(req.Context())
			if !ok {
				t.Fatal("nothing was recorded for a 500")
			}

			if _, isIt := errors.AsType[*noCallerError](recorded.Err); !isIt {
				t.Fatalf("recorded %T (%v), want the missing caller: the request did not get as far as asking for it", recorded.Err, recorded.Err)
			}

			if respond.IsRefusal(recorded.Err) {
				t.Error("a missing caller is counted as a refused request")
			}
		})
	}

	// Every site, and the next one somebody adds.
	call := regexp.MustCompile(`\bget(UserID|Token)FromContext\(`)

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	inSource := 0

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		inSource += len(call.FindAll(src, -1))
	}

	// The two definitions are among the matches.
	if inSource-2 != len(sites) {
		t.Errorf("%d call sites of the two helpers in the package, %d in this test: add the new one here", inSource-2, len(sites))
	}
}

// TestEverySiteAnswersAMissingCallerWithTheFixed500: the same rule read from
// the source, for the shape the test above cannot see: a site that asks for
// the caller and answers its absence through anything but
// respond.WriteInternalError.
func TestEverySiteAnswersAMissingCallerWithTheFixed500(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(string(src), "\n")

		for i, line := range lines {
			asks := strings.Contains(line, "getUserIDFromContext(") || strings.Contains(line, "getTokenFromContext(")
			if !asks || strings.HasPrefix(line, "func ") {
				continue
			}

			// The answer is within the `if err != nil { … }` that follows.
			window := strings.Join(lines[i+1:min(i+8, len(lines))], "\n")
			if !strings.Contains(window, "respond.WriteInternalError(w, r, e)") {
				t.Errorf("%s:%d asks for the caller and does not answer its absence with respond.WriteInternalError", file, i+1)
			}
		}
	}
}
