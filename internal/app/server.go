package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/p2p-b2b/go-rest-api-service-template/docs"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/server"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/jwtvalidator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
	"github.com/p2p-b2b/ratelimiter"
	"golang.org/x/time/rate"
)

// initHTTPServer initializes the HTTP server with all registered routes
func (a *App) initHTTPServer(ctx context.Context) error {
	// Configure server URL information
	serverProtocol := "http"
	if a.configs.HTTPServer.TLSEnabled.Value {
		serverProtocol = "https"
	}

	serverURL := fmt.Sprintf("%s://%s:%d/%s",
		serverProtocol,
		a.configs.HTTPServer.Address.Value,
		a.configs.HTTPServer.Port.Value,
		apiPrefix,
	)

	statusURL := fmt.Sprintf("%s/status", serverURL)
	serverHost := fmt.Sprintf("%s:%d", a.configs.HTTPServer.Address.Value, a.configs.HTTPServer.Port.Value)
	swaggerURLIndex := fmt.Sprintf("%s/swagger/index.html", serverURL)

	slog.Info("server endpoints",
		"api", serverURL,
		"status", statusURL,
		"swagger", swaggerURLIndex,
	)

	// Configure Swagger metadata
	if err := configureSwaggerMetadata(serverHost, apiPrefix, serverProtocol); err != nil {
		return err
	}

	// Create a new router for API endpoints
	apiRouter := http.NewServeMux()

	// Setup common middlewares
	apiCommonMdws := []middleware.Middleware{
		middleware.RewriteStandardErrorsAsJSON,
		middleware.Logging,
		middleware.HeaderAPIVersion(apiVersion),
		middleware.OtelTextMapPropagation,
	}

	// Add CORS middleware if enabled
	if a.configs.HTTPServer.CorsEnabled.Value {
		corsOpts := a.getCorsOptions()
		apiCommonMdws = append(apiCommonMdws, middleware.Cors(corsOpts))
	}

	// Add rate limiter middleware if enabled
	if a.configs.HTTPServer.IPRateLimiterEnabled.Value {
		rateLimiter := a.createRateLimiter()
		apiCommonMdws = append(apiCommonMdws, middleware.IPRateLimiter(rateLimiter))
	}

	// Create JWT validators
	jwtValidators := a.createJWTValidators()

	// Create middleware chains
	apiCommonMiddlewares := middleware.Chain(apiCommonMdws...)
	accessTokenMiddlewares := middleware.Chain(
		middleware.CheckAccessToken(jwtValidators),
		middleware.CheckAuthz(a.services.Authz),
	)
	refreshTokenMiddlewares := middleware.Chain(
		middleware.CheckRefreshToken(jwtValidators),
		middleware.CheckAuthz(a.services.Authz),
	)

	// Register public routes
	a.handlers.Swagger.RegisterRoutes(apiRouter)
	a.handlers.Health.RegisterRoutes(apiRouter)
	a.handlers.Version.RegisterRoutes(apiRouter)

	// Register protected routes
	a.handlers.Users.RegisterRoutes(apiRouter, accessTokenMiddlewares)
	a.handlers.Policies.RegisterRoutes(apiRouter, accessTokenMiddlewares)
	a.handlers.Resources.RegisterRoutes(apiRouter, accessTokenMiddlewares)
	a.handlers.Roles.RegisterRoutes(apiRouter, accessTokenMiddlewares)
	a.handlers.Authn.RegisterRoutes(apiRouter, accessTokenMiddlewares, refreshTokenMiddlewares)
	a.handlers.Projects.RegisterRoutes(apiRouter, accessTokenMiddlewares)
	a.handlers.Products.RegisterRoutes(apiRouter, accessTokenMiddlewares)

	// Create the main router
	mainRouter := http.NewServeMux()
	mainRouter.Handle(fmt.Sprintf("/%s/", apiPrefix),
		http.StripPrefix(fmt.Sprintf("/%s", apiPrefix), apiCommonMiddlewares(apiRouter)),
	)

	// Create HTTP server
	a.httpServer = server.NewHTTPServer(server.HTTPServerConfig{
		Ctx:         ctx,
		HTTPHandler: mainRouter,
		Config:      a.configs.HTTPServer,
	})

	return nil
}

