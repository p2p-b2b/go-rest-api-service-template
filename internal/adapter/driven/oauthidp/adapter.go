// Package oauthidp is the driven adapter that satisfies the oauth.Provider port.
//
// Two kinds of provider, decided by the IdP type's Kind and never by its name:
//
//   - oidc: discovery at <issuer>/.well-known/openid-configuration (cached per
//     issuer for the life of the process; the key set refreshes itself), the
//     authorization code flow with PKCE (S256) and a nonce, and the ID token
//     verified with coreos/go-oidc against the discovered JWKS, the issuer and
//     this client id. The identity is the token's `sub`. The user-info
//     endpoint is read only when the ID token carries no email, which Entra ID
//     does when the optional claim is not configured.
//   - github: GitHub has no OpenID Connect for users, so this is
//     golang.org/x/oauth2 against GitHub's fixed endpoints (with PKCE, which
//     GitHub accepts), /user for the identity and /user/emails for the primary
//     VERIFIED address -- /user.email is null for anyone whose email is private.
//
// # What a provider's word is worth
//
// UserInfo.EmailVerified is what gates account creation in the use case, and
// it is decided here, per provider: Google and Okta assert `email_verified`
// and it is passed through; GitHub's /user/emails says `verified` per address
// and only a verified primary is used; Entra ID has no such claim, and a
// single-tenant registration -- the only shape this service supports, the
// issuer pins the tenant -- means the email is the directory's own attribute,
// so it is treated as verified.
//
// # Errors
//
// The provider's own wording -- `oauth2: "invalid_client" ...`, go-oidc's
// "oidc: id token issued by a different provider" -- is never returned to the
// caller: it would publish a dependency's text as this API's contract and
// tell whoever is probing how our client is registered with the provider. It
// travels in the domain error's Detail, which Error() never prints (an
// OAuth refusal's code, description and status; any other failure's first
// line, cut at 200 bytes, because go-oidc puts the provider's page into its
// errors), and the
// one line written about the failure carries it (domain.FailureDetail): the
// request's access line, at ERROR for a 503. This adapter writes no line of
// its own; it wrote one at WARN for each failure, beside the access line, and
// then at DEBUG, where an operator at INFO never saw the reason.
package oauthidp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/port/driven/oauth"
)

// Config configures [New].
type Config struct {
	// HTTPClient is the outbound client every request to a provider goes
	// through: discovery, the token exchange, JWKS, user-info. Required, and it
	// must carry a timeout -- golang.org/x/oauth2 falls back to
	// http.DefaultClient, which has none, and the token exchange used to run
	// on exactly that.
	HTTPClient *http.Client
}

// githubEndpoint is GitHub's fixed OAuth2 endpoint pair. A variable so the
// contract test can point it at a fake GitHub; nothing else assigns it.
var githubEndpoint = github.Endpoint

// Provider implements oauth.Provider.
type Provider struct {
	httpClient *http.Client

	// providers caches one go-oidc Provider per issuer. Discovery is one HTTP
	// round trip that never changes for a given issuer, and the provider
	// holds the remote key set, which refreshes itself when a signature does
	// not verify. Keyed on the issuer string exactly as stored, so two rows
	// with the same tenant share one.
	providers sync.Map // issuer string -> *oidc.Provider
}

var _ oauth.Provider = (*Provider)(nil)

// New constructs a Provider.
func New(conf Config) (*Provider, error) {
	if conf.HTTPClient == nil {
		return nil, fmt.Errorf("oauthidp: HTTPClient is required")
	}

	if conf.HTTPClient.Timeout <= 0 {
		return nil, fmt.Errorf("oauthidp: HTTPClient must carry a timeout; a provider that hangs must not hold a sign-in open forever")
	}

	return &Provider{httpClient: conf.HTTPClient}, nil
}

// AuthCodeURL implements oauth.Provider.
func (p *Provider) AuthCodeURL(ctx context.Context, idp *domain.IDP, req oauth.AuthRequest) (string, error) {
	if req.State == "" || req.CodeVerifier == "" {
		return "", &domain.InvalidIdentityProvidersError{Message: "a sign-in needs a state and a PKCE verifier"}
	}

	cfg, err := p.configFor(ctx, idp)
	if err != nil {
		return "", err
	}

	opts := []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(req.CodeVerifier)}

	if idp.IsOIDC() {
		if req.Nonce == "" {
			return "", &domain.InvalidIdentityProvidersError{Message: "an OpenID Connect sign-in needs a nonce"}
		}

		opts = append(opts, oidc.Nonce(req.Nonce))
	}

	// No AccessTypeOffline and no ApprovalForce. Both are Google-specific: the
	// first asks for a refresh token this service never uses, the second forces
	// the consent screen on every sign-in, which Entra ID and Okta honour as
	// prompt=consent.
	return cfg.AuthCodeURL(req.State, opts...), nil
}

