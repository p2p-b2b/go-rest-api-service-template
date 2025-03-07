package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
)

// WriteJSONData writes the given data to the client as a JSON response.
func WriteJSONData(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

// WriteJSONMessage writes a success log and response to the client with the given status code and message.
func WriteJSONMessage(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var success model.HTTPMessage
	success.Timestamp = time.Now()
	success.StatusCode = statusCode
	success.Message = message
	success.Method = r.Method
	success.Path = r.URL.Path

	if err := json.NewEncoder(w).Encode(success); err != nil {
		slog.Error("failed to write JSON response", "error", err)

		http.Error(w, "failed to write JSON response", http.StatusInternalServerError)
	}

	slog.Debug(message,
		"status_code", statusCode,
		"method", r.Method,
		"url", r.URL.Path,
		"query", r.URL.RawQuery,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)
}
