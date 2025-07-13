package jwtvalidator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RefreshTokenValidator validates JWT tokens issued by Google Identity Platform.
type RefreshTokenValidator struct {
	PublicKey []byte
	ClientID  string
}

// Validate validates a JWT token and returns the claims if the token is valid.
func (ref *RefreshTokenValidator) Validate(ctx context.Context, token string) (claims map[string]any, err error) {
	validator, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return jwt.ParseECPublicKeyFromPEM(ref.PublicKey)
	})
	if err != nil {
		return nil, &InvalidTokenError{Message: fmt.Sprintf("failed to parse token: %v", err)}
	}

	if !validator.Valid {
		slog.Error("RefreshTokenValidator.Validate: token is invalid")
		return nil, &InvalidClaimsError{Message: "token is invalid"}
	}

	claims, ok := validator.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("RefreshTokenValidator.Validate: token claims are invalid")
		return nil, &InvalidClaimsError{Message: "token claims are invalid"}
	}

	// jti is only present in refresh tokens
	if _, ok := claims["jti"]; !ok {
		slog.Error("RefreshTokenValidator.Validate: jti claim is missing")
		return nil, &InvalidClaimsError{Message: "jti claim is missing"}
	}

	if claims["jti"] == nil {
		slog.Error("RefreshTokenValidator.Validate: jti claim is nil")
		return nil, &InvalidClaimsError{Message: "jti claim is nil"}
	}

	// Type assertion for jti claim to ensure it's a string
	jtiStr, ok := claims["jti"].(string)
	if !ok {
		slog.Error("RefreshTokenValidator.Validate: jti claim is not a string")
		return nil, &InvalidClaimsError{Message: "jti claim is not a string"}
	}

	jti, err := uuid.Parse(jtiStr)
	if err != nil {
		slog.Error("RefreshTokenValidator.Validate: failed to parse jti", "error", err)
		return nil, &InvalidClaimsError{Message: "failed to parse jti claim"}
	}

	if jti == uuid.Nil {
		slog.Error("RefreshTokenValidator.Validate: jti cannot be nil")
		return nil, &InvalidClaimsError{Message: "jti cannot be nil"}
	}

	return claims, nil
}

// GetClientID returns the client ID of the RefreshTokenValidator.
func (ref *RefreshTokenValidator) GetClientID() string {
	return ref.ClientID
}
