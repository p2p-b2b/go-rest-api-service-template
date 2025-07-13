package jwtvalidator

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenValidator_GetClientID(t *testing.T) {
	validator := AccessTokenValidator{
		ClientID: "test-client-id",
	}

	assert.Equal(t, "test-client-id", validator.GetClientID())
}

func TestAccessTokenValidator_Validate(t *testing.T) {
	// Generate a key pair for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Create PEM encoded public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	validator := AccessTokenValidator{
		PublicKey: publicKeyPEM,
		ClientID:  "test-client-id",
	}

	t.Run("valid token", func(t *testing.T) {
		// Create a valid token
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"aud": "test-client-id",
		})

		// Sign the token with our private key
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate the token
		claims, err := validator.Validate(context.Background(), signedToken)
		require.NoError(t, err)
		require.NotNil(t, claims)

		// Check claims
		assert.Equal(t, "user123", claims["sub"])
		assert.Equal(t, "test-client-id", claims["aud"])
	})

	t.Run("invalid signature", func(t *testing.T) {
		// Generate another key pair (different from the one in validator)
		otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		// Create a token signed with the other key
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
		})

		signedToken, err := token.SignedString(otherKey)
		require.NoError(t, err)

		// Validation should fail
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create an expired token
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		})

		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validation should fail due to expiration
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("malformed token", func(t *testing.T) {
		// Test with completely invalid token
		claims, err := validator.Validate(context.Background(), "not.a.jwt.token")
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("invalid token format", func(t *testing.T) {
		// Test with invalid token format
		claims, err := validator.Validate(context.Background(), "invalid-token-format")
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("empty token", func(t *testing.T) {
		// Test with empty token
		claims, err := validator.Validate(context.Background(), "")
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("nil public key", func(t *testing.T) {
		nilKeyValidator := AccessTokenValidator{
			PublicKey: nil,
			ClientID:  "test-client-id",
		}

		// Create a valid token
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"aud": "test-client-id",
		})

		// Sign the token with our private key
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate the token
		claims, err := nilKeyValidator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("invalid public key", func(t *testing.T) {
		invalidKeyValidator := AccessTokenValidator{
			PublicKey: []byte("invalid-public-key"),
			ClientID:  "test-client-id",
		}

		// Create a valid token
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"aud": "test-client-id",
		})

		// Sign the token with our private key
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate the token
		claims, err := invalidKeyValidator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("token with custom claims", func(t *testing.T) {
		// Create a token with custom claims
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub":          "user123",
			"iat":          time.Now().Unix(),
			"exp":          time.Now().Add(time.Hour).Unix(),
			"aud":          "test-client-id",
			"custom_claim": "custom_value",
		})

		// Sign the token with our private key
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate the token
		claims, err := validator.Validate(context.Background(), signedToken)
		require.NoError(t, err)
		require.NotNil(t, claims)

		// Check claims
		assert.Equal(t, "user123", claims["sub"])
		assert.Equal(t, "test-client-id", claims["aud"])
		assert.Equal(t, "custom_value", claims["custom_claim"])
	})
}
