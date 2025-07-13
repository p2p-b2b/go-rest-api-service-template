package jwtvalidator

import (
	"context"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenValidator validates JWT tokens issued by Google Identity Platform.
type AccessTokenValidator struct {
	PublicKey []byte
	ClientID  string
}

// Validate validates a JWT token and returns the claims if the token is valid.
func (ref *AccessTokenValidator) Validate(ctx context.Context, token string) (claims map[string]any, err error) {
	validator, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return jwt.ParseECPublicKeyFromPEM(ref.PublicKey)
	})
	if err != nil {
		return nil, err
	}

	if !validator.Valid {
		slog.Error("AccessTokenValidator.Validate: token is invalid")
		return nil, &InvalidTokenError{Message: "token is invalid"}
	}

	claims, ok := validator.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("AccessTokenValidator.Validate: token claims are invalid")
		return nil, &InvalidClaimsError{Message: "failed to cast token claims to jwt.MapClaims"}
	}

	return claims, nil
}

// GetClientID returns the client ID of the AccessTokenValidator.
func (ref *AccessTokenValidator) GetClientID() string {
	return ref.ClientID
}