// Exchange implements oauth.Provider.
func (p *Provider) Exchange(ctx context.Context, idp *domain.IDP, code string, req oauth.AuthRequest) (*domain.UserInfo, error) {
	if code == "" || req.CodeVerifier == "" {
		return nil, &domain.InvalidIdentityProvidersError{Message: "the exchange needs the code and the PKCE verifier"}
	}

	cfg, err := p.configFor(ctx, idp)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.httpClient)

	token, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(req.CodeVerifier))
	if err != nil {
		return nil, exchangeFailure(err)
	}

	if idp.IsOIDC() {
		return p.identityFromIDToken(ctx, idp, cfg, token, req.Nonce)
	}

	return p.identityFromGithub(ctx, idp, cfg, token)
}

// configFor builds the oauth2.Config for one provider row.
func (p *Provider) configFor(ctx context.Context, idp *domain.IDP) (*oauth2.Config, error) {
	if idp == nil {
		return nil, &domain.InvalidIdentityProvidersError{Message: "unknown idp"}
	}

	if idp.ClientID == "" || idp.ClientSecret == "" || idp.CallbackURL == "" {
		return nil, notConfigured("the client id, the client secret or the callback URL is empty")
	}

	switch idp.IDPType.Kind {
	case domain.IDPTypeKindOIDC:
		if idp.IssuerURL == "" {
			return nil, notConfigured("an OpenID Connect provider has no issuer URL")
		}

		provider, err := p.discover(ctx, idp)
		if err != nil {
			return nil, err
		}

		endpoint := provider.Endpoint()
		endpoint.AuthStyle = tokenEndpointAuthStyle(provider)

		return &oauth2.Config{
			Endpoint:     endpoint,
			ClientID:     idp.ClientID,
			ClientSecret: idp.ClientSecret,
			RedirectURL:  idp.CallbackURL,
			Scopes:       withOpenID(idp.IDPType.Scopes),
		}, nil

	case domain.IDPTypeKindGithub:
		scopes := idp.IDPType.Scopes
		if len(scopes) == 0 {
			scopes = []string{"read:user", "user:email"}
		}

		// GitHub takes the client's credentials in the form, which is what
		// its documentation shows.
		endpoint := githubEndpoint
		if endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
			endpoint.AuthStyle = oauth2.AuthStyleInParams
		}

		return &oauth2.Config{
			Endpoint:     endpoint,
			ClientID:     idp.ClientID,
			ClientSecret: idp.ClientSecret,
			RedirectURL:  idp.CallbackURL,
			Scopes:       scopes,
		}, nil

	default:
		// The kind, not the name. A provider the adapter does not know how to
		// talk to used to be one whose NAME the switch did not list, so a row
		// called anything but "Google" or "Github" failed at the callback.
		return nil, notConfigured("the provider's kind is not one this service can sign in with")
	}
}

// notConfigured is a provider row this service cannot sign in with: the API
// refuses every one of these shapes, so the row was written some other way.
// It is the service's side, a 503 in our words. It was a 400 that named the
// provider, on routes anyone may call: the caller had sent nothing wrong.
// What is wrong is the failure's detail.
func notConfigured(reason string) error {
	return &domain.InvalidAuthnServiceError{Message: domain.IDPNotConfiguredMessage, Detail: reason}
}

