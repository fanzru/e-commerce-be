package formatter

import (
	"encoding/json"
	"net/http"
	"time"
)

// StandardResponse is the standard API response format
type StandardResponse struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	ServerTime time.Time   `json:"server_time"`
	Data       interface{} `json:"data,omitempty"`
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return e.Message
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) StandardResponse {
	return StandardResponse{
		Code:       "SUCCESS",
		Message:    message,
		ServerTime: time.Now(),
		Data:       data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, data interface{}) StandardResponse {
	return StandardResponse{
		Code:       "ERROR",
		Message:    message,
		ServerTime: time.Now(),
		Data:       data,
	}
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// SuccessJSON writes a success JSON response
func SuccessJSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := NewSuccessResponse(message, data)
	JSON(w, statusCode, response)
}

// ErrorJSON writes an error JSON response
func ErrorJSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := NewErrorResponse(message, data)
	JSON(w, statusCode, response)
}
