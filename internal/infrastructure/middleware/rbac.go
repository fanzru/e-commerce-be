package middleware

import (
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
)

// OperationAccess defines which roles can access a specific operation
type OperationAccess struct {
	OperationID  string     // Name of the operation (e.g., "ListProducts")
	AllowedRoles []AuthType // Roles that can access this operation
}

// RBACMiddleware provides role-based access control at the operation level
type RBACMiddleware struct {
	factory         *Factory
	operationMap    map[string]OperationAccess
	defaultAccess   []AuthType                   // Default allowed roles if operation not found
	pathToOperation map[string]map[string]string // Method -> Path -> OperationID
}

// NewRBACMiddleware creates a new role-based access control middleware
func NewRBACMiddleware(factory *Factory) *RBACMiddleware {
	return &RBACMiddleware{
		factory:         factory,
		operationMap:    make(map[string]OperationAccess),
		defaultAccess:   []AuthType{AuthTypeRoleAdmin}, // By default, only admins can access undefined operations
		pathToOperation: make(map[string]map[string]string),
	}
}

// RegisterPathPattern registers a path pattern for an operation
// Used to dynamically register paths from the OpenAPI definition
func (rm *RBACMiddleware) RegisterPathPattern(method, path, operationID string) *RBACMiddleware {
	// Ensure the method map exists
	if _, exists := rm.pathToOperation[method]; !exists {
		rm.pathToOperation[method] = make(map[string]string)
	}

	// Register the path pattern
	rm.pathToOperation[method][path] = operationID

	Logger.Info("RBAC: Registered path pattern",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("operation", operationID))

	return rm
}

// WithOperation adds an operation with its allowed roles
func (rm *RBACMiddleware) WithOperation(operationID string, allowedRoles ...AuthType) *RBACMiddleware {
	rm.operationMap[operationID] = OperationAccess{
		OperationID:  operationID,
		AllowedRoles: allowedRoles,
	}

	// Log the operation registration
	Logger.Info("RBAC: Registered operation access control",
		slog.String("operation", operationID),
		slog.Any("allowed_roles", allowedRoles))

	return rm
}

// WithDefaultRoles sets the default roles allowed for operations not explicitly defined
func (rm *RBACMiddleware) WithDefaultRoles(allowedRoles ...AuthType) *RBACMiddleware {
	rm.defaultAccess = allowedRoles

	// Log the default roles
	Logger.Info("RBAC: Set default roles",
		slog.Any("default_roles", allowedRoles))

	return rm
}

// Wrap wraps an HTTP handler with RBAC middleware
func (rm *RBACMiddleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID for correlation
		requestID := GetRequestID(r.Context())

		// Determine the operation from the request
		operationID := rm.getOperationFromRequest(r)
		Logger.Debug("RBAC: Handling request",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("operation", operationID))

		// Get access control for this operation
		access, exists := rm.operationMap[operationID]
		allowedRoles := rm.defaultAccess
		if exists {
			allowedRoles = access.AllowedRoles
			Logger.Debug("RBAC: Found operation config",
				slog.String("request_id", requestID),
				slog.String("operation", operationID),
				slog.Any("allowed_roles", allowedRoles))
		} else {
			Logger.Debug("RBAC: Operation not explicitly configured, using defaults",
				slog.String("request_id", requestID),
				slog.String("operation", operationID),
				slog.Any("default_roles", allowedRoles))
		}

		// Check if public access is allowed
		publicAllowed := containsAuthType(allowedRoles, AuthTypePublic)
		if publicAllowed {
			Logger.Info("RBAC: Public access allowed",
				slog.String("request_id", requestID),
				slog.String("operation", operationID))

			// For public operations, just apply default middleware
			middleware := rm.factory.DefaultMiddleware()
			Chain(handler, middleware...).ServeHTTP(w, r)
			return
		}

		// For protected operations, we need to check the user's role
		// First, extract token and validate
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			Logger.Debug("RBAC: Authorization header missing",
				slog.String("request_id", requestID),
				slog.String("operation", operationID))

			respondWithError(w, http.StatusUnauthorized, "Authorization header missing")
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			Logger.Debug("RBAC: Invalid authorization format",
				slog.String("request_id", requestID),
				slog.String("operation", operationID),
				slog.String("auth_header", authHeader))

			respondWithError(w, http.StatusUnauthorized, "Invalid authorization format, expected 'Bearer TOKEN'")
			return
		}

		tokenString := parts[1]

		// Validate token using the TokenValidator from the factory
		claims, err := rm.factory.tokenValidator.ValidateToken(tokenString)
		if err != nil {
			Logger.Debug("RBAC: Invalid token",
				slog.String("request_id", requestID),
				slog.String("operation", operationID),
				slog.String("error", err.Error()))

			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Determine the user's role auth type
		var userAuthType AuthType
		switch claims.Role {
		case entity.RoleAdmin:
			userAuthType = AuthTypeRoleAdmin
		case entity.RoleCustomer:
			userAuthType = AuthTypeRoleCustomer
		default:
			userAuthType = AuthTypeBearer // Just authenticated, no specific role
		}

		Logger.Debug("RBAC: User authenticated",
			slog.String("request_id", requestID),
			slog.String("operation", operationID),
			slog.String("user_id", claims.UserID),
			slog.String("email", claims.Email),
			slog.String("role", string(claims.Role)),
			slog.String("auth_type", string(userAuthType)))

		// Check if the user's role is allowed
		if !containsAuthType(allowedRoles, userAuthType) {
			Logger.Debug("RBAC: Insufficient permissions",
				slog.String("request_id", requestID),
				slog.String("operation", operationID),
				slog.String("user_role", string(userAuthType)),
				slog.Any("allowed_roles", allowedRoles))

			respondWithError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}

		// Role is allowed, apply appropriate middleware chain
		Logger.Debug("RBAC: Access granted",
			slog.String("request_id", requestID),
			slog.String("operation", operationID),
			slog.String("user_id", claims.UserID),
			slog.String("role", string(claims.Role)))

		middleware := rm.factory.AuthMiddleware(userAuthType)
		Chain(handler, middleware...).ServeHTTP(w, r)
	})
}

