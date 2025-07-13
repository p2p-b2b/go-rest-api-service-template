package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/templates"
	"github.com/p2p-b2b/mailer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// AuthnServiceConf represents the configuration for the auth service.
// this implement the UsersRepository interface
type AuthnServiceConf struct {
	Repository                  UsersRepository
	MailQueueService            mailer.MailQueueService
	PrivateKey                  []byte
	PublicKey                   []byte
	AccessTokenDuration         time.Duration
	RefreshTokenDuration        time.Duration
	Issuer                      string
	SenderEmail                 string
	SenderName                  string
	UserVerificationAPIEndpoint string
	UserVerificationTokenTTL    time.Duration
	OT                          *o11y.OpenTelemetry
	MetricsPrefix               string
}

type authServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type AuthnService struct {
	repository                  UsersRepository
	mailQueueService            mailer.MailQueueService
	privateKey                  []byte
	publicKey                   []byte
	issuer                      string
	accessTokenDuration         time.Duration
	refreshTokenDuration        time.Duration
	senderEmail                 string
	senderName                  string
	userVerificationAPIEndpoint string
	userVerificationTokenTTL    time.Duration
	ot                          *o11y.OpenTelemetry
	metricsPrefix               string
	metrics                     authServiceMetrics
}

// NewAuthnService creates a new AuthnService.
func NewAuthnService(conf AuthnServiceConf) (*AuthnService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for AuthnService"}
	}

	if conf.MailQueueService == nil {
		return nil, &model.InvalidMailQueueServiceError{Message: "MailQueueService is nil, but it is required for AuthnService"}
	}

	if len(conf.PrivateKey) == 0 {
		return nil, &model.InvalidPrivateKeyError{Message: "PrivateKey is nil, but it is required for AuthnService"}
	}

	if len(conf.PublicKey) == 0 {
		return nil, &model.InvalidPublicKeyError{Message: "PublicKey is nil, but it is required for AuthnService"}
	}

	if len(conf.Issuer) <= 2 || len(conf.Issuer) > 100 {
		return nil, &model.InvalidIssuerError{Message: "Issuer is invalid, but it is required for AuthnService"}
	}

	if conf.AccessTokenDuration < 1*time.Minute || conf.AccessTokenDuration > 168*time.Hour {
		return nil, &model.InvalidAccessTokenDurationError{Message: "AccessTokenDuration is invalid, but it is required for AuthnService"}
	}

	if conf.RefreshTokenDuration < 5*time.Minute || conf.RefreshTokenDuration > 720*time.Hour {
		return nil, &model.InvalidRefreshTokenDurationError{Message: "RefreshTokenDuration is invalid, but it is required for AuthnService"}
	}

	if conf.UserVerificationAPIEndpoint == "" {
		return nil, &model.InvalidVerificationEndpointError{Message: "UserVerificationAPIEndpoint is required"}
	}

	if conf.UserVerificationTokenTTL < 1*time.Hour || conf.UserVerificationTokenTTL > 72*time.Hour {
		return nil, &model.InvalidJWTError{Message: "UserVerificationTokenTTL is invalid"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for AuthnService"}
	}

	if conf.SenderEmail == "" {
		return nil, &model.InvalidSenderError{Message: "SenderEmail is required"}
	}

	if conf.SenderName == "" {
		return nil, &model.InvalidSenderError{Message: "SenderName is required"}
	}

	ref := &AuthnService{
		repository:                  conf.Repository,
		mailQueueService:            conf.MailQueueService,
		privateKey:                  conf.PrivateKey,
		publicKey:                   conf.PublicKey,
		issuer:                      conf.Issuer,
		senderEmail:                 conf.SenderEmail,
		senderName:                  conf.SenderName,
		userVerificationAPIEndpoint: conf.UserVerificationAPIEndpoint,
		userVerificationTokenTTL:    conf.UserVerificationTokenTTL,
		accessTokenDuration:         conf.AccessTokenDuration,
		refreshTokenDuration:        conf.RefreshTokenDuration,
		ot:                          conf.OT,
	}
	if conf.MetricsPrefix != "" {
		ref.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		ref.metricsPrefix += "_"
	}

	serviceCalls, err := ref.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", ref.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the auth service"),
	)
	if err != nil {
		return nil, err
	}

	ref.metrics.serviceCalls = serviceCalls

	return ref, nil
}

