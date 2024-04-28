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
