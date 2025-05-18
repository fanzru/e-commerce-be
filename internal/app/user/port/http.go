package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/fanzru/e-commerce-be/internal/app/user/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/user/usecase"
	"github.com/fanzru/e-commerce-be/pkg/formatter"
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
		Code:    &code,
		Message: &message,
		Data: &genhttp.User{
			Id:        &id,
			Email:     &email,
			Name:      &user.Name,
			Role:      &role,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		},
		ServerTime: &now,
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
	now := time.Now()

	response := genhttp.TokenResponse{
		Code:    &code,
		Message: &message,
		Data: &genhttp.TokenPair{
			AccessToken:  &accessToken,
			RefreshToken: &refreshToken,
			ExpiresIn:    &expiresIn,
		},
		ServerTime: &now,
	}

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

// RefreshToken handles POST /auth/refresh requests
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Map request to params
	refreshParams := params.RefreshTokenParams{
		RefreshToken: reqBody.RefreshToken,
	}

	// Call use case
	tokenPair, err := h.userUseCase.RefreshToken(ctx, refreshParams)
	if err != nil {
		switch err {
		case errs.ErrInvalidRefreshToken, errs.ErrRefreshTokenExpired:
			handleError(w, formatter.NewHTTPError(http.StatusUnauthorized, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	// Create response
	code := "SUCCESS"
	message := "Token refreshed successfully"
	accessToken := tokenPair.AccessToken
	refreshToken := tokenPair.RefreshToken
	expiresIn := tokenPair.ExpiresIn
	now := time.Now()

	response := genhttp.TokenResponse{
		Code:    &code,
		Message: &message,
		Data: &genhttp.TokenPair{
			AccessToken:  &accessToken,
			RefreshToken: &refreshToken,
			ExpiresIn:    &expiresIn,
		},
		ServerTime: &now,
	}

	respondJSON(w, http.StatusOK, response)
}

// ListUsers handles GET /users requests
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request, params genhttp.ListUsersParams) {
	ctx := r.Context()

	// Set default pagination values
	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	limit := 10
	if params.Limit != nil {
		limit = *params.Limit
	}

	// Parse role filter
	var roleFilter *entity.UserRole
	if params.Role != nil {
		var role entity.UserRole
		switch *params.Role {
		case "admin":
			role = entity.RoleAdmin
		case "customer":
			role = entity.RoleCustomer
		default:
			handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid role"))
			return
		}
		roleFilter = &role
	}

	// Call use case
	users, total, err := h.userUseCase.ListUsers(ctx, page, limit, roleFilter)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map users to response format
	userList := make([]genhttp.User, len(users))
	for i, user := range users {
		id := user.ID
		email := openapi_types.Email(user.Email)
		role := genhttp.UserRole(user.Role)
		createdAt := user.CreatedAt
		updatedAt := user.UpdatedAt

		userList[i] = genhttp.User{
			Id:        &id,
			Email:     &email,
			Name:      &user.Name,
			Role:      &role,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		}
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit

	// Create response
	code := "SUCCESS"
	message := "Users retrieved successfully"
	now := time.Now()
	currentPage := page
	perPage := limit
	totalItems := total
	totalPagesVal := totalPages

	// Build data struct for response
	meta := struct {
		CurrentPage *int `json:"current_page,omitempty"`
		PerPage     *int `json:"per_page,omitempty"`
		Total       *int `json:"total,omitempty"`
		TotalPages  *int `json:"total_pages,omitempty"`
	}{
		CurrentPage: &currentPage,
		PerPage:     &perPage,
		Total:       &totalItems,
		TotalPages:  &totalPagesVal,
	}

	data := struct {
		Meta *struct {
			CurrentPage *int `json:"current_page,omitempty"`
			PerPage     *int `json:"per_page,omitempty"`
			Total       *int `json:"total,omitempty"`
			TotalPages  *int `json:"total_pages,omitempty"`
		} `json:"meta,omitempty"`
		Users *[]genhttp.User `json:"users,omitempty"`
	}{
		Meta:  &meta,
		Users: &userList,
	}

	response := genhttp.UserListResponse{
		Code:       &code,
		Message:    &message,
		Data:       &data,
		ServerTime: &now,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetUser handles GET /users/{id} requests
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	// Convert openapi_types.UUID to uuid.UUID (which should be the same type)
	userID := id

	// Call use case
	user, err := h.userUseCase.GetUserByID(ctx, userID)
	if err != nil {
		switch err {
		case errs.ErrUserNotFound:
			handleError(w, formatter.NewHTTPError(http.StatusNotFound, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	// Create response
	code := "SUCCESS"
	message := "User retrieved successfully"
	userId := user.ID
	email := openapi_types.Email(user.Email)
	role := genhttp.UserRole(user.Role)
	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt
	now := time.Now()

	response := genhttp.UserResponse{
		Code:    &code,
		Message: &message,
		Data: &genhttp.User{
			Id:        &userId,
			Email:     &email,
			Name:      &user.Name,
			Role:      &role,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		},
		ServerTime: &now,
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdateUser handles PATCH /users/{id} requests
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	// Convert openapi_types.UUID to uuid.UUID (which should be the same type)
	userID := id

	var reqBody genhttp.UpdateUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Map request to params
	updateParams := params.UpdateUserParams{}
	if reqBody.Name != nil {
		updateParams.Name = reqBody.Name
	}

	// Call use case
	user, err := h.userUseCase.UpdateUser(ctx, userID, updateParams)
	if err != nil {
		switch err {
		case errs.ErrUserNotFound:
			handleError(w, formatter.NewHTTPError(http.StatusNotFound, err.Error()))
		case errs.ErrEmailAlreadyExists:
			handleError(w, formatter.NewHTTPError(http.StatusConflict, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	// Create response
	code := "SUCCESS"
	message := "User updated successfully"
	userId := user.ID
	email := openapi_types.Email(user.Email)
	role := genhttp.UserRole(user.Role)
	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt
	now := time.Now()

	response := genhttp.UserResponse{
		Code:    &code,
		Message: &message,
		Data: &genhttp.User{
			Id:        &userId,
			Email:     &email,
			Name:      &user.Name,
			Role:      &role,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		},
		ServerTime: &now,
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteUser handles DELETE /users/{id} requests
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	// Convert openapi_types.UUID to uuid.UUID (which should be the same type)
	userID := id

	// Call use case
	if err := h.userUseCase.DeleteUser(ctx, userID); err != nil {
		switch err {
		case errs.ErrUserNotFound:
			handleError(w, formatter.NewHTTPError(http.StatusNotFound, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePassword handles PUT /users/{id}/password requests
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	// Convert openapi_types.UUID to uuid.UUID (which should be the same type)
	userID := id

	var reqBody genhttp.UpdatePasswordJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		handleError(w, formatter.NewHTTPError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Map request to params
	updateParams := params.UpdatePasswordParams{
		CurrentPassword: reqBody.CurrentPassword,
		NewPassword:     reqBody.NewPassword,
	}

	// Call use case
	if err := h.userUseCase.UpdatePassword(ctx, userID, updateParams); err != nil {
		switch err {
		case errs.ErrUserNotFound:
			handleError(w, formatter.NewHTTPError(http.StatusNotFound, err.Error()))
		case errs.ErrInvalidCurrentPassword:
			handleError(w, formatter.NewHTTPError(http.StatusUnauthorized, err.Error()))
		default:
			handleError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

// handleError handles API errors
func handleError(w http.ResponseWriter, err error) {
	// Check if it's an HTTP error from formatter package
	if httpErr, ok := err.(*formatter.HTTPError); ok {
		code := "ERROR"
		now := time.Now()

		response := formatter.StandardResponse{
			Code:       code,
			Message:    httpErr.Message,
			ServerTime: now,
			Data:       nil,
		}

		respondJSON(w, httpErr.StatusCode, response)
		return
	}

	// Default to internal server error
	code := "ERROR"
	message := "Internal server error"
	now := time.Now()

	response := formatter.StandardResponse{
		Code:       code,
		Message:    message,
		ServerTime: now,
		Data:       nil,
	}

	if err != nil {
		response.Message = err.Error()
	}

	respondJSON(w, http.StatusInternalServerError, response)
}

// respondJSON responds with JSON
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	formatter.JSON(w, status, payload)
}
