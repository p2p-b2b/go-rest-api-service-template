//go:build unit

package app

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"
	"strings"
	"testing"
)

// withoutComments is a Go file as the compiler reads it: parsed, and printed
// again with no comment. What is read from it is code, never a sentence
// about code, and a check that was commented out is not there.
func withoutComments(t *testing.T, path string) string {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	var out bytes.Buffer

	if err := (&printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}).Fprint(&out, fset, file); err != nil {
		t.Fatalf("print %s: %v", path, err)
	}

	return out.String()
}

// A token that verifies says nothing about its account. The check that
// refuses a disabled or deleted account (middleware.RequireActiveAccount) was
// on the /me chain alone, so after an account was disabled its tokens kept
// answering on every other route.
//
// This reads server.go for the chains themselves: every chain that verifies
// an access token asks about the account, after the token checks and before
// the grants are looked up, except logout, where a disabled account must
// still be able to end the tokens it holds. A chain added without it, or the
// check dropped from one, compiles and passes every other unit test.
func TestEveryAccessTokenChainAsksAboutTheAccount(t *testing.T) {
	t.Parallel()

	raw := withoutComments(t, "server.go")

	// Each chain: its name, and the middlewares it lists, in order.
	chain := regexp.MustCompile(`(?s)\n\t(\w+) := middleware\.Chain\(\n(.*?)\n\t\)\n`)
	call := regexp.MustCompile(`middleware\.(\w+)\(`)

	withAccount := map[string]bool{"accessTokenMiddlewares": true, "meEndpointMiddlewares": true}
	without := map[string]bool{"logoutMiddlewares": true}

	seen := 0

	for _, m := range chain.FindAllStringSubmatch(raw, -1) {
		name, body := m[1], m[2]

		// Comments name middlewares too; only the calls count.
		var code []string

		for line := range strings.SplitSeq(body, "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "//") {
				code = append(code, line)
			}
		}

		var order []string
		for _, c := range call.FindAllStringSubmatch(strings.Join(code, "\n"), -1) {
			order = append(order, c[1])
		}

		position := func(of string) int {
			for i, n := range order {
				if n == of {
					return i
				}
			}

			return -1
		}

		if position("CheckAccessToken") < 0 {
			continue // refresh, password reset, verification: not an access token
		}

		seen++

		account := position("RequireActiveAccount")

		switch {
		case withAccount[name]:
			if account < 0 {
				t.Errorf("%s verifies an access token and never asks whether its account may act: %v", name, order)

				continue
			}

			if account < position("CheckAccessToken") {
				t.Errorf("%s asks about the account before the token is verified: %v", name, order)
			}

			if authz := position("CheckAuthz"); authz < 0 || account > authz {
				t.Errorf("%s looks the grants up before it knows the account may act: %v", name, order)
			}
		case without[name]:
			if account >= 0 {
				t.Errorf("%s refuses a disabled account: it could not end the tokens it holds: %v", name, order)
			}
		default:
			t.Errorf("%s verifies an access token and this test does not know it: say here whether it asks about the account", name)
		}
	}

	if seen != len(withAccount)+len(without) {
		t.Errorf("found %d chains that verify an access token, want %d: the chains changed, or this parser no longer reads them", seen, len(withAccount)+len(without))
	}

	source := raw

	// The loop above reads chains written as a middleware.Chain(...) literal.
	// A chain built any other way (by append, through a helper, through
	// another name for the check) that verifies an access token is one it
	// never saw: the token check is named three times in this file, once
	// for each chain above, and each time it is called.
	if got := strings.Count(source, "middleware.CheckAccessToken"); got != seen {
		t.Errorf("middleware.CheckAccessToken is named %d times and %d chains were read: a chain verifies an access token where this test does not look", got, seen)
	}

	if got := strings.Count(source, "middleware.CheckAccessToken("); got != seen {
		t.Errorf("middleware.CheckAccessToken is called %d times and named %d: it is handed on under another name", got, seen)
	}

	// A chain is what it was built as: assigned once, where it is built,
	// and never made to be another afterwards.
	for name := range map[string]bool{"accessTokenMiddlewares": true, "meEndpointMiddlewares": true, "logoutMiddlewares": true} {
		if got := len(regexp.MustCompile(`\b`+name+`\s*:?=[^=]`).FindAllString(source, -1)); got != 1 {
			t.Errorf("%s is assigned %d times, want once: a chain was made to be another", name, got)
		}
	}

	// The one chain that does not ask about the account is named twice:
	// where it is built, and where it is handed to the one handler that
	// uses it. A third mention is another name for it, or another route.
	if got := len(regexp.MustCompile(`\blogoutMiddlewares\b`).FindAllString(source, -1)); got != 2 {
		t.Errorf("logoutMiddlewares is named %d times, want twice: built, and handed to the authentication handler", got)
	}

	// The chains are handed to the routes by their names. The one chain that
	// does not ask about the account is given to one registration, the
	// authentication handler's, which puts the logout route on it: whatever
	// else were on it would not be asked whether its account may act.
	handed := regexp.MustCompile(`(?s)\n\ta\.handlers\.(\w+)\.RegisterRoutes\((.*?)\)\n`)

	var withLogout []string

	for _, m := range handed.FindAllStringSubmatch(source, -1) {
		handler, args := m[1], m[2]

		if strings.Contains(args, "logoutMiddlewares") {
			withLogout = append(withLogout, handler)
		}

		// /me is read on its own chain, and it is one that asks.
		if handler == "Me" && !strings.Contains(args, "meEndpointMiddlewares") {
			t.Errorf("the /me routes are not on meEndpointMiddlewares: %s", args)
		}
	}

	if len(withLogout) != 1 || withLogout[0] != "Authn" {
		t.Errorf("the logout chain is given to %v, want the authentication handler alone", withLogout)
	}

	// And that handler puts one route on it, the logout: the chain reaches
	// it as a parameter, which is named where it is declared and where the
	// logout route takes it, and nowhere else.
	authn := withoutComments(t, "../adapter/driving/http/handler/authn.go")

	if got := len(regexp.MustCompile(`\blogoutMiddleware\b`).FindAllString(authn, -1)); got != 2 {
		t.Errorf("the authentication handler names its logout chain %d times, want twice: its parameter, and the logout route", got)
	}

	if !strings.Contains(authn, `mux.Handle("DELETE /auth/logout", logoutMiddleware.ThenFunc(ref.logout))`) {
		t.Error("the logout route is not the one on the logout chain: this test needs reading again")
	}
}
