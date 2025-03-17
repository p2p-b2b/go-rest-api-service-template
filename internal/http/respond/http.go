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
	w.WriteHeader(statusCode)

	// var success model.HTTPMessage
	mgs := httpMessagePool.Get().(*model.HTTPMessage)

	mgs.Timestamp = time.Now()
	mgs.StatusCode = statusCode
	mgs.Message = message
	mgs.Method = r.Method
	mgs.Path = r.URL.Path

	if err := json.NewEncoder(w).Encode(mgs); err != nil {
		slog.Error("failed to write JSON response", "error", err)

		http.Error(w, "failed to write JSON response", http.StatusInternalServerError)
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
