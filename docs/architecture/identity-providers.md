# Identity providers

How a person signs in through Google, GitHub, Microsoft Entra ID or Okta, what
this service trusts a provider for, and what it deliberately does not.

The short version: a provider proves a **subject**, never an email. The
identity is `(provider, subject)`, stored in `users_identities`; the email is a
hint used once, to create an account when the provider vouches for it. The
browser lands on the **frontend**, which talks to this API server-to-server;
this API sets no cookie and issues no redirect. OpenID Connect providers are
spoken to through discovery, with PKCE, a nonce and a verified ID token; GitHub
is plain OAuth2 with the primary verified email.

## The rows

```mermaid
erDiagram
    idp_types {
        uuid id PK
        varchar name "Google | Github | EntraID | Okta -- matched in Go for claim quirks"
        varchar kind "oidc | github -- HOW the adapter talks to it"
        text[] scopes
        varchar user_info_api_url "github only; oidc discovers it"
        varchar issuer_hint "what an instance's issuer should look like"
    }
    idps {
        uuid id PK
        uuid idp_types FK
        varchar name
        varchar callback_url "the FRONTEND's callback route"
        varchar issuer_url "oidc only; ONE tenant per row"
        text client_secret "AES-encrypted"
        boolean enabled
        boolean auto_provision
    }
    users_identities {
        uuid users_id FK
        uuid idps_id FK
        varchar subject "OIDC sub, GitHub numeric id"
        varchar email "as seen at link time; never matched on"
        timestamptz linked_at
    }
    users {
        uuid id PK
        varchar email
        boolean local_account "has a password"
    }
    idp_types ||--o{ idps : "kind of"
    idps ||--o{ users_identities : "identity at"
    users ||--o{ users_identities : "belongs to"
```

**A type is a kind plus claim quirks.** `kind` decides the protocol; the
adapter used to switch on the *name* instead, so a row called anything but
`Google` or `Github` could be created and could never sign anybody in. The
name still matters for one thing: the per-vendor claim mapping (Entra ID's
`preferred_username`, see below).

**The issuer lives on the instance.** Google's is a constant; Entra ID's and
Okta's contain the tenant or the org, so a type cannot carry it. One tenant per
row, deliberately: the issuer is what the ID token's `iss` must equal, and a
second tenant is a second row rather than a `common` endpoint plus an
allow-list this service would have to maintain.

**`client_secret` is `TEXT`.** It holds ciphertext, and Okta's secrets are
already 64 characters in the clear.

## A sign-in, end to end

```mermaid
sequenceDiagram
    participant B as Browser
    participant F as Frontend (SvelteKit server)
    participant A as API (AuthnIDPsHandler)
    participant U as AuthnIDPsService
    participant O as oauthidp (go-oidc)
    participant P as Provider

    B->>F: click "Sign in with Okta"
    F->>A: GET /auth/idp/{id}/login
    A->>U: GetLoginURL(login)
    U->>U: mint verifier + nonce, seal them in the state (AES), sign the state (JWT, single-use jti)
    U->>O: AuthCodeURL(idp, state, nonce, verifier)
    O->>P: discovery (cached per issuer)
    O-->>U: authorize URL + S256 challenge + nonce
    A-->>F: {redirect_url}
    F-->>B: 302 to the provider
    B->>P: sign in, consent
    P-->>B: 302 to callback_url = https://frontend/auth/idp/{id}/callback?state&code
    B->>F: GET /auth/idp/{id}/callback?state&code
    F->>A: GET /auth/idp/{id}/callback?state&code (server-to-server)
    A->>U: Callback(id, state, code)
    U->>U: verify state, spend jti, unseal verifier + nonce
    alt state spent, expired, or for another idp
        U-->>A: InvalidJWTError → 400
    end
    U->>O: Exchange(idp, code, verifier, nonce)
    O->>P: POST token (code + code_verifier)
    P-->>O: id_token (+ access_token)
    O->>O: verify iss, aud, exp, signature (JWKS), nonce
    alt bad token, wrong nonce, other tenant
        O-->>U: "did not return a valid identity" → 401
    end
    O-->>U: UserInfo{Subject, Email, EmailVerified, ...}
    U->>U: resolve (idp, subject) in users_identities
    alt known identity
        U->>U: LoginUserByID(linked account)
    else unknown, auto_provision on, email verified, no account with that email
        U->>U: RegisterUser + Link + LoginUserByID
    else unknown, anything else
        U-->>A: IDPIdentityNotLinkedError → 401 (one wording for every reason)
    end
    A-->>F: 200 {event, login:{access_token, refresh_token, ...}}
    F->>F: saveBackendSession → its own cookies
    F-->>B: 302 to the app
```