// configureSwaggerMetadata sets up the Swagger documentation metadata
func configureSwaggerMetadata(serverHost, apiPrefix, serverProtocol string) error {
	docs.SwaggerInfo.Host = serverHost
	docs.SwaggerInfo.BasePath = fmt.Sprintf("/%s", apiPrefix)
	docs.SwaggerInfo.Schemes = []string{serverProtocol}
	docs.SwaggerInfo.Version = version.Version

	return nil
}

// getCorsOptions creates CORS configuration options
func (a *App) getCorsOptions() middleware.CorsOpts {
	slog.Warn("CORS enabled",
		"allowed_origins", a.configs.HTTPServer.CorsAllowedOrigins.Value,
		"allowed_methods", a.configs.HTTPServer.CorsAllowedMethods.Value,
		"allowed_headers", a.configs.HTTPServer.CorsAllowedHeaders.Value,
		"allow_credentials", a.configs.HTTPServer.CorsAllowCredentials.Value,
	)

	return middleware.CorsOpts{
		AllowedOrigins:   strings.Split(strings.Trim(a.configs.HTTPServer.CorsAllowedOrigins.Value, " "), ","),
		AllowedMethods:   strings.Split(strings.Trim(a.configs.HTTPServer.CorsAllowedMethods.Value, " "), ","),
		AllowedHeaders:   strings.Split(strings.Trim(a.configs.HTTPServer.CorsAllowedHeaders.Value, " "), ","),
		AllowCredentials: a.configs.HTTPServer.CorsAllowCredentials.Value,
	}
}

// createRateLimiter creates a rate limiter for HTTP requests
func (a *App) createRateLimiter() *ratelimiter.BucketLimiter {
	slog.Warn("ip rate limiter enabled",
		"period", "1s",
		"limit", a.configs.HTTPServer.IPRateLimiterLimit.Value,
		"burst", a.configs.HTTPServer.IPRateLimiterBurst.Value,
		"delete_after", a.configs.HTTPServer.IPRateLimiterDeleteAfter.Value,
	)

	// Create a storage system
	storage := ratelimiter.NewInMemoryStorage()

	// Create a base rate limiter
	baseLimiter := rate.NewLimiter(
		rate.Limit(float64(a.configs.HTTPServer.IPRateLimiterLimit.Value)),
		a.configs.HTTPServer.IPRateLimiterBurst.Value,
	)

	// Create a bucket limiter
	return ratelimiter.NewBucketLimiter(
		baseLimiter,
		a.configs.HTTPServer.IPRateLimiterDeleteAfter.Value,
		storage,
	)
}

// createJWTValidators creates JWT validators for access and refresh tokens
func (a *App) createJWTValidators() map[string]jwtvalidator.Validator {
	jwtValidators := make(map[string]jwtvalidator.Validator)

	// Read the public key
	jwtPublicKey, err := os.ReadFile(a.configs.Authn.PublicKeyFile.Value.Name())
	if err != nil {
		slog.Error("error reading JWT public key file", "file", a.configs.Authn.PublicKeyFile.Value.Name(), "error", err)
		return jwtValidators
	}

	jwtValidators["accessToken"] = &jwtvalidator.AccessTokenValidator{
		PublicKey: jwtPublicKey,
		ClientID:  a.configs.Authn.Issuer.Value,
	}

	jwtValidators["refreshToken"] = &jwtvalidator.RefreshTokenValidator{
		PublicKey: jwtPublicKey,
		ClientID:  a.configs.Authn.Issuer.Value,
	}

	return jwtValidators
}
