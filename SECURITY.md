# Security policy

## Reporting a vulnerability

Report it privately, through GitHub: **Security → Report a vulnerability** on
this repository
([direct link](https://github.com/slashdevops/go-rest-api-service-template/security/advisories/new)).
Please do not open a public issue or a pull request for it.

Say what you found, where (a route, a file, a commit), and how you observed
it. A request and the answer it got are worth more than a description. You
will get an answer in the advisory thread, and the fix is discussed there
until it is released.

## What is supported

This repository is a template: you copy it, and from then on the code is
yours. So there are no maintained release branches, and one supported line:

| Version | Supported |
| --- | --- |
| `main` | yes: fixes land here |
| the latest tag | it gets no patch releases; the next tag carries the fixes |
| anything older | no |

## How a fork learns of a fix

A fork does not receive a fix by itself. To know what to carry over:

- **[`CHANGELOG.md`](./CHANGELOG.md)** has a **Security** section in every
  release. Each entry names what changed, in the terms of the behaviour, and
  the pull request that changed it, so that the diff can be read and ported.
- **[Security advisories](https://github.com/slashdevops/go-rest-api-service-template/security/advisories)**
  are published for what a fork should not leave unpatched.
- **Releases**: watch the repository for releases to be told of both.

Dependency advisories are a separate matter: `govulncheck` runs weekly
(`.github/workflows/security-scan.yaml`) and Dependabot opens the updates. A
fork inherits both workflows and has to keep them running.

## What is in scope

The service as it is in this repository: its routes, its authentication and
authorization, its handling of tokens and of stored secrets, its SQL, and
the defaults it ships with. The design and the reasons behind each rule are
in [`docs/architecture/security.md`](./docs/architecture/security.md),
[`authentication.md`](./docs/architecture/authentication.md) and
[`authorization.md`](./docs/architecture/authorization.md); a report that
shows one of the invariants written there does not hold is exactly what this
policy is for.

Not in scope: the development stack's fixed credentials and keys (they are
for a laptop, and the documentation says what to replace), and a deployment
that changed the shipped defaults.
