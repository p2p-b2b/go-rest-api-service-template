// Package respond provides functions to write JSON responses to HTTP requests.
// It includes a sync.Pool to reuse HTTP message objects, reducing memory allocations.
// It also provides a function to write JSON data directly to the response writer.
// The package is designed to be used in HTTP handlers to respond with structured JSON messages.
// It includes logging capabilities to log the response details for debugging and monitoring purposes.
package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
)

// WriteJSONData writes the given data to the client as a JSON response.
func WriteJSONData(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

// httpMessagePool is a sync.Pool for HTTP messages to reduce memory allocations.
// in the case of high load, this can help reduce the number of allocations and improve performance.
var httpMessagePool = sync.Pool{
	New: func() any {
		return new(model.HTTPMessage)
	},
}

// WriteJSONMessage writes a success log and response to the client with the given status code and message.
func WriteJSONMessage(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")

	mgs := httpMessagePool.Get().(*model.HTTPMessage)
	mgs.Timestamp = time.Now()
	mgs.StatusCode = statusCode
	mgs.Message = message
	mgs.Method = r.Method
	mgs.Path = r.URL.Path

	// Set the status code before writing the response body
	w.WriteHeader(statusCode)

	// Now try to write the data
	if err := json.NewEncoder(w).Encode(mgs); err != nil {
		slog.Error("failed to write JSON response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	httpMessagePool.Put(mgs)

	slog.Debug(message,
		"status_code", statusCode,
		"method", r.Method,
		"url", r.URL.Path,
		"query", r.URL.RawQuery,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)
}
