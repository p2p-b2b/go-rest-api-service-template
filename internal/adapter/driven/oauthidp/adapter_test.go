//go:build unit

package oauthidp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/oauth"
)

// fakeOIDC is an OpenID Connect provider in a test server: discovery, JWKS, a
// token endpoint that mints an RS256 ID token, and user-info. It is a REAL
// provider as far as go-oidc can tell, which is the point -- a mock replaying
// this package's own structs would assert nothing about discovery, PKCE, the
// nonce or the signature.
type fakeOIDC struct {
	srv    *httptest.Server
	key    *rsa.PrivateKey
	claims map[string]any // what the next ID token and user-info carry
	// what the token endpoint saw, so a test can assert PKCE reached it
	gotVerifier string
	gotCode     string
	nonce       string
	// tokenAnswer, when set, is what the token endpoint does instead of
	// issuing a token.
	tokenAnswer func(w http.ResponseWriter)
	// userInfoStatus, when set, is what the user-info endpoint answers.
	userInfoStatus int
	// authMethods is what the discovery document declares for the token
	// endpoint; nil leaves the key out.
	authMethods []string
	// How the client presented itself, and how often.
	tokenHits     int
	gotBasicAuth  bool
	gotFormSecret string
}

func newFakeOIDC(t *testing.T) *fakeOIDC {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	f := &fakeOIDC{key: key, claims: map[string]any{}}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		document := map[string]any{
			"issuer":                                f.srv.URL,
			"authorization_endpoint":                f.srv.URL + "/authorize",
			"token_endpoint":                        f.srv.URL + "/token",
			"userinfo_endpoint":                     f.srv.URL + "/userinfo",
			"jwks_uri":                              f.srv.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"code_challenge_methods_supported":      []string{"S256"},
		}
		if f.authMethods != nil {
			document["token_endpoint_auth_methods_supported"] = f.authMethods
		}

		_ = json.NewEncoder(w).Encode(document)
	})

	mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, _ *http.Request) {
		pub := &key.PublicKey
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{
			"kty": "RSA", "kid": "test-key", "use": "sig", "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}}})
	})

	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.gotVerifier = r.Form.Get("code_verifier")
		f.gotCode = r.Form.Get("code")
		f.tokenHits++
		_, _, f.gotBasicAuth = r.BasicAuth()
		f.gotFormSecret = r.Form.Get("client_secret")

		if f.tokenAnswer != nil {
			f.tokenAnswer(w)

			return
		}

		claims := jwt.MapClaims{
			"iss": f.srv.URL, "aud": "client-id", "sub": "subject-1",
			"iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(), "nonce": f.nonce,
		}
		for k, v := range f.claims {
			claims[k] = v
		}

		tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tok.Header["kid"] = "test-key"

		signed, err := tok.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "at-1", "token_type": "Bearer", "expires_in": 3600, "id_token": signed,
		})
	})

	mux.HandleFunc("GET /userinfo", func(w http.ResponseWriter, _ *http.Request) {
		if f.userInfoStatus != 0 {
			w.WriteHeader(f.userInfoStatus)

			return
		}

		out := map[string]any{"sub": "subject-1"}
		for k, v := range f.claims {
			out[k] = v
		}

		_ = json.NewEncoder(w).Encode(out)
	})

	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)

	return f
}

