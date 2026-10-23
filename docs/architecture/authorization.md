# Authorization: roles, policies, and the decision

Who may call what. This document is the model as it is. It was ported from
`svc-qu3ry-core`'s authorization review of 2026-09-06, and each section says
what changed there and why, because the same code shipped here.

## The model

```mermaid
erDiagram
    users ||--o{ users_roles : has
    roles ||--o{ users_roles : ""
    roles ||--o{ roles_policies : composes
    policies ||--o{ roles_policies : ""
    resources ||--o{ policies : "names an endpoint"
    roles {
        uuid id
        text name
        bool system
        bool auto_assign
    }
    policies {
        uuid id
        text allowed_action "GET | POST | ... | *"
        text allowed_resource "/users/* | * | /me"
        bool system
    }
    resources {
        uuid id
        text action
        text resource "/users/{user_id}, from swagger"
    }
```

- A **policy** is one grant: an action and a resource pattern. The pattern
  grammar is a whitelist (`/[a-z_]{1,50}`, a UUID, or `*` per segment, at
  most eight), which is why the engine needs no escaping.
- A **role** composes policies; a **user** holds roles. `system` rows cannot
  be edited or deleted; copy one to tune it.
- **`resources`** are the API's endpoints, generated from the Swagger spec.
  They are referential: a policy must name a real endpoint when it is
  created, and nothing reads them at decision time.

### The seeded roles

| Role | Holds | Meant for |
| --- | --- | --- |
| `Administrator` | `*` on `*` | the bootstrap account |
| `AuthenticatedUser` (auto-assigned) | logout, refresh, password reset, `/me`, `/me/authz`, `/me/identities`, `/me/resources_limits`, IdP linking | every account |
| `ReadOnly` | `GET` on `*` | auditors, dashboards |
| `ProjectMember` | the projects the account belongs to, and their products | the ordinary user |
| `ProjectManager` | `ProjectMember` plus creating projects and managing their members (listing users to pick them) | a team lead |
| `UserManager` | users, their roles, projects and password resets; reading roles | an account administrator |

`GET /me` and `PUT /me` were seeded from the start and linked to no role, so
every non-administrator got 403 on their own profile; `AuthenticatedUser`
holds them now. The seed is `00008`; the default roles are its last block,
and `products` is the one entity here that the core does not have, so
`ProjectMember` holds its six routes in place of the RAG ones.

## The decision

```mermaid
sequenceDiagram
    participant M as middleware.CheckAuthz
    participant U as UsersService.SelectAuthz
    participant C as cache authz:<uid>
    participant E as policyopa.Engine
    M->>M: sub from the claims, HEAD folded to GET
    M->>U: SelectAuthz(uid)
    U->>C: get
    C-->>U: hit, or fetch users→roles→policies once
    U-->>M: {"permissions": {"users": {uid: {pattern: [actions]}}}}
    M->>E: IsAllowed(uid, action, path, permissions)
    E->>E: prepared query, permissions in input
    E-->>M: allow, or an error
    M-->>M: error → 500 · false → 403, logged · true → next
```

**OPA, used properly.** The policy is a Rego module compiled **once** in
`policyopa.New`; each request evaluates the prepared query with the decision
as input. The caller's permission set travels in `input`, the idiomatic place
for request-scoped data, so no store is built per request. It used to be
parsed and compiled on every authenticated request, with the permissions
loaded as OPA's data document each time.

The policy itself (`policyopa/rego/bundle/authorization/policy.rego`):

- `*` as a resource is the global grant; a literal path matches exactly; a
  path with `*` expands each `*` to one lowercase UUID segment, anchored, so
  `/users/*` admits `/users/<uuid>` and refuses `/users/me`, `/roles/*`
  refuses `/roles/<uuid>/users`, `/rate_limits/*` refuses
  `/rate_limits/effective`.
- `*` as an action means every method, wherever it appears. It used to mean
  that only on the global resource, so `*` on `/roles` validated, inserted
  and admitted nothing.
- The administrator rule is membership, `"*" in actions`, not list equality.
  It used to be `== ["*"]`, so attaching one more `*`-resource policy turned
  the list into `["*", "GET"]` and locked every administrator out of every
  non-GET call, including the unlink that would have undone it. Measured.