### What the callback used to do, and why it moved

The API was the provider's redirect target. It set three cookies **with the
frontend's cookie names on its own host** and redirected to a per-IdP
`login_redirect_url`. That worked in development only because both ran on
`localhost` — cookies ignore ports — and would have failed the first time the
API and the frontend had different hostnames: the SvelteKit server never sees
a cookie set for another host. It also contradicted the rule the frontend is
built on, that the browser never holds a token the API issued to it.

The callback now answers JSON, sets nothing and redirects nowhere. The
provider's `redirect_uri` is the frontend route, which exchanges state and
code with the API the same way it exchanges a password with `/auth/login`,
and stores the tokens with the same code. The two redirect URLs left the
`idps` row: where the browser lands is the frontend's decision.

## Identity, not email

```mermaid
flowchart TD
    CB[callback: UserInfo] --> R{"(idp, subject) in users_identities?"}
    R -- known --> L[LoginUserByID → session]
    R -- unknown --> E{event == link?}
    E -- yes --> LK["Link to the account in the state<br/>(the user who started it, signed in)"]
    E -- no --> AP{auto_provision?}
    AP -- off --> REF[refused: not linked]
    AP -- on --> EV{email verified<br/>by the provider?}
    EV -- no --> REF
    EV -- yes --> EX{account with<br/>that email exists?}
    EX -- yes --> REF
    EX -- no --> NEW[RegisterUser + Link → session]
```

**The refusal is one message for every branch.** Which addresses have
accounts here is not something whoever controls a provider account gets to
learn from the difference between "unverified" and "already registered".

**What it replaces.** `LoginUser` used to be called with the provider's email
and no password: it looked the account up by email, signed in, and if the
account was a local one set `local_account = false`, disabling the password.
Nothing proved the email belonged to the account — the provider's word was
enough, and an Entra administrator, a GitHub user with an unverified address,
or anyone at a provider that lets emails be changed could sign in as any
account whose address they could make a provider assert. `LoginUser` keeps its
one job, passwords; IdP sign-ins go through `LoginUserByID` after resolution,
and no login path changes the account on the way through.

**Linking is the holder's act.** A signed-in user starts
`GET /auth/idp/{id}/link`; the state carries their account id, sealed; the
callback attaches the identity to that account. The session proves the
account, the provider proves the identity, and it is the only moment both are
in hand. `GET /me/identities` lists them; `DELETE /me/identities/{idp_id}`
removes one — refused when it is the only way into an account with no
password, because nothing could sign in to it afterwards.

## What each provider is trusted for

| Provider | Kind | Subject | Email | Verified? |
| --- | --- | --- | --- | --- |
| Google | oidc, issuer `https://accounts.google.com` | `sub` | `email` | `email_verified` claim |
| Okta | oidc, issuer `https://<org>.okta.com/oauth2/default` | `sub` | `email` | `email_verified` claim |
| Entra ID | oidc, issuer `https://login.microsoftonline.com/<tenant>/v2.0` | `sub` (pairwise per app registration, stable for this row) | `email` if the optional claim is configured, else `preferred_username` when it is an address | **by construction**: one tenant per row, the address is the directory's own attribute |
| GitHub | github, fixed endpoints | numeric `id` (survives a rename; the login does not) | primary address from `/user/emails` (`/user.email` is null when private) | that address's `verified` flag |

Entra ID's trust is the one worth restating: there is no `email_verified`
claim, and the choice to treat a single tenant's addresses as verified is a
policy, recorded here and in the adapter, not a fact the provider asserts. The
scopes seeded for GitHub are `read:user user:email`; the old seed carried
`github:email` and `github:profile`, which are not GitHub scopes.

## The state token

