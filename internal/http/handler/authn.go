package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/authn.go -source=authn.go AuthnService

// AuthnService is the interface that must be implemented by the service that
// the AuthnHandler will use to authenticate users.
type AuthnService interface {
	LoginUser(ctx context.Context, input *model.LoginUserInput) (*model.LoginUserOutput, error)
	LoggingOutUser(ctx context.Context, userID string) error
	RefreshAccessToken(ctx context.Context, input *model.RefreshAccessTokenInput) (*model.RefreshAccessTokenOutput, error)
	RegisterUser(ctx context.Context, input *model.RegisterUserInput) error
	VerifyUser(ctx context.Context, jwtToken string) error
	ReVerifyUser(ctx context.Context, email string) error
}

// AuthnHandlerConf is the configuration struct for the AuthnHandler.
type AuthnHandlerConf struct {
	Service       AuthnService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type authnHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// AuthnHandler is the handler that will handle the authentication of users.
type AuthnHandler struct {
	service       AuthnService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       authnHandlerMetrics
}

// NewAuthnHandler creates a new AuthnHandler.
func NewAuthnHandler(conf AuthnHandlerConf) (*AuthnHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "AuthnService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	ref := &AuthnHandler{
		service:       conf.Service,
		ot:            conf.OT,
		metricsPrefix: conf.MetricsPrefix,
	}

	if conf.MetricsPrefix != "" {
		ref.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		ref.metricsPrefix += "_"
	}

	handlerCalls, err := ref.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", ref.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the auth handler"),
	)
	if err != nil {
		return nil, err
	}

	ref.metrics.handlerCalls = handlerCalls

	return ref, nil
}

// RegisterRoutes registers the routes for the AuthnHandler.
func (ref *AuthnHandler) RegisterRoutes(mux *http.ServeMux, accessTokenMiddleware, refreshTokenMiddleware middleware.Middleware) {
	mux.Handle("DELETE /auth/logout", accessTokenMiddleware.ThenFunc(ref.logout))
	mux.Handle("POST /auth/refresh", refreshTokenMiddleware.ThenFunc(ref.refreshAccessToken))

	mux.HandleFunc("POST /auth/login", ref.loginUser)
	mux.HandleFunc("POST /auth/register", ref.registerUser)
	mux.HandleFunc("GET /auth/verify/{jwt}", ref.verifyUser)
	mux.HandleFunc("POST /auth/verify", ref.reVerifyUser)
}

// loginUser login a user and return its JWT tokens.
//
//	@Id				0198042a-f9c5-7547-a6e7-567af5db26cd
//	@Summary		Login user
//	@Description	Authenticate user credentials and return JWT access and refresh tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.LoginUserRequest	true	"The information of the user to login"
//	@Success		200		{object}	model.LoginUserResponse
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		401		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/auth/login [post]
func (ref *AuthnHandler) loginUser(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Authn.loginUser")
	defer span.End()

	var req model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorType := &model.InvalidRequestError{Message: "failed to decode request"}
		e := recordError(ctx, span, errorType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.loginUser")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.loginUser")
		respond.WriteJSONMessage(w, r, http.StatusUnauthorized, e.Error())
		return
	}

	loginUserInput := &model.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	out, err := ref.service.LoginUser(ctx, loginUserInput)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.loginUser")
		respond.WriteJSONMessage(w, r, http.StatusUnauthorized, e.Error())
		return
	}

	resp := model.LoginUserResponse{
		UserID:       out.UserID,
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		Resources:    out.Resources,
		TokenType:    out.TokenType,
	}

	if err := respond.WriteJSONData(w, http.StatusOK, resp); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.loginUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Authn.loginUser: user logged in", "user_id", resp.UserID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "user logged in")
}

// registerUser Register a new user and send a confirmation email.
//
//	@Id				0198042a-f9c5-75c8-9231-ad5fc9e7b32e
//	@Summary		Register user
//	@Description	Create a new user account and send email verification
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.RegisterUserRequest	true	"The information of the user to register"
//	@Success		201		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/auth/register [post]
func (ref *AuthnHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Authn.registerUser")
	defer span.End()

	var req model.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorType := &model.InvalidRequestError{Message: "failed to decode request"}
		e := recordError(ctx, span, errorType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.registerUser")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if req.ID == uuid.Nil {
		var err error
		req.ID, err = uuid.NewV7()
		if err != nil {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.registerUser")
			respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
			return
		}
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.registerUser")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.RegisterUserInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	if err := ref.service.RegisterUser(ctx, input); err != nil {
		var errorTypeInvalidMail *model.InvalidEmailError
		var errorTypeInvalidPassword *model.InvalidPasswordError
		var errorTypeUserAlreadyExistsError *model.UserAlreadyExistsError
		var errorTypeUserEmailAlreadyExistsError *model.UserEmailAlreadyExistsError

		if errors.As(err, &errorTypeInvalidMail) || errors.As(err, &errorTypeInvalidPassword) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.registerUser")
			respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
			return
		}

		if errors.As(err, &errorTypeUserAlreadyExistsError) || errors.As(err, &errorTypeUserEmailAlreadyExistsError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Authn.registerUser")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.registerUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Authn.registerUser: user created", "user_id", req.ID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "user created")

	respond.WriteJSONMessage(w, r, http.StatusCreated, model.AuthnUserRegisteredSuccessfully)
}

