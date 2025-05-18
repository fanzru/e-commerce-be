package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/promotion/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/promotion/usecase"
	"github.com/fanzru/e-commerce-be/pkg/errors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// PromotionHandler handles HTTP requests for promotions
type PromotionHandler struct {
	promotionUseCase usecase.PromotionUseCase
}

// NewPromotionHandler creates a new promotion HTTP handler
func NewPromotionHandler(promotionUseCase usecase.PromotionUseCase) *PromotionHandler {
	return &PromotionHandler{
		promotionUseCase: promotionUseCase,
	}
}

// NewHTTPServer creates a new HTTP server for promotions
func NewHTTPServer(promotionUseCase usecase.PromotionUseCase) http.Handler {
	handler := NewPromotionHandler(promotionUseCase)
	return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			handleError(w, err)
		},
	})
}

// GetPromotion handles GET /api/v1/promotions/{id} requests
func (h *PromotionHandler) GetPromotion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	promotionID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid promotion ID"))
		return
	}

	promotion, err := h.promotionUseCase.GetByID(ctx, promotionID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapPromotionToResponse(promotion))
}

// ListPromotions handles GET /api/v1/promotions requests
func (h *PromotionHandler) ListPromotions(w http.ResponseWriter, r *http.Request, params genhttp.ListPromotionsParams) {
	ctx := r.Context()

	// Set default values if not provided
	page := 1
	limit := 10
	var active *bool

	if params.Page != nil {
		page = *params.Page
	}

	if params.Limit != nil {
		limit = *params.Limit
	}

	if params.Active != nil {
		active = params.Active
	}

	// Call the use case
	promotions, total, err := h.promotionUseCase.List(ctx, page, limit, active)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	promotionsData := make([]genhttp.Promotion, len(promotions))
	for i, promo := range promotions {
		promotionType := genhttp.PromotionType(promo.Type)
		promotionsData[i] = genhttp.Promotion{
			Id:          &promo.ID,
			Type:        &promotionType,
			Description: &promo.Description,
			Active:      &promo.Active,
			CreatedAt:   &promo.CreatedAt,
			UpdatedAt:   &promo.UpdatedAt,
		}
	}

	meta := genhttp.PaginationMeta{
		CurrentPage: &page,
		PerPage:     &limit,
		Total:       &total,
	}

	totalPages := (total + limit - 1) / limit
	meta.TotalPages = &totalPages

	response := genhttp.PromotionListResponse{
		Code:    "success",
		Message: "Promotion list retrieved successfully",
		Data: struct {
			Meta       *genhttp.PaginationMeta `json:"meta,omitempty"`
			Promotions *[]genhttp.Promotion    `json:"promotions,omitempty"`
		}{
			Promotions: &promotionsData,
			Meta:       &meta,
		},
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// CreatePromotion handles POST /api/v1/promotions requests
func (h *PromotionHandler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	promotionType, ok := requestBody["type"].(string)
	if !ok {
		handleError(w, errors.NewBadRequest("missing promotion type"))
		return
	}

	description, ok := requestBody["description"].(string)
	if !ok {
		handleError(w, errors.NewBadRequest("missing description"))
		return
	}

	var promotion *entity.Promotion
	var err error

	// Create the appropriate promotion type based on the request
	switch promotionType {
	case string(genhttp.PromotionTypeBUYONEGETONEFREE):
		triggerSku, ok := requestBody["trigger_sku"].(string)
		if !ok {
			handleError(w, errors.NewBadRequest("missing trigger_sku for buy one get one free promotion"))
			return
		}
		freeSku, ok := requestBody["free_sku"].(string)
		if !ok {
			handleError(w, errors.NewBadRequest("missing free_sku for buy one get one free promotion"))
			return
		}

		// Attempt to get integer values with default fallbacks
		triggerQuantity := 1
		if tq, ok := requestBody["trigger_quantity"].(float64); ok {
			triggerQuantity = int(tq)
		}

		freeQuantity := 1
		if fq, ok := requestBody["free_quantity"].(float64); ok {
			freeQuantity = int(fq)
		}

		active := true
		if a, ok := requestBody["active"].(bool); ok {
			active = a
		}

		promotion, err = h.promotionUseCase.CreateBuyOneGetOneFree(ctx, description, triggerSku, freeSku, triggerQuantity, freeQuantity, active)

	case string(genhttp.PromotionTypeBUY3PAY2):
		sku, ok := requestBody["sku"].(string)
		if !ok {
			handleError(w, errors.NewBadRequest("missing sku for 3 for 2 promotion"))
			return
		}

		minQuantity := 3
		if mq, ok := requestBody["min_quantity"].(float64); ok {
			minQuantity = int(mq)
		}

		paidQuantityDivisor := 2
		if pqd, ok := requestBody["paid_quantity_divisor"].(float64); ok {
			paidQuantityDivisor = int(pqd)
		}

		freeQuantityDivisor := 1
		if fqd, ok := requestBody["free_quantity_divisor"].(float64); ok {
			freeQuantityDivisor = int(fqd)
		}

		active := true
		if a, ok := requestBody["active"].(bool); ok {
			active = a
		}

		promotion, err = h.promotionUseCase.CreateBuy3Pay2(ctx, description, sku, minQuantity, paidQuantityDivisor, freeQuantityDivisor, active)

	case string(genhttp.PromotionTypeBULKDISCOUNT):
		sku, ok := requestBody["sku"].(string)
		if !ok {
			handleError(w, errors.NewBadRequest("missing sku for bulk discount promotion"))
			return
		}

		minQuantity := 3
		if mq, ok := requestBody["min_quantity"].(float64); ok {
			minQuantity = int(mq)
		}

		discountPercentage := 10.0
		if dp, ok := requestBody["discount_percentage"].(float64); ok {
			discountPercentage = dp
		}

		active := true
		if a, ok := requestBody["active"].(bool); ok {
			active = a
		}

		promotion, err = h.promotionUseCase.CreateBulkDiscount(ctx, description, sku, minQuantity, discountPercentage, active)

	default:
		handleError(w, errors.NewBadRequest("invalid promotion type"))
		return
	}

	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, mapPromotionToResponse(promotion))
}

// UpdatePromotionStatus handles PATCH /api/v1/promotions/{id} requests
func (h *PromotionHandler) UpdatePromotionStatus(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	promotionID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid promotion ID"))
		return
	}

	var requestBody genhttp.UpdatePromotionStatusJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	err = h.promotionUseCase.UpdateStatus(ctx, promotionID, requestBody.Active)
	if err != nil {
		handleError(w, err)
		return
	}

	// Get updated promotion to return in response
	promotion, err := h.promotionUseCase.GetByID(ctx, promotionID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapPromotionToResponse(promotion))
}

// DeletePromotion handles DELETE /api/v1/promotions/{id} requests
func (h *PromotionHandler) DeletePromotion(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	promotionID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid promotion ID"))
		return
	}

	err = h.promotionUseCase.Delete(ctx, promotionID)
	if err != nil {
		handleError(w, err)
		return
	}

	response := genhttp.StandardResponse{
		Code:       "success",
		Message:    "Promotion deleted successfully",
		ServerTime: time.Now(),
		Data:       map[string]interface{}{},
	}

	respondJSON(w, http.StatusNoContent, response)
}

// Helper functions

// mapPromotionToResponse maps a promotion entity to a promotion response
func mapPromotionToResponse(promotion *entity.Promotion) genhttp.PromotionResponse {
	promotionType := genhttp.PromotionType(promotion.Type)
	promotionData := genhttp.Promotion{
		Id:          &promotion.ID,
		Type:        &promotionType,
		Description: &promotion.Description,
		Active:      &promotion.Active,
		CreatedAt:   &promotion.CreatedAt,
		UpdatedAt:   &promotion.UpdatedAt,
	}

	return genhttp.PromotionResponse{
		Code:       "success",
		Message:    "Promotion retrieved successfully",
		Data:       promotionData,
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
