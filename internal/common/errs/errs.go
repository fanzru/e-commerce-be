package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// Error codes
const (
	// Common error codes
	CodeInternalServerError = "internal_server_error"
	CodeNotFound            = "not_found"
	CodeBadRequest          = "bad_request"
	CodeUnauthorized        = "unauthorized"
	CodeForbidden           = "forbidden"
	CodeConflict            = "conflict"
	CodeValidationError     = "validation_error"
	CodeOutOfStock          = "out_of_stock"
)

// ErrorResponse is the standard API error response format
type ErrorResponse struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	ServerTime time.Time   `json:"server_time"`
	Source     string      `json:"source,omitempty"` // Contains file:line information
}

// AppError represents an application error with HTTP status code and caller information
type AppError struct {
	Err        error
	Code       string
	Status     int
	Message    string
	Data       map[string]interface{}
	StackTrace string
	File       string
	Line       int
}

// Error returns the error message with file and line information
func (e *AppError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s [%s:%d]", e.Message, e.File, e.Line)
	}
	return fmt.Sprintf("%s [%s:%d]", e.Err.Error(), e.File, e.Line)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Source returns a string with file and line information
func (e *AppError) Source() string {
	return fmt.Sprintf("%s:%d", e.File, e.Line)
}

// GetErrorResponse formats the error as an ErrorResponse
func (e *AppError) GetErrorResponse() ErrorResponse {
	code := e.Code
	if code == "" {
		code = CodeInternalServerError
	}

	return ErrorResponse{
		Code:       code,
		Message:    e.Message,
		Data:       e.Data,
		ServerTime: time.Now(),
		Source:     e.Source(),
	}
}

// ValidationError represents validation errors with field details
type ValidationError struct {
	mainError      error
	code           string
	details        []error
	detailMessages []string
	file           string
	line           int
}

// NewValidationError creates a new validation error
func NewValidationError(err error, code string, message string) *ValidationError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	validationErr := &ValidationError{
		mainError: err,
		code:      code,
		details:   []error{},
		file:      file,
		line:      line,
	}

	if message != "" {
		validationErr.detailMessages = []string{message}
	}
	return validationErr
}

// AddDetail adds a detail error to the validation error
func (e *ValidationError) AddDetail(err error) {
	e.details = append(e.details, err)
	e.detailMessages = append(e.detailMessages, err.Error())
}

// GetCode returns the error code
func (e *ValidationError) GetCode() string {
	if e.code == "" {
		return CodeValidationError
	}
	return e.code
}

// HasDetail checks if the validation error has a specific detail error
func (e *ValidationError) HasDetail(err error) bool {
	for _, detailErr := range e.details {
		if errors.Is(err, detailErr) {
			return true
		}
	}
	return false
}

// DetailLength returns the number of detail errors
func (e *ValidationError) DetailLength() int {
	return len(e.details)
}

// Error returns the error message with details and caller information
func (e *ValidationError) Error() string {
	if len(e.detailMessages) == 0 {
		return fmt.Sprintf("%s [%s:%d]", e.mainError.Error(), e.file, e.line)
	}
	return fmt.Sprintf("%s: %s [%s:%d]",
		e.mainError.Error(),
		strings.Join(e.detailMessages, ","),
		e.file,
		e.line,
	)
}

// Unwrap returns the original error
func (e *ValidationError) Unwrap() error {
	return e.mainError
}

// Source returns a string with file and line information
func (e *ValidationError) Source() string {
	return fmt.Sprintf("%s:%d", e.file, e.line)
}

// GetValidationError extracts ValidationError from wrapped errors
func GetValidationError(err error) *ValidationError {
	if e, ok := err.(*ValidationError); ok {
		return e
	}

	err = errors.Unwrap(err)
	if nil == err {
		return nil
	}

	return GetValidationError(err)
}

