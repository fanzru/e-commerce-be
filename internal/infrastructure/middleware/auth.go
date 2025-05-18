package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/fanzru/e-commerce-be/pkg/formatter"
	"github.com/golang-jwt/jwt/v5"
)

// AuthType defines the type of authentication required for a route
type AuthType string

const (
	// AuthTypePublic indicates no authentication is required
	AuthTypePublic AuthType = "public"
	// AuthTypeBearer indicates JWT bearer token authentication is required
	AuthTypeBearer AuthType = "bearer"
	// AuthTypeRoleAdmin indicates admin role is required
	AuthTypeRoleAdmin AuthType = "role:admin"
	// AuthTypeRoleCustomer indicates customer role is required
	AuthTypeRoleCustomer AuthType = "role:customer"

	// ContextUserKey is the key used to store user information in request context
	ContextUserKey = "user"
	// ContextTokenClaimsKey is the key used to store token claims in request context
	ContextTokenClaimsKey = "token_claims"
)

// JWTConfig contains JWT configuration
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
}

// AuthConfig is the configuration for the authentication middleware
type AuthConfig struct {
	JWT JWTConfig
}

// TokenValidator defines an interface for JWT token validation
type TokenValidator interface {
	ValidateToken(token string) (*params.TokenClaims, error)
}

// Auth middleware provides authentication and authorization
func Auth(validator TokenValidator, authType AuthType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Public routes don't require authentication
			if authType == AuthTypePublic {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization format, expected 'Bearer TOKEN'")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := validator.ValidateToken(tokenString)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
				return
			}

			// Check role-based access
			if authType == AuthTypeRoleAdmin && claims.Role != entity.RoleAdmin {
				respondWithError(w, http.StatusForbidden, "Admin access required")
				return
			}

			if authType == AuthTypeRoleCustomer && claims.Role != entity.RoleCustomer {
				respondWithError(w, http.StatusForbidden, "Customer access required")
				return
			}

			// Store user info in context
			ctx := context.WithValue(r.Context(), ContextTokenClaimsKey, claims)

			// If we need the full user object, we could fetch it here:
			// userID, _ := uuid.Parse(claims.UserID)
			// user, _ := userUC.GetUserByID(ctx, userID)
			// ctx = context.WithValue(ctx, ContextUserKey, user)

			// Continue with the next handler with the enhanced context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(ContextUserKey).(*entity.User)
	if !ok || user == nil {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// GetTokenClaimsFromContext retrieves the token claims from the request context
func GetTokenClaimsFromContext(ctx context.Context) (*params.TokenClaims, error) {
	claims, ok := ctx.Value(ContextTokenClaimsKey).(*params.TokenClaims)
	if !ok || claims == nil {
		return nil, errors.New("token claims not found in context")
	}
	return claims, nil
}

// GenerateJWT creates a new JWT token for the user
func GenerateJWT(user *entity.User, config JWTConfig) (string, int, error) {
	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(config.ExpirationHours) * time.Hour)
	expiresIn := int(time.Until(expirationTime).Seconds())

	// Create claims
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string, secretKey string) (*params.TokenClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	expClaim, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid expiration claim")
	}

	if time.Now().Unix() > int64(expClaim) {
		return nil, errors.New("token expired")
	}

	// Extract user information from claims
	var userID string

	// Try to get user_id first, then fall back to sub if needed
	if userIDVal, hasUserID := claims["user_id"]; hasUserID {
		var ok bool
		userID, ok = userIDVal.(string)
		if !ok {
			return nil, errors.New("invalid user ID claim")
		}
	} else if subVal, hasSub := claims["sub"]; hasSub {
		var ok bool
		userID, ok = subVal.(string)
		if !ok {
			return nil, errors.New("invalid user ID claim")
		}
	} else {
		return nil, errors.New("missing user ID claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email claim")
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid role claim")
	}

	var role entity.UserRole
	switch roleStr {
	case string(entity.RoleAdmin), "role:admin":
		role = entity.RoleAdmin
	case string(entity.RoleCustomer), "role:customer":
		role = entity.RoleCustomer
	default:
		return nil, errors.New("unknown role")
	}

	return &params.TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
	}, nil
}

// JWTValidator implements the TokenValidator interface using JWT
type JWTValidator struct {
	SecretKey string
}

// NewJWTValidator creates a new JWTValidator
func NewJWTValidator(secretKey string) *JWTValidator {
	return &JWTValidator{
		SecretKey: secretKey,
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *JWTValidator) ValidateToken(tokenString string) (*params.TokenClaims, error) {
	return ValidateJWT(tokenString, v.SecretKey)
}

// respondWithError sends an error response in JSON format
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	response := formatter.StandardResponse{
		Code:       "ERROR",
		Message:    message,
		ServerTime: time.Now(),
		Data:       nil,
	}

	formatter.JSON(w, statusCode, response)
}
