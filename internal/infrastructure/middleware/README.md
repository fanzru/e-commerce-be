# Middleware Package

This package provides middleware components for the e-commerce API, including:

## Authentication Middleware

The authentication middleware (`auth.go`) provides role-based access control:

- **Public Access**: No authentication required
- **Bearer Authentication**: JWT token validation
- **Role-based Access**: Admin and Customer role checks

### Usage

```go
// For a public endpoint (no auth required)
http.Handle("/public", middlewareFactory.WrapFunc(middleware.AuthTypePublic, publicHandler))

// For an authenticated endpoint (any user)
http.Handle("/user", middlewareFactory.WrapFunc(middleware.AuthTypeBearer, userHandler))

// For an admin-only endpoint
http.Handle("/admin", middlewareFactory.WrapFunc(middleware.AuthTypeRoleAdmin, adminHandler))

// For a customer-only endpoint
http.Handle("/customer", middlewareFactory.WrapFunc(middleware.AuthTypeRoleCustomer, customerHandler))
```

## Handler-Level Authentication

The handler-level authentication (`handler.go`) provides more fine-grained control over authentication requirements at the handler level rather than the endpoint level:

### Protected Handler

Use `ProtectedHandler` when you want to set different auth requirements for different path patterns within the same handler:

```go
// Create a protected handler with default authentication type
protected := middleware.NewProtectedHandler(middlewareFactory).
    WithDefaultAuth(middleware.AuthTypeBearer).
    // Set specific auth requirements for particular paths
    WithPathAuth("/users/public/*", middleware.AuthTypePublic).
    WithPathAuth("/users/admin/*", middleware.AuthTypeRoleAdmin)

// Wrap your handler with the authentication middleware
http.Handle("/users/", protected.Wrap(userHandler))
```

### Method-Specific Authentication

Use `MethodAuthHandler` when different HTTP methods require different authentication levels:

```go
// Create a method-auth handler with default and method-specific authentication
methodAuth := middleware.NewMethodAuthHandler(middlewareFactory).
    WithDefaultAuth(middleware.AuthTypePublic).
    // GET is public, POST and DELETE require admin role
    WithMethodAuth(http.MethodGet, middleware.AuthTypePublic).
    WithMethodAuth(http.MethodPost, middleware.AuthTypeRoleAdmin).
    WithMethodAuth(http.MethodDelete, middleware.AuthTypeRoleAdmin).
    // Specific path+method combinations
    WithPathMethodAuth("/resources/special", http.MethodPut, middleware.AuthTypeRoleAdmin)

// Wrap your handler with the method-specific authentication middleware
http.Handle("/resources/", methodAuth.Wrap(resourceHandler))
```

## Request ID Middleware

The request ID middleware (`requestid.go`) ensures each request has a unique ID:

- Preserves existing X-Request-ID header if present
- Generates a new UUID if no request ID is provided
- Adds the request ID to the response headers
- Stores the request ID in the request context

### Usage

The request ID middleware is automatically applied as part of the middleware chain.

To retrieve the request ID in your handlers:

```go
requestID := middleware.GetRequestID(r.Context())
```

## Logging Middleware

The logging middleware (`logger.go`) provides structured logging using Go's `slog` package:

- Logs detailed information about requests and responses
- Uses appropriate log levels based on response status
- Masks sensitive information in request/response bodies
- Includes request ID, method, path, headers, timing information, etc.
- Structured JSON output for easy parsing and analysis
- OpenTelemetry integration for distributed tracing

### Usage

The logging middleware is automatically applied as part of the middleware chain.

Access the logger in your code:

```go
middleware.Logger.Info("Custom log message",
    slog.String("key", "value"),
    slog.Int("count", 42))
```

## Middleware Factory

The middleware factory (`middleware.go`) provides a convenient way to apply middleware chains:

```go
// Create a middleware factory
factory := middleware.NewFactory(userUseCase, config)

// Apply middleware to a handler with specific auth type
handler := factory.Apply(myHandler, middleware.AuthTypeBearer)

// Or wrap a http.HandlerFunc
handler := factory.WrapFunc(middleware.AuthTypePublic, myHandlerFunc)
```

## Tracing Middleware

The tracing middleware (`otel.go`) provides distributed tracing with OpenTelemetry:

- Automatically creates spans for each HTTP request
- Captures request details (method, path, status code)
- Integrates with the logging system to include trace context in logs
- Supports exporting traces to OpenTelemetry collectors

### Usage

Initialize OpenTelemetry at application startup:

```go
shutdown, err := middleware.InitOTEL()
if err != nil {
    log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
}
defer shutdown(context.Background())
```

To log with trace context:

```go
middleware.LogWithContext(ctx, slog.LevelInfo, "Message with trace context")
```

## Configuration

Middleware configuration is loaded from environment variables:

```
# JWT Configuration
JWT_SECRET_KEY=your-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24

# OpenTelemetry Configuration
OTEL_ENABLED=true
OTEL_SERVICE_NAME=e-commerce-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```
