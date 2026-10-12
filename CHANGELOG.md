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
- A filter or a sort field is one the list can serve. (#116, #118)
- The Rego bundle passes Regal, which `make lint` now runs. (#113, #114)
- The development stack syncs the container VM's clock, which freezes while
  a laptop sleeps. (#110)

## [0.0.3] - 2026-09-06

This file starts after this release. What it contains is in
[its release notes](https://github.com/slashdevops/go-rest-api-service-template/releases/tag/v0.0.3).

[Unreleased]: https://github.com/slashdevops/go-rest-api-service-template/compare/v0.0.3...HEAD
[0.0.3]: https://github.com/slashdevops/go-rest-api-service-template/releases/tag/v0.0.3
