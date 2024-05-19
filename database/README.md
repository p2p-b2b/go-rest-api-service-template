# db

Documentation for the database.

This folder is here because go:embed does not work with files outside the module root.

## how to use Goose

Install goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Export the database connection string

```bash
export DATABASE_DSN="host=localhost port=5432 user=username password=password dbname=go-rest-api-service-template sslmode=disable TimeZone=UTC"
````

```bash
Run goose status

```bash
goose -dir database/migrations postgres $DATABASE_DSN status
```

Run goose up

```bash
goose -dir database/migrations postgres $DATABASE_DSN up
```

Run goose down

```bash
goose -dir database/migrations postgres $DATABASE_DSN down
```

Create a new migration

```bash
goose -dir database/migrations create create_user_table sql
```
