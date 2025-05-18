package middleware

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/fanzru/e-commerce-be/internal/common/errs"
)

// ErrorHandlerMiddleware is a middleware that handles errors
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer that can capture the response
		crw := &captureResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(crw, r)

		// If the status code is an error, log it
		if crw.statusCode >= 400 {
			Logger.Error("HTTP error",
				"status", crw.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
				"client_ip", r.RemoteAddr,
			)
		}
	})
}

// captureResponseWriter is a custom response writer that captures the response
type captureResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (crw *captureResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// RespondWithError responds with a JSON error message
func RespondWithError(w http.ResponseWriter, err error) {
	statusCode, errResp := errs.HandleError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResp)

	// Also log the error
	Logger.Error("Application error",
		"code", errResp.Code,
		"message", errResp.Message,
		"source", errResp.Source,
	)
}

// RespondWithJSON responds with a JSON message
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	// Check if payload already has a standard response format
	// by checking if it has the Code, Message, and ServerTime fields
	if hasStandardFormat(payload) {
		// If payload already has the standard format, use it directly
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			Logger.Error("Failed to encode response", "error", err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Otherwise, wrap it in the standard format
	response := struct {
		Code       string      `json:"code"`
		Message    string      `json:"message"`
		Data       interface{} `json:"data,omitempty"`
		ServerTime time.Time   `json:"server_time"`
	}{
		Code:       "success",
		Message:    getMessageForStatusCode(statusCode),
		Data:       payload,
		ServerTime: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Logger.Error("Failed to encode response", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// hasStandardFormat checks if a value has the standard response format
func hasStandardFormat(v interface{}) bool {
	// Use reflection to check if v has Code, Message, and ServerTime fields
	vValue := reflect.ValueOf(v)

	// If v is a pointer, get its underlying value
	if vValue.Kind() == reflect.Ptr && !vValue.IsNil() {
		vValue = vValue.Elem()
	}

	// Only struct types can have our standard format
	if vValue.Kind() != reflect.Struct {
		return false
	}

	// Check for the presence of required fields
	codeField := vValue.FieldByName("Code")
	messageField := vValue.FieldByName("Message")
	serverTimeField := vValue.FieldByName("ServerTime")

	return codeField.IsValid() && messageField.IsValid() && serverTimeField.IsValid()
}

// getMessageForStatusCode returns a message for a status code
func getMessageForStatusCode(statusCode int) string {
	switch statusCode {
	case http.StatusOK:
		return "Request successful"
	case http.StatusCreated:
		return "Resource created successfully"
	case http.StatusAccepted:
		return "Request accepted"
	case http.StatusNoContent:
		return "Request successful, no content to return"
	default:
		return "Request successful"
	}
}
