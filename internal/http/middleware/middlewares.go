package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
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
func (m Middleware) ThenFunc(h http.HandlerFunc) http.Handler {
	return m(http.HandlerFunc(h))
}

// Then wraps an http.Handler with a middleware
// This is a convenience method to allow chaining middlewares
func (m Middleware) Then(h http.Handler) http.Handler {
	return m(h)
}

// Apply applies the middleware to an http.Handler
func (mws Middleware) Apply(h http.Handler) http.Handler {
	return mws(h)
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