func newProvider(t *testing.T) *Provider {
	t.Helper()

	p, err := New(Config{HTTPClient: &http.Client{Timeout: 5 * time.Second}})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func oidcIDP(issuer, typeName string) *domain.IDP {
	return &domain.IDP{
		Name: "Test " + typeName, ClientID: "client-id", ClientSecret: "secret",
		CallbackURL: "https://app.example/auth/idp/x/callback", IssuerURL: issuer, Enabled: true,
		IDPType: domain.IDPTypes{Name: typeName, Kind: domain.IDPTypeKindOIDC, Scopes: []string{"openid", "email", "profile"}},
	}
}

// The authorization URL carries what the flow needs on the way out: the
// state, an S256 challenge derived from the verifier (never the verifier), the
// nonce, and the openid scope -- and none of the Google-only flags.
func TestOIDCAuthCodeURLCarriesPKCEAndNonce(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	p := newProvider(t)

	raw, err := p.AuthCodeURL(t.Context(), oidcIDP(f.srv.URL, "Okta"), oauth.AuthRequest{State: "st", Nonce: "nn", CodeVerifier: "verifier-verifier-verifier-verifier-verifier"})
	if err != nil {
		t.Fatalf("AuthCodeURL: %v", err)
	}

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}

	q := u.Query()

	if !strings.HasPrefix(raw, f.srv.URL+"/authorize") {
		t.Errorf("must go to the DISCOVERED authorization endpoint, got %s", raw)
	}

	if q.Get("state") != "st" || q.Get("nonce") != "nn" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
		t.Errorf("missing state/nonce/PKCE: %s", q.Encode())
	}

	if strings.Contains(raw, "verifier-verifier") {
		t.Error("the PKCE verifier must never appear in the authorization URL")
	}

	if !strings.Contains(q.Get("scope"), "openid") {
		t.Errorf("openid scope missing: %q", q.Get("scope"))
	}

	if q.Get("access_type") != "" || q.Get("prompt") != "" {
		t.Errorf("Google-only options leaked into a generic provider's URL: %s", q.Encode())
	}
}

// The exchange presents the verifier, verifies the ID token against the
// discovered keys, checks the nonce, and maps the claims.
func TestOIDCExchangeVerifiesTheIDTokenAndMapsClaims(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "nonce-1"
	f.claims = map[string]any{"email": "Jane@Example.com", "email_verified": true, "given_name": "Jane", "family_name": "Doe"}

	p := newProvider(t)

	info, err := p.Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1", oauth.AuthRequest{Nonce: "nonce-1", CodeVerifier: "verifier-verifier-verifier-verifier-verifier"})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if f.gotVerifier != "verifier-verifier-verifier-verifier-verifier" || f.gotCode != "code-1" {
		t.Errorf("token endpoint saw verifier=%q code=%q", f.gotVerifier, f.gotCode)
	}

	want := domain.UserInfo{Subject: "subject-1", Email: "jane@example.com", EmailVerified: true, FirstName: "Jane", LastName: "Doe"}
	if *info != want {
		t.Errorf("UserInfo = %+v, want %+v", *info, want)
	}
}

// A nonce that does not match is a token minted for another request.
func TestOIDCExchangeRefusesAWrongNonce(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "minted-for-someone-else"
	f.claims = map[string]any{"email": "jane@example.com", "email_verified": true}

	p := newProvider(t)

	_, err := p.Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1", oauth.AuthRequest{Nonce: "mine", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})
	if _, ok := errors.AsType[*domain.IDPSignInRefusedError](err); !ok {
		t.Fatalf("a nonce mismatch is the caller's side, a refused sign-in: got %T (%v)", err, err)
	}

	if strings.Contains(err.Error(), "oidc:") {
		t.Errorf("the library's wording leaked: %v", err)
	}
}

// An ID token from another issuer -- same vendor, another tenant -- does not
// verify against the row's issuer.
func TestOIDCExchangeRefusesAnotherIssuer(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "n"
	f.claims = map[string]any{"iss": "https://login.microsoftonline.com/another-tenant/v2.0", "email": "x@example.com"}

	p := newProvider(t)

	_, err := p.Exchange(t.Context(), oidcIDP(f.srv.URL, "EntraID"), "c", oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})
	if _, ok := errors.AsType[*domain.InvalidAuthnServiceError](err); !ok {
		t.Fatalf("a token that does not verify is the provider's side: got %T (%v)", err, err)
	}
}

