package middleware

import (
	"net/http"

	"github.com/fanzru/e-commerce-be/internal/infrastructure/config"
)

// Factory creates and configures middleware
type Factory struct {
	tokenValidator TokenValidator
	config         *config.Config
}

// NewFactory creates a new middleware factory
func NewFactory(config *config.Config) *Factory {
	// Create a JWT validator
	validator := NewJWTValidator(config.JWT.SecretKey)

	return &Factory{
		tokenValidator: validator,
		config:         config,
	}
}

// Chain applies a chain of middleware to a handler
func Chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	// Apply middleware in reverse so they execute in the order they are passed
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// DefaultMiddleware returns the default middleware chain
func (f *Factory) DefaultMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		RequestID,
		TraceMiddleware,
		APILogger,
		Recoverer,
	}
}

// AuthMiddleware returns middleware for protected routes
func (f *Factory) AuthMiddleware(authType AuthType) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		RequestID,
		TraceMiddleware,
		APILogger,
		Recoverer,
		Auth(f.tokenValidator, authType),
	}
}

// Apply applies middleware to a handler
func (f *Factory) Apply(h http.Handler, authType AuthType) http.Handler {
	var middleware []func(http.Handler) http.Handler

	if authType == AuthTypePublic {
		middleware = f.DefaultMiddleware()
	} else {
		middleware = f.AuthMiddleware(authType)
	}

	return Chain(h, middleware...)
}

// WrapFunc wraps an http.HandlerFunc with middleware
func (f *Factory) WrapFunc(authType AuthType, fn http.HandlerFunc) http.Handler {
	return f.Apply(fn, authType)
}
