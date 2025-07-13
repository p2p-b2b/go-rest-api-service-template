package service

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func Encrypt(plaintext, symmetricKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func Decrypt(ciphertext, symmetricKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func CiphertextToString(ciphertext []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func StringToCiphertext(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// HashAndSaltPassword hashes and salts the password
// It uses bcrypt to hash the password with a cost of 10.
// The hashed password is returned as a string.
func HashAndSaltPassword(password string, cost ...int) (string, error) {
	var costVal int
	if len(cost) > 0 {
		if cost[0] < bcrypt.MinCost || cost[0] > bcrypt.MaxCost {
			return "", fmt.Errorf("cost value must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
		}
		costVal = cost[0]
	} else {
		costVal = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), costVal)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// ComparePasswords compares the hashed password and the plain password
// It uses bcrypt to compare the hashed password with the plain password.
// It returns true if the passwords match, false otherwise.
func ComparePasswords(hashedPwd string, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}

// createJWT creates a JWT token.
// It uses the private key to sign the token.
// The token is signed using the ES256 algorithm.
// The token contains the user email, the token ID, the token type, the issuer, the audience, the subject, the issued at and the expiration time.
func createJWT(claims model.JWTClaims, privateKey []byte) (string, error) {
	if claims.Subject == "" {
		return "", fmt.Errorf("subject is required")
	}

	if !claims.TokenType.IsValid() {
		return "", fmt.Errorf("invalid token type")
	}

	if claims.Issuer == "" {
		return "", fmt.Errorf("issuer is required")
	}

	// Generate a access token
	type tokenCustomClaims struct {
		Email     string          `json:"email,omitempty"`
		TokenType model.TokenType `json:"token_type"`
		jwt.RegisteredClaims
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	tokenClaims := tokenCustomClaims{
		Email:     claims.Email,
		TokenType: claims.TokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uid.String(),
			Issuer:    claims.Issuer,
			Audience:  jwt.ClaimStrings{claims.Issuer},
			Subject:   claims.Subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(claims.TokenDuration)),
		},
	}

	if claims.TokenDuration > time.Second {
		tokenClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(claims.TokenDuration))
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodES256, tokenClaims)
	signKey, err := jwt.ParseECPrivateKeyFromPEM(privateKey)
	if err != nil {
		slog.Error("service.createAccessToken", "error", err)
		return "", err
	}

	// get the key kid
	kid := signKey.Params().N.String()
	// add the kid to the header
	accessToken.Header["kid"] = kid

	tokenSigned, err := accessToken.SignedString(signKey)
	if err != nil {
		slog.Error("service.createAccessToken", "error", err)
		return "", err

	}

	return tokenSigned, nil
}

// verifyJWT verifies a JWT token and returns the claims.
// It uses the public key to verify the token.
func verifyJWT(token string, publicKey []byte) (jwt.MapClaims, error) {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, &model.InvalidJWTError{Message: "invalid JWT kid not in header"}
		}

		// get the public key
		publicKey, err := jwt.ParseECPublicKeyFromPEM(publicKey)
		if err != nil {
			return nil, err
		}

		// get the key from the kid
		if kid != publicKey.Params().N.String() {
			return nil, &model.InvalidJWTError{Message: "invalid JWT kid"}
		}

		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, &model.InvalidJWTError{Message: "token is invalid"}
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &model.InvalidJWTError{Message: "claims are invalid"}
	}

	return claims, nil
}

func CountWords(text string) int {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanWords)
	count := 0

	for scanner.Scan() {
		count++
	}

	return count
}

// convertToSQLRegex replaces UUIDs and * in a resource string with SQL regex patterns.
// It converts UUIDs to a regex pattern that matches any string of characters (.*).
// Example: "projects/123e4567-e89b-12d3-a456-426614174000/details" becomes "projects/.*?/details".
// The function also adds ^ at the beginning and $ at the end of the string to ensure it matches the entire string.
func convertToSQLRegex(resource string) string {
	reUUID := regexp.MustCompile(model.ValidUUIDOrStarRegex)

	// https://regex101.com/r/4bn9da/1
	resource = reUUID.ReplaceAllString(resource, "\\{[a-z_]{1,50}\\}")

	resource = `^` + resource + `$`

	return resource
}
