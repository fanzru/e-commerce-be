package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/checkout/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/checkout/usecase"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	appmiddleware "github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/fanzru/e-commerce-be/pkg/errors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// HTTPHandler defines the interface for checkout HTTP handlers
type HTTPHandler interface {
	// GetCheckout handles the GET /checkouts/:id endpoint
	GetCheckout(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)

	// ListCheckouts handles the GET /checkouts endpoint
	ListCheckouts(w http.ResponseWriter, r *http.Request)

	// ProcessCart handles the POST /checkouts endpoint
	ProcessCart(w http.ResponseWriter, r *http.Request)
}

// CheckoutHandler handles HTTP requests for checkouts
type CheckoutHandler struct {
	checkoutUseCase usecase.CheckoutUseCase
}

// NewCheckoutHandler creates a new checkout HTTP handler
func NewCheckoutHandler(checkoutUseCase usecase.CheckoutUseCase) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutUseCase: checkoutUseCase,
	}
}

// NewHTTPServer creates a new HTTP server for checkouts
func NewHTTPServer(checkoutUseCase usecase.CheckoutUseCase) http.Handler {
	handler := NewCheckoutHandler(checkoutUseCase)
	return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			handleError(w, err)
		},
	})
}

// GetApiV1CheckoutsId handles GET /v1/checkouts/{id} requests
func (h *CheckoutHandler) GetApiV1CheckoutsId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	checkoutID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid checkout ID"))
		return
	}

	checkout, err := h.checkoutUseCase.GetByID(ctx, checkoutID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapCheckoutToResponse(checkout))
}

