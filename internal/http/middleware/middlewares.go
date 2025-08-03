// Package middleware provides a set of HTTP middleware functions
// that can be used to enhance the functionality of HTTP handlers.
// It includes middlewares for logging, CORS, JWT validation, rate limiting, and more.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/jwtvalidator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"github.com/p2p-b2b/ratelimiter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// ContextKey is a type for context keys
type ContextKey string

func (k ContextKey) String() string {
	return string(k)
}

const (
	JwtClaims ContextKey = "jwt_claims"
)

// Middleware is a function that wraps an http.Handler
// to provide additional functionality
type Middleware func(http.Handler) http.Handler

// ThenFunc wraps an http.HandlerFunc with a middleware
// This is a convenience method to allow chaining middlewares
func (m Middleware) ThenFunc(h http.HandlerFunc) http.Handler {
	return m(http.HandlerFunc(h))
}

// Then wraps an http.Handler with a middleware
// This is a convenience method to allow chaining middlewares
func (m Middleware) Then(h http.Handler) http.Handler {
	return m(h)
}

// Apply applies the middleware to an http.Handler
func (m Middleware) Apply(h http.Handler) http.Handler {
	return m(h)
}

// Chain applies middlewares to an http.Handler
// in the order they are provided
func Chain(mws ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := range mws {
			h = mws[len(mws)-1-i](h)
		}
		return h
	}
}

// Append appends a middleware to the chain
func Append(m Middleware, mws ...Middleware) []Middleware {
	return append(mws, m)
}

// HeaderAPIVersion adds the API version to the response headers
func HeaderAPIVersion(version string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if version == "" {
				version = "v1"
			}

			w.Header().Set("X-API-Version", version)
			next.ServeHTTP(w, r)
		})
	}
}

// Logging middleware logs the request and response
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &wrappedResponseWriter{
			w,
			http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		slog.Info("request", "method", r.Method, "path", r.URL.Path, "address", r.RemoteAddr, "status", wrapped.status)
	})
}

// OtelTextMapPropagation middleware propagates the OpenTelemetry context
// from incoming requests to outgoing requests
func OtelTextMapPropagation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(
			r.Context(), propagation.HeaderCarrier(r.Header),
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// customResponseWriter is a custom response writer that handles custom error responses.
type customResponseWriter struct {
	*wrappedResponseWriter
	method string
	path   string
}

// Write writes the response data.
func (w *customResponseWriter) Write(data []byte) (n int, err error) {
	var apiResponse model.HTTPMessage

	switch w.status {
	case http.StatusNotFound:
		if err := json.Unmarshal(data, &apiResponse); err != nil {
			data, err = json.Marshal(
				model.HTTPMessage{
					Timestamp:  time.Now().UTC(),
					StatusCode: http.StatusNotFound,
					Message:    "Not Found",
					Method:     w.method,
					Path:       w.path,
				},
			)
			if err != nil {
				return 0, err
			}
		}

	case http.StatusMethodNotAllowed:
		if err := json.Unmarshal(data, &apiResponse); err != nil {
			data, err = json.Marshal(
				model.HTTPMessage{
					Timestamp:  time.Now().UTC(),
					StatusCode: http.StatusMethodNotAllowed,
					Message:    "Method Not Allowed",
					Method:     w.method,
					Path:       w.path,
				},
			)
			if err != nil {
				return 0, err
			}
		}
	}

	return w.wrappedResponseWriter.Write(data)
}

// RewriteStandardErrorsAsJSON is a middleware that rewrites standard HTTP errors as JSON responses.
func RewriteStandardErrorsAsJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newW := &customResponseWriter{
			wrappedResponseWriter: newWrappedResponseWriter(w),
			method:                r.Method,
			path:                  r.URL.Path,
		}

		h.ServeHTTP(newW, r)
	})
}

// CorsOpts is the configuration for the CORS middleware
// Options are:
// AllowedOrigins is a list of origins a cross-domain request can be executed from
// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests
// AllowedHeaders is a list of non-simple headers the client is allowed to use with cross-domain requests
// AllowCredentials indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates
type CorsOpts struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// Cors middleware adds CORS headers to the response
func Cors(opts CorsOpts) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if slices.Contains(opts.AllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if len(opts.AllowedOrigins) == 0 {
				opts.AllowedOrigins = []string{"*"}
			}

			if len(opts.AllowedMethods) == 0 {
				opts.AllowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}
			}

			if len(opts.AllowedHeaders) == 0 {
				opts.AllowedHeaders = []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"}
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(opts.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(opts.AllowedHeaders, ", "))

			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Credentials", "false")
			}

			// TODO: remove this block of code and implement a better way
			// to handle OPTIONS requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckAccessToken checks the JWTs created and signed by the application