// Entra ID: no email_verified claim, email possibly absent from the token and
// present as preferred_username. Single tenant means the address is the
// directory's own attribute, so it is verified by construction.
func TestEntraIDMapsPreferredUsernameAndTrustsTheTenant(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "n"
	f.claims = map[string]any{"preferred_username": "jane@corp.example", "name": "Jane Doe"}

	p := newProvider(t)

	info, err := p.Exchange(t.Context(), oidcIDP(f.srv.URL, "EntraID"), "c", oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if info.Email != "jane@corp.example" || !info.EmailVerified || info.FirstName != "Jane" || info.LastName != "Doe" {
		t.Errorf("UserInfo = %+v", *info)
	}
}

// Google and Okta assert email_verified; false is passed through as false.
func TestOIDCPassesAnUnverifiedEmailThroughAsUnverified(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "n"
	f.claims = map[string]any{"email": "x@example.com", "email_verified": false, "name": "Cher"}

	p := newProvider(t)

	info, err := p.Exchange(t.Context(), oidcIDP(f.srv.URL, "Google"), "c", oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if info.EmailVerified {
		t.Error("email_verified=false must not become true")
	}

	// A single-word name is a first name, not an error.
	if info.FirstName != "Cher" || info.LastName != "Cher" {
		t.Errorf("name split = %q %q", info.FirstName, info.LastName)
	}
}

// GitHub: the identity is the numeric id, the email the primary VERIFIED
// address from /user/emails, and a null /user.email is not a failure.
func TestGithubReadsThePrimaryVerifiedEmail(t *testing.T) {
	// Not parallel: it swaps the package-level GitHub endpoint for the fake.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/oauth/access_token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "gh-at", "token_type": "bearer"})
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 4242, "login": "octocat", "name": "The Octocat", "email": nil})
	})
	mux.HandleFunc("GET /user/emails", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"email": "old@example.com", "primary": false, "verified": true},
			{"email": "Octo@Example.com", "primary": true, "verified": true},
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := newProvider(t)

	// The github kind uses fixed endpoints; point them at the fake.
	saved := githubEndpoint
	githubEndpoint = oauth2.Endpoint{AuthURL: srv.URL + "/login/oauth/authorize", TokenURL: srv.URL + "/login/oauth/access_token"}
	t.Cleanup(func() { githubEndpoint = saved })

	idp := &domain.IDP{
		Name: "GH", ClientID: "cid", ClientSecret: "sec", CallbackURL: "https://app.example/cb", Enabled: true,
		IDPType: domain.IDPTypes{Name: "Github", Kind: domain.IDPTypeKindGithub, UserInfoAPIURL: srv.URL + "/user", Scopes: []string{"read:user", "user:email"}},
	}

	info, err := p.Exchange(t.Context(), idp, "code", oauth.AuthRequest{CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}

	want := domain.UserInfo{Subject: "4242", Email: "octo@example.com", EmailVerified: true, FirstName: "The", LastName: "Octocat"}
	if *info != want {
		t.Errorf("UserInfo = %+v, want %+v", *info, want)
	}
}

// The kind decides, never the name: a row named anything works for its kind,
// and an unknown kind is refused before any network call.
func TestKindDecidesNotTheName(t *testing.T) {
	t.Parallel()

	p := newProvider(t)

	idp := oidcIDP("http://127.0.0.1:1", "Company SSO")
	idp.IDPType.Kind = "saml"

	if _, err := p.AuthCodeURL(t.Context(), idp, oauth.AuthRequest{State: "s", Nonce: "n", CodeVerifier: "v"}); err == nil {
		t.Fatal("an unknown kind must be refused")
	}
}