**Fail closed, everywhere.** No grants is a refusal; a store fault, an
engine error or a non-boolean answer is a 500 and a refusal; a denial is a
403 logged at `Warn` with the request id and counted with an `authorized`
label on the metric.

**Tests that run.** `make lint` runs `opa check --strict`, `opa fmt`,
`opa test --coverage --threshold 100` and `regal lint` on the bundle; `policyopa.TestEngineDecisions`
is the Go twin of `policy_test.rego`, driving the same cases through the
port with the wire-shaped map; `usecase.TestIsAuthorized` covers the use
case's failure paths. The Rego tests existed before and were never run.

**After the grant**: `RequireProjectMembership` (see
[security.md](./security.md)) asks whether the caller belongs to the project
named in the path, because a grant on `/projects/*/…` cannot know that.

**Disabled accounts.** A token issued before an account was disabled keeps
verifying: a token is checked for its signature, its expiry and the denylist,
and none of that asks about the account. `RequireActiveAccount` does, on
every chain that verifies an access token but logout, after the token checks
and before the grants are looked up.

```mermaid
flowchart TD
    T["an access token that verifies"] --> S{"RequireActiveAccount:<br/>the account's standing,<br/>read from the database"}
    S -- "active" --> Z["CheckAuthz, and on"]
    S -- "disabled" --> R["401, code token_revoked:<br/>the session is over"]
    S -- "not there any more" --> R
    S -- "the read failed" --> F["500: fail closed"]
    L["DELETE /auth/logout"] -. "not asked: a disabled account<br/>can still end its tokens" .-> Z2["logout"]
```

- **It was on the `/me` chain alone** (`CheckUserExists`), so a disabled
  account's access token was not asked again on any other route until it
  expired.
- **401 with the code `token_revoked`**, for a disabled account and for one
  that is not there, on every chain. It is the one answer a client must not
  meet with refresh-and-retry: the session is over. `/me` answered 403
  "account is disabled" and 404 "user not found" for the two; it answers
  this now.
- **Read from the database on every request**, not from the cache
  (`UsersService.Standing`). The cached account lives twelve hours and an
  invalidation is best effort: a failed one would leave a disabled account
  working for half a day. One read by primary key per authenticated request,
  of the one column it needs (`SelectDisabled`): it read the whole row, the
  password hash included, for one boolean.
- **A read that fails is not permission to continue**: a 500. It fails
  closed.
- **Not on logout.** And nothing is revoked at the disable: the refusal is
  what makes "disabled" hold, and an account that is enabled again has its
  tokens back.
- **The refresh, and a sign-in through an identity provider, ask the
  database too** (`UserServiceConsumer.Standing`). Both read the account
  through the twelve-hour cache, for its data, and decided `disabled` from
  that copy, which an invalidation that failed leaves saying "enabled".
- **What it costs.** One read by primary key, a span and a count in each of
  two layers, on every authenticated request; the users use-case's call
  count now follows the traffic. A request the post-auth limiter refuses
  has paid for that read first. And while the database is unreachable every
  authenticated request is a 500 at this check. That includes
  `/health/detailed`, the readiness target: its 503 with the per-component
  payload is what the handler answers when it is reached, and with the
  database down the account's read fails before it. A probe still sees a
  failure and takes the replica out; it does not get the payload. The public
  `/health/status` answers 500 for a failed database ping, and
  `/health/live` checks nothing.
- **Not asked:** the password-reset and verification chains, which carry
  no access token.

`TestADisabledAccountsTokensAreRefusedOnEveryRoute` holds it against a
running service (a route of each chain, enabled again, a logout that ends
the session, a deleted account, and an account disabled behind the cache's
back, on a request and on a refresh) and fails with the check off the
general chain. `TestEveryAccessTokenChainAsksAboutTheAccount` reads the
wiring: the three chains, that nothing else in the file verifies an access
token, and that the logout chain is given to the authentication handler
alone. A chain added without the check, or the check dropped from one,
compiles and passes every other unit test.

## Who may grant what

