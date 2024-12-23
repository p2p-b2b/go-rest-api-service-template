package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var APIVersion = "v1"

type Middleware func(http.Handler) http.Handler

// Thanks to:
// - https://github.com/denpeshkov/greenlight/blob/c68f5a2111adcd5b1a65a06595acc93a02b6380e/internal/http/middleware.go#L16-L71
// - https://github.com/golang/go/issues/65648
type wrappedResponseWriter struct {
	http.ResponseWriter
	status int
}

// newWrappedResponseWriter creates a new statusResponseWriter.
func newWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	// WriteHeader() is not called if our response implicitly returns 200 OK, so we default to that status code.
	return &wrappedResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (w *wrappedResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

// Unwrap is used by a [http.ResponseController].
func (w *wrappedResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Chain applies middlewares to an http.Handler
// in the order they are provided
func Chain(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := range mws {
			next = mws[len(mws)-1-i](next)
		}

		return next
	}
}

// HeaderAPIVersion adds the API version to the response headers
// Configurable via the APIVersion variable
// Defaults to "v1"
// Set the header X-API-Version
func HeaderAPIVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-API-Version", APIVersion)
		next.ServeHTTP(w, r)
	})
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
	var apiResponse APIResponse

	switch w.wrappedResponseWriter.status {
	case http.StatusNotFound:
		if err := json.Unmarshal(data, &apiResponse); err != nil {
			data, err = json.Marshal(
				APIResponse{
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
				APIResponse{
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