// Helper function to check if an auth type is in a slice
func containsAuthType(slice []AuthType, item AuthType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getOperationFromRequest extracts the operation ID from the request
func (rm *RBACMiddleware) getOperationFromRequest(r *http.Request) string {
	path := r.URL.Path
	method := r.Method

	// Normalize the path for API versioning
	normalizedPath := path

	// TODO: enable this when we have a way to handle api versioning
	// if strings.HasPrefix(path, "/api/v1") {
	// 	normalizedPath = strings.TrimPrefix(path, "/api/v1")
	// }

	// First, try exact path match
	if methodMap, exists := rm.pathToOperation[method]; exists {
		if operationID, pathExists := methodMap[normalizedPath]; pathExists {
			return operationID
		}

		// If exact match fails, try pattern matching for paths with parameters
		for patternPath, opID := range methodMap {
			if matchPathPattern(patternPath, normalizedPath) {
				Logger.Debug("RBAC: Matched path pattern",
					slog.String("pattern", patternPath),
					slog.String("actual_path", normalizedPath),
					slog.String("operation", opID))
				return opID
			}
		}
	}

	// Log that we couldn't identify the operation
	Logger.Warn("RBAC: Could not identify operation",
		slog.String("method", method),
		slog.String("path", normalizedPath),
		slog.String("original_path", path))

	// If no specific operation is identified, return a default
	return "Unknown" + method + normalizedPath
}

// matchPathPattern checks if an actual path matches a path pattern with parameters
// Pattern example: /api/v1/products/{id}
// Actual path: /api/v1/products/123
func matchPathPattern(pattern, actualPath string) bool {
	// Split both paths into segments
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")
	actualSegments := strings.Split(strings.Trim(actualPath, "/"), "/")

	// If they have different number of segments, they don't match
	if len(patternSegments) != len(actualSegments) {
		return false
	}

	// Check each segment
	for i, patternSeg := range patternSegments {
		actualSeg := actualSegments[i]

		// If pattern segment is a parameter (enclosed in {}), it matches any value
		if strings.HasPrefix(patternSeg, "{") && strings.HasSuffix(patternSeg, "}") {
			continue // Parameter segment matches anything
		}

		// Otherwise, segments must match exactly
		if patternSeg != actualSeg {
			return false
		}
	}

	// All segments matched
	return true
}

// RegisterServerInterface analyzes a server interface and automatically registers its methods
func (rm *RBACMiddleware) RegisterServerInterface(iface interface{}) *RBACMiddleware {
	ifaceType := reflect.TypeOf(iface)

	// Check if iface is a pointer
	if ifaceType.Kind() == reflect.Ptr {
		ifaceType = ifaceType.Elem()
	}

	// Log what we're registering
	Logger.Info("RBAC: Registering server interface",
		slog.String("interface", ifaceType.Name()))

	// For each method in the interface
	for i := 0; i < ifaceType.NumMethod(); i++ {
		method := ifaceType.Method(i)

		// Skip non-exported methods
		if !method.IsExported() {
			continue
		}

		// Get the method name as the operation ID
		operationID := method.Name

		Logger.Info("RBAC: Found interface method",
			slog.String("method", operationID))

		// If this operation hasn't already been registered, register it with default roles
		if _, exists := rm.operationMap[operationID]; !exists {
			rm.WithOperation(operationID, rm.defaultAccess...)
		}
	}

	return rm
}

// GetHandlerName gets the name of a handler function
// This can be used to extract operation names from http.HandlerFunc
func GetHandlerName(handler interface{}) string {
	// Get the function's name
	funcName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

	// Extract just the method name part
	parts := strings.Split(funcName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return funcName
}
