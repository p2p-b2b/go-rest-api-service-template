package service

import (
	"golang.org/x/crypto/bcrypt"
)

// hashAndSaltPassword hashes and salts the password.
func hashAndSaltPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// comparePasswords compares the hashed password and the plain password.
func comparePasswords(hashedPwd string, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}

// TokenType represents the type of token.
// It can be either an access token, a refresh token, an email verification token or a password reset token.
type TokenType string

const (
	TypeAccessToken            TokenType = "access"
	TypeRefreshToken           TokenType = "refresh"
	TypeEmailVerificationToken TokenType = "email_verification"
	TypePasswordResetToken     TokenType = "password_reset"
)

// String returns the string representation of the token type.
func (tt TokenType) String() string {
	return string(tt)
}

// IsValid checks if the token type is valid.
func (tt TokenType) IsValid() bool {
	switch tt {
	case TypeAccessToken, TypeRefreshToken, TypeEmailVerificationToken, TypePasswordResetToken:
		return true
	}
	return false
}