// A provider row this service cannot sign in with is the service's side,
// in its words: a 503 that names nothing. It was a 400 naming the provider,
// on the public routes that start a sign-in and take its callback.
func TestAProviderThatCannotBeUsedIsTheServicesSide(t *testing.T) {
	t.Parallel()

	const name = "Company SSO"

	for shape, spoil := range map[string]func(*domain.IDP){
		"no_client_secret": func(idp *domain.IDP) { idp.ClientSecret = "" },
		"no_issuer":        func(idp *domain.IDP) { idp.IssuerURL = "" },
		"an_unknown_kind":  func(idp *domain.IDP) { idp.IDPType.Kind = "saml" },
	} {
		t.Run(shape, func(t *testing.T) {
			t.Parallel()

			idp := oidcIDP("http://127.0.0.1:1", name)
			spoil(idp)

			_, startErr := newProvider(t).AuthCodeURL(t.Context(), idp, oauth.AuthRequest{State: "s", Nonce: "n", CodeVerifier: "v"})
			_, exchangeErr := newProvider(t).Exchange(t.Context(), idp, "code", oauth.AuthRequest{Nonce: "n", CodeVerifier: "v"})

			for step, err := range map[string]error{"start": startErr, "exchange": exchangeErr} {
				refused, ok := errors.AsType[*domain.InvalidAuthnServiceError](err)
				if !ok {
					t.Fatalf("%s: got %T (%v), want the service's side", step, err, err)
				}

				if refused.Message != domain.IDPNotConfiguredMessage || strings.Contains(err.Error(), name) {
					t.Errorf("%s: answered %q, want %q and no provider name", step, err, domain.IDPNotConfiguredMessage)
				}
			}
		})
	}
}

// New refuses a client without a timeout: the token exchange used to run on
// http.DefaultClient, which has none.
func TestNewRequiresATimeout(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{HTTPClient: &http.Client{}}); err == nil {
		t.Fatal("a client without a timeout must be refused")
	}

	if _, err := New(Config{}); err == nil {
		t.Fatal("a nil client must be refused")
	}
}

// A failed code exchange is one of three things, and they are answered
// differently: the person's sign-in is refused (401), the provider or this
// service's registration at it is wrong (503), the provider did not answer
// (503). They were one error answered 401, so a provider that was down read
// as "sign-in refused" and a cancelled sign-in was counted as a failed call.
func TestExchangeFailureSaysWhoseItIs(t *testing.T) {
	t.Parallel()

	const (
		refused     = "*domain.IDPSignInRefusedError"
		misbehaving = "*domain.InvalidAuthnServiceError"
		unreachable = "*domain.IDPUnreachableError"
	)

	for name, tt := range map[string]struct {
		err  error
		want string
	}{
		"a_code_that_is_spent_or_expired":       {&oauth2.RetrieveError{ErrorCode: "invalid_grant"}, refused},
		"githubs_word_for_the_same":             {&oauth2.RetrieveError{ErrorCode: "bad_verification_code"}, refused},
		"github_will_not_vouch_for_the_address": {&oauth2.RetrieveError{ErrorCode: "unverified_user_email"}, refused},
		"this_services_client_is_rejected":      {&oauth2.RetrieveError{ErrorCode: "invalid_client"}, misbehaving},
		"a_client_not_allowed_this_grant":       {&oauth2.RetrieveError{ErrorCode: "unauthorized_client"}, misbehaving},
		"a_5xx_with_no_oauth_error":             {&oauth2.RetrieveError{}, misbehaving},
		"an_answer_that_is_not_a_token":         {errors.New("oauth2: cannot parse json: invalid character"), misbehaving},
		"a_policy_that_turns_the_person_away":   {&oauth2.RetrieveError{ErrorCode: "access_denied"}, refused},
		"githubs_word_for_a_rejected_client":    {&oauth2.RetrieveError{ErrorCode: "incorrect_client_credentials"}, misbehaving},
		"no_answer":                             {&url.Error{Op: "Post", URL: "https://idp.example/token", Err: errors.New("connection refused")}, unreachable},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := exchangeFailure(tt.err)
			if typ := fmt.Sprintf("%T", got); typ != tt.want {
				t.Fatalf("got %s (%v), want %s", typ, got, tt.want)
			}

			// The provider's own words, and the URL it was reached at, stay
			// in the log.
			for _, leak := range []string{"oauth2", "invalid_", "idp.example", "connection refused"} {
				if strings.Contains(got.Error(), leak) {
					t.Errorf("the answer carries %q: %v", leak, got)
				}
			}
		})
	}
}

