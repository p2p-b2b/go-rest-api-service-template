package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	ErrAuthnInvalidPrivateKey                  = errors.New("invalid private key. Private key must be a valid file and the name must be between [" + strconv.Itoa(ValidAuthnPrivateKeyMinLength) + "] and [" + strconv.Itoa(ValidAuthnPrivateKeyMinLength) + "] characters")
	ErrAuthnInvalidPublicKey                   = errors.New("invalid public key. Public key must be a valid file and the name must be between [" + strconv.Itoa(ValidAuthnPublicKeyMinLength) + "] and [" + strconv.Itoa(ValidAuthnPublicKeyMinLength) + "] characters")
	ErrAuthnInvalidIssuer                      = errors.New("invalid issuer. Issuer must be between [" + strconv.Itoa(ValidAuthnIssuerMinLength) + "] and [" + strconv.Itoa(ValidAuthnIssuerMaxLength) + "] characters")
	ErrAuthnInvalidAccessTokenDuration         = errors.New("invalid access token duration. Duration must be between [" + ValidAuthnAccessTokenMinDuration.String() + " and " + ValidAuthnAccessTokenMaxDuration.String())
	ErrAuthnInvalidRefreshTokenDuration        = errors.New("invalid refresh token duration. Duration must be between [" + ValidAuthnRefreshTokenMinDuration.String() + " and " + ValidAuthnRefreshTokenMaxDuration.String())
	ErrAuthnInvalidSymmetricKey                = errors.New("invalid symmetric key. Symmetric key must be a valid file and the name must be between [" + strconv.Itoa(ValidAuthnSymmetricKeyMinLength) + "] and [" + strconv.Itoa(ValidAuthnSymmetricKeyMinLength) + "] characters")
	ErrAuthnInvalidUserVerificationAPIEndpoint = errors.New("invalid user verification API endpoint. API endpoint must be a valid URL")
	ErrAuthnInvalidUserVerificationTokenTTL    = errors.New("invalid user verification token TTL. Token TTL must be between [" + ValidAuthnMinUserVerificationTokenTTL.String() + "] and [" + ValidAuthnMaxUserVerificationTokenTTL.String() + "]")
)

const (
	ValidAuthnPrivateKeyMinLength         = 3
	ValidAuthnPublicKeyMinLength          = 3
	ValidAuthnSymmetricKeyMinLength       = 3
	ValidAuthnIssuerMinLength             = 3
	ValidAuthnIssuerMaxLength             = 100
	ValidAuthnMaxEntitiesCacheTTL         = 72 * time.Hour
	ValidAuthnMinEntitiesCacheTTL         = 1 * time.Hour
	ValidAuthnAccessTokenMinDuration      = 1 * time.Minute
	ValidAuthnAccessTokenMaxDuration      = 7 * 24 * time.Hour
	ValidAuthnRefreshTokenMinDuration     = 5 * time.Minute
	ValidAuthnRefreshTokenMaxDuration     = 30 * 24 * time.Hour
	ValidAuthnMaxUserVerificationTokenTTL = 3 * 24 * time.Hour
	ValidAuthnMinUserVerificationTokenTTL = 1 * time.Hour

	// DefaultAuthnIssuer is the default issuer of the JWT tokens
	DefaultAuthnIssuer = "https://qu3ry.me"

	// DefaultAuthnAccessTokenDuration is the default duration of the access token
	DefaultAuthnAccessTokenDuration = 5 * time.Minute

	// DefaultAuthnRefreshTokenDuration is the default duration of the refresh token
	DefaultAuthnRefreshTokenDuration = 24 * time.Hour
)

var (
	// DefaultAuthnPrivateKeyFile is the default private key file used to sign the JWT tokens
	// DefaultAuthnPrivateKeyFile = "jwt.key"
	DefaultAuthnPrivateKeyFile = FileVar{os.NewFile(0, "jwt.key"), os.O_RDONLY}

	// DefaultAuthnPublicKeyFile
	DefaultAuthnPublicKeyFile = FileVar{os.NewFile(0, "jwt.pub"), os.O_RDONLY}

	// DefaultSymmetricKeyFile
	DefaultAuthnSymmetricKeyFile = FileVar{os.NewFile(0, "aes-256-symmetric.key"), os.O_RDONLY}

	DefaultAuthnUserVerificationTokenTTL    = 24 * time.Hour
	DefaultAuthnUserVerificationAPIEndpoint = "http://localhost:8080/api/v1/auth/verify"
)

type AuthnConfig struct {
	PrivateKeyFile              Field[FileVar]
	PublicKeyFile               Field[FileVar]
	SymmetricKeyFile            Field[FileVar]
	Issuer                      Field[string]
	AccessTokenDuration         Field[time.Duration]
	RefreshTokenDuration        Field[time.Duration]
	UserVerificationAPIEndpoint Field[string]
	UserVerificationTokenTTL    Field[time.Duration]
}