// GetApiV1Checkouts handles GET /v1/checkouts requests
func (h *CheckoutHandler) GetApiV1Checkouts(w http.ResponseWriter, r *http.Request, params genhttp.GetApiV1CheckoutsParams) {
	ctx := r.Context()

	// Set default values if not provided
	page := 1
	limit := 10

	if params.Page != nil {
		page = *params.Page
	}

	if params.Limit != nil {
		limit = *params.Limit
	}

	// Call the use case
	checkouts, total, err := h.checkoutUseCase.ListCheckouts(ctx, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	checkoutSummaries := make([]genhttp.CheckoutSummary, len(checkouts))
	for i, checkout := range checkouts {
		subtotal := float32(checkout.Subtotal)
		totalDiscount := float32(checkout.TotalDiscount)
		total := float32(checkout.Total)

		// Convert payment status and order status
		paymentStatus := genhttp.CheckoutSummaryPaymentStatus(checkout.PaymentStatus)
		status := genhttp.CheckoutSummaryStatus(checkout.Status)

		checkoutSummaries[i] = genhttp.CheckoutSummary{
			Id:            &checkout.ID,
			UserId:        checkout.UserID,
			PaymentStatus: &paymentStatus,
			Status:        &status,
			Subtotal:      &subtotal,
			TotalDiscount: &totalDiscount,
			Total:         &total,
			CreatedAt:     &checkout.CreatedAt,
		}
	}

	meta := genhttp.PaginationMeta{
		CurrentPage: &page,
		PerPage:     &limit,
		Total:       &total,
	}

	totalPages := (total + limit - 1) / limit
	meta.TotalPages = &totalPages

	response := genhttp.CheckoutListResponse{
		Code:    "success",
		Message: "Checkout list retrieved successfully",
		Data: struct {
			Checkouts *[]genhttp.CheckoutSummary `json:"checkouts,omitempty"`
			Meta      *genhttp.PaginationMeta    `json:"meta,omitempty"`
		}{
			Checkouts: &checkoutSummaries,
			Meta:      &meta,
		},
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// PostApiV1Checkouts handles POST /v1/checkouts requests
func (h *CheckoutHandler) PostApiV1Checkouts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Get user ID from token claims in context
	claims, ok := ctx.Value("token_claims").(*params.TokenClaims)
	if !ok {
		handleError(w, errors.NewUnauthorized("unauthorized: missing token claims"))
		return
	}
	userIDStr := claims.UserID

	// Parse user ID from string to UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid user ID"))
		return
	}

	// Process cart checkout
	checkout, err := h.checkoutUseCase.ProcessCart(ctx, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Log that we're clearing the cart (it's already cleared in ProcessCart)
	logger := appmiddleware.Logger.With(
		"method", "CheckoutHandler.PostApiV1Checkouts",
		"user_id", userID.String(),
	)
	logger.Info("User cart cleared after checkout")

	respondJSON(w, http.StatusCreated, mapCheckoutToResponse(checkout))
}

// GetApiV1UsersUserIdOrders handles GET /api/v1/users/{user_id}/orders requests
func (h *CheckoutHandler) GetApiV1UsersUserIdOrders(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID, params genhttp.GetApiV1UsersUserIdOrdersParams) {
	ctx := r.Context()

	// Convert UUID
	userID, err := uuid.Parse(userId.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid user ID"))
		return
	}

	// Set default values if not provided
	page := 1
	limit := 10

	if params.Page != nil {
		page = *params.Page
	}

	if params.Limit != nil {
		limit = *params.Limit
	}

	// Call the use case
	orders, total, err := h.checkoutUseCase.GetUserOrders(ctx, userID, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	orderSummaries := make([]genhttp.OrderSummary, len(orders))
	for i, order := range orders {
		subtotal := float32(order.Subtotal)
		totalDiscount := float32(order.TotalDiscount)
		total := float32(order.Total)
		itemCount := len(order.Items)

		// Convert payment status and order status
		paymentStatus := genhttp.OrderSummaryPaymentStatus(order.PaymentStatus)
		status := genhttp.OrderSummaryStatus(order.Status)

		orderSummaries[i] = genhttp.OrderSummary{
			Id:            &order.ID,
			PaymentStatus: &paymentStatus,
			Status:        &status,
			Subtotal:      &subtotal,
			TotalDiscount: &totalDiscount,
			Total:         &total,
			ItemCount:     &itemCount,
			CreatedAt:     &order.CreatedAt,
			CompletedAt:   order.CompletedAt,
		}
	}

	meta := genhttp.PaginationMeta{
		CurrentPage: &page,
		PerPage:     &limit,
		Total:       &total,
	}

	totalPages := (total + limit - 1) / limit
	meta.TotalPages = &totalPages

	response := genhttp.OrderListResponse{
		Code:    "success",
		Message: "User orders retrieved successfully",
		Data: struct {
			Meta   *genhttp.PaginationMeta `json:"meta,omitempty"`
			Orders *[]genhttp.OrderSummary `json:"orders,omitempty"`
		}{
			Orders: &orderSummaries,
			Meta:   &meta,
		},
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// PutApiV1CheckoutsIdPayment handles PUT /api/v1/checkouts/{id}/payment requests
func (h *CheckoutHandler) PutApiV1CheckoutsIdPayment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	// Parse checkout ID
	checkoutID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid checkout ID"))
		return
	}

	// Parse request body
	var requestBody genhttp.PaymentStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	// Validate status
	paymentStatus := entity.PaymentStatus(requestBody.Status)
	if !isValidPaymentStatus(paymentStatus) {
		handleError(w, errors.NewBadRequest("invalid payment status"))
		return
	}

	// Get payment method and reference
	paymentMethod := ""
	paymentReference := ""
	if requestBody.PaymentMethod != nil {
		paymentMethod = *requestBody.PaymentMethod
	}
	if requestBody.PaymentReference != nil {
		paymentReference = *requestBody.PaymentReference
	}

	// Update payment status
	err = h.checkoutUseCase.UpdatePaymentStatus(ctx, checkoutID, paymentStatus, paymentMethod, paymentReference)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return success response
	response := genhttp.SuccessResponse{
		Code:       "success",
		Message:    "Payment status updated successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// PutApiV1CheckoutsIdStatus handles PUT /api/v1/checkouts/{id}/status requests
func (h *CheckoutHandler) PutApiV1CheckoutsIdStatus(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	// Parse checkout ID
	checkoutID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid checkout ID"))
		return
	}

	// Parse request body
	var requestBody genhttp.OrderStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	// Validate status
	orderStatus := entity.OrderStatus(requestBody.Status)
	if !isValidOrderStatus(orderStatus) {
		handleError(w, errors.NewBadRequest("invalid order status"))
		return
	}

	// Update order status
	err = h.checkoutUseCase.UpdateOrderStatus(ctx, checkoutID, orderStatus)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return success response
	response := genhttp.SuccessResponse{
		Code:       "success",
		Message:    "Order status updated successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// Helper functions

// isValidPaymentStatus checks if a payment status is valid
func isValidPaymentStatus(status entity.PaymentStatus) bool {
	validStatuses := []entity.PaymentStatus{
		entity.PaymentStatusPending,
		entity.PaymentStatusPaid,
		entity.PaymentStatusFailed,
		entity.PaymentStatusRefunded,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}

	return false
}

// isValidOrderStatus checks if an order status is valid
func isValidOrderStatus(status entity.OrderStatus) bool {
	validStatuses := []entity.OrderStatus{
		entity.OrderStatusCreated,
		entity.OrderStatusProcessing,
		entity.OrderStatusShipped,
		entity.OrderStatusDelivered,
		entity.OrderStatusCancelled,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}

	return false
}

// mapCheckoutToResponse maps a checkout entity to a checkout response
func mapCheckoutToResponse(checkout *entity.Checkout) genhttp.CheckoutResponse {
	subtotal := float32(checkout.Subtotal)
	totalDiscount := float32(checkout.TotalDiscount)
	total := float32(checkout.Total)

	// Convert payment status and order status
	paymentStatus := genhttp.CheckoutPaymentStatus(checkout.PaymentStatus)
	status := genhttp.CheckoutStatus(checkout.Status)

	checkoutData := genhttp.Checkout{
		Id:               &checkout.ID,
		UserId:           checkout.UserID,
		PaymentStatus:    &paymentStatus,
		PaymentMethod:    checkout.PaymentMethod,
		PaymentReference: checkout.PaymentReference,
		Notes:            checkout.Notes,
		Status:           &status,
		Subtotal:         &subtotal,
		TotalDiscount:    &totalDiscount,
		Total:            &total,
		CreatedAt:        &checkout.CreatedAt,
		UpdatedAt:        &checkout.UpdatedAt,
		CompletedAt:      checkout.CompletedAt,
	}

	if len(checkout.Items) > 0 {
		items := make([]genhttp.CheckoutItem, len(checkout.Items))
		for i, item := range checkout.Items {
			unitPrice := float32(item.UnitPrice)
			subtotal := float32(item.Subtotal)
			discount := float32(item.Discount)
			total := float32(item.Total)
			quantity := item.Quantity

			items[i] = genhttp.CheckoutItem{
				Id:          &item.ID,
				CheckoutId:  &item.CheckoutID,
				ProductId:   &item.ProductID,
				ProductSku:  &item.ProductSKU,
				ProductName: &item.ProductName,
				Quantity:    &quantity,
				UnitPrice:   &unitPrice,
				Subtotal:    &subtotal,
				Discount:    &discount,
				Total:       &total,
			}
		}
		checkoutData.Items = &items
	}

	if len(checkout.Promotions) > 0 {
		promotions := make([]genhttp.PromotionApplied, len(checkout.Promotions))
		for i, promo := range checkout.Promotions {
			discount := float32(promo.Discount)

			promotions[i] = genhttp.PromotionApplied{
				Id:          &promo.ID,
				CheckoutId:  &promo.CheckoutID,
				PromotionId: &promo.PromotionID,
				Description: &promo.Description,
				Discount:    &discount,
			}
		}
		checkoutData.Promotions = &promotions
	}

	return genhttp.CheckoutResponse{
		Code:       "success",
		Message:    "Checkout retrieved successfully",
		Data:       checkoutData,
		ServerTime: time.Now(),
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}

// handleError handles errors and sends appropriate HTTP responses
func handleError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	case errors.IsNotFound(err):
		status = http.StatusNotFound
		message = err.Error()
	case errors.IsBadRequest(err):
		status = http.StatusBadRequest
		message = err.Error()
	default:
		status = http.StatusInternalServerError
		message = "internal server error"
	}

	errorResponse := genhttp.ErrorResponse{
		Code:       "error",
		Message:    message,
		ServerTime: time.Now(),
	}

	respondJSON(w, status, errorResponse)
}
