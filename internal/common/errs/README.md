# Error Handling System

This package provides a centralized, standardized error handling system for the application. It includes:

1. **Error Types**: Various error types with detailed information
2. **Error Codes**: Standardized error codes for consistent client-side handling
3. **File/Line Info**: Each error includes file and line information for better debugging
4. **ValidationError**: Special error type for field-level validation errors
5. **JSON Formatting**: Easy conversion to JSON for API responses

## Basic Usage

### Creating and Using Errors

```go
// Create a simple error
err := errs.NewBadRequest("Invalid input parameter")

// Create a not found error
err := errs.NewNotFound("Product not found")

// Create a custom error with specific code and status
err := errs.New(
    errors.New("custom error"),
    "custom_error_code",
    http.StatusBadRequest,
    "Human-readable error message",
)

// Add data to an error
err := errs.NewWithData(
    errors.New("order failed"),
    "order_failed",
    http.StatusBadRequest,
    "Order processing failed",
    map[string]interface{}{
        "order_id": "12345",
        "reason": "payment_declined",
    },
)

// Wrap an error with additional context
err := errs.Wrap(err, "Failed while processing order")
```

### Validation Errors

```go
// Create a validation error
validationErr := errs.NewValidationError(
    errors.New("validation failed"),
    errs.CodeValidationError,
    "The form contains errors",
)

// Add field-specific errors
if len(product.Name) == 0 {
    validationErr.AddDetail(errors.New("name is required"))
}

if product.Price <= 0 {
    validationErr.AddDetail(errors.New("price must be greater than zero"))
}

// Only return the error if there are details
if validationErr.DetailLength() > 0 {
    return validationErr
}
```

### Using Predefined Errors

```go
// Use a predefined error
return errs.ErrProductNotFound

// Use a predefined validation error
return errs.ErrPaymentMethodNotAllowed
```

## Error Handling in HTTP Handlers

In HTTP handlers, use the middleware package's `RespondWithError` function:

```go
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request, id string) {
    ctx := r.Context()

    // Validate ID
    productID, err := uuid.Parse(id)
    if err != nil {
        middleware.RespondWithError(w, errs.NewBadRequest("Invalid product ID"))
        return
    }

    // Call use case
    product, err := h.productUseCase.GetByID(ctx, productID)
    if err != nil {
        middleware.RespondWithError(w, err)
        return
    }

    // Return success response
    middleware.RespondWithJSON(w, http.StatusOK, product)
}
```

## Updating Existing HTTP Handlers

To update an existing HTTP handler to use the new error system, follow these steps:

1. Update the imports:

```go
import (
    // Remove old error package
    // "github.com/fanzru/e-commerce-be/pkg/errors"

    // Add new error package
    "github.com/fanzru/e-commerce-be/internal/common/errs"
    "github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
)
```

2. Update the error handler function:

```go
// Replace this
func handleError(w http.ResponseWriter, err error) {
    var status int
    var code string
    var message string

    // Extract status, code, message from error
    // ...

    // Send error response
    respondJSON(w, status, errorResponse)
}

// With this
func handleError(w http.ResponseWriter, err error) {
    middleware.RespondWithError(w, err)
}
```

3. Update the JSON response function:

```go
// Replace this
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// With this
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    middleware.RespondWithJSON(w, status, data)
}
```

4. Update error creation calls:

```go
// Replace these
handleError(w, errors.NewBadRequest("Invalid parameter"))
handleError(w, errors.NewNotFound("Resource not found"))

// With these
handleError(w, errs.NewBadRequest("Invalid parameter"))
handleError(w, errs.NewNotFound("Resource not found"))
```

5. Update the HTTP server initialization:

```go
// Replace this
return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
    BaseRouter: http.NewServeMux(),
    ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
        handleError(w, err)
    },
})

// With this
return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
    BaseRouter: http.NewServeMux(),
    ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
        middleware.RespondWithError(w, err)
    },
})
```

## Checking Error Types

```go
// Check if an error is a not found error
if errs.IsNotFound(err) {
    // Handle not found case
}

// Check if an error is a validation error
if errs.IsValidationError(err) {
    validationErr := errs.GetValidationError(err)
    // Handle validation errors
}

// Check if an error is an internal server error
if errs.IsInternalError(err) {
    // Log the error and handle it
}
```

## Standard JSON Response Format

Errors are formatted as JSON with this structure:

```json
{
  "code": "error_code",
  "message": "Human-readable error message",
  "data": null,
  "server_time": "2023-05-19T01:23:37.329879+07:00",
  "source": "file.go:123"
}
```

The `source` field shows the file and line where the error occurred, making debugging easier.

## Adding New Error Codes

To add new error codes, edit the constants in `errs.go` and add predefined errors in `errors_def.go`.

## Integration with Middleware

The error handling system integrates with the middleware package, which provides:

1. `ErrorHandlerMiddleware`: Middleware to handle and log errors
2. `RespondWithError`: Function to respond with an error
3. `RespondWithJSON`: Function to respond with a JSON success response
