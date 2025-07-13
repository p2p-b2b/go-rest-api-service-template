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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenValidator_GetClientID(t *testing.T) {
	validator := RefreshTokenValidator{
		ClientID: "test-client-id",
	}

	assert.Equal(t, "test-client-id", validator.GetClientID())
}

func TestRefreshTokenValidator_Validate(t *testing.T) {
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

	validator := RefreshTokenValidator{
		PublicKey: publicKeyPEM,
		ClientID:  "test-client-id",
	}

	t.Run("valid token", func(t *testing.T) {
		// Create a valid token with jti claim
		jtiValue, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"jti": jtiValue.String(),
			"aud": "test-client-id", // Add audience claim matching the ClientID
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
		assert.Equal(t, jtiValue.String(), claims["jti"])
		assert.Equal(t, "test-client-id", claims["aud"])
	})

	t.Run("invalid signature", func(t *testing.T) {
		// Generate another key pair (different from the one in validator)
		otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		// Create a token signed with the other key
		uid, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"jti": uid.String(),
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
		uid, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"jti": uid.String(),
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		})

		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validation should fail due to expiration
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("missing jti", func(t *testing.T) {
		// Create a token without jti claim
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
		})

		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validation should fail due to missing jti
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)

		var invalidClaimsError *InvalidClaimsError
		assert.ErrorAs(t, err, &invalidClaimsError)
	})

	t.Run("invalid jti format", func(t *testing.T) {
		// Create a token with invalid jti format
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"jti": "not-a-uuid",
		})

		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validation should fail due to invalid jti format
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)

		var invalidClaimsError *InvalidClaimsError
		assert.ErrorAs(t, err, &invalidClaimsError)
	})

	t.Run("nil jti", func(t *testing.T) {
		// Create a token with nil UUID as jti
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"jti": uuid.Nil.String(),
		})

		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validation should fail due to nil jti
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)

		var invalidClaimsError *InvalidClaimsError
		assert.ErrorAs(t, err, &invalidClaimsError)
	})

	t.Run("malformed token", func(t *testing.T) {
		// Test with completely invalid token
		claims, err := validator.Validate(context.Background(), "not.a.jwt.token")
		assert.Nil(t, claims)
		assert.Error(t, err)

		var invalidTokenError *InvalidTokenError
		assert.ErrorAs(t, err, &invalidTokenError)
	})

	t.Run("invalid token format", func(t *testing.T) {
		// Test with invalid token format
		claims, err := validator.Validate(context.Background(), "invalid-token-format")
		assert.Nil(t, claims)
		assert.Error(t, err)

		var invalidTokenError *InvalidTokenError
		assert.ErrorAs(t, err, &invalidTokenError)
	})

	t.Run("empty token", func(t *testing.T) {
		// Test with empty token
		claims, err := validator.Validate(context.Background(), "")
		assert.Nil(t, claims)
		assert.Error(t, err)

		var invalidTokenError *InvalidTokenError
		assert.ErrorAs(t, err, &invalidTokenError)
	})

	t.Run("nil public key", func(t *testing.T) {
		validator := RefreshTokenValidator{
			PublicKey: nil,
			ClientID:  "test-client-id",
		}

		// Create a valid token with jti claim
		jtiValue, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"jti": jtiValue.String(),
			"aud": "test-client-id", // Add audience claim matching the ClientID
		})

		// Sign the token with our private key
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate the token
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("invalid public key", func(t *testing.T) {
		validator := RefreshTokenValidator{
			PublicKey: []byte("invalid-public-key"),
			ClientID:  "test-client-id",
		}

		// Create a valid token with jti claim
		jtiValue, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"jti": jtiValue.String(),
			"aud": "test-client-id", // Add audience claim matching the ClientID
		})

		// Sign the token with our private key
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate the token
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)
		assert.Error(t, err)
	})

	t.Run("valid token with audience claim", func(t *testing.T) {
		// Create a valid token with jti claim and audience claim
		jtiValue, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"jti": jtiValue.String(),
			"aud": "test-client-id", // Add audience claim matching the ClientID
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
		assert.Equal(t, jtiValue.String(), claims["jti"])
		assert.Equal(t, "test-client-id", claims["aud"])
	})

	t.Run("valid token with custom claims", func(t *testing.T) {
		// Create a valid token with custom claims
		jtiValue, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub":          "user123",
			"iat":          time.Now().Unix(),
			"exp":          time.Now().Add(time.Hour).Unix(),
			"jti":          jtiValue.String(),
			"aud":          "test-client-id", // Add audience claim matching the ClientID
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
		assert.Equal(t, jtiValue.String(), claims["jti"])
		assert.Equal(t, "test-client-id", claims["aud"])
		assert.Equal(t, "custom_value", claims["custom_claim"])
	})

	t.Run("jti is not a string", func(t *testing.T) {
		// Create a token with jti claim as an integer
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub": "user123",
			"jti": 12345, // jti is an integer
		})

		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validation should fail due to jti not being a string
		claims, err := validator.Validate(context.Background(), signedToken)
		assert.Nil(t, claims)

		var invalidClaimsError *InvalidClaimsError
		assert.ErrorAs(t, err, &invalidClaimsError)
	})
}
