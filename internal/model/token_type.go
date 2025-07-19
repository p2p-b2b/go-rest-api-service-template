package model

// TokenType represents the type of token.
// It can be either an access token, a refresh token, an email verification token or a password reset token.
type TokenType string

const (
	TokenTypeAccess            TokenType = "access"
	TokenTypeRefresh           TokenType = "refresh"
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypePersonalAccess    TokenType = "personal_access"
	TokenTypeBearer            TokenType = "Bearer" // This is used for the HTTP Authorization header
)

// String returns the string representation of the token type.
func (tt TokenType) String() string {
	return string(tt)
}

// IsValid checks if the token type is valid.
func (tt TokenType) IsValid() bool {
	switch tt {
	case TokenTypeAccess, TokenTypeRefresh, TokenTypeEmailVerification, TokenTypePasswordReset, TokenTypePersonalAccess:
		return true
	}
	return false
}
