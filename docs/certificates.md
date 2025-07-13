# Certificates

## Asymmetric Private & Public Key Pair for JWT

This service use a `private and public key pair` to `sign and validate` JWT tokens.

### Requirements

- [OpenSSL 3.x](https://www.openssl.org/)

### Ed25519 Key Pair

Generate the private and public key pair:

```bash
# Create the directory to store the certificates
mkdir -p certs

# Generate the JWT Private Key
openssl ecparam -genkey -name prime256v1 -noout -out certs/jwt.key

# Generate the JWT Public Key
openssl ec -in certs/jwt.key -pubout -out certs/jwt.pub
```

## Symmetric Key for encryption and decryption of Application and API tokens

This service use a `symmetric key` to `encrypt and decrypt` application tokens stored in the database.
Application tokens are used to authenticate the application to the service and these
are long-lived JWT or JOT tokens without expiration time.

This also use to encrypt and decrypt API tokens stored in the database.

### Requirements for AES-256 Key

- [OpenSSL 3.x](https://www.openssl.org/)

### AES-256 Key

Generate the symmetric key:

```bash
# Create the directory to store the certificates
mkdir -p certs

# Generate the AES-256 Key, hexadecimal format is important!
openssl rand -hex 32 | tr -d '\n' > certs/aes-256-symmetric-hex.key
```

## Self-Signed Certificates

This service could use self-signed certificates.

### Requirements for Self-Signed Certificates

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
O = Peer to Peer and Business to Business SL
OU = qu3ry.me
CN = *.qu3ry.me
emailAddress = info@qu3ry.me

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
O = Peer to Peer and Business to Business SL
OU = qu3ry.me
CN = *.qu3ry.me
emailAddress = intermediate@qu3ry.me

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
cat <<EOF > req.qu3ry.me.cnf
[ req ]
prompt = no
default_bits = 2048
default_keyfile = qu3ry.me.key
encrypt_key = no
default_md = sha256
utf8 = yes
distinguished_name = dn
req_extensions = v3_req

[ dn ]
C = ES
ST = Barcelona
L = Barcelona
O = Peer to Peer and Business to Business SL
OU = qu3ry.me
CN = *.qu3ry.me
emailAddress = info@qu3ry.me

[ v3_req ]
subjectKeyIdentifier = hash
keyUsage = critical, digitalSignature, keyEncipherment
basicConstraints = critical, CA:FALSE
extendedKeyUsage = critical, serverAuth, clientAuth
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = localhost
DNS.2 = 127.0.0.1
DNS.3 = *.qu3ry.me
DNS.4 = qu3ry.me
EOF

# Generate the Domain Key Pair and Certificate Signing Request (CSR):
# Generate the Domain Key Pair
# openssl ecparam -genkey -name secp256k1 -out qu3ry.me.key

# Generate the Domain Certificate Signing Request (CSR)
openssl req -new -out qu3ry.me.csr -config req.qu3ry.me.cnf

# Sign the Domain Certificate Signing Request (CSR) with the Intermediate CA Certificate
# NOTES:
# + This will request the Intermediate CA pass phrase
# + This will request validation of the Domain Certificate Signing Request (CSR)
# + This will request confirmation to sign the Domain Certificate Signing Request (CSR)
openssl ca -config sign.ca.cnf -extfile req.qu3ry.me.cnf -extensions v3_req -in qu3ry.me.csr -out qu3ry.me.crt

# Check the Domain Certificate Signing Request (CSR)
openssl req -in qu3ry.me.csr -noout -text

# Generate the public keys and certificates in PEM format
# NOTES:
# + The public keys are used to verify the signature of the certificates
# + The certificates are used to verify the public keys
# + This will request the pass phrase of the CA and Intermediate CA
openssl ec -in ca.key -pubout -out ca.pub
openssl ec -in intermediate_ca.key -pubout -out intermediate_ca.pub
openssl ec -in qu3ry.me.key -pubout -out qu3ry.me.pub

# Generate the public keys and certificates in DER format
openssl ec -in ca.key -pubout -outform DER -out ca.pub.der
openssl ec -in intermediate_ca.key -pubout -outform DER -out intermediate_ca.pub.der
openssl ec -in qu3ry.me.key -pubout -outform DER -out qu3ry.me.pub.der

# Generate PEM format certificates
openssl x509 -in ca.crt -outform PEM -out ca.pem
openssl x509 -in intermediate_ca.crt -outform PEM -out intermediate_ca.pem
openssl x509 -in qu3ry.me.crt -outform PEM -out qu3ry.me.pem
```
