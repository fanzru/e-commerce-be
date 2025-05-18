package port

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/cart/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/cart/usecase"
	promotionUseCase "github.com/fanzru/e-commerce-be/internal/app/promotion/usecase"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/fanzru/e-commerce-be/pkg/errors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// HTTPHandler defines the interface for cart HTTP handlers
type HTTPHandler interface {
	// GetCurrentUserCart handles the GET /carts/me endpoint
	GetCurrentUserCart(w http.ResponseWriter, r *http.Request)

	// AddItemToCurrentUserCart handles the POST /carts/me endpoint
	AddItemToCurrentUserCart(w http.ResponseWriter, r *http.Request)

	// UpdateCartItem handles the PUT /carts/me/items/:itemId endpoint
	UpdateCartItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID)

	// RemoveCartItem handles the DELETE /carts/me/items/:itemId endpoint
	RemoveCartItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID)

	// ClearUserCart handles the DELETE /carts/me/clear endpoint
	ClearUserCart(w http.ResponseWriter, r *http.Request)
}

// CartHandler handles HTTP requests for carts
type CartHandler struct {
	cartUseCase      usecase.CartUseCase
	promotionUseCase promotionUseCase.PromotionUseCase
}

// NewCartHandler creates a new cart HTTP handler
func NewCartHandler(cartUseCase usecase.CartUseCase, promotionUseCase promotionUseCase.PromotionUseCase) *CartHandler {
	return &CartHandler{
		cartUseCase:      cartUseCase,
		promotionUseCase: promotionUseCase,
	}
}