// ToAppError converts a ValidationError to an AppError
func (e *ValidationError) ToAppError() *AppError {
	return &AppError{
		Err:     e.mainError,
		Code:    e.GetCode(),
		Status:  http.StatusBadRequest, // Validation errors are usually bad requests
		Message: e.Error(),
		Data: map[string]interface{}{
			"details": e.detailMessages,
		},
		File: e.file,
		Line: e.line,
	}
}

// New creates a new AppError with caller information
func New(err error, code string, status int, message string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     err,
		Code:    code,
		Status:  status,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// NewWithData creates a new AppError with additional data and caller information
func NewWithData(err error, code string, status int, message string, data map[string]interface{}) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     err,
		Code:    code,
		Status:  status,
		Message: message,
		Data:    data,
		File:    file,
		Line:    line,
	}
}

// NewNotFound creates a not found error with caller information
func NewNotFound(message string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     errors.New("resource not found"),
		Code:    CodeNotFound,
		Status:  http.StatusNotFound,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// NewBadRequest creates a bad request error with caller information
func NewBadRequest(message string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     errors.New("bad request"),
		Code:    CodeBadRequest,
		Status:  http.StatusBadRequest,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// NewInternalError creates an internal server error with caller information
func NewInternalError(err error) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	message := "An unexpected error occurred"
	if err != nil {
		message = err.Error()
	}

	return &AppError{
		Err:     err,
		Code:    CodeInternalServerError,
		Status:  http.StatusInternalServerError,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// NewUnauthorized creates an unauthorized error with caller information
func NewUnauthorized(message string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     errors.New("unauthorized"),
		Code:    CodeUnauthorized,
		Status:  http.StatusUnauthorized,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// NewForbidden creates a forbidden error with caller information
func NewForbidden(message string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     errors.New("forbidden"),
		Code:    CodeForbidden,
		Status:  http.StatusForbidden,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// NewConflict creates a conflict error with caller information
func NewConflict(message string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return &AppError{
		Err:     errors.New("conflict"),
		Code:    CodeConflict,
		Status:  http.StatusConflict,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// Wrap wraps an error with a message and caller information
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return fmt.Errorf("%s: %w [%s:%d]", message, err, file, line)
}

// IsAppError checks if the error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// IsValidationError checks if the error is a ValidationError
func IsValidationError(err error) bool {
	return GetValidationError(err) != nil
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == CodeNotFound
	}
	return false
}

// IsBadRequest checks if the error is a bad request error
func IsBadRequest(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == CodeBadRequest
	}
	return false
}

// IsInternalError checks if the error is an internal server error
func IsInternalError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == CodeInternalServerError
	}
	return false
}

// HandleError handles an error and returns the appropriate error response
func HandleError(err error) (int, ErrorResponse) {
	// Check if it's an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status, appErr.GetErrorResponse()
	}

	// Check if it's a ValidationError
	if validationErr := GetValidationError(err); validationErr != nil {
		// Convert to AppError
		appErr = validationErr.ToAppError()
		return appErr.Status, appErr.GetErrorResponse()
	}

	// Default to internal server error
	_, file, line, _ := runtime.Caller(1)
	file = getShortFilePath(file)

	return http.StatusInternalServerError, ErrorResponse{
		Code:       CodeInternalServerError,
		Message:    "An unexpected error occurred",
		ServerTime: time.Now(),
		Source:     fmt.Sprintf("%s:%d", file, line),
	}
}

// Helper function to get the short file path
func getShortFilePath(file string) string {
	parts := strings.Split(file, "/")
	if len(parts) <= 2 {
		return file
	}

	// Return at most the last 3 parts of the path
	start := len(parts) - 3
	if start < 0 {
		start = 0
	}

	return strings.Join(parts[start:], "/")
}

// ToJSON converts an error to JSON
func ToJSON(err error) string {
	status, response := HandleError(err)
	response.Data = map[string]interface{}{"status": status}

	b, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		return fmt.Sprintf(`{"code":"internal_server_error","message":"Error marshaling error response","server_time":"%s"}`, time.Now().Format(time.RFC3339))
	}

	return string(b)
}