// tokenEndpointAuthStyle says how this client presents its credentials at
// the token endpoint, from what the provider's discovery document declares.
//
// Left unset, the OAuth library probes: it sends the exchange with the
// credentials in the Authorization header and, on ANY failure, sends it
// again with them in the form. So every exchange that failed for a reason of
// its own -- a spent code, a provider that was down -- was presented twice,
// the second time with a code the first had already used up.
//
// `client_secret_basic` (the header) is the method when the document lists
// it, and also when it lists nothing: OpenID Connect Discovery makes it the
// default for an absent `token_endpoint_auth_methods_supported`. The form
// (`client_secret_post`) is used when it is the only one of the two listed.
func tokenEndpointAuthStyle(provider *oidc.Provider) oauth2.AuthStyle {
	var document struct {
		Methods []string `json:"token_endpoint_auth_methods_supported"`
	}

	if err := provider.Claims(&document); err != nil {
		return oauth2.AuthStyleInHeader
	}

	if !slices.Contains(document.Methods, "client_secret_basic") && slices.Contains(document.Methods, "client_secret_post") {
		return oauth2.AuthStyleInParams
	}

	return oauth2.AuthStyleInHeader
}

// discover returns the cached go-oidc provider for the issuer, fetching the
// discovery document on first use.
func (p *Provider) discover(ctx context.Context, idp *domain.IDP) (*oidc.Provider, error) {
	if cached, ok := p.providers.Load(idp.IssuerURL); ok {
		return cached.(*oidc.Provider), nil
	}

	ctx = oidc.ClientContext(ctx, p.httpClient)

	provider, err := oidc.NewProvider(ctx, idp.IssuerURL)
	if err != nil {
		return nil, &domain.IDPUnreachableError{Message: "the identity provider is not reachable", Detail: "discovery: " + providerText(err)}
	}

	actual, _ := p.providers.LoadOrStore(idp.IssuerURL, provider)

	return actual.(*oidc.Provider), nil
}

// exchangeFailure says whose failure a failed code exchange is, because the
// three are answered differently.
//
// The token endpoint answered with an OAuth error about the CODE or the
// PERSON: their sign-in is refused (401). `invalid_grant` is RFC 6749's word
// for a code that is spent, expired or was issued to another client or
// redirect; GitHub says `bad_verification_code`, and `unverified_user_email`
// for an account whose address it will not vouch for; `access_denied` is a
// provider's sign-on policy turning this person away.
//
// It answered with anything else -- `invalid_client`, `unauthorized_client`,
// GitHub's `incorrect_client_credentials` and `redirect_uri_mismatch`, a 5xx,
// a body that is not a token -- and it is this service's registration at the
// provider, or the provider itself, that is wrong: an outage of the sign-in
// (503), not something the person can fix by trying again.
//
// It did not answer: unreachable (503).
//
// And when the request was cancelled because whoever called THIS service
// went away, it is none of the three: the context's own error goes back as
// it is, and the provider is not blamed for a closed tab.
func exchangeFailure(err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}

	if refusal, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
		switch refusal.ErrorCode {
		case "invalid_grant", "bad_verification_code", "unverified_user_email", "access_denied":
			return &domain.IDPSignInRefusedError{Message: "the identity provider did not accept this sign-in", Detail: providerDetail(err)}
		}

		return &domain.InvalidAuthnServiceError{Message: "the identity provider did not accept this service's request", Detail: providerDetail(err)}
	}

	if _, ok := errors.AsType[*url.Error](err); ok {
		return &domain.IDPUnreachableError{Message: "the identity provider is not reachable", Detail: providerDetail(err)}
	}

	return &domain.InvalidAuthnServiceError{Message: "the identity provider did not return a token", Detail: providerDetail(err)}
}

// providerDetail is what a failed call to the provider says, for the
// failure's detail: an OAuth refusal's code, description and status, never
// its body, which for an answer with no OAuth error is whatever page the
// provider served (oauth2 prints it whole); any other failure's text, cut
// by providerText.
func providerDetail(err error) string {
	refusal, ok := errors.AsType[*oauth2.RetrieveError](err)
	if !ok {
		return providerText(err)
	}

	status := 0
	if refusal.Response != nil {
		status = refusal.Response.StatusCode
	}

	if refusal.ErrorCode == "" {
		return fmt.Sprintf("the token endpoint answered %d with no OAuth error", status)
	}

	return fmt.Sprintf("the token endpoint answered %d: %s: %s", status, refusal.ErrorCode, refusal.ErrorDescription)
}

// maxProviderText bounds providerText.
const maxProviderText = 200

// providerText is a failure's own text, cut at its first line and at
// maxProviderText bytes. go-oidc puts the provider's answer into its errors
// -- discovery, the account details and the keys are "<status>: <body>" --
// so the text may begin a page the provider served: what is kept is the
// status and the start of the page, which tells a 404 from a login form, and
// no more.
func providerText(err error) string {
	text, _, _ := strings.Cut(err.Error(), "\n")
	if len(text) <= maxProviderText {
		return text
	}

	cut := maxProviderText
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}

	return text[:cut] + "…"
}

