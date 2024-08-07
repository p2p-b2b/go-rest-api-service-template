# go-rest-api-service-template

This is a template for a Go HTTP REST API Service.

## Features

- [x] Create a new repository from this template
- [x] Change the module name in `go.mod` using the command `go mod edit -module github.com/your-username/your-repo`
- [x] Change the service name in `cmd/go-rest-api-service-template` to the name of your service, e.g. `cmd/your-service-name`
- [x] Change the service name in `Makefile` to the name of your service
- [x] Hot reload with [air](https://github.com/cosmtrek/air), use `make install-air` to install it, then `air` to run it
- [x] Ready to use Certificates for HTTPS, see [Certificates](#certificates)
- [x] Containerize your service with Podman, see [Building](#building)
- [x] Podman pod file for development [dev-service-pod.yaml](dev-service-pod.yaml), see [Running](#running). This is something like `docker-compose/podman-compose` but more powerful
- [x] Database migrations with [goose](https://github.com/pressly/goose). Check the [db](db/README.md) documentation
- [x] Validation with [go-playground/validator](https://github.com/go-playground/validator)
- [x] Allow filtering, sorting, pagination and partial responses
- [x] Middleware for logging and headers versioning

## References

- [RFC-7223 -> Hypertext Transfer Protocol (HTTP/1.1): Semantics and Content](https://www.rfc-editor.org/rfc/rfc7231)
- [RESTful web API design](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design)
- [API pagination techniques](https://samu.space/api-pagination/#uuid-primary-keys)
- [How To Do Pagination in Postgres with Golang in 4 Common Ways](https://medium.easyread.co/how-to-do-pagination-in-postgres-with-golang-in-4-common-ways-12365b9fb528)
- [“Cursor Pagination” Profile](https://jsonapi.org/profiles/ethanresnick/cursor-pagination/)
- [Cursor-Based Pagination in Go](https://mtekmir.com/blog/golang-cursor-pagination/)
- [Migrations in YDB using “goose”](https://blog.ydb.tech/migrations-in-ydb-using-goose-58137bc5c303)

## Swagger Documentation

The Swagger documentation is available at `/swagger/index.html`.

To generate the Swagger documentation, you need to install [swag](https://github.com/swaggo/swag) and run the following command:

```bash
swag init \
  --dir "./cmd/go-rest-api-service-template/,./internal/handler" \
  --output ./docs \
  --parseDependency true \
  --parseInternal true
```

## Self-Signed Certificates

This service uses self-signed certificates for HTTPS. You can generate new certificates using the following command:

```bash
# ECDSA recommendation key ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl req -x509 -nodes -newkey ec:secp384r1 -keyout server.key -out server.crt -days 3650
# openssl req -x509 -nodes -newkey ec:<(openssl ecparam -name secp384r1) -keyout server.ecdsa.key -out server.ecdsa.crt -days 3650
# -pkeyopt ec_paramgen_curve:… / ec:<(openssl ecparam -name …) / -newkey ec:…
ln -sf server.ecdsa.key server.key
ln -sf server.ecdsa.crt server.crt

# RSA recommendation key ≥ 2048-bit
openssl req -x509 -nodes -newkey rsa:2048 -keyout server.key -out server.crt -days 3650
ln -sf server.rsa.key server.key
ln -sf server.rsa.crt server.crt
```

## Running

### Requirements

By default Podman machine (Macbook ARM processors) adds only you $HOME directory to the container.

Reference: [https://github.com/containers/podman/issues/14815]

To add more directories you need to create a new machine with the following command:

```bash
podman machine stop
podman machine rm
podman machine init -v $HOME:$HOME -v /tmp:/tmp
podman machine start
```

__WARNING:__ This will remove the current machine and all the containers.

### Development environment

start

```bash
# create the necessary containers
make start-dev-env

# start the development environment
air
```

stop

```bash
# create the necessary containers
make stop-dev-env
```

[OPTIONAL] Connect to the PostgreSQL database from the host

```bash
PGPASSWORD=password psql -h localhost -p 5432 -U username
```

Database migrations

see [database](database/README.md)
