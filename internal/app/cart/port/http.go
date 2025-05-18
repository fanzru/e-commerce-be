package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/cart/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/cart/usecase"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/fanzru/e-commerce-be/pkg/errors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// HTTPHandler defines the interface for cart HTTP handlers
type HTTPHandler interface {
	// CreateCart handles the POST /carts endpoint
	CreateCart(w http.ResponseWriter, r *http.Request)

	// GetCart handles the GET /carts/:id endpoint
	GetCart(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)

	// DeleteCart handles the DELETE /carts/:id endpoint
	DeleteCart(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)

	// AddItemToCart handles the POST /carts/:id/items endpoint
	AddItemToCart(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)

	// UpdateCartItem handles the PUT /carts/:cartId/items/:itemId endpoint
	UpdateCartItem(w http.ResponseWriter, r *http.Request, cartId openapi_types.UUID, itemId openapi_types.UUID)

	// RemoveCartItem handles the DELETE /carts/:cartId/items/:itemId endpoint
	RemoveCartItem(w http.ResponseWriter, r *http.Request, cartId openapi_types.UUID, itemId openapi_types.UUID)

	// GetCurrentUserCart handles the GET /carts/me endpoint
	GetCurrentUserCart(w http.ResponseWriter, r *http.Request)

	// AddItemToCurrentUserCart handles the POST /carts/me endpoint
	AddItemToCurrentUserCart(w http.ResponseWriter, r *http.Request)
}

// CartHandler handles HTTP requests for carts
type CartHandler struct {
	cartUseCase usecase.CartUseCase
}

// NewCartHandler creates a new cart HTTP handler
func NewCartHandler(cartUseCase usecase.CartUseCase) *CartHandler {
	return &CartHandler{
		cartUseCase: cartUseCase,
	}
}

// NewHTTPServer creates a new HTTP server for carts
func NewHTTPServer(cartUseCase usecase.CartUseCase) http.Handler {
	handler := NewCartHandler(cartUseCase)
	return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			handleError(w, err)
		},
	})
}

// CreateCart handles POST /api/v1/carts requests
func (h *CartHandler) CreateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT token in context
	userClaims, ok := ctx.Value(middleware.ContextTokenClaimsKey).(*params.TokenClaims)
	if !ok || userClaims == nil {
		// Create anonymous cart if no user is authenticated
		cart, err := h.cartUseCase.Create(ctx)
		if err != nil {
			handleError(w, err)
			return
		}

		// Map to response
		response := mapCartToResponse(cart, "Cart created successfully")
		respondJSON(w, http.StatusCreated, response)
		return
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid user ID"))
		return
	}

	// Create cart for the authenticated user
	cart, err := h.cartUseCase.CreateForUser(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	response := mapCartToResponse(cart, "Cart created successfully")
	respondJSON(w, http.StatusCreated, response)
}

// GetCart handles GET /api/v1/carts/{id} requests
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	cartID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid cart ID"))
		return
	}

	cart, err := h.cartUseCase.GetByID(ctx, cartID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	response := mapCartToResponse(cart, "Cart retrieved successfully")
	respondJSON(w, http.StatusOK, response)
}

