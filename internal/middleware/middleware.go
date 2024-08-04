package middleware

import (
	"log/slog"
	"net/http"
)

var APIVersion = "v1"

type Middleware func(http.Handler) http.Handler

type wrappedResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrappedResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
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

// AddAPIVersion adds the API version to the response headers
// Configurable via the APIVersion variable
// Defaults to "v1"
// Set the header X-API-Version
func AddAPIVersion(next http.Handler) http.Handler {
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