Every request in the escalation chain was, on its own, one the caller was
allowed: `POST /policies`, `POST /roles`, `POST /roles/{id}/policies`,
`POST /users/{id}/roles`. The policy engine cannot see that together they
mint an allow-all policy and hand it to the caller. `usecase.GrantGuard`
can: before a policy is created, before an edit changes a policy's action or
resource, before policies are linked to a role, before roles are linked to a
user, the caller's **live** grants (the repository, never the cache) must
cover every grant being handed out. Renaming a policy or rewording its
description widens nothing and is not guarded: a UserManager may tidy a
policy they do not hold, but may not change what it grants.

```mermaid
flowchart LR
    H[handler: caller from the claims] --> S[use case]
    S --> G{GrantGuard: does the caller hold each grant?}
    G -- "* on *" --> OK[proceed]
    G -- same action on the same pattern, or on * --> OK
    G -- otherwise --> R[403 grant refused: the caller does not hold X]
```

"Holds" is literal on purpose: the same action on the same pattern, or on
`*`. A holder of `GET /projects/*` may not hand out `GET /projects/<one id>`
though it would be sound; narrowing patterns is the engine's job, and doing
it twice, differently, is how the two would drift. The seeded `UserManager` holds only
user, role-read and token routes, so on its own it can assign no role whose
grants it does not hold; an administrator who wants a `UserManager` to hand
out `ProjectMember` gives it `ProjectMember` too. That is the intended
shape: the right to assign a role is the right to hand out its contents.

The two links that make the bootstrap administrator an administrator
(`Administrator` → `Full Access`, the seeded account → `Administrator`) are
refused deletion by a trigger on the link tables, whoever asks. One
`DELETE /roles/<Administrator>/policies` used to lock every administrator out
for good. `PUT /policies/{id}` now cross-checks the `resources` table as
`POST` does, so an edit cannot name an endpoint that does not exist.

`TestGrantGuardStopsTheEscalationChain` keeps the proof: each step of the
chain answers 403 for the regular account and succeeds for an
administrator. `TestBootstrapLinksCannotBeSevered` covers the trigger.

## The cache, and why a revoke is never stale

A user's effective permissions are cached under `authz:<uid>` for hours, with
the user, every role and every policy as dependencies. Each path that
widens or narrows a grant invalidates; the two unlink paths used to
invalidate fewer keys than their link twins (`PoliciesService.UnlinkRoles`
dropped only the policy's, `RolesService.UnlinkPolicies` only the role's),
leaving a revoke to the dependency cascade, which nothing tested. Both
invalidate the same keys as their link now, and
`TestARevokedGrantBitesOnTheNextRequest` drives all three unlink routes and
asserts the next request is refused. One `authzCacheKey` helper spells the
key; eight sites used to spell it by hand beside an interface that nothing
implemented.

Grants written straight into the database bypass every invalidation and
stay invisible until the TTL; that is by design, and it bit two live
proofs during the review. Grant through the API.

## Adding a project-scoped route

A grant names a route; membership names a project. OPA expands the `*` in
`/projects/*/…` to any uuid and cannot know membership, so
`middleware.RequireProjectMembership` asks the database once per
project-scoped request, keyed on the `project_id` path value, **after**
`CheckAuthz`: no grant is 403, wrong project is 404, the same as a missing
project. An administrator bypasses and is logged as one.

A new project-scoped route needs nothing. A new path wildcard that is not
named `project_id` gets **no** check, so name it `project_id` or add one.

## A route about a user's projects is held to its caller

A grant names a route and membership names a project in the path. Three
routes name a user in the path and their projects somewhere else, so neither
check holds them to a project: `GET`, `POST` and `DELETE
/users/{user_id}/projects`. The repository holds them to the account that
asks.

"The caller's projects" means what `GET /projects` means by it:
`view_projects_users` for the caller, which is every project for an account
with the `admin` flag and the memberships of anyone else.

### The list

`GET /users/{user_id}/projects` answers the projects of the user in its path
**that are the caller's too**.