// DeleteCart handles DELETE /api/v1/carts/{id} requests
func (h *CartHandler) DeleteCart(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	cartID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid cart ID"))
		return
	}

	err = h.cartUseCase.Delete(ctx, cartID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return a standard success response
	response := genhttp.StandardResponse{
		Code:       "success",
		Data:       map[string]interface{}{},
		Message:    "Cart deleted successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// AddItemToCart handles POST /carts/{id}/items requests
func (h *CartHandler) AddItemToCart(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	cartID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid cart ID"))
		return
	}

	var params genhttp.AddItemToCartJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	productID, err := uuid.Parse(params.ProductId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	cartItem, err := h.cartUseCase.AddItem(ctx, cartID, productID, params.Quantity)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, mapCartItemToResponse(cartItem, "Item added to cart successfully"))
}

// UpdateCartItem handles PUT /carts/{cartId}/items/{itemId} requests
func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request, cartId openapi_types.UUID, itemId openapi_types.UUID) {
	ctx := r.Context()

	cartID, err := uuid.Parse(cartId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid cart ID"))
		return
	}

	itemID, err := uuid.Parse(itemId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid item ID"))
		return
	}

	var params genhttp.UpdateCartItemJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	err = h.cartUseCase.UpdateItemQuantity(ctx, cartID, itemID, params.Quantity)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveCartItem handles DELETE /carts/{cartId}/items/{itemId} requests
func (h *CartHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request, cartId openapi_types.UUID, itemId openapi_types.UUID) {
	ctx := r.Context()

	cartID, err := uuid.Parse(cartId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid cart ID"))
		return
	}

	itemID, err := uuid.Parse(itemId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid item ID"))
		return
	}

	err = h.cartUseCase.RemoveItem(ctx, cartID, itemID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCurrentUserCart handles GET /api/v1/carts/me requests
func (h *CartHandler) GetCurrentUserCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT token in context
	userClaims, ok := ctx.Value(middleware.ContextTokenClaimsKey).(*params.TokenClaims)
	if !ok || userClaims == nil {
		handleError(w, errors.NewUnauthorized("authentication required"))
		return
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid user ID"))
		return
	}

	// Get or create cart for the user
	cart, err := h.cartUseCase.GetByUserID(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	response := mapCartToResponse(cart, "Cart retrieved successfully")
	respondJSON(w, http.StatusOK, response)
}

// AddItemToCurrentUserCart handles POST /api/v1/carts/me requests
func (h *CartHandler) AddItemToCurrentUserCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT token in context
	userClaims, ok := ctx.Value(middleware.ContextTokenClaimsKey).(*params.TokenClaims)
	if !ok || userClaims == nil {
		handleError(w, errors.NewUnauthorized("authentication required"))
		return
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid user ID"))
		return
	}

	// Parse request body
	var params genhttp.AddItemToCartJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	productID, err := uuid.Parse(params.ProductId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	// Add item to the user's cart
	cartItem, err := h.cartUseCase.AddItemToUserCart(ctx, userID, productID, params.Quantity)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, mapCartItemToResponse(cartItem, "Item added to cart successfully"))
}

// Helper functions

// mapCartToResponse maps a cart entity to a cart response
func mapCartToResponse(cart *entity.Cart, message string) genhttp.CartResponse {
	// Convert cart entity to response format
	cartData := genhttp.Cart{}

	// Set cart ID
	cartId := openapi_types.UUID(cart.ID)
	cartData.Id = &cartId

	// Set cart timestamps
	createdAt := cart.CreatedAt
	updatedAt := cart.UpdatedAt
	cartData.CreatedAt = &createdAt
	cartData.UpdatedAt = &updatedAt

	// Calculate subtotal and total items
	var subtotal float32
	totalItems := 0

	// Convert cart items
	if len(cart.Items) > 0 {
		items := make([]genhttp.CartItem, len(cart.Items))
		for i, item := range cart.Items {
			items[i] = convertCartItemToGenHTTP(item)

			// Add to subtotal and total items count
			if item.UnitPrice > 0 && item.Quantity > 0 {
				itemSubtotal := float32(item.UnitPrice) * float32(item.Quantity)
				subtotal += itemSubtotal
				totalItems += item.Quantity
			}
		}
		cartData.Items = &items
	}

	cartData.Subtotal = &subtotal
	cartData.TotalItems = &totalItems

	// Create the response
	return genhttp.CartResponse{
		Code:       "success",
		Data:       cartData,
		Message:    message,
		ServerTime: time.Now(),
	}
}

// mapCartItemToResponse maps a cart item entity to a cart item response
func mapCartItemToResponse(item *entity.CartItem, message string) genhttp.CartItemResponse {
	itemData := convertCartItemToGenHTTP(item)
	return genhttp.CartItemResponse{
		Code:       "success",
		Data:       itemData,
		Message:    message,
		ServerTime: time.Now(),
	}
}

// convertCartItemToGenHTTP converts a cart item entity to a genhttp cart item
func convertCartItemToGenHTTP(item *entity.CartItem) genhttp.CartItem {
	itemId := openapi_types.UUID(item.ID)
	productId := openapi_types.UUID(item.ProductID)
	cartId := openapi_types.UUID(item.CartID)
	unitPrice := float32(item.UnitPrice)
	quantity := item.Quantity
	createdAt := item.CreatedAt
	updatedAt := item.UpdatedAt

	cartItem := genhttp.CartItem{
		Id:          &itemId,
		CartId:      &cartId,
		ProductId:   &productId,
		ProductName: &item.ProductName,
		ProductSku:  &item.ProductSKU,
		Quantity:    &quantity,
		UnitPrice:   &unitPrice,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	subtotal := float32(item.UnitPrice) * float32(item.Quantity)
	cartItem.Subtotal = &subtotal

	return cartItem
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleError handles an error and sends an appropriate response
func handleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_server_error"
	message := "An unexpected error occurred"

	// Check for domain errors
	if appErr, ok := err.(*errors.AppError); ok {
		status = appErr.Status
		message = appErr.Message

		// Set error code based on HTTP status
		switch status {
		case http.StatusNotFound:
			code = "not_found"
		case http.StatusBadRequest:
			code = "bad_request"
		case http.StatusConflict:
			code = "conflict"
		}
	}

	// Send error response
	errorResponse := genhttp.ErrorResponse{
		Code:       code,
		Message:    message,
		ServerTime: time.Now(),
	}

	respondJSON(w, status, errorResponse)
}