Signed by this service, single-use, and it carries three things sealed with
the AES key in a `data` claim: the PKCE verifier, the nonce, and — for a link —
the account. Sealed because the state is readable by whoever sees the URL, and
the verifier must not be. Spent **before** the code is exchanged, on the
`revoked_tokens` denylist, so a state that survives a failed exchange cannot be
retried. A replay is refused with the same wording as any other bad state.

## What can go wrong

| Failure | Answer | Why |
| --- | --- | --- |
| provider cancelled (`error=access_denied`) | 401, "did not complete the sign-in" | the provider's description goes to the log, not the caller |
| state missing, spent, expired, for another idp, unsealable | 400 | one wording |
| discovery unreachable | 503 at start or callback | cached per issuer afterwards |
| ID token fails iss/aud/exp/signature, or wrong nonce | 401 | go-oidc's text is logged, never returned |
| identity unknown and not provisionable | 401, one wording | see above |
| identity already linked elsewhere (link) | 409 | the other account keeps it |
| provider disabled | 404 at start, at the callback, and absent from `/auth/idp/available` | a half-configured provider is never offered |

## Configuring a provider

| Field | Meaning |
| --- | --- |
| `idp_type_id` | one of the seeded types; decides kind and claim mapping |
| `issuer_url` | required for the oidc kind, refused for github; the type's `issuer_hint` shows the shape |
| `callback_url` | `https://<frontend>/auth/idp/<idp_id>/callback`, registered verbatim with the provider |
| `client_id`, `client_secret` | from the provider's app registration |
| `enabled` | offered and accepted; off while setting up |
| `auto_provision` | may a first sign-in with a verified email create an account |

At the provider: register the callback URL exactly; for Entra ID add the
`email` optional claim to the ID token or accept the UPN fallback; for Okta
assign the app to the users who should sign in.

## Deleting or disabling a provider

A provider's delete takes its identities with it: `users_identities.idps_id`
cascades. It asked nothing first. An account that signs in through that
provider and has no password (`local_account` false) was left
with no way in at all, for nobody to sign in to. A provider created again
does not give it back, since its sign-in finds an account with the address
that is not linked, which it refuses. Recovery by e-mail sends such an
account nothing (`silentRecovery`): a password is a local account's to
recover.

