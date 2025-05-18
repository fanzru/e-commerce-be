package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// Logger is the global logger instance
	Logger *slog.Logger
)

func init() {
	// Configure the logger with JSON handler for structured logging
	// Get log level from environment variable
	logLevelStr := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch strings.ToLower(logLevelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo // Default to INFO
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// Add custom attributes to all log entries
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(Logger)

	// Log the configured level
	Logger.Info("Logger initialized", slog.String("level", level.String()))
}

// responseWriter is a custom responseWriter that captures the response status code and body
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	buffer     *bytes.Buffer
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body
func (rw *responseWriter) Write(buf []byte) (int, error) {
	rw.buffer.Write(buf)
	return rw.ResponseWriter.Write(buf)
}

// APILogEntry represents a log entry for API request/response
type APILogEntry struct {
	RequestID   string                 `json:"request_id"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	QueryParams map[string][]string    `json:"query_params,omitempty"`
	Headers     map[string][]string    `json:"headers,omitempty"`
	RemoteAddr  string                 `json:"remote_addr"`
	RequestBody map[string]interface{} `json:"request_body,omitempty"`
	Status      int                    `json:"status"`
	Duration    time.Duration          `json:"duration"`
	Response    interface{}            `json:"response,omitempty"`
}

// Recoverer middleware recovers from panics
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())

				// Log the panic with structured information
				Logger.Error("panic recovered",
					slog.String("request_id", requestID),
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// APILogger middleware logs detailed information about API requests and responses
func APILogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create a copy of the request body
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Create a custom responseWriter to capture the response
		customWriter := &responseWriter{
			ResponseWriter: w,
			buffer:         &bytes.Buffer{},
			statusCode:     http.StatusOK, // Default status code
		}

		// Get or create request ID
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set(RequestIDHeader, requestID)
		}

		// Process the request
		next.ServeHTTP(customWriter, r)

		// Calculate duration
		duration := time.Since(startTime)

		// Parse request body if it's JSON
		var reqBodyJSON map[string]interface{}
		if len(reqBody) > 0 && isJSONContent(r.Header.Get("Content-Type")) {
			if err := json.Unmarshal(reqBody, &reqBodyJSON); err == nil {
				reqBodyJSON = maskSensitiveFields(reqBodyJSON)
			}
		}

		// Parse response body if it's JSON
		var respBodyJSON interface{}
		if isJSONContent(customWriter.Header().Get("Content-Type")) {
			_ = json.Unmarshal(customWriter.buffer.Bytes(), &respBodyJSON)
		}

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if customWriter.statusCode >= 400 && customWriter.statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if customWriter.statusCode >= 500 {
			logLevel = slog.LevelError
		}

		// Create attribute list for the log entry
		attrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status", customWriter.statusCode),
			slog.Duration("duration", duration),
		}

		// Add query parameters if present
		if len(r.URL.Query()) > 0 {
			attrs = append(attrs, slog.Any("query_params", r.URL.Query()))
		}

		// Add filtered headers
		attrs = append(attrs, slog.Any("headers", filterHeaders(r.Header)))

		// Add request body if present
		if reqBodyJSON != nil {
			attrs = append(attrs, slog.Any("request_body", reqBodyJSON))
		}

		// Add response body if present and not too large
		if respBodyJSON != nil {
			// Consider limiting the size of the logged response
			attrs = append(attrs, slog.Any("response", respBodyJSON))
		}

		// Log the request with appropriate level
		if customWriter.statusCode >= 400 {
			logMessage := "request error"
			if customWriter.statusCode >= 500 {
				logMessage = "request failed"
			}
			Logger.LogAttrs(context.Background(), logLevel, logMessage, attrs...)
		} else {
			Logger.LogAttrs(context.Background(), logLevel, "request completed", attrs...)
		}
	})
}

// filterHeaders filters out sensitive headers
func filterHeaders(headers http.Header) map[string][]string {
	filtered := make(map[string][]string)
	for key, values := range headers {
		// Skip sensitive headers
		if isSensitiveHeader(key) {
			continue
		}
		filtered[key] = values
	}
	return filtered
}

// isSensitiveHeader checks if a header is sensitive
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-api-secret",
	}

	headerLower := strings.ToLower(header)
	for _, sensitive := range sensitiveHeaders {
		if headerLower == sensitive {
			return true
		}
	}

	return false
}

// isJSONContent checks if the content type is JSON
func isJSONContent(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// maskSensitiveFields masks sensitive fields in the request body
func maskSensitiveFields(data map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{
		"password",
		"current_password",
		"new_password",
		"token",
		"access_token",
		"refresh_token",
		"secret",
		"card_number",
		"cvv",
	}

	result := make(map[string]interface{})
	for key, value := range data {
		if contains(sensitiveFields, strings.ToLower(key)) {
			result[key] = "********" // Mask sensitive data
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively mask nested maps
			result[key] = maskSensitiveFields(nestedMap)
		} else {
			result[key] = value
		}
	}

	return result
}

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
