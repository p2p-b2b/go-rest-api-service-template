# Changelog

What changed, by release, for whoever runs this service or carries its
changes into a copy of it. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Every release has
a **Security** section when it has anything to say there: read that one
first if you maintain a fork. Each entry names the pull request, so that the
change can be read and ported.

Reporting a vulnerability: [`SECURITY.md`](./SECURITY.md).

## [Unreleased]

Since `v0.0.3`.

### Security

- **An identity provider's delete strands nobody, and neither its delete
  nor its disable shuts the last door.** The delete took the provider's
  identities with it and left every account that signed in through it, with
  no password, without a way in. It is a `409` while an account's only way
  in is an identity at that provider; disable the provider instead, or
  delete those accounts or make them local accounts first. Disabling or
  deleting a provider is a `403` when it would leave no administrator
  account that can sign in, and an account's own unlink keeps an identity
  at an enabled provider. The design is in
  [`docs/architecture/identity-providers.md`](./docs/architecture/identity-providers.md#deleting-or-disabling-a-provider).
  ([#132](https://github.com/slashdevops/go-rest-api-service-template/pull/132))
- **An administrator account is locked, and unlocked, only by one.** An
  account with the `admin` flag is disabled or enabled, made a federated or
  a local account, stripped of a role or deleted only by an account that
  has the flag, and the last one that can still sign in (with a password,
  or through an enabled provider) by nobody. A role
  that manages users managed that account like any other. From anyone else
  such a change is a `403`; a form that sends back the values it read is
  unaffected. The design is in
  [`docs/architecture/security.md`](./docs/architecture/security.md#an-administrator-account-is-locked-only-by-one).
  ([#131](https://github.com/slashdevops/go-rest-api-service-template/pull/131))
- **A list that fails says nothing of the cause.** A list answered an
  unexpected failure with `500` and the error's own text as the message;
  that text is the driver's or the database's, and a connection that cannot
  be opened names where the service connects to and as whom. A `500` from a list is now
  `internal server error`, with the cause on the request's log line under
  its request id. A `sort`, `filter` or `fields` that the service refuses
  is still a `400` that says what to correct.
  ([#128](https://github.com/slashdevops/go-rest-api-service-template/pull/128))
- **The three routes about a user's projects are held to the projects of
  whoever asks.** `GET /users/{user_id}/projects` answers the projects of
  that user that are the caller's too, and `POST` and `DELETE
  /users/{user_id}/projects` change a link only on a project the caller is
  linked to. For an account with the `admin` flag that is every project.
  The design is in
  [`docs/architecture/authorization.md`](./docs/architecture/authorization.md#a-route-about-a-users-projects-is-held-to-its-caller).
  ([#127](https://github.com/slashdevops/go-rest-api-service-template/pull/127))
- **An account an administrator disabled is not opened by its verification
  link.** A registration is disabled until its link is followed, and the
  link must open only that. An account created through `POST /users` is
  stored verified, and a `disabled` set through `PUT /users/{user_id}` is
  remembered in a new column, `users.disabled_by_admin`; the verification
  enables only a row without it.
  ([#125](https://github.com/slashdevops/go-rest-api-service-template/pull/125))
- **A disabled or deleted account is refused on every route.** A token that
  verifies says nothing about its account: every chain that verifies an
  access token now asks whether the account may act, and answers a disabled
  or deleted one `401` with the code `token_revoked`. The refresh and a
  sign-in through an identity provider ask the database too.
  ([#124](https://github.com/slashdevops/go-rest-api-service-template/pull/124))
- **Only the administrator account changes an account's e-mail address.**
  The address is where a password reset is sent. From any other caller a
  change of it through `PUT /users/{user_id}` is a `403`.
  ([#123](https://github.com/slashdevops/go-rest-api-service-template/pull/123))
- **A list expression over its length is refused before it is parsed.**
  `filter`, `sort` and `fields` are bounded: a query string is not covered
  by the request-body limit.
  ([#121](https://github.com/slashdevops/go-rest-api-service-template/pull/121))
- **The filter parser bounds its nesting** (`qfv` v1.0.3).
  ([#117](https://github.com/slashdevops/go-rest-api-service-template/pull/117))
- **A repository trace no longer carries an address and a password hash**
  at a level that is exported.
  ([#104](https://github.com/slashdevops/go-rest-api-service-template/pull/104))
- **Listing users needs a grant for it**, and is not every account's.
  ([#101](https://github.com/slashdevops/go-rest-api-service-template/pull/101))
- **The hardening phases**: the middleware chain and its order, response
  headers, body and content-type bounds, error hygiene, project membership,
  single-use links, the outbound address guard. The design is in
  [`docs/architecture/security.md`](./docs/architecture/security.md).
  ([#94](https://github.com/slashdevops/go-rest-api-service-template/pull/94))

### Upgrade notes

- **`DELETE /auth/idps/{idp_id}` can answer `409`**, and it and
  `PUT /auth/idps/{idp_id}` with `enabled: false` can answer `403`. A client
  that deleted providers to retire them disables them now. (#132)

- **The `users` table has a new column**, added to the file that creates
  the table (`00002_users_tables.sql`), as this template does before a first
  deploy. The migration tool does not see an edit: **a database created from
  an earlier version starts, and then fails every write to `users`.**
  Recreate it, or run
  `ALTER TABLE users ADD COLUMN disabled_by_admin BOOLEAN NOT NULL DEFAULT FALSE;`
  and set it by hand on the unverified, disabled rows an administrator
  disabled. (#125)
- **`/me` answers `401` with `token_revoked`** for a disabled or a deleted
  account, where it answered `403` and `404`. A client that treated those
  two apart has one case now, and it is "the session is over". (#124)
- **`PUT /users/{user_id}` can answer `403` for a body that carries
  `email`.** A form that sends back the address it read is unaffected. (#123)
- **Four routes can answer `403` for an account with the `admin` flag**:
  `PUT /users/{user_id}` for a change of `disabled` or `local_account`,
  `DELETE /users/{user_id}`, `DELETE /users/{user_id}/roles` and
  `DELETE /roles/{role_id}/users`, when the caller does not have the flag,
  and for everybody when it is the last such account that can sign in.
  Deleting the seeded administrator, and taking the Administrator role from
  it through either route, is a `403` where it was a `500`. (#131)
- **An account created through `POST /users` is `verified`** in what the
  API answers. (#125)
- **`GET /users/{user_id}/projects` answers fewer rows to a caller without
  the `admin` flag**: the ones the caller is linked to too. And `POST` and
  `DELETE /users/{user_id}/projects` answer `400` for a project that is not
  one of the caller's, **and for one that does not exist**, where a link
  answered `404` "User not found" and an unlink `200`. A role that manages
  users and is a member of no project links nobody. (#127)
- `/health/detailed` answers `500`, not its `503` with the per-component
  payload, while the database is unreachable: the account check in front of
  it needs the database. A probe still sees a failure. (#124)
- Go 1.27.2. (#122)

### Changed

- **The filter parser is `qfv` v1.0.4**, which reads three inputs as
  PostgreSQL does. A number or a boolean as a pattern (`name LIKE 1`,
  `SIMILAR TO 1`, the `~~` operators) is refused, as PostgreSQL has no such
  operator for any column. `YES` and `NO` are names and no boolean literals:
  `disabled = yes` is refused, where it reached PostgreSQL as a column that
  does not exist. `IS` in lower case is the keyword (`id is null`), which was
  refused in that one spelling. A `400` before and after, in other words for
  the first two. (#129)
- A failed request is one log line, at the level of its answer: `ERROR` for
  a `5xx`, `INFO` for a `4xx`. (#115)
- Logs go through OpenTelemetry as a third signal, beside the local output
  and never instead of it; every request-path log carries its trace; the
  instruments are one pair for every layer, and the dashboards are rebuilt.
  (#102, #103, #105, #106, #107)
- Authorization: the policy is compiled once and the caller's grants travel
  as input; a caller may grant only what they hold; default roles. (#96)
- `verified` is its own column, beside `disabled`; a password is 8 to 72
  bytes, and login does not apply the rule for choosing one. (#99, #108)

### Fixed

- Deleting a user or an identity provider that does not exist is a `404`
  and refunds nothing. (#119)
- `POST /users/{user_id}/projects` with a project named twice in its body
  is a `200`, where it was a `500`. (#127)
- Starting a sign-in through an identity provider that cannot be reached
  (`GET /auth/idp/{idp_id}/login`, `/register`, `/link`) answers `503`, as
  those routes document, where it answered `500`. (#128)
- A filter or a sort field is one the list can serve. (#116, #118)
- The Rego bundle passes Regal, which `make lint` now runs. (#113, #114)
- The development stack syncs the container VM's clock, which freezes while
  a laptop sleeps. (#110)

## [0.0.3] - 2026-09-06

This file starts after this release. What it contains is in
[its release notes](https://github.com/slashdevops/go-rest-api-service-template/releases/tag/v0.0.3).

[Unreleased]: https://github.com/slashdevops/go-rest-api-service-template/compare/v0.0.3...HEAD
[0.0.3]: https://github.com/slashdevops/go-rest-api-service-template/releases/tag/v0.0.3