func NewAuthConfig() *AuthnConfig {
	return &AuthnConfig{
		PrivateKeyFile:              NewField("authn.private.key.file", "AUTHN_PRIVATE_KEY_FILE", "Auth Private Key File used to sign the JWT tokens. Using Elliptic Curve keys (prime256v1)", DefaultAuthnPrivateKeyFile),
		PublicKeyFile:               NewField("authn.public.key.file", "AUTHN_PUBLIC_KEY_FILE", "Auth Public Key File used to verify the JWT tokens", DefaultAuthnPublicKeyFile),
		SymmetricKeyFile:            NewField("authn.symmetric.key.file", "AUTHN_SYMMETRIC_KEY_FILE", "Auth Symmetric Key File used to encrypt/decrypt Application tokens and API tokens", DefaultAuthnSymmetricKeyFile),
		Issuer:                      NewField("authn.issuer", "AUTHN_ISSUER", "Issuer of the JWT tokens", DefaultAuthnIssuer),
		AccessTokenDuration:         NewField("authn.access.token.duration", "AUTHN_ACCESS_TOKEN_DURATION", "Duration of the access token", DefaultAuthnAccessTokenDuration),
		RefreshTokenDuration:        NewField("authn.refresh.token.duration", "AUTHN_REFRESH_TOKEN_DURATION", "Duration of the refresh token", DefaultAuthnRefreshTokenDuration),
		UserVerificationAPIEndpoint: NewField("authn.user.verification.api.endpoint", "AUTHN_USER_VERIFICATION_API_ENDPOINT", "User Verification API Endpoint", DefaultAuthnUserVerificationAPIEndpoint),
		UserVerificationTokenTTL:    NewField("authn.user.verification.token.ttl", "AUTHN_USER_VERIFICATION_TOKEN_TTL", "User Verification Token TTL", DefaultAuthnUserVerificationTokenTTL),
	}
}

// ParseEnvVars reads the server configuration from environment variables
// and sets the values in the configuration
func (ref *AuthnConfig) ParseEnvVars() {
	ref.PrivateKeyFile.Value = GetEnv(ref.PrivateKeyFile.EnVarName, ref.PrivateKeyFile.Value)
	ref.PublicKeyFile.Value = GetEnv(ref.PublicKeyFile.EnVarName, ref.PublicKeyFile.Value)
	ref.SymmetricKeyFile.Value = GetEnv(ref.SymmetricKeyFile.EnVarName, ref.SymmetricKeyFile.Value)
	ref.Issuer.Value = GetEnv(ref.Issuer.EnVarName, ref.Issuer.Value)
	ref.AccessTokenDuration.Value = GetEnv(ref.AccessTokenDuration.EnVarName, ref.AccessTokenDuration.Value)
	ref.RefreshTokenDuration.Value = GetEnv(ref.RefreshTokenDuration.EnVarName, ref.RefreshTokenDuration.Value)
	ref.UserVerificationAPIEndpoint.Value = GetEnv(ref.UserVerificationAPIEndpoint.EnVarName, ref.UserVerificationAPIEndpoint.Value)
	ref.UserVerificationTokenTTL.Value = GetEnv(ref.UserVerificationTokenTTL.EnVarName, ref.UserVerificationTokenTTL.Value)
}

func (ref *AuthnConfig) Validate() error {
	if len(ref.PrivateKeyFile.Value.Name()) <= ValidAuthnPrivateKeyMinLength {
		return ErrAuthnInvalidPrivateKey
	}

	if len(ref.PublicKeyFile.Value.Name()) <= ValidAuthnPublicKeyMinLength {
		return ErrAuthnInvalidPublicKey
	}

	if len(ref.SymmetricKeyFile.Value.Name()) <= ValidAuthnSymmetricKeyMinLength {
		return ErrAuthnInvalidSymmetricKey
	}

	if ref.Issuer.Value == "" || len(ref.Issuer.Value) < ValidAuthnIssuerMinLength || len(ref.Issuer.Value) > ValidAuthnIssuerMaxLength {
		return ErrAuthnInvalidIssuer
	}

	if ref.AccessTokenDuration.Value <= ValidAuthnAccessTokenMinDuration || ref.AccessTokenDuration.Value > ValidAuthnAccessTokenMaxDuration {
		return ErrAuthnInvalidAccessTokenDuration
	}

	if ref.RefreshTokenDuration.Value <= ValidAuthnRefreshTokenMinDuration || ref.RefreshTokenDuration.Value > ValidAuthnRefreshTokenMaxDuration {
		return ErrAuthnInvalidRefreshTokenDuration
	}

	if _, err := url.Parse(ref.UserVerificationAPIEndpoint.Value); err != nil {
		return ErrAuthnInvalidUserVerificationAPIEndpoint
	}

	if ref.UserVerificationTokenTTL.Value < ValidAuthnMinUserVerificationTokenTTL || ref.UserVerificationTokenTTL.Value > ValidAuthnMaxUserVerificationTokenTTL {
		return ErrAuthnInvalidUserVerificationTokenTTL
	}

	return nil
}
