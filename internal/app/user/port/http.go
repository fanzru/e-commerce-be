package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/fanzru/e-commerce-be/internal/app/user/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/user/usecase"
	"github.com/fanzru/e-commerce-be/pkg/formatter"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userUseCase usecase.UserUseCase
}

// NewUserHandler creates a new user HTTP handler
func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// NewHTTPServer creates a new HTTP server for users
func NewHTTPServer(userUseCase usecase.UserUseCase) http.Handler {
	handler := NewUserHandler(userUseCase)
	return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			handleError(w, err)
		},
	})
}

// RegisterUser handles POST /auth/register requests
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody genhttp.RegisterUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Map request to params
	registerParams := params.RegisterUserParams{
		Email:    string(reqBody.Email),
		Password: reqBody.Password,
		Name:     reqBody.Name,
	}

	// Call use case
	user, err := h.userUseCase.Register(ctx, registerParams)
	if err != nil {
		switch err {
		case errs.ErrEmailAlreadyExists:
			handleError(w, formatter.NewHTTPError(http.StatusConflict, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	// Create response
	code := "SUCCESS"
	message := "User registered successfully"
	id := user.ID
	email := openapi_types.Email(user.Email)
	role := genhttp.UserRole(user.Role)
	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt
	now := time.Now()

	response := genhttp.UserResponse{
		Code:       code,
		Message:    message,
		ServerTime: now,
		Data: genhttp.User{
			Id:        &id,
			Email:     &email,
			Name:      &user.Name,
			Role:      &role,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		},
	}

	respondJSON(w, http.StatusCreated, response)
}

// LoginUser handles POST /auth/login requests
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody genhttp.LoginUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Map request to params
	loginParams := params.LoginUserParams{
		Email:    string(reqBody.Email),
		Password: reqBody.Password,
	}

	// Call use case
	tokenPair, err := h.userUseCase.Login(ctx, loginParams)
	if err != nil {
		switch err {
		case errs.ErrInvalidCredentials:
			handleError(w, formatter.NewHTTPError(http.StatusUnauthorized, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	// Create response
	code := "SUCCESS"
	message := "Login successful"
	accessToken := tokenPair.AccessToken
	refreshToken := tokenPair.RefreshToken
	expiresIn := tokenPair.ExpiresIn
	tokenType := "bearer"
	now := time.Now()

	response := genhttp.TokenResponse{
		Code:       code,
		Message:    message,
		ServerTime: now,
	}

	// Set data field
	response.Data.AccessToken = &accessToken
	response.Data.RefreshToken = &refreshToken
	response.Data.ExpiresIn = &expiresIn
	response.Data.TokenType = &tokenType

	respondJSON(w, http.StatusOK, response)
}

// LogoutUser handles POST /auth/logout requests
func (h *UserHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Call use case
	if err := h.userUseCase.Logout(ctx, reqBody.RefreshToken); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCurrentUser handles GET /users/me requests
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (added by auth middleware)
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		handleError(w, formatter.NewHTTPError(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	// Parse user ID from string to UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid user ID"))
		return
	}

	// Call use case to get user by ID
	user, err := h.userUseCase.GetUserByID(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Create response
	code := "SUCCESS"
	message := "User retrieved successfully"
	id := user.ID
	email := openapi_types.Email(user.Email)
	role := genhttp.UserRole(user.Role)
	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt
	now := time.Now()

	response := genhttp.UserResponse{
		Code:       code,
		Message:    message,
		ServerTime: now,
		Data: genhttp.User{
			Id:        &id,
			Email:     &email,
			Name:      &user.Name,
			Role:      &role,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// Helper functions

// handleError handles errors and sends appropriate HTTP responses
func handleError(w http.ResponseWriter, err error) {
	var status int
	var message string
	var code string

	switch e := err.(type) {
	case *formatter.HTTPError:
		status = e.StatusCode
		message = e.Message
		code = "ERROR"
	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
		code = "ERROR"
		if err != nil {
			message = err.Error()
		}
	}

	now := time.Now()
	errorResponse := genhttp.ErrorResponse{
		Code:       code,
		Message:    message,
		ServerTime: now,
		Data:       nil,
	}

	respondJSON(w, status, errorResponse)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Error encoding response","code":"ERROR"}`))
	}
}
