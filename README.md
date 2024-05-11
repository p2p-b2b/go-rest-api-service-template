# go-service-template

This is a template for a Go HTTP REST API Service.

## Features

- [x] Create a new repository from this template
- [x] Change the module name in `go.mod` using the command `go mod edit -module github.com/your-username/your-repo`
- [x] Change the service name in `cmd/go-service-template` to the name of your service, e.g. `cmd/your-service-name`
- [x] Change the service name in `Makefile` to the name of your service
- [x] Hot reload with [air](https://github.com/cosmtrek/air), use `make install-air` to install it, then `air` to run it
- [x] Ready to use Certificates for HTTPS, see [Certificates](#certificates)
- [x] Containerize your service with Podman, see [Building](#building)
- [x] Podman pod file for development [dev-service-pod.yaml](dev-service-pod.yaml), see [Running](#running). This is something like `docker-compose/podman-compose` but more powerful
- [x] Database migrations with [goose](https://github.com/pressly/goose). Check the [db](db/README.md) documentation

## References

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
  --dir "./cmd/go-service-template/,./internal/handler" \
  --output ./docs \
  --parseDependency true \
  --parseInternal true
```

## Certificates

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

## Building

using Docker:

```bash
docker build \
   --platform linux/arm64 \
   --tag p2p-b2b/go-service-template:v1.0.0-linux-arm64 \
   --tag p2p-b2b/go-service-template:v1.0.0-linux-arm64 \
    --build-arg SERVICE_NAME=go-service-template
    --build-arg GOOS=linux
    --build-arg GOARCH=arm64
    --build-arg BUILD_DATE=2024-04-28T15:26:26
    --build-arg BUILD_VERSION=v1.0.0
   --file ./Containerfile .
```

Using Podman:

```bash
podman build
    --platform linux/arm64
    --manifest p2p-b2b/go-service-template:v1.0.0
    --tag p2p-b2b/go-service-template:v1.0.0-linux-arm64
    --build-arg SERVICE_NAME=go-service-template
    --build-arg GOOS=linux
    --build-arg GOARCH=arm64
    --build-arg BUILD_DATE=2024-04-28T15:26:26
    --build-arg BUILD_VERSION=v1.0.0
    --file ./Containerfile .
```

## Running

### With podman pod file

Requirements:

By default Podman machine (Macbook ARM processors) adds only you $HOME directory to the container.

Reference: [https://github.com/containers/podman/issues/14815]

To add more directories you need to create a new machine with the following command:

```bash
podman machine stop
podman machine rm
podman machine init -v /tmp:/tmp
podman machine start
```

__WARNING:__ This will remove the current machine and all the containers.

Start the pod:

```bash
mkdir -p /tmp/go-service-template-db-volume-host
podman play kube dev-service-pod.yaml
```

Stop the pod:

```bash
# WARNING: --force remove the volumes declared in the pod file
podman play kube --force --down dev-service-pod.yaml
```

### PostgreSQL

Start a PostgreSQL container:

```bash
mkdir -p /tmp/go-service-template-db-volume-host
podman volume create \
      -o device=/tmp/go-service-template-db-volume-host \
      -o=o=bind \
      go-service-template-db-volume-host

podman run --name postgres \
  --env POSTGRES_USER=username \
  --env POSTGRES_PASSWORD=password \
  --publish 5432:5432 \
  --volume go-service-template-db-volume-host:/var/lib/postgresq/data \
  --detach postgres:16
```

Stop the PostgreSQL container:

```bash
podman stop postgres
```

Remove the PostgreSQL container and volume:

```bash
podman rm postgres
podman volume rm local-db-volume-host-0
```

#### [OPTIONAL] Connect to the PostgreSQL container

```bash
podman exec -it postgres psql -U username
```

#### [OPTIONAL] Connect to the PostgreSQL database from the host

```bash
PGPASSWORD=password psql -h localhost -p 5432 -U username
```

```bash
# this is a personal access token (classic)
export CR_PAT=ghp_uxxxxxxx
podman login ghcr.io -u p2p-b2b -p $CR_PAT

make container-publish
```

From ghcr.io:

```bash
podman run --name go-service-template --rm ghcr.io/p2p-b2b/go-service-template:first-implementation
```

## Pagination

```sql
-- all users
SELECT *
FROM users
ORDER BY created_at DESC, id DESC;

-- first query
WITH usrs AS (
 SELECT *
 FROM users usrs
 ORDER BY usrs.created_at DESC, id DESC
 LIMIT 3
)
SELECT * FROM usrs ORDER BY created_at DESC, id DESC;

-- moving forward
WITH usrs AS (
 SELECT *
 FROM users usrs
 WHERE usrs.created_at < '2024-03-08T08:00:00Z'
  AND (usrs.id < 'c3b11505-9606-4046-b1f2-7a2a5cf6df58'
  OR usrs.created_at < '2024-03-08T08:00:00Z')
 ORDER BY usrs.created_at DESC, usrs.id DESC
 LIMIT 3
)
SELECT * FROM usrs ORDER BY created_at DESC, id DESC;

-- moving backward
WITH usrs AS (
 SELECT *
 FROM users usrs
 WHERE usrs.created_at > '2024-03-07T07:00:00Z'
  AND (usrs.id > '085d16e2-0200-47f9-8bdc-732dd12677be'
  OR usrs.created_at > '2024-03-07T07:00:00Z')
 ORDER BY usrs.created_at ASC, usrs.id ASC
 LIMIT 3
)
SELECT * FROM usrs ORDER BY created_at DESC, id DESC;
```
