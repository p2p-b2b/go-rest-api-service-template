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
- [x] Return JSON as default response, even for standard http library errors

## References

- [RFC-7223 -> Hypertext Transfer Protocol (HTTP/1.1): Semantics and Content](https://www.rfc-editor.org/rfc/rfc7231)
- [RESTful web API design](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design)
- [API pagination techniques](https://samu.space/api-pagination/#uuid-primary-keys)
- [How To Do Pagination in Postgres with Golang in 4 Common Ways](https://medium.easyread.co/how-to-do-pagination-in-postgres-with-golang-in-4-common-ways-12365b9fb528)
- [“Cursor Pagination” Profile](https://jsonapi.org/profiles/ethanresnick/cursor-pagination/)
- [Cursor-Based Pagination in Go](https://mtekmir.com/blog/golang-cursor-pagination/)
- [Migrations in YDB using “goose”](https://blog.ydb.tech/migrations-in-ydb-using-goose-58137bc5c303)
- [Multi-hop tracing with OpenTelemetry in Golang](https://faun.pub/multi-hop-tracing-with-opentelemetry-in-golang-792df5feb37c)
- [How to setup your own CA with OpenSSL](https://gist.github.com/soarez/9688998)
- [How to setup your own self signed Certificate Authority (CA) and certificates with OpenSSL 3.x and using ED25519](https://gist.github.com/christiangda/aaa1c5b58dfa17f4d1cf6e42d60f9273#file-howto-self-signed-ca-certs-ed25519-md)

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

## Self-Signed Certificates for Mutual TLS (mTLS)

This service could use self-signed certificates for mutual TLS (mTLS) authentication.

### Requirements

- [OpenSSL 3.x](https://www.openssl.org/)

### Ed25519 Certificates

Generate the Certificate Authority (CA) key and certificate:

```bash
mkdir -p certs/newcerts
cd certs/

# Generate the CA Key Pair
# Reference: https://docs.openssl.org/3.4/man1/openssl-ecparam/
# to get -name parameter user -> openssl ecparam -list_curves
openssl ecparam -genkey -name secp256k1 -out ca.key

# Generate the CA Certificate configuration file
cat <<EOF > ca.cnf
[req]
prompt = no
default_bits = 2048
default_keyfile = ca.key
default_days = 3650
default_md = sha256
utf8 = yes
distinguished_name = dn
x509_extensions = v3_ca

[dn]
C = ES
ST = Barcelona
L = Barcelona
O = ACME
OU = ACME CA
CN = *.acme.com
emailAddress = info@acme.com

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, cRLSign, keyCertSign, digitalSignature, keyEncipherment
EOF

# Create the CA Certificate (10 years)
# NOTE: This will request a PEM pass phrase
openssl req -new -x509 -out ca.crt -config ca.cnf

# Check the CA Certificate
openssl x509 -in ca.crt -text -noout

# Generate the Intermediate CA Key Pair (Optional)
openssl ecparam -genkey -name secp256k1 -out intermediate_ca.key

# Generate the Intermediate CA Certificate configuration file (Optional)
cat <<EOF > intermediate_ca.cnf
[req]
prompt = no
default_bits = 2048
default_keyfile = intermediate_ca.key
default_days = 3650
default_md = sha256
distinguished_name = dn
x509_extensions = v3_ca

[dn]
C = ES
ST = Barcelona
L = Barcelona
O = ACME
OU = ACME Intermediate CA
CN = *.acme.com
emailAddress = intermediate@acme.com

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, cRLSign, keyCertSign, digitalSignature, keyEncipherment
EOF

# Create the Intermediate CA Certificate (10 years) (Optional)
# NOTES:
# + This is to protect the CA private key and Certificate because this could be used to sign other certificates and validate them
# + This will request the CA pass phrase
openssl req -new -out intermediate_ca.csr -config intermediate_ca.cnf

# Check the Intermediate CA Certificate (Optional)
openssl req -in intermediate_ca.csr -noout -text

# Sign the Intermediate CA Certificate with the CA Certificate (Optional)
openssl x509 -req -in intermediate_ca.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out intermediate_ca.crt -days 3650 -sha256

# Generate the infrastructure to Create Private Self-Signed CA Certificates
touch index.txt
echo 1000 > serial

# Generate Sign CA Certificate configuration file
cat <<EOF > sign.ca.cnf
[ ca ]
default_ca = CA_default

[ CA_default ]
new_certs_dir = ./newcerts
database = ./index.txt
serial = ./serial
RANDFILE = ./.rand
certificate = ./intermediate_ca.crt
private_key = ./intermediate_ca.key
default_days = 365
default_md = sha256
policy = policy_any
x509_extensions = v3_ca

[ policy_any ]
# optional, supplied or match
countryName = match
stateOrProvinceName = match
organizationName = match
organizationalUnitName = optional
commonName = supplied
emailAddress = optional

[ v3_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, cRLSign, keyCertSign, digitalSignature, keyEncipherment
EOF

# Generate base domain configuration file
cat <<EOF > req.acme.com.cnf
[ req ]
prompt = no
default_bits = 2048
default_keyfile = acme.com.key
encrypt_key = no
default_md = sha256
utf8 = yes
distinguished_name = dn
req_extensions = v3_req

[ dn ]
C = ES
ST = Barcelona
L = Barcelona
O = ACME
OU = ACME Domain
CN = *.acme.com
emailAddress = info@acme.com

[ v3_req ]
subjectKeyIdentifier = hash
keyUsage = critical, digitalSignature, keyEncipherment
basicConstraints = critical, CA:FALSE
extendedKeyUsage = critical, serverAuth, clientAuth
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = *.acme.com
DNS.2 = acme.com
EOF

# Generate the Domain Key Pair and Certificate Signing Request (CSR):
# Generate the Domain Key Pair
# openssl ecparam -genkey -name secp256k1 -out acme.com.key

# Generate the Domain Certificate Signing Request (CSR)
openssl req -new -out acme.com.csr -config req.acme.com.cnf

# Sign the Domain Certificate Signing Request (CSR) with the Intermediate CA Certificate
# NOTES:
# + This will request the Intermediate CA pass phrase
# + This will request validation of the Domain Certificate Signing Request (CSR)
# + This will request confirmation to sign the Domain Certificate Signing Request (CSR)
openssl ca -config sign.ca.cnf -extfile req.acme.com.cnf -extensions v3_req -in acme.com.csr -out acme.com.crt

# Check the Domain Certificate Signing Request (CSR)
openssl req -in acme.com.csr -noout -text

# Generate the public keys and certificates in PEM format
# NOTES:
# + The public keys are used to verify the signature of the certificates
# + The certificates are used to verify the public keys
# + This will request the pass phrase of the CA and Intermediate CA
openssl ec -in ca.key -pubout -out ca.pub
openssl ec -in intermediate_ca.key -pubout -out intermediate_ca.pub
openssl ec -in acme.com.key -pubout -out acme.com.pub

# Generate the public keys and certificates in DER format
openssl ec -in ca.key -pubout -outform DER -out ca.pub.der
openssl ec -in intermediate_ca.key -pubout -outform DER -out intermediate_ca.pub.der
openssl ec -in acme.com.key -pubout -outform DER -out acme.com.pub.der

# Generate PEM format certificates
openssl x509 -in ca.crt -outform PEM -out ca.pem
openssl x509 -in intermediate_ca.crt -outform PEM -out intermediate_ca.pem
openssl x509 -in acme.com.crt -outform PEM -out acme.com.pem
```

## Building and Running

### Requirement

By default Podman machine (Macbook ARM processors) adds only you $HOME directory to the container.

Reference: <https://github.com/containers/podman/issues/14815>

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