// and validates the token_type claim
// The token_type claim is used to identify the type of token
// and this validate the "access" or "personal_access" token
func CheckAccessToken(validator map[string]jwtvalidator.Validator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Missing header: Authorization")
				return
			}

			// avoid panic if authHeader is too short
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Authorization header must start with Bearer ")
				return
			}

			// Extract Bearer from authHeader
			token := authHeader[len("Bearer "):]
			if token == "" {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Token is empty")
				return
			}

			if validator == nil {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// check validator has the idp
			validator, ok := validator["accessToken"]
			if !ok {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Unauthorized")
				return
			}

			claims, err := validator.Validate(r.Context(), token)
			if err != nil {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, err.Error())
				return
			}

			if len(claims) == 0 {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Claims is empty")
				return
			}

			// validate the token_type claim
			var tokenType any
			if tokenType, ok = claims["token_type"]; !ok {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Token type field not found in claims")
				return
			}

			// validate "access" or "personal_access" token
			if tokenType != model.TokenTypeAccess.String() && tokenType != model.TokenTypePersonalAccess.String() {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "invalid token type access or personal_access")
				return
			}

			// Add the claims to the request context
			r = r.WithContext(context.WithValue(r.Context(), JwtClaims, claims))

			// Check if the provider ClientID is the same as the one in the token audience (aud) string
			// if !strings.Contains(claims["aud"].([]string), validator.GetClientID()) {
			// 	WriteJSONMessage(w, r, http.StatusUnauthorized, "Token audience does not match provider ClientID")
			// 	return
			// }

			next.ServeHTTP(w, r)
		})
	}
}

// CheckRefreshToken checks the JWTs created and signed by the application
func CheckRefreshToken(validator map[string]jwtvalidator.Validator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Missing header: Authorization")
				return
			}

			// avoid panic if authHeader is too short
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Authorization header must start with Bearer ")
				return
			}

			// Extract Bearer from authHeader
			token := authHeader[len("Bearer "):]
			if token == "" {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Token is empty")
				return
			}

			if validator == nil {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// check validator has the idp
			validator, ok := validator["refreshToken"]
			if !ok {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Unauthorized")
				return
			}

			claims, err := validator.Validate(r.Context(), token)
			if err != nil {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, err.Error())
				return
			}

			if len(claims) == 0 {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Claims is empty")
				return
			}

			// validate the token_type claim
			var tokenType any
			if tokenType, ok = claims["token_type"]; !ok {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Token type field not found in claims")
				return
			}

			if tokenType != model.TokenTypeRefresh.String() {
				respond.WriteJSONMessage(w, r, http.StatusUnauthorized, "Token type is not refresh")
				return
			}

			// Add the claims to the request context
			r = r.WithContext(context.WithValue(r.Context(), JwtClaims, claims))

			// Check if the provider ClientID is the same as the one in the token audience (aud) string
			// if !strings.Contains(claims["aud"].([]string), validator.GetClientID()) {
			// 	WriteJSONMessage(w, r, http.StatusUnauthorized, "Token audience does not match provider ClientID")
			// 	return
			// }

			next.ServeHTTP(w, r)
		})
	}
}

// CheckAuthz middleware checks if the user_id (sub) in the JWT is authorized to access the resource
// through the OPA policy engine
func CheckAuthz(service *service.AuthzService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get the sub claim from the context
			claims, ok := r.Context().Value(JwtClaims).(map[string]any)
			if !ok {
				respond.WriteJSONMessage(w, r, http.StatusForbidden, "Claims not found in context")
				return
			}

			subStr, ok := claims["sub"].(string)
			if !ok {
				respond.WriteJSONMessage(w, r, http.StatusForbidden, "sub claim not found in claims")
				return
			}

			// sub to uuid
			sub, err := uuid.Parse(subStr)
			if err != nil {
				respond.WriteJSONMessage(w, r, http.StatusForbidden, "Invalid sub claim")
				return
			}

			ok, err = service.IsAuthorized(r.Context(), sub, r.Method, r.URL.Path)
			if err != nil {
				respond.WriteJSONMessage(w, r, http.StatusForbidden, fmt.Sprintf("Unauthorized access to %s %s", r.Method, r.URL.Path))
				return
			}

			if !ok {
				respond.WriteJSONMessage(w, r, http.StatusForbidden, fmt.Sprintf("Unauthorized access to %s %s", r.Method, r.URL.Path))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPRateLimiter is a middleware that limits the number of requests
// from a single IP address
// The rate limiter is a token bucket algorithm
// https://en.wikipedia.org/wiki/Token_bucket
func IPRateLimiter(limiter *ratelimiter.BucketLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]

			lim := limiter.GetOrAdd(ip)
			if !lim.Allow() {
				respond.WriteJSONMessage(w, r, http.StatusTooManyRequests, fmt.Sprintf("too many requests from ip address %s", ip))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