// NewHTTPServer creates a new HTTP server for carts
func NewHTTPServer(cartUseCase usecase.CartUseCase, promotionUseCase promotionUseCase.PromotionUseCase) http.Handler {
	handler := NewCartHandler(cartUseCase, promotionUseCase)
	return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			handleError(w, err)
		},
	})
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

	// Get cart info for the user with product details
	cartInfo, err := h.cartUseCase.GetUserCartInfo(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Debug: Log cart info details
	middleware.Logger.Info("Cart details before promotion application",
		"user_id", cartInfo.UserID.String(),
		"item_count", len(cartInfo.Items),
		"subtotal", cartInfo.Subtotal)

	// Log each cart item detail
	for i, item := range cartInfo.Items {
		middleware.Logger.Info(fmt.Sprintf("Cart item %d", i+1),
			"product_id", item.ProductID.String(),
			"product_sku", item.ProductSKU,
			"product_name", item.ProductName,
			"quantity", item.Quantity,
			"unit_price", item.UnitPrice,
			"subtotal", item.Subtotal)
	}

	// Calculate potential promotions
	applicablePromotions, totalDiscount, err := h.promotionUseCase.ApplyPromotions(ctx, cartInfo)
	if err != nil {
		// Log the error but don't fail the request
		middleware.Logger.Error("Failed to calculate promotions", "error", err.Error())
		// Continue without promotions
	}

	// Debug: Log promotion application results
	middleware.Logger.Info("Promotion application results",
		"applicable_promotions_count", len(applicablePromotions),
		"total_discount", totalDiscount)

	// Log each applicable promotion in detail
	for i, promo := range applicablePromotions {
		middleware.Logger.Info(fmt.Sprintf("Applicable promotion %d", i+1),
			"promotion_id", promo.PromotionID.String(),
			"promotion_type", promo.PromotionType,
			"description", promo.Description,
			"discount", promo.Discount)
	}

	// Map to response
	response := mapCartInfoToResponseWithPromotions(cartInfo, applicablePromotions, totalDiscount, "Cart retrieved successfully")
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
	var params struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int       `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	// Add item to the user's cart
	cartItem, err := h.cartUseCase.AddItemToUserCart(ctx, userID, params.ProductID, params.Quantity)
	if err != nil {
		handleError(w, err)
		return
	}

	// Get cart info to get the item with product details
	cartInfo, err := h.cartUseCase.GetUserCartInfo(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Calculate potential promotions
	applicablePromotions, totalDiscount, err := h.promotionUseCase.ApplyPromotions(ctx, cartInfo)
	if err != nil {
		// Log the error but don't fail the request
		middleware.Logger.Error("Failed to calculate promotions", "error", err.Error())
		// Continue without promotions
	}

	// Find the corresponding item in the cart info
	var itemInfo *entity.CartItemInfo
	for _, item := range cartInfo.Items {
		if item.ID == cartItem.ID {
			itemInfo = item
			break
		}
	}

	if itemInfo == nil {
		handleError(w, errors.NewInternalServerError(fmt.Errorf("could not find added item in cart")))
		return
	}

	// Map to response with promotion data
	response := mapCartInfoToResponseWithPromotions(cartInfo, applicablePromotions, totalDiscount, "Item added to cart successfully")
	respondJSON(w, http.StatusCreated, response)
}

// UpdateCartItem handles PUT /carts/me/items/{itemId} requests
func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	ctx := r.Context()

	// Authenticate user
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

	// Validate item ID
	itemID, err := uuid.Parse(itemId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid item ID"))
		return
	}

	var params struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	err = h.cartUseCase.UpdateItemQuantity(ctx, userID, itemID, params.Quantity)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveCartItem handles DELETE /carts/me/items/{itemId} requests
func (h *CartHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request, itemId openapi_types.UUID) {
	ctx := r.Context()

	// Authenticate user
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

	itemID, err := uuid.Parse(itemId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid item ID"))
		return
	}

	err = h.cartUseCase.RemoveItem(ctx, userID, itemID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearUserCart handles DELETE /carts/me/clear requests
func (h *CartHandler) ClearUserCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Authenticate user
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

	err = h.cartUseCase.ClearUserCart(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return a standard success response
	response := genhttp.StandardResponse{
		Code:       "success",
		Data:       map[string]interface{}{},
		Message:    "Cart cleared successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// Helper functions

// mapCartInfoToResponseWithPromotions maps a cart entity to a cart response with promotions
func mapCartInfoToResponseWithPromotions(cartInfo *entity.CartInfo, promotions []promotionUseCase.PromotionDiscount, totalDiscount float64, message string) genhttp.CartResponse {
	// First convert the basic cart info
	cartData := genhttp.Cart{}

	// Set user ID
	userID := openapi_types.UUID(cartInfo.UserID)
	cartData.UserId = &userID

	// Set cart timestamps
	createdAt := cartInfo.CreatedAt
	updatedAt := cartInfo.UpdatedAt
	cartData.CreatedAt = &createdAt
	cartData.UpdatedAt = &updatedAt

	// Prepare the items and calculate totals
	subtotal := float32(cartInfo.Subtotal)
	totalItems := 0

	// Convert cart items
	if len(cartInfo.Items) > 0 {
		items := make([]genhttp.CartItem, len(cartInfo.Items))
		for i, item := range cartInfo.Items {
			items[i] = convertCartItemInfoToGenHTTP(item)
			totalItems += item.Quantity
		}
		cartData.Items = &items
	}

	cartData.Subtotal = &subtotal
	cartData.TotalItems = &totalItems

	// Add promotions if any
	if len(promotions) > 0 {
		applicablePromotions := make([]genhttp.ApplicablePromotion, len(promotions))
		for i, promo := range promotions {
			id := openapi_types.UUID(promo.PromotionID)
			promoType := promo.PromotionType
			description := promo.Description
			discount := float32(promo.Discount)

			applicablePromotions[i] = genhttp.ApplicablePromotion{
				Id:          &id,
				Type:        &promoType,
				Description: &description,
				Discount:    &discount,
			}
		}
		cartData.ApplicablePromotions = &applicablePromotions

		// Calculate potential discount and total
		potentialDiscount := float32(totalDiscount)
		cartData.PotentialDiscount = &potentialDiscount

		potentialTotal := subtotal - potentialDiscount
		if potentialTotal < 0 {
			potentialTotal = 0
		}
		cartData.PotentialTotal = &potentialTotal
	}

	// Create the response
	return genhttp.CartResponse{
		Code:       "success",
		Data:       cartData,
		Message:    message,
		ServerTime: time.Now(),
	}
}

// mapCartItemToResponse maps a cart item entity to a cart item response
func mapCartItemToResponse(item *entity.CartItemInfo, message string) genhttp.CartItemResponse {
	itemData := convertCartItemInfoToGenHTTP(item)
	return genhttp.CartItemResponse{
		Code:       "success",
		Data:       itemData,
		Message:    message,
		ServerTime: time.Now(),
	}
}

// convertCartItemInfoToGenHTTP converts a cart item info entity to a genhttp cart item
func convertCartItemInfoToGenHTTP(item *entity.CartItemInfo) genhttp.CartItem {
	itemId := openapi_types.UUID(item.ID)
	productId := openapi_types.UUID(item.ProductID)
	userId := openapi_types.UUID(item.UserID)
	unitPrice := float32(item.UnitPrice)
	quantity := item.Quantity
	createdAt := item.CreatedAt
	updatedAt := item.UpdatedAt

	cartItem := genhttp.CartItem{
		Id:          &itemId,
		UserId:      &userId,
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