// LoginUser logs in a user.
func (ref *AuthnService) LoginUser(ctx context.Context, input *model.LoginUserInput) (*model.LoginUserOutput, error) {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Authn.LoginUser")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser", "input is nil")
	}

	span.SetAttributes(attribute.String("user.email", input.Email))

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser", "failed to validate input")
	}

	user, err := ref.repository.SelectByEmail(ctx, input.Email)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser", "failed to get user by email")
	}

	// check if user is active
	if user.Disabled != nil && *user.Disabled {
		errorType := &model.UserDisabledError{Username: user.Email}
		return nil, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser")
	}

	// check if the password matches
	if !ComparePasswords(user.PasswordHash, input.Password) {
		errValue := &model.InvalidPasswordError{Message: "invalid password"}
		return nil, o11y.RecordError(ctx, span, errValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser")
	}

	accessTokenJWTClaims := model.JWTClaims{
		Email:         user.Email,
		Subject:       user.ID.String(),
		Issuer:        ref.issuer,
		TokenType:     model.TokenTypeAccess,
		TokenDuration: ref.accessTokenDuration,
	}

	accessToken, err := createJWT(accessTokenJWTClaims, ref.privateKey)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser", "failed to create access token")
	}

	refreshTokenJWTClaims := model.JWTClaims{
		Email:         user.Email,
		Subject:       user.ID.String(),
		Issuer:        ref.issuer,
		TokenType:     model.TokenTypeRefresh,
		TokenDuration: ref.refreshTokenDuration,
	}

	refreshToken, err := createJWT(refreshTokenJWTClaims, ref.privateKey)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser", "failed to create refresh token")
	}

	// get the user permissions
	permissions, err := ref.repository.SelectAuthz(ctx, user.ID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser", "failed to get user permissions")
	}

	if permissions == nil || permissions["permissions"] == nil {
		slog.Warn("service.Authn.LoginUser: user does not have any permissions")
		permissions = map[string]any{
			"permissions": map[string]any{},
		}
	}

	// remove the first level of the permissions map which is the key "permissions"
	permissionsL1, ok := permissions["permissions"].(map[string]any)
	if !ok {
		err := fmt.Errorf("failed to cast permissions to []string")
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.LoginUser")
	}

	result := &model.LoginUserOutput{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Resources:    permissionsL1,
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "login successful")
	return result, nil
}

// RegisterUser creates a new user.
func (ref *AuthnService) RegisterUser(ctx context.Context, input *model.RegisterUserInput) error {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Authn.RegisterUser")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser")
	}

	if input.ID == uuid.Nil {
		var err error
		input.ID, err = uuid.NewV7()
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to generate user ID")
		}
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to validate input")
	}

	hashPwd, err := HashAndSaltPassword(input.Password)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to hash password")
	}

	user := &model.InsertUserInput{
		ID:           input.ID,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PasswordHash: hashPwd,
	}

	if err := ref.repository.Insert(ctx, user); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to insert user")
	}

	jwtClaims := model.JWTClaims{
		Email:         input.Email,
		Subject:       input.ID.String(),
		Issuer:        ref.issuer,
		TokenType:     model.TokenTypeEmailVerification,
		TokenDuration: ref.userVerificationTokenTTL,
	}

	emailToken, err := createJWT(jwtClaims, ref.privateKey)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to create email verification token")
	}

	emailContent, err := templates.NewEmailAccountVerification(&templates.EmailAccountVerificationConf{
		VerificationAPIEndpoint: ref.userVerificationAPIEndpoint,
		VerificationToken:       emailToken,
		VerificationTTL:         ref.userVerificationTokenTTL.String(),
		UserName:                fmt.Sprintf("%s %s", input.FirstName, input.LastName),
		HTML:                    true,
	})
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to create email content")
	}

	mailContent, err := mailer.NewMailContentBuilder().
		WithFromName(ref.senderName).
		WithFromAddress(ref.senderEmail).
		WithToName(fmt.Sprintf("%s %s", input.FirstName, input.LastName)).
		WithToAddress(input.Email).
		WithMimeType(mailer.MimeTypeTextHTML).
		WithSubject("Account Verification").
		WithBody(emailContent.Render()).
		Build()
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to create mail content")
	}

	if err := ref.mailQueueService.Enqueue(mailContent); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RegisterUser", "failed to enqueue mail content")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "user created successfully")
	return nil
}