// lookupFailure is exchangeFailure's twin for the calls made with the token:
// the account details and the addresses. There is no OAuth error about the
// person here, so a call that fails is the provider's side or nobody's.
func lookupFailure(err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}

	if _, ok := errors.AsType[*url.Error](err); ok {
		return &domain.IDPUnreachableError{Message: "the identity provider is not reachable", Detail: providerDetail(err)}
	}

	return &domain.InvalidAuthnServiceError{Message: "the identity provider did not return the account details", Detail: providerDetail(err)}
}

// identityFromIDToken verifies the ID token and maps its claims.
func (p *Provider) identityFromIDToken(ctx context.Context, idp *domain.IDP, cfg *oauth2.Config, token *oauth2.Token, nonce string) (*domain.UserInfo, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, &domain.InvalidAuthnServiceError{Message: "the identity provider did not return an identity", Detail: "the token response carried no id_token"}
	}

	provider, err := p.discover(ctx, idp)
	if err != nil {
		return nil, err
	}

	// Issuer, audience (this client id), expiry and signature against the
	// discovered keys. The issuer is the one the row names, so a token from
	// another tenant of the same vendor -- same keys, different iss -- is
	// refused here.
	idToken, err := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}).Verify(oidc.ClientContext(ctx, p.httpClient), rawIDToken)
	if err != nil {
		return nil, &domain.InvalidAuthnServiceError{Message: "the identity provider did not return a valid identity", Detail: "the ID token did not verify: " + providerText(err)}
	}

	if idToken.Nonce == "" {
		// A provider that does not echo the nonce at all is not following
		// the protocol: every sign-in fails, and that is its side.
		return nil, &domain.InvalidAuthnServiceError{Message: "the identity provider did not return a valid identity", Detail: "the ID token carries no nonce"}
	}

	if idToken.Nonce != nonce {
		// The token is the provider's and it verifies; it answers another
		// sign-in than this one. That is the caller's side.
		return nil, &domain.IDPSignInRefusedError{Message: "the identity provider's answer is not for this sign-in", Detail: "the ID token's nonce does not match this sign-in"}
	}

	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, &domain.InvalidAuthnServiceError{Message: "the identity provider did not return a valid identity", Detail: "the ID token's claims could not be read: " + providerText(err)}
	}

	// Entra ID sends `email` only when the optional claim is configured on
	// the app registration; the user-info endpoint has it in the same shape.
	// One extra round trip, only when the token was short.
	var lookupErr error

	if claims.Email == "" {
		info, err := provider.UserInfo(oidc.ClientContext(ctx, p.httpClient), oauth2.StaticTokenSource(token))
		if err != nil {
			lookupErr = err
		} else {
			var extra oidcClaims
			if err := info.Claims(&extra); err == nil {
				claims.merge(extra)
			}
		}
	}

	info, err := mapOIDCClaims(idp, idToken.Subject, claims)

	// "This account has no address" is the person's side only when the
	// provider was asked and said so. When the question could not be put,
	// nobody knows, and that is the provider's side.
	if _, noAddress := errors.AsType[*domain.IDPSignInRefusedError](err); noAddress && lookupErr != nil {
		return nil, lookupFailure(lookupErr)
	}

	return info, err
}

// oidcClaims are the claims this service reads from an ID token or user-info
// document. Everything else the provider sends is ignored.
type oidcClaims struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	EmailVerified     bool   `json:"email_verified"`
}

func (c *oidcClaims) merge(o oidcClaims) {
	if c.Email == "" {
		c.Email, c.EmailVerified = o.Email, o.EmailVerified
	}

	if c.PreferredUsername == "" {
		c.PreferredUsername = o.PreferredUsername
	}

	if c.Name == "" {
		c.Name = o.Name
	}

	if c.GivenName == "" {
		c.GivenName = o.GivenName
	}

	if c.FamilyName == "" {
		c.FamilyName = o.FamilyName
	}
}