Unlinking the last identity of such an account was already refused ("this
is the only way into the account"). The provider's delete is the same act
for all of them at once.

**The delete is a 409 while an account's only way in is an identity at that
provider. A disable is allowed, unless it would leave no administrator
account that can sign in.**

```mermaid
flowchart TD
    D["DELETE /auth/idps/{idp_id}"] --> A{"would an administrator account<br/>still be able to sign in?<br/>(keepsAnAdministrator)"}
    A -- "no, and one can now" --> P["403 IDPProtectedError:<br/>the last enabled administrator account<br/>signs in through it"]
    A -- "yes, or none can as it is" --> S["one statement:<br/>DELETE ... WHERE NOT EXISTS<br/>(an account that is not local,<br/>with an identity here and at no other provider)"]
    S -- "a row was deleted" --> OK["200; its identities go with it,<br/>its slot is given back"]
    S -- "nothing was deleted" --> R{"is the provider there?"}
    R -- no --> N["404"]
    R -- yes --> C["409 IDPInUseError:<br/>still used by at least one account<br/>that has no other way to sign in"]
    U["PUT /auth/idps/{idp_id}<br/>enabled: false, on an enabled provider"] --> A2{"keepsAnAdministrator"}
    A2 -- no --> P2["403 IDPProtectedError"]
    A2 -- yes --> W["the update is written:<br/>sign-ins through it stop until it is enabled again"]
```

- **The rule is in the delete's own statement** (`queryDeleteIDP`), so the
  provider is kept by the rows as they are when it runs. Not by the foreign
  key: `ON DELETE NO ACTION` would keep the provider while any identity
  exists, also the one of an account that signs in with its password.
- **Which accounts keep it**: not a local account, an identity at this
  provider, and none at another. A disabled account counts: enabled again,
  it would have no way in. An account with a password does not, and neither
  does one with an identity at another provider, whatever that provider's
  state.
- **What an administrator does instead**: disables the provider. Or deletes
  those accounts, or makes them local accounts (`local_account: true`), which
  lets recovery mail them a link to choose a password; but no route lists
  the accounts that sign in through a provider, so with more than one
  provider those accounts are found in the database, and the disable is the
  practical answer.
- **A disable is not held to that rule.** It takes away nothing that enabling
  does not give back, and it is the way out of the 409. It is held to one
  thing: an administrator account must be left that can sign in.
- **The last door** (`IDPsService.keepsAnAdministrator`). The `admin` flag
  may be held by an account that signs in through a provider and has no
  password. With every other administrator account disabled, which is
  allowed while that one can sign in, its provider is the last door:
  disabling it, or deleting it under an account that also has an identity at
  a disabled provider, left no administrator account that could sign in, and
  so nobody to change an address or unlock another administrator account.
  (The provider itself can still be enabled by an account with its route's
  grant: what is kept is the account.) Both are a 403 for everybody: the
  routes are the Administrator role's, and the role is not the flag. When no
  administrator account can sign in as it is, there is nothing left to keep,
  and nothing is refused.
- **"A way in" means two things, and each is where it belongs.** *Can sign
  in now*: a password login, or an identity at a provider that is enabled.
  That is the count of administrator accounts (`enabledAdministrator`, one
  predicate for "how many", "how many without this provider" and "is this
  account one of them"), and it is what an account must keep when it
  unlinks an identity of its own (`keepsAWayIn`): an identity at a disabled
  provider is not one, and counted as one it let an account unlink its only
  door. *Could sign in again without the database*: any other identity,
  whatever its provider's state. That is the delete's rule, because
  enabling a provider gives its identities back and deleting one does not.
- **The count and the write are two statements**, as every check of this
  kind in the service is.
- **A delete waits for a sign-in through the provider.** An identity's
  insert (a first sign-in's, a link) holds a `KEY SHARE` lock on the
  provider's row, through its foreign key, until it commits. The delete's
  statement read its condition on a snapshot without that identity, waited
  for the lock, and went ahead after it without reading the condition again
  (the row had been locked, not changed): its cascade removed the identity
  just written, and an account that had just signed in had no way in. The
  delete is one transaction now and takes the row `FOR UPDATE` first
  (`queryLockIDPForDelete`), which waits for the insert; its statement then
  reads the identities as they are (READ COMMITTED takes a snapshot per
  statement). The lock also says whether the provider is there.

```mermaid
sequenceDiagram
    participant S as an identity's insert (a sign-in, a link)
    participant DB as idps row
    participant D as the provider's delete (one transaction)
    alt the insert first
        S->>DB: INSERT users_identities (KEY SHARE on the row)
        D->>DB: SELECT ... FOR UPDATE: waits
        S->>DB: COMMIT
        DB-->>D: the lock
        D->>DB: DELETE ... WHERE NOT EXISTS (only way in): sees the identity
        D-->>D: 409, nothing deleted
    else the delete first
        D->>DB: SELECT ... FOR UPDATE (the row is there)
        S->>DB: INSERT users_identities: its foreign key waits
        D->>DB: DELETE, COMMIT
        DB-->>S: no provider: the insert fails, no identity is stored
    else the provider is not there
        D->>DB: SELECT ... FOR UPDATE: no row
        D-->>D: 404
    end
```

  A first sign-in here is still two writes, the account and then its
  identity: when the delete comes first, the account the sign-in registered
  stays, with no identity. Writing both in one transaction is what closes
  that, and it is not done yet.
- **An account's ways in are removed one at a time.** The delete's rule
  reads an account's other identities, and it held only its own provider's
  row: the two providers of an account that has no other way in, deleted in
  the same instant, each saw the other's identity and both went, and so did
  a delete and an unlink of the account's other identity. Every write that
  removes a way in takes the account's row first, `FOR NO KEY UPDATE`: the
  provider's delete takes the rows of the accounts linked to it, in the
  order of their ids, after its own row (`queryLockAccountsOfIDP`), and the
  unlink, which is one transaction now, takes the account's row and then
  deletes with the rule in its statement (`queryUnlinkIdentity`; the
  use-case's `keepsAWayIn` is still asked first, and answers the ordinary
  case). The second write waits for the first and reads what it left: a
  409, and the account keeps its last way in. Each keeps its own rule (the
  delete's counts an identity at a disabled provider, which enabling the
  provider gives back; the unlink's does not, above): what is serialised is
  the reading, not the rules. A provider's disable takes no account's row,
  and the count of administrator accounts that can sign in
  (`keepsAnAdministrator`) is a statement of its own before either write:
  two deletes or disables of an administrator account's two providers can
  each pass it. `FOR NO KEY UPDATE` does not conflict with the `KEY SHARE`
  an identity's insert takes on the account, so a sign-in into the account
  is not held, and the id order keeps two deletes that share accounts from
  deadlocking: the second waits for the first.

```mermaid
sequenceDiagram
    participant A as delete of provider A
    participant X as account X's row
    participant B as an unlink of B, or delete of provider B
    A->>X: SELECT ... FOR NO KEY UPDATE (after A's own row)
    B->>X: SELECT ... FOR NO KEY UPDATE: waits
    A->>A: DELETE: X still has its identity at B, so A goes
    A-->>X: COMMIT
    X-->>B: the lock
    B->>B: DELETE with its rule: X has no other way in now
    B-->>B: 409, nothing removed: X keeps B
```

`TestAnIdentityProvidersDeleteStrandsNobody` holds the delete against a
running service (an account whose only way in it is, a disabled one, one
with a password, one with an identity elsewhere, several accounts of which
one keeps it, a provider nobody uses, a missing one, and the disable that is
allowed instead). `TestAnIdentityAtADisabledProviderIsNoWayIn` holds the
count, the count without a provider, the question about one account and
what an identity says of its provider. In the unit suite:
`TestAProviderIsNotDisabledUnderTheLastAdministrator`,
`TestAProviderIsNotDeletedUnderTheLastAdministrator`,
`TestAProvidersGuardDoesNotGuessTheCount` (either count failing),
`TestTheProvidersServiceNeedsItsCount`, `TestAnUnlinkKeepsAWayIn`,
`TestADeleteThatDeletedNothingSaysWhy`,
`TestTheProvidersDeleteKeepsItForAnAccountWithNoOtherWayIn` and
`TestWhoCountsAsAnAdministratorThatCanSignIn` (the statements, whole),
`TestAProvidersDeleteLocksItsRowBeforeItsCheck` (the two locks, and that the
delete runs after them in its transaction) and
`TestAnUnlinkTakesTheAccountBeforeItsRule` (the unlink's lock, its statement
and their order). Against a running service,
`TestAProvidersDeleteWaitsForASignInThroughIt` plays the first race: a
transaction of its own holds an account's identity uncommitted, the delete
is seen waiting on its lock, and after the commit it is a 409 with the
identity kept (it was a 200, and the account had no way in).
`TestAnAccountsWaysInAreRemovedOneAtATime` plays the four others the same
way (a provider's delete under an unlink of the account's other identity
and under the delete of its other provider, an unlink under each of the
two): a 409 each, the account keeping one identity; on the code before, a
200 and an account with none.

## Pointing a provider elsewhere

An identity is (provider, subject), and a sign-in is "this provider says
this subject", then a session for the account. The issuer and the kind are
what "this provider" means. `PUT /auth/idps/{idp_id}` could change both
while the provider's identities stayed, keyed by its id: whoever holds that
route (the Administrator role, which is not the flag) pointed the provider
at an issuer of their own, asserted a subject that is linked, and was signed
in as that account. A subject is not shown by the API, and at many
providers it is no secret. It also shut, with no refusal, the door of an
administrator account that signs in through the provider, and the counts
went on taking the repointed provider for a door.

**Another issuer or another type is a 409 while the provider has identities
(a new provider instead). Where its sign-in goes, the client and the
callback, may change, and while accounts are linked that change is the
administrator account's to make.** The client and the callback together are
a takeover too: set to an application and a host of one's own, at the real
issuer, a linked user's ordinary sign-in sends the code to that host, and at
a provider whose subjects are global (GitHub, Google) that code is a session
as the user. The state does not help: it is sealed by the service and
carries the PKCE verifier, and it is not bound to the browser that started
the flow.

```mermaid
flowchart TD
    P["PUT /auth/idps/{idp_id}"] --> RD{"the body carries the type,<br/>the issuer, a disable, the client<br/>or the callback?"}
    RD -- yes --> G["the provider read once"]
    G -- "not there" --> N1["404"]
    G --> K["the type and the issuer:<br/>an oidc type needs an issuer (400);<br/>a disable keeps an administrator (403)"]
    RD -- no --> S
    K --> CL{"a change of client_id<br/>or callback_url?"}
    CL -- no --> S
    CL -- yes --> H{"are accounts linked to it?<br/>(HasIdentities)"}
    H -- no --> S
    H -- yes --> A{"has the caller the flag,<br/>and is it not disabled?"}
    A -- no --> F["403 AdministratorOnlyError:<br/>client_id, or callback_url"]
    A -- "no caller reached the use-case" --> W["500"]
    A -- yes --> S["one statement: UPDATE ... WHERE id AND<br/>(type and issuer as stored, or sent back as stored,<br/>OR no identity at the provider)"]
    S -- "a row was written" --> OK["200"]
    S -- "nothing was written" --> R{"is the provider there?"}
    R -- no --> N["404"]
    R -- yes --> C["409 IDPInUseError: cannot be pointed at<br/>another issuer or type: still used by<br/>at least one linked account"]
```

- **The type and the issuer, in the update's own statement**
  (`queryUpdateIDP`), as the delete's rule is in its: the provider is kept
  by the rows as they are when it runs. A value equal to the stored one is
  no change: a form sends the whole provider back, and a github provider's
  sends an empty issuer, which is its stored one. A provider nobody is
  linked to may be pointed anywhere.
- **For everybody, the administrator account too.** The identities are the
  accounts', and the issuer is what says who they are; a provider for
  another issuer is a provider of its own.
- **The client and the callback are the administrator account's while
  accounts are linked** (`signInMayChange`). An application registered
  again, or a frontend that moved, is ordinary, and it is the account that
  answers for every account that does it. It may shut doors, its own
  included: whether the identities match afterwards is the provider's to
  say (some issue one subject per application), and that is not refused.
  The 403 tells a caller with the providers' grant that accounts are linked
  to the provider, which the grant otherwise does not show; it says nothing
  of which. The check and the write are two statements.
- **The secret, the name, the logo, the switches** change as before. A
  secret alone does not move the sign-in: the code is issued to the stored
  client.
- **Not closed here:** the check reads the identities as they are when the
  update runs. A sign-in already in flight, or one that reads the provider
  through a cached copy, may store an identity under a provider that was
  just repointed because nobody was linked to it. Unlike the delete, the
  update does not wait for an identity's insert: an `UPDATE` that changes no
  key column takes `FOR NO KEY UPDATE`, which does not conflict with the
  insert's `KEY SHARE`.

`TestAProviderIsNotRepointedUnderItsIdentities` holds it against a running
service: another issuer under an identity, by the role and by the flag;
another type; the type and issuer sent back with a new name; a github
provider's form; a provider nobody is linked to; the client and the callback
under an identity, by the role and by the flag, and where nobody is linked;
a missing provider. In the unit suite:
`TestAProvidersUpdateKeepsItForItsIdentities` (the statement, whole),
`TestAnUpdateThatWroteNothingSaysWhy`,
`TestWhereAProvidersSignInGoesIsTheAdministratorsWhileAccountsAreLinked`
and `TestAProvidersUpdatePassesTheCaller`.

## What it used to do, and why that was wrong

- endpoints chosen by the type's *name*, so `parseUserInfo(raw, idp.Name)`
  failed for any row not literally named `Google` or `Github`;
- no PKCE, no nonce, no ID token: plain OAuth2 and a user-info fetch, with
  the token exchange on `http.DefaultClient` and no timeout;
- `AccessTypeOffline` + `ApprovalForce` on every provider — Google-only
  options, forcing consent on every sign-in elsewhere;
- GitHub sign-in failed for anyone with a private email or a one-word name;
- accounts matched by email, `local_account` flipped on the way in;
- cookies set by the API on its host, redirect to a per-row URL;
- callback failures answered as JSON to a browser navigation;
- no `enabled` flag, no URL validation, every listing decrypting every secret.
