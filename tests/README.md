# integration

This folder contains a set of tests

## Requirements

- The `certs` folder and certificates in the root project folder, the README.md file in the root project folder explains how to generate the certificates.
- The `integration.env` file in the `test/integration` folder, this file contains the environment variables used to run the integration tests

```bash
MAIL_SMTP_HOST=localhost
MAIL_SMTP_PORT=1025
MAIL_SMTP_USERNAME=welcome@qu3ry.me
MAIL_SMTP_PASSWORD=new_secure_password
DB_USERNAME=username
DB_PASSWORD=password

```

## How to run integration tests

The `Makefile` in the root project folder contains a targets to prepare and run the integration tests.

Build the integration tests image:

```bash
make container-build-integration-test
```

Start the integration tests environment:

```bash
make start-integration-test
```

Run the integration tests:

This target will execute the previous target if the integration tests environment is not running.

```bash
make test-integration
```

## Run test manually

You can run the integration tests manually using the following command:

```bash
go test -v -race -tags=integration ./test/integration
```

Run certain test manually:

```bash
go test -v -race -tags=integration ./test/integration -run 'TestUser*'
```
