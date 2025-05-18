package middleware

import (
	"net/http"
	"strings"
)

// HandlerMiddleware represents a middleware that can be applied to a specific HTTP handler
type HandlerMiddleware func(http.Handler) http.Handler

// HandlerFunc is a function that handles an HTTP request
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// ProtectedHandler wraps an HTTP handler with authentication middleware
type ProtectedHandler struct {
	factory   *Factory
	authType  AuthType
	pathRoles map[string]AuthType // Maps path patterns to required roles
}

// NewProtectedHandler creates a new protected handler
func NewProtectedHandler(factory *Factory) *ProtectedHandler {
	return &ProtectedHandler{
		factory:   factory,
		authType:  AuthTypePublic, // Default to public
		pathRoles: make(map[string]AuthType),
	}
}

// WithDefaultAuth sets the default authentication type for all routes
func (ph *ProtectedHandler) WithDefaultAuth(authType AuthType) *ProtectedHandler {
	ph.authType = authType
	return ph
}

// WithPathAuth sets the authentication type for a specific path pattern
func (ph *ProtectedHandler) WithPathAuth(pathPattern string, authType AuthType) *ProtectedHandler {
	ph.pathRoles[pathPattern] = authType
	return ph
}

// Wrap wraps an HTTP handler with the appropriate authentication middleware
func (ph *ProtectedHandler) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine the auth type based on path
		authType := ph.getAuthTypeForPath(r.URL.Path)

		// Apply the middleware chain
		if authType == AuthTypePublic {
			// For public endpoints, just apply default middleware
			middleware := ph.factory.DefaultMiddleware()
			Chain(handler, middleware...).ServeHTTP(w, r)
		} else {
			// For authenticated endpoints, apply auth middleware
			middleware := ph.factory.AuthMiddleware(authType)
			Chain(handler, middleware...).ServeHTTP(w, r)
		}
	})
}

// WrapFunc wraps an HTTP handler function with the appropriate authentication middleware
func (ph *ProtectedHandler) WrapFunc(handlerFunc http.HandlerFunc) http.Handler {
	return ph.Wrap(http.HandlerFunc(handlerFunc))
}

// getAuthTypeForPath determines the authentication type required for a path
func (ph *ProtectedHandler) getAuthTypeForPath(path string) AuthType {
	// Check for specific path patterns
	for pattern, authType := range ph.pathRoles {
		if matchesPattern(pattern, path) {
			return authType
		}
	}

	// Default to the handler's default auth type
	return ph.authType
}

// matchesPattern checks if a path matches a pattern
func matchesPattern(pattern, path string) bool {
	// Simple wildcard matching (can be enhanced for more complex patterns)
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

// MethodAuthHandler allows different auth types for different HTTP methods
type MethodAuthHandler struct {
	factory     *Factory
	defaultAuth AuthType
	methodAuths map[string]AuthType            // Maps HTTP methods to required auth types
	pathMethods map[string]map[string]AuthType // Maps paths to method-specific auth types
}

// NewMethodAuthHandler creates a new method-auth handler
func NewMethodAuthHandler(factory *Factory) *MethodAuthHandler {
	return &MethodAuthHandler{
		factory:     factory,
		defaultAuth: AuthTypePublic,
		methodAuths: make(map[string]AuthType),
		pathMethods: make(map[string]map[string]AuthType),
	}
}

// WithDefaultAuth sets the default authentication type
func (mh *MethodAuthHandler) WithDefaultAuth(authType AuthType) *MethodAuthHandler {
	mh.defaultAuth = authType
	return mh
}

// WithMethodAuth sets the authentication type for a specific HTTP method
func (mh *MethodAuthHandler) WithMethodAuth(method string, authType AuthType) *MethodAuthHandler {
	mh.methodAuths[strings.ToUpper(method)] = authType
	return mh
}

// WithPathMethodAuth sets the authentication type for a specific path and HTTP method
func (mh *MethodAuthHandler) WithPathMethodAuth(path, method string, authType AuthType) *MethodAuthHandler {
	if _, ok := mh.pathMethods[path]; !ok {
		mh.pathMethods[path] = make(map[string]AuthType)
	}
	mh.pathMethods[path][strings.ToUpper(method)] = authType
	return mh
}

// Wrap wraps an HTTP handler with method-specific authentication middleware
func (mh *MethodAuthHandler) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine the auth type based on method and path
		authType := mh.getAuthTypeForRequest(r)

		// Apply the middleware chain
		if authType == AuthTypePublic {
			// For public endpoints, just apply default middleware
			middleware := mh.factory.DefaultMiddleware()
			Chain(handler, middleware...).ServeHTTP(w, r)
		} else {
			// For authenticated endpoints, apply auth middleware
			middleware := mh.factory.AuthMiddleware(authType)
			Chain(handler, middleware...).ServeHTTP(w, r)
		}
	})
}

// WrapFunc wraps an HTTP handler function with method-specific authentication middleware
func (mh *MethodAuthHandler) WrapFunc(handlerFunc http.HandlerFunc) http.Handler {
	return mh.Wrap(http.HandlerFunc(handlerFunc))
}

// getAuthTypeForRequest determines the authentication type required for a request
func (mh *MethodAuthHandler) getAuthTypeForRequest(r *http.Request) AuthType {
	method := strings.ToUpper(r.Method)
	path := r.URL.Path

	// Check for specific path+method combinations
	if methodAuths, ok := mh.pathMethods[path]; ok {
		if authType, ok := methodAuths[method]; ok {
			return authType
		}
	}

	// Check for wildcard path patterns with method combinations
	for pattern, methodAuths := range mh.pathMethods {
		if matchesPattern(pattern, path) {
			if authType, ok := methodAuths[method]; ok {
				return authType
			}
		}
	}

	// Check for method-specific auth types
	if authType, ok := mh.methodAuths[method]; ok {
		return authType
	}

	// Default to the handler's default auth type
	return mh.defaultAuth
}
