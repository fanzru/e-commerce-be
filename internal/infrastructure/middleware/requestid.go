package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header key for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDCtxKey is the context key for request ID
	RequestIDCtxKey = "request_id"
)

// RequestID middleware ensures each request has a unique ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			// Generate a new request ID
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		w.Header().Set(RequestIDHeader, requestID)

		// Store request ID in context
		ctx := context.WithValue(r.Context(), RequestIDCtxKey, requestID)

		// Call the next handler with the enhanced context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDCtxKey).(string); ok {
		return reqID
	}
	return ""
}