// The same through a real exchange against a token endpoint: what the
// library hands back for each answer is what exchangeFailure reads.
func TestOIDCExchangeAgainstATokenEndpointThatRefuses(t *testing.T) {
	t.Parallel()

	oauthError := func(status int, code string) func(http.ResponseWriter) {
		return func(w http.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": code, "error_description": "secret detail from the provider"})
		}
	}

	for name, tt := range map[string]struct {
		answer func(http.ResponseWriter)
		want   string
	}{
		"invalid_grant":                     {oauthError(http.StatusBadRequest, "invalid_grant"), "*domain.IDPSignInRefusedError"},
		"an_error_in_a_200_as_github_sends": {oauthError(http.StatusOK, "bad_verification_code"), "*domain.IDPSignInRefusedError"},
		"invalid_client":                    {oauthError(http.StatusUnauthorized, "invalid_client"), "*domain.InvalidAuthnServiceError"},
		"a_500":                             {func(w http.ResponseWriter) { w.WriteHeader(http.StatusInternalServerError) }, "*domain.InvalidAuthnServiceError"},
		"a_connection_closed_without_answer": {func(w http.ResponseWriter) {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_ = conn.Close()
			}
		}, "*domain.IDPUnreachableError"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakeOIDC(t)
			f.nonce = "n"
			f.tokenAnswer = tt.answer

			_, err := newProvider(t).Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1",
				oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

			if typ := fmt.Sprintf("%T", err); typ != tt.want {
				t.Fatalf("got %s (%v), want %s", typ, err, tt.want)
			}

			if strings.Contains(err.Error(), "secret detail") {
				t.Errorf("the provider's description is in the answer: %v", err)
			}
		})
	}
}

// An account the provider gives no address for cannot be provisioned or
// matched: the sign-in is refused, and the provider is not at fault.
func TestOIDCExchangeWithoutAnAddressIsARefusedSignIn(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "n"
	f.claims = map[string]any{"name": "No Address"}

	_, err := newProvider(t).Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1",
		oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

	if _, ok := errors.AsType[*domain.IDPSignInRefusedError](err); !ok {
		t.Fatalf("got %T (%v), want a refused sign-in", err, err)
	}
}