// mapOIDCClaims turns verified claims into a UserInfo, with the per-vendor
// quirks in one place.
func mapOIDCClaims(idp *domain.IDP, subject string, c oidcClaims) (*domain.UserInfo, error) {
	if subject == "" {
		return nil, &domain.InvalidAuthnServiceError{Message: "the identity provider did not return a valid identity", Detail: "the ID token has no subject"}
	}

	email, verified := c.Email, c.EmailVerified

	switch domain.IDPTypeName(idp.IDPType.Name) {
	case domain.IDPTypeNameEntraID:
		// No email_verified claim exists. The row pins ONE tenant through its
		// issuer, so the address is the directory's own attribute for a member
		// of that tenant, set by its administrators: verified by construction.
		// preferred_username is the UPN, which is an address in most tenants
		// and is the fallback when the optional email claim is not configured.
		if email == "" && isEmail(c.PreferredUsername) {
			email = c.PreferredUsername
		}

		verified = email != ""
	default:
		// Google and Okta assert it; anything else is taken at its word too.
	}

	if email == "" {
		return nil, &domain.IDPSignInRefusedError{Message: "the identity provider did not return an email address for this account"}
	}

	first, last := c.GivenName, c.FamilyName
	if first == "" && last == "" {
		first, last = splitName(c.Name, email)
	}

	return &domain.UserInfo{
		Subject:       subject,
		Email:         strings.ToLower(email),
		EmailVerified: verified,
		FirstName:     first,
		LastName:      last,
	}, nil
}

// identityFromGithub reads /user for the identity and /user/emails for the
// primary verified address.
func (p *Provider) identityFromGithub(ctx context.Context, idp *domain.IDP, cfg *oauth2.Config, token *oauth2.Token) (*domain.UserInfo, error) {
	client := cfg.Client(ctx, token)

	userURL := idp.IDPType.UserInfoAPIURL
	if userURL == "" {
		userURL = "https://api.github.com/user"
	}

	var user struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := getJSON(ctx, client, userURL, &user); err != nil {
		return nil, lookupFailure(err)
	}

	if user.ID == 0 {
		return nil, &domain.InvalidAuthnServiceError{Message: "the identity provider did not return a valid identity", Detail: "the account details carry no id"}
	}

	// /user.email is null for anyone whose email is private, which is most
	// people. /user/emails lists every address with `primary` and `verified`;
	// only a verified primary is an address this service will act on.
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	email, verified := "", false

	var lookupErr error

	if err := getJSON(ctx, client, strings.TrimSuffix(userURL, "/")+"/emails", &emails); err != nil {
		lookupErr = err
	} else {
		for _, e := range emails {
			if e.Primary {
				email, verified = e.Email, e.Verified

				break
			}
		}
	}

	if email == "" && user.Email != "" {
		// The public profile email, whose verification GitHub does not state.
		email, verified = user.Email, false
	}

	if email == "" {
		// Asked and answered "none" is the person's side. Not answered (the
		// call failed, or this provider row lacks the scope to make it) is
		// not.
		if lookupErr != nil {
			return nil, lookupFailure(lookupErr)
		}

		return nil, &domain.IDPSignInRefusedError{Message: "the identity provider did not return an email address for this account"}
	}

	first, last := splitName(user.Name, user.Login)

	return &domain.UserInfo{
		// The numeric id, which survives a rename; the login does not.
		Subject:       strconv.FormatInt(user.ID, 10),
		Email:         strings.ToLower(email),
		EmailVerified: verified,
		FirstName:     first,
		LastName:      last,
	}, nil
}

// splitName makes a first and last name out of whatever the provider gave.
// A single word is the first name and the last is the local part of the
// fallback, so a person called "Cher" or with a private name can still be
// registered; it used to be an error.
func splitName(name, fallback string) (string, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		name = fallback
		if at := strings.Index(name, "@"); at > 0 {
			name = name[:at]
		}
	}

	first, last, found := strings.Cut(name, " ")
	if !found || strings.TrimSpace(last) == "" {
		return first, first
	}

	return first, strings.TrimSpace(last)
}

func withOpenID(scopes []string) []string {
	if slices.Contains(scopes, oidc.ScopeOpenID) {
		return scopes
	}

	return append([]string{oidc.ScopeOpenID}, scopes...)
}

func isEmail(s string) bool {
	if s == "" {
		return false
	}

	_, err := mail.ParseAddress(s)

	return err == nil
}

// userInfoAPITimeout caps one call to a github endpoint, on top of the
// client's own timeout.
const userInfoAPITimeout = 10 * time.Second

func getJSON(ctx context.Context, client *http.Client, apiURL string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, userInfoAPITimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
