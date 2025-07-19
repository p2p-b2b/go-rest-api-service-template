package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	AuthnUserRegisteredSuccessfully = "User registered successfully"
	AuthnUserVerifiedSuccessfully   = "User verified successfully"
	AuthnUserVerificationEmailSent  = "User verification email sent"
	AuthnUserLoggedOutSuccessfully  = "User logged out successfully"
)

// JWTClaims represents the claims in a JWT token.
//
//	@Description	JWTClaims represents the claims in a JWT token.
type JWTClaims struct {
	Email         string        `json:"email,omitempty"`
	Subject       string        `json:"sub"`
	TokenType     TokenType     `json:"token_type"`
	Issuer        string        `json:"iss"`
	TokenDuration time.Duration `json:"token_duration,omitempty"`
}

// LoginUserInput is the input struct for the LoginUser service.
type LoginUserInput struct {
	Email    string
	Password string
}

// Validate validates the LoginUserRequest.
func (req *LoginUserInput) Validate() error {
	// Validate email with comprehensive validation
	normalizedEmail, err := ValidateEmail(req.Email, "email")
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidEmailError{Email: req.Email, Message: valErr.Message}
		}
		return &InvalidEmailError{Email: req.Email}
	}
	req.Email = normalizedEmail

	// Validate password with enhanced security checks
	if err := ValidatePassword(req.Password, "password"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidPasswordError{Message: valErr.Message, MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
		return &InvalidPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
	}

	return nil
}

// LoginUserOutput is the output struct for the LoginUser service.
type LoginUserOutput struct {
	AccessToken  string
	RefreshToken string
	TokenType    TokenType
	UserID       uuid.UUID
	Resources    map[string]any
}

// RefreshAccessTokenInput is the input struct for the RefreshAccessToken service.
type RefreshAccessTokenInput struct {
	RefreshToken string
}

// Validate validates the RefreshAccessTokenRequest.
func (req *RefreshAccessTokenInput) Validate() error {
	// Validate refresh token with string validation
	token, err := ValidateString(req.RefreshToken, StringValidationOptions{
		MinLength:      1,
		MaxLength:      2048,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,
		FieldName:      "refresh_token",
	})
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidRefreshTokenError{Message: valErr.Message}
		}
		return &InvalidRefreshTokenError{Message: "refresh token cannot be empty"}
	}
	req.RefreshToken = token

	return nil
}

type RefreshAccessTokenOutput struct {
	AccessToken  string
	RefreshToken string
	TokenType    TokenType
}

type RegisterUserInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	ID        uuid.UUID
	Disabled  bool
}

func (ref *RegisterUserInput) Validate() error {
	// Validate UUID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidUserIDError{ID: ref.ID, Message: valErr.Message}
		}
		return &InvalidUserIDError{ID: ref.ID}
	}

	// Validate first name
	firstName, err := ValidateString(ref.FirstName, StringValidationOptions{
		MinLength:      ValidUserFirstNameMinLength,
		MaxLength:      ValidUserFirstNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,
		FieldName:      "first_name",
	})
	if err != nil {
		return &InvalidFirstNameError{MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength}
	}
	ref.FirstName = firstName

	// Validate last name
	lastName, err := ValidateString(ref.LastName, StringValidationOptions{
		MinLength:      ValidUserLastNameMinLength,
		MaxLength:      ValidUserLastNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,
		FieldName:      "last_name",
	})
	if err != nil {
		return &InvalidLastNameError{MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength}
	}
	ref.LastName = lastName

	// Validate email
	normalizedEmail, err := ValidateEmail(ref.Email, "email")
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidEmailError{Email: ref.Email, Message: valErr.Message}
		}
		return &InvalidEmailError{Email: ref.Email}
	}
	ref.Email = normalizedEmail

	// Validate password
	if err := ValidatePassword(ref.Password, "password"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidPasswordError{Message: valErr.Message, MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
		return &InvalidPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
	}

	return nil
}

// LoginUserRequest is the request struct for the LoginUser handler.
//
//	@Description	LoginUserRequest is the request struct for the LoginUser handler.
type LoginUserRequest struct {
	Email    string `json:"email" example:"admin@qu3ry.me" format:"email" validate:"required"`
	Password string `json:"password" example:"ThisIsApassw0rd.," format:"string" validate:"required"`
}

// Validate validates the LoginUserRequest.
func (req *LoginUserRequest) Validate() error {
	// Validate email with comprehensive validation
	normalizedEmail, err := ValidateEmail(req.Email, "email")
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidEmailError{Email: req.Email, Message: valErr.Message}
		}
		return &InvalidEmailError{Email: req.Email}
	}
	req.Email = normalizedEmail

	// Validate password with enhanced security checks
	if err := ValidatePassword(req.Password, "password"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidPasswordError{Message: valErr.Message, MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
		return &InvalidPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
	}

	return nil
}