// verifyUser Verify a user using the JWT token.
//
//	@Id				0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a
//	@Summary		Verify user
//	@Description	Verify user account using JWT verification token
//	@Tags			Auth
//	@Produce		json
//	@Param			jwt	path		string	true	"The JWT token to use"	Format(jwt)
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		401	{object}	model.HTTPMessage
//	@Failure		404	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/auth/verify/{jwt} [get]
func (ref *AuthnHandler) verifyUser(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Authn.verifyUser")
	defer span.End()

	jwt, err := parseJWTQueryParams(r.PathValue("jwt"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.verifyUser")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := ref.service.VerifyUser(ctx, jwt); err != nil {
		var errorTypeInvalidJWT *model.InvalidJWTError
		var errorInvalidJWT *model.InvalidJWTError

		if errors.As(err, &errorTypeInvalidJWT) ||
			errors.As(err, &errorInvalidJWT) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusUnauthorized, "handler.Authn.verifyUser")
			respond.WriteJSONMessage(w, r, http.StatusUnauthorized, e.Error())
			return
		}

		var errorTypeUserNotFound *model.UserNotFoundError
		if errors.As(err, &errorTypeUserNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Authn.verifyUser")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.verifyUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Authn.verifyUser: user verified")
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "user verified")

	respond.WriteJSONMessage(w, r, http.StatusOK, model.AuthnUserVerifiedSuccessfully)
}

// reVerifyUser Re-verify a user using the JWT token.
//
//	@Id				0198042a-f9c5-75d0-8c20-fea31b65587f
//	@Summary		Resend verification
//	@Description	Resend account verification email to user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.ReVerifyUserRequest	true	"The email of the user to re-verify"
//	@Success		200		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		401		{object}	model.HTTPMessage
//	@Failure		404		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/auth/verify [post]
func (ref *AuthnHandler) reVerifyUser(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Authn.reVerifyUser")
	defer span.End()

	var req model.ReVerifyUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorType := &model.InvalidRequestError{Message: "failed to decode request"}
		e := recordError(ctx, span, errorType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.reVerifyUser")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.reVerifyUser")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := ref.service.ReVerifyUser(ctx, req.Email); err != nil {
		var errorTypeInvalidJWT *model.InvalidJWTError
		var errorInvalidJWT *model.InvalidJWTError

		if errors.As(err, &errorTypeInvalidJWT) ||
			errors.As(err, &errorInvalidJWT) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusUnauthorized, "handler.Authn.reVerifyUser")
			respond.WriteJSONMessage(w, r, http.StatusUnauthorized, e.Error())
			return
		}

		// gracefully handle the case where the user is not found, securely
		// without exposing any information about the user
		var errorTypeUserNotFound *model.UserNotFoundError
		if errors.As(err, &errorTypeUserNotFound) {
			// gratefully handle the case where the user is not found
			recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "verification email sent")
			respond.WriteJSONMessage(w, r, http.StatusOK, model.AuthnUserVerificationEmailSent)
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.reVerifyUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Authn.reVerifyUser: verification email sent")
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "verification email sent")

	respond.WriteJSONMessage(w, r, http.StatusOK, model.AuthnUserVerificationEmailSent)
}

// logout Log out the current user
//
//	@Id				0198042a-f9c5-75d4-afa6-fe658744c80f
//	@Summary		Logout user
//	@Description	Logout user and invalidate session tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		401	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/auth/logout [delete]
//	@Security		AccessToken
func (ref *AuthnHandler) logout(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Authn.logoutUser")
	defer span.End()

	// get the user id from the context
	jwtClaims, ok := r.Context().Value(middleware.JwtClaims).(map[string]any)
	if !ok {
		errorMsg := "failed to get user id from context"
		recordContextError(ctx, span, errorMsg, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.logoutUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, errorMsg)
		return
	}

	userID, ok := jwtClaims["sub"].(string)
	if !ok {
		errorMsg := "failed to get user id from context"
		recordContextError(ctx, span, errorMsg, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.logoutUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, errorMsg)
		return
	}

	if err := ref.service.LoggingOutUser(ctx, userID); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.logoutUser")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Authn.logoutUser: user logged out")
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "user logged out")

	respond.WriteJSONMessage(w, r, http.StatusOK, model.AuthnUserLoggedOutSuccessfully)
}

// refreshAccessToken Retrieve a new access token using the refresh token.
//
//	@Id				0198042a-f9c5-75d8-aa7b-37524ea4f124
//	@Summary		Refresh access token
//	@Description	Generate new access token using valid refresh token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.RefreshTokenRequest	true	"The refresh token to use"
//	@Success		200		{object}	model.RefreshTokenResponse
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		401		{object}	model.HTTPMessage
//	@Failure		404		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/auth/refresh [post]
//	@Security		RefreshToken
func (ref *AuthnHandler) refreshAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Authn.refreshAccessToken")
	defer span.End()

	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorType := &model.InvalidRequestError{Message: "failed to decode request"}
		e := recordError(ctx, span, errorType, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.refreshAccessToken")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Authn.refreshAccessToken")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.RefreshAccessTokenInput{
		RefreshToken: req.RefreshToken,
	}

	out, err := ref.service.RefreshAccessToken(ctx, input)
	if err != nil {
		var errorTypeInvalidJWT *model.InvalidJWTError

		if errors.As(err, &errorTypeInvalidJWT) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusUnauthorized, "handler.Authn.refreshAccessToken")
			respond.WriteJSONMessage(w, r, http.StatusUnauthorized, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.refreshAccessToken")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	resp := &model.RefreshTokenResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		TokenType:    out.TokenType,
	}

	if err := respond.WriteJSONData(w, http.StatusOK, resp); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Authn.refreshAccessToken")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Authn.refreshAccessToken: access token refreshed")
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "access token refreshed")
}