// A request cancelled because whoever called this service went away is not
// the provider's failure: it was counted as "provider unreachable", a 503
// and a red span for every closed tab.
func TestACallerWhoWentAwayIsNotAProviderOutage(t *testing.T) {
	t.Parallel()

	gone := &url.Error{Op: "Post", URL: "https://idp.example/token", Err: context.Canceled}

	for name, classify := range map[string]func(error) error{"the_exchange": exchangeFailure, "a_lookup": lookupFailure} {
		if got := classify(gone); !errors.Is(got, context.Canceled) {
			t.Errorf("%s: got %T (%v), want the context's own error", name, got, got)
		}
	}

	// And through a real exchange.
	f := newFakeOIDC(t)
	f.nonce = "n"

	ctx, cancel := context.WithCancel(t.Context())
	f.tokenAnswer = func(http.ResponseWriter) { cancel(); time.Sleep(50 * time.Millisecond) }

	_, err := newProvider(t).Exchange(ctx, oidcIDP(f.srv.URL, "Okta"), "code-1",
		oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

	if _, blamed := errors.AsType[*domain.IDPUnreachableError](err); blamed || !errors.Is(err, context.Canceled) {
		t.Fatalf("got %T (%v), want the cancellation itself", err, err)
	}
}

// A provider that never echoes the nonce fails every sign-in: that is its
// side, not a refusal of this person.
func TestOIDCExchangeWithNoNonceAtAllIsTheProvidersSide(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = ""
	f.claims = map[string]any{"email": "jane@example.com", "email_verified": true}

	_, err := newProvider(t).Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1",
		oauth.AuthRequest{Nonce: "mine", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

	if _, ok := errors.AsType[*domain.InvalidAuthnServiceError](err); !ok {
		t.Fatalf("got %T (%v), want the provider's side", err, err)
	}
}

// "This account has no address" is the person's side when the provider was
// asked and said so. When the question could not be put -- the user-info
// call failed -- nobody knows, and it is the provider's side: a user-info
// outage used to tell every visitor their account had no address.
func TestOIDCAFailedAddressLookupIsNotAnAccountWithoutAnAddress(t *testing.T) {
	t.Parallel()

	f := newFakeOIDC(t)
	f.nonce = "n"
	f.claims = map[string]any{"name": "No Address In The Token"}
	f.userInfoStatus = http.StatusServiceUnavailable

	_, err := newProvider(t).Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1",
		oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

	if _, refused := errors.AsType[*domain.IDPSignInRefusedError](err); refused || err == nil {
		t.Fatalf("got %T (%v), want the provider's side", err, err)
	}
}

// githubFake is GitHub's three endpoints, each answering what a test says.
func githubFake(t *testing.T, token, user, emails http.HandlerFunc) *domain.IDP {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/oauth/access_token", token)
	mux.HandleFunc("GET /user", user)
	mux.HandleFunc("GET /user/emails", emails)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// The github kind uses fixed endpoints; point them at the fake.
	saved := githubEndpoint
	githubEndpoint = oauth2.Endpoint{AuthURL: srv.URL + "/login/oauth/authorize", TokenURL: srv.URL + "/login/oauth/access_token"}
	t.Cleanup(func() { githubEndpoint = saved })

	return &domain.IDP{
		Name: "GH", ClientID: "cid", ClientSecret: "sec", CallbackURL: "https://app.example/cb", Enabled: true,
		IDPType: domain.IDPTypes{Name: "Github", Kind: domain.IDPTypeKindGithub, UserInfoAPIURL: srv.URL + "/user", Scopes: []string{"read:user", "user:email"}},
	}
}

// GitHub's side of the same split. Its token endpoint answers an error
// inside a 200, form-encoded unless asked otherwise; and its addresses come
// from a second call, which can fail by itself.
func TestGithubARefusedSignInIsNotAnOutage(t *testing.T) {
	// Not parallel: githubFake swaps the package-level GitHub endpoint.
	issue := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "gh-at", "token_type": "bearer"})
	}
	formError := func(code string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
			_, _ = w.Write([]byte("error=" + code + "&error_description=the+provider%27s+own+words"))
		}
	}
	account := func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 4242, "login": "octocat", "name": "The Octocat", "email": nil})
	}
	noEmails := func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode([]map[string]any{}) }
	status := func(code int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(code) }
	}

	const (
		refused     = "*domain.IDPSignInRefusedError"
		misbehaving = "*domain.InvalidAuthnServiceError"
	)

	for _, tt := range []struct {
		name                string
		token, user, emails http.HandlerFunc
		want                string
	}{
		{name: "a_code_github_does_not_accept", token: formError("bad_verification_code"), user: account, emails: noEmails, want: refused},
		{name: "an_address_github_will_not_vouch_for", token: formError("unverified_user_email"), user: account, emails: noEmails, want: refused},
		{name: "this_services_client_is_rejected", token: formError("incorrect_client_credentials"), user: account, emails: noEmails, want: misbehaving},
		{name: "an_account_with_no_address", token: issue, user: account, emails: noEmails, want: refused},
		{name: "the_addresses_cannot_be_read", token: issue, user: account, emails: status(http.StatusInternalServerError), want: misbehaving},
		{name: "a_provider_row_without_the_scope_for_addresses", token: issue, user: account, emails: status(http.StatusNotFound), want: misbehaving},
		{name: "the_account_cannot_be_read", token: issue, user: status(http.StatusBadGateway), emails: noEmails, want: misbehaving},
	} {
		t.Run(tt.name, func(t *testing.T) {
			idp := githubFake(t, tt.token, tt.user, tt.emails)

			_, err := newProvider(t).Exchange(t.Context(), idp, "code", oauth.AuthRequest{CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

			if typ := fmt.Sprintf("%T", err); typ != tt.want {
				t.Fatalf("got %s (%v), want %s", typ, err, tt.want)
			}

			if strings.Contains(err.Error(), "own words") {
				t.Errorf("the provider's description is in the answer: %v", err)
			}
		})
	}
}

// TestAFailedExchangeIsPresentedOnce: an authorization code is good for one
// exchange. Left to probe, the OAuth library sends a failed exchange a second
// time with the client's credentials in the other place, whatever the
// failure was: a spent code was presented twice, and a provider that was
// down was waited for twice. The place is decided before the exchange, from
// what the provider's discovery document declares.
func TestAFailedExchangeIsPresentedOnce(t *testing.T) {
	t.Parallel()

	refuse := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
	}

	for name, tt := range map[string]struct {
		declared  []string
		wantBasic bool
	}{
		// OpenID Connect Discovery: client_secret_basic when nothing is said.
		"nothing_declared":   {nil, true},
		"both_declared":      {[]string{"client_secret_post", "client_secret_basic"}, true},
		"only_the_form":      {[]string{"client_secret_post", "private_key_jwt"}, false},
		"neither_of_the_two": {[]string{"private_key_jwt"}, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakeOIDC(t)
			f.nonce = "n"
			f.authMethods = tt.declared
			f.tokenAnswer = refuse

			_, err := newProvider(t).Exchange(t.Context(), oidcIDP(f.srv.URL, "Okta"), "code-1",
				oauth.AuthRequest{Nonce: "n", CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

			if _, ok := errors.AsType[*domain.IDPSignInRefusedError](err); !ok {
				t.Fatalf("got %T (%v), want the refused sign-in", err, err)
			}

			if f.tokenHits != 1 {
				t.Errorf("the token endpoint was asked %d times, want once", f.tokenHits)
			}

			if f.gotBasicAuth != tt.wantBasic || (f.gotFormSecret != "") == tt.wantBasic {
				t.Errorf("credentials in the header: %v, in the form: %v; want the header: %v", f.gotBasicAuth, f.gotFormSecret != "", tt.wantBasic)
			}
		})
	}
}

// GitHub takes the credentials in the form, and a refused code is presented
// to it once as well.
func TestGithubAFailedExchangeIsPresentedOnce(t *testing.T) {
	// Not parallel: githubFake swaps the package-level GitHub endpoint.
	hits, inForm, inHeader := 0, false, false

	idp := githubFake(t,
		func(w http.ResponseWriter, r *http.Request) {
			hits++

			_ = r.ParseForm()
			inForm = r.Form.Get("client_secret") != ""
			_, _, inHeader = r.BasicAuth()

			w.Header().Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
			_, _ = w.Write([]byte("error=bad_verification_code"))
		},
		func(http.ResponseWriter, *http.Request) {}, func(http.ResponseWriter, *http.Request) {})

	_, err := newProvider(t).Exchange(t.Context(), idp, "code", oauth.AuthRequest{CodeVerifier: "v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v-v"})

	if _, ok := errors.AsType[*domain.IDPSignInRefusedError](err); !ok {
		t.Fatalf("got %T (%v), want the refused sign-in", err, err)
	}

	if hits != 1 || !inForm || inHeader {
		t.Errorf("asked %d times, credentials in the form: %v, in the header: %v; want once, in the form", hits, inForm, inHeader)
	}
}