```mermaid
flowchart TD
    R["GET /users/{user_id}/projects"] --> H["the handler takes the caller<br/>the middleware verified"]
    H -- "no caller on the request" --> W["500: a route wired<br/>without its middleware"]
    H --> S{"SelectByUserID"}
    S --> U["the projects of the user in the path<br/>(view_projects_users)"]
    U --> C{"is the project the caller's too?<br/>(the same view, for the caller)"}
    C -- "the caller has the admin flag" --> A["every project of that user"]
    C -- "the caller is that user" --> O["their own list, whole"]
    C -- "anyone else" --> I["the projects the two share"]
    S -- "no caller reached<br/>the repository" --> X["refused: never the<br/>user's whole list"]
```

- **A list asked for with no caller is refused twice.** The handler answers
  the 500 of a route wired without its middleware and asks nothing; the
  repository refuses a nil caller on its own, so that a second caller of it
  written without one reads nothing instead of everything.
- **The list is the intersection of two accounts' row sets in one view.** A
  caller asking about themselves gets what `GET /projects` gives them.
- **Both predicates are in the page's statement**, so a filter and a walk by
  page token stay inside them.
- **An empty list says only that nothing is shared.** A user who does not
  exist, a user with no project and a user whose projects are none of the
  caller's are answered alike.
- **What this means for a `UserManager`.** The role holds this route's
  grant. Without the `admin` flag it reads, of another user's projects, the
  ones it is linked to itself: an account that manages users and is a member
  of nothing reads none.

### The links

`POST` and `DELETE /users/{user_id}/projects` take their projects from the
body: **a caller links or unlinks a user only on a project the caller is
linked to.**

```mermaid
flowchart TD
    R["POST or DELETE /users/{user_id}/projects<br/>body: project_ids"] --> H["the handler takes the caller<br/>the middleware verified"]
    H -- "no caller on the request" --> W["500: a route wired<br/>without its middleware"]
    H --> C{"projectsOfCaller:<br/>is every id in the caller's<br/>row set of view_projects_users?"}
    C -- "one is not: another account's<br/>project, or none at all" --> B["400 on project_ids, naming the id;<br/>nothing is written"]
    C -- "all are (every project<br/>for the admin flag)" --> X["the statement runs"]
    X -- "link: the user is not there" --> N["404 User not found"]
    X -- "link: the project went<br/>between check and insert" --> B2["400 on project_ids"]
    X --> OK["200"]
```

- **The whole request is refused**, as a foreign key would refuse it: a body
  that names one project that is not the caller's links none of the others.
- **Another account's project and one that does not exist are refused in the
  same words**, with the id the caller sent
  (`domain.ProjectNotOfCallerError`). Two answers follow from that: linking
  a project that does not exist is this 400 and no longer a 404 "User not
  found", and unlinking one is this 400 and no longer a 200 that unlinked
  nothing.
- **A link's two foreign keys are told apart by name**: the project's is the
  same refusal, for a project that was the caller's at the check and is gone
  at the insert; the user's stays the 404. Migration `00005` writes the two
  names out; they are the ones Postgres gives them, so the schema is the
  same.
- **The check and the write are two statements**, as every check of this
  kind in the service is: a caller unlinked from a project between the two
  still makes the one change that was checked. The window is two statements
  wide; a project deleted in it is the foreign key's refusal above.
- **Unlinking a user who is not in the project** is still a 200: the project
  is the caller's, and there is nothing to remove.
- The other side, `POST` and `DELETE /projects/{project_id}/users`, has its
  project in the path and is held by the membership check.

### Where it is held

Against a running service: `TestAnotherUsersProjectsAreOnlyTheCallersToo`
and `TestAUsersProjectLinksChangeOnlyOnTheCallersProjects`. The unit suite
has no database, so the wiring is held where it can run:
`TestListByUserIDPassesThePathsUserAndTheCaller` and
`TestAUsersProjectLinksPassThePathsUserAndTheCaller` (the handlers pass the
path's user and the caller as what each is),
`TestAUsersProjectLinksReachTheRepositoryWithTheirCaller` and
`TestAUsersProjectsReachTheRepositoryWithTheirCaller` (the use-cases hand
both on, and the refusal with its type),
`TestAUsersProjectsAreReadHeldToTheCaller` (the list's statement),
`TestProjectsOfCaller` (the check),
`TestAUsersProjectLinksAreCheckedBeforeTheyAreWritten` (both writes ask
first, with the caller) and `TestALinksTwoForeignKeysAreToldApart`.