// VerifyUser verifies a user.
func (ref *AuthnService) VerifyUser(ctx context.Context, jwtToken string) error {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Authn.VerifyUser")
	defer span.End()

	if jwtToken == "" {
		errorType := &model.InvalidJWTError{Message: "JWT token is empty"}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	claims, err := verifyJWT(jwtToken, ref.publicKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenInvalidClaims) {
			errValue := &model.InvalidJWTError{Value: jwtToken, Message: "invalid JWT claims"}
			return o11y.RecordError(ctx, span, errValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
		}

		if errors.Is(err, jwt.ErrTokenUsedBeforeIssued) {
			errValue := &model.InvalidJWTError{Value: jwtToken, Message: "token used before issued"}
			return o11y.RecordError(ctx, span, errValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			errorValue := &model.InvalidJWTError{Message: "token expired"}
			return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
		}

		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			errorValue := &model.InvalidJWTError{Message: "token signature invalid"}
			return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
		}

		if errors.Is(err, jwt.ErrTokenUnverifiable) {
			errorValue := &model.InvalidJWTError{Message: "token unverifiable"}
			return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
		}

		if errors.Is(err, jwt.ErrTokenMalformed) {
			errorValue := &model.InvalidJWTError{Message: "token malformed"}
			return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
		}

		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok {
		errorValue := &model.InvalidJWTError{Message: "token_type claim is missing"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	if tokenType != model.TokenTypeEmailVerification.String() {
		errorValue := &model.InvalidJWTError{Message: "token_type claim is invalid"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	email, ok := claims["email"].(string)
	if !ok {
		errorValue := &model.InvalidJWTError{Message: "email claim is missing"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	// check the expiration time
	exp, ok := claims["exp"].(float64)
	if !ok {
		errorValue := &model.InvalidJWTError{Message: "exp claim is missing"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	if time.Now().Unix() > int64(exp) {
		errorValue := &model.InvalidJWTError{Value: jwtToken, Message: "token expired"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	user, err := ref.repository.SelectByID(ctx, userID)
	if err != nil {
		// grateful answer when user not found, because security reason
		var userNotFoundError *model.UserNotFoundError
		if errors.As(err, &userNotFoundError) {
			return nil
		}

		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	if user == nil {
		ErrorType := &model.UserNotFoundError{ID: userID.String()}
		return o11y.RecordError(ctx, span, ErrorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	if email != user.Email {
		errorValue := &model.InvalidJWTError{Message: "email claim is invalid"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	if user.Disabled == nil || !*user.Disabled {
		errorType := &model.UserAlreadyVerifiedError{Email: email}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	isDisabled := new(bool)
	*isDisabled = false

	updateInput := &model.UpdateUserInput{
		ID:       user.ID,
		Disabled: isDisabled,
	}

	if err := ref.repository.UpdateByID(ctx, updateInput); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.VerifyUser")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "user verified successfully")
	return nil
}

// ReVerifyUser re-verifies a user.
func (ref *AuthnService) ReVerifyUser(ctx context.Context, email string) error {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Authn.ReVerifyUser")
	defer span.End()

	if email == "" {
		errorType := &model.InvalidEmailError{Email: email, Message: "The email is empty. Please provide a valid email address."}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser")
	}

	if len(email) < model.ValidUserEmailMinLength || len(email) > model.ValidUserEmailMaxLength {
		errorType := &model.InvalidEmailError{Email: email, Message: fmt.Sprintf("The email must be between %d and %d characters long.", model.ValidUserEmailMinLength, model.ValidUserEmailMaxLength)}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		errorType := &model.InvalidEmailError{Email: email, Message: fmt.Sprintf("The email '%s' is not valid.", email)}
		return o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser")
	}

	user, err := ref.repository.SelectByEmail(ctx, email)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser")
	}

	// grateful answer when user is already verified, because security reason
	if user.Disabled != nil && !*user.Disabled {
		return nil
	}

	jwtClaims := model.JWTClaims{
		Email:         user.Email,
		Subject:       user.ID.String(),
		Issuer:        ref.issuer,
		TokenType:     model.TokenTypeEmailVerification,
		TokenDuration: ref.userVerificationTokenTTL,
	}

	emailToken, err := createJWT(jwtClaims, ref.privateKey)
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser", "failed to create email verification token")
	}

	emailContent, err := templates.NewEmailAccountVerification(
		&templates.EmailAccountVerificationConf{
			VerificationAPIEndpoint: ref.userVerificationAPIEndpoint,
			VerificationToken:       emailToken,
			VerificationTTL:         ref.userVerificationTokenTTL.String(),
			UserName:                fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			HTML:                    true,
		})
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser", "failed to create email content")
	}

	mailContent, err := mailer.NewMailContentBuilder().
		WithFromName(ref.senderName).
		WithFromAddress(ref.senderEmail).
		WithToName(fmt.Sprintf("%s %s", user.FirstName, user.LastName)).
		WithToAddress(user.Email).
		WithMimeType(mailer.MimeTypeTextHTML).
		WithSubject("Account Verification").
		WithBody(emailContent.Render()).
		Build()
	if err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser", "failed to create mail content")
	}

	if err := ref.mailQueueService.Enqueue(mailContent); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.ReVerifyUser", "failed to enqueue mail content")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "user re-verification email sent successfully")
	return nil
}

// LoggingOutUser logs out a user.
func (ref *AuthnService) LoggingOutUser(ctx context.Context, userID string) error {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Authn.LoggingOutUser")
	defer span.End()

	// TBI - The method is currently just a placeholder, but now uses our helper methods pattern
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "user logged out successfully")
	return nil
}

// RefreshAccessToken refreshes an access token.
func (ref *AuthnService) RefreshAccessToken(ctx context.Context, input *model.RefreshAccessTokenInput) (*model.RefreshAccessTokenOutput, error) {
	ctx, span, metricAttrs := ref.setupContext(ctx, "service.Authn.RefreshAccessToken")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidRefreshTokenError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken")
	}

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "failed to validate input")
	}

	refreshToken, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (any, error) {
		return jwt.ParseECPublicKeyFromPEM(ref.publicKey)
	})
	if err != nil {
		invalid := jwt.ErrTokenInvalidClaims
		if errors.Is(err, invalid) {
			if errors.Is(err, jwt.ErrTokenExpired) {
				errorValue := &model.InvalidJWTError{Value: input.RefreshToken, Message: "token expired"}
				return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken")
			}

			if errors.Is(err, jwt.ErrTokenUsedBeforeIssued) {
				errValue := &model.InvalidJWTError{Message: "token used before issued"}
				return nil, o11y.RecordError(ctx, span, errValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken")
			}

			errorValue := &model.InvalidJWTError{Value: input.RefreshToken, Message: "invalid JWT"}

			return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken")
		}

		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "failed to parse refresh token")
	}

	if !refreshToken.Valid {
		errorValue := &model.InvalidRefreshTokenError{Message: "refresh token is invalid"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "refresh token is invalid")
	}

	claims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok {
		errorValue := &model.InvalidRefreshTokenError{Message: "failed to get claims from refresh token"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "failed to get claims from refresh token")
	}

	// The jti claim is required for a refresh token only and difference it from an access token
	if claims["jti"] == nil || claims["jti"] == "" {
		errorValue := &model.InvalidRefreshTokenError{Message: "jti claim is missing"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "jti claim is missing")
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "failed to parse user ID")
	}

	slog.Debug("service.Authn.RefreshAccessToken", "userID", userID)

	userEmail, ok := claims["email"].(string)
	if !ok {
		errorValue := &model.InvalidRefreshTokenError{Message: "email claim is missing"}
		// The email claim is required for a refresh token only and difference it from an access token
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "email claim is missing")
	}
	slog.Debug("service.Authn.RefreshAccessToken", "userEmail", userEmail)

	tokenType, ok := claims["token_type"].(string)
	if !ok {
		errorValue := &model.InvalidRefreshTokenError{Message: "token_type claim is missing"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "token_type claim is missing")
	}

	if tokenType != model.TokenTypeRefresh.String() {
		errorValue := &model.InvalidRefreshTokenError{Message: "token_type is not a refresh token"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "token_type is not a refresh token")
	}

	user := &model.User{
		ID:    userID,
		Email: userEmail,
	}

	jwtClaims := model.JWTClaims{
		Email:         user.Email,
		Subject:       user.ID.String(),
		Issuer:        ref.issuer,
		TokenType:     model.TokenTypeAccess,
		TokenDuration: ref.accessTokenDuration,
	}

	accessTokenSigned, err := createJWT(jwtClaims, ref.privateKey)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricAttrs, "service.Authn.RefreshAccessToken", "failed to create access token")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricAttrs, "access token refreshed successfully")
	return &model.RefreshAccessTokenOutput{
		AccessToken: accessTokenSigned,
		TokenType:   "Bearer",
	}, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *AuthnService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}
