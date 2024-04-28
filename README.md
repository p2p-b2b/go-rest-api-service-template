# go-service-template

This is a template for a Go service. It includes a basic structure for a Go service, with a Makefile for building and running the service, and a Dockerfile for building a Docker image.

## Getting started

- [] Create a new repository from this template
- [] Change the module name in `go.mod` using the command `go mod edit -module github.com/your-username/your-repo`
- [] Change the service name in `cmd/go-service-template` to the name of your service, e.g. `cmd/your-service-name`
- [] Change the service name in `Makefile` to the name of your service

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

using

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