// LoginUserResponse is the response when a user logs in.
//
//	@Description	LoginUserResponse is the response when a user logs in.
type LoginUserResponse struct {
	AccessToken  string `json:"access_token" format:"string"`
	RefreshToken string `json:"refresh_token" format:"string"`
	TokenType    TokenType
	UserID       uuid.UUID      `json:"user_id" example:"01980434-b7ff-7a54-a71f-34868a34e51e" format:"uuid"`
	Resources    map[string]any `json:"permissions" format:"object"`
}

const (
	ValidJWTRefreshTokenMinLength = 50
	ValidJWTRefreshTokenMaxLength = 2048
)

// RefreshTokenRequest is the request struct for the RefreshToken handler.
//
//	@Description	RefreshTokenRequest is the request struct for the RefreshToken handler.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" format:"string"`
}

// Validate validates the RefreshTokenRequest.
func (req *RefreshTokenRequest) Validate() error {
	// Validate refresh token with string validation
	token, err := ValidateString(req.RefreshToken, StringValidationOptions{
		MinLength:      ValidJWTRefreshTokenMinLength,
		MaxLength:      ValidJWTRefreshTokenMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,
		FieldName:      "refresh_token",
	})
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidRefreshTokenError{Message: valErr.Message}
		}
		return &InvalidRefreshTokenError{Message: "refresh token cannot be empty"}
	}
	req.RefreshToken = token

	return nil
}

// RefreshTokenResponse is the response when a user refreshes their token.
//
//	@Description	RefreshTokenResponse is the response when a user refreshes their token.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token" format:"string"`
	RefreshToken string `json:"refresh_token" format:"string"`
	TokenType    TokenType
}

// RegisterUserRequest is the request struct for the RegisterUser handler.
//
//	@Description	RegisterUserRequest is the request struct for the RegisterUser handler.
type RegisterUserRequest struct {
	FirstName string    `json:"first_name" example:"John" format:"string" validate:"required"`
	LastName  string    `json:"last_name" example:"Doe" format:"string" validate:"required"`
	Email     string    `json:"email" example:"john.doe@email.com" format:"email" validate:"required"`
	Password  string    `json:"password" example:"ThisIsApassw0rd.," format:"string" validate:"required"`
	ID        uuid.UUID `json:"id" example:"01980434-b7ff-7a8b-b8e9-144341357314" format:"uuid" validate:"optional"`
}

// Validate validates the RegisterUserRequest.
func (req *RegisterUserRequest) Validate() error {
	// Validate UUID
	if err := ValidateUUID(req.ID, 7, "id"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidUserIDError{ID: req.ID, Message: valErr.Message}
		}
		return &InvalidUserIDError{ID: req.ID}
	}

	// Validate first name
	firstName, err := ValidateString(req.FirstName, StringValidationOptions{
		MinLength:      ValidUserFirstNameMinLength,
		MaxLength:      ValidUserFirstNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,
		FieldName:      "first_name",
	})
	if err != nil {
		return &InvalidFirstNameError{MinLength: ValidUserFirstNameMinLength, MaxLength: ValidUserFirstNameMaxLength}
	}
	req.FirstName = firstName

	// Validate last name
	lastName, err := ValidateString(req.LastName, StringValidationOptions{
		MinLength:      ValidUserLastNameMinLength,
		MaxLength:      ValidUserLastNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,
		FieldName:      "last_name",
	})
	if err != nil {
		return &InvalidLastNameError{MinLength: ValidUserLastNameMinLength, MaxLength: ValidUserLastNameMaxLength}
	}
	req.LastName = lastName

	// Validate email
	normalizedEmail, err := ValidateEmail(req.Email, "email")
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidEmailError{Email: req.Email, Message: valErr.Message}
		}
		return &InvalidEmailError{Email: req.Email}
	}
	req.Email = normalizedEmail

	// Validate password
	if err := ValidatePassword(req.Password, "password"); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidPasswordError{Message: valErr.Message, MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
		}
		return &InvalidPasswordError{MinLength: ValidUserPasswordMinLength, MaxLength: ValidUserPasswordMaxLength}
	}

	return nil
}

// ReVerifyUserRequest is the request struct for the ReVerifyUser handler.
//
//	@Description	ReVerifyUserRequest is the request struct for the ReVerifyUser handler.
type ReVerifyUserRequest struct {
	Email string `json:"email" format:"email" example:"user@mail.com" required:"true" validate:"required"`
}

// Validate validates the ReVerifyUserRequest.
func (req *ReVerifyUserRequest) Validate() error {
	// Validate email with comprehensive validation
	normalizedEmail, err := ValidateEmail(req.Email, "email")
	if err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			return &InvalidEmailError{Email: req.Email, Message: valErr.Message}
		}
		return &InvalidEmailError{Email: req.Email}
	}
	req.Email = normalizedEmail

	return nil
}
