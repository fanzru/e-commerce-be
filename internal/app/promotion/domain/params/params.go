package params

import (
	"github.com/google/uuid"
)

// PromotionTypeEnum enumerates the available promotion types
type PromotionTypeEnum string

const (
	BuyOneGetOneFree PromotionTypeEnum = "BUY_ONE_GET_ONE_FREE"
	Buy3Pay2         PromotionTypeEnum = "BUY_3_PAY_2"
	BulkDiscount     PromotionTypeEnum = "BULK_DISCOUNT"
)

// CreateBuyOneGetOneFreeParams defines the parameters for creating a buy one get one free promotion
type CreateBuyOneGetOneFreeParams struct {
	Description string `json:"description" binding:"required"`
	TriggerSKU  string `json:"trigger_sku" binding:"required"`
	FreeSKU     string `json:"free_sku" binding:"required"`
	TriggerQty  int    `json:"trigger_quantity" binding:"required,gt=0"`
	FreeQty     int    `json:"free_quantity" binding:"required,gt=0"`
}

// CreateBuy3Pay2Params defines the parameters for creating a buy 3 pay 2 promotion
type CreateBuy3Pay2Params struct {
	Description         string `json:"description" binding:"required"`
	SKU                 string `json:"sku" binding:"required"`
	MinQuantity         int    `json:"min_quantity" binding:"required,gt=0"`
	PaidQuantityDivisor int    `json:"paid_quantity_divisor" binding:"required,gt=0"`
	FreeQuantityDivisor int    `json:"free_quantity_divisor" binding:"required,gt=0"`
}

// CreateBulkDiscountParams defines the parameters for creating a bulk discount promotion
type CreateBulkDiscountParams struct {
	Description        string  `json:"description" binding:"required"`
	SKU                string  `json:"sku" binding:"required"`
	MinQuantity        int     `json:"min_quantity" binding:"required,gt=0"`
	DiscountPercentage float64 `json:"discount_percentage" binding:"required,gt=0,lte=100"`
}

// UpdatePromotionStatusParams defines the parameters for updating a promotion status
type UpdatePromotionStatusParams struct {
	Active bool `json:"active" binding:"required"`
}

// PromotionResponse defines the response structure for a promotion
type PromotionResponse struct {
	ID          uuid.UUID         `json:"id"`
	Type        PromotionTypeEnum `json:"type"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
}

// PromotionListResponse defines the response structure for a list of promotions
type PromotionListResponse struct {
	Promotions []PromotionResponse `json:"promotions"`
	Total      int                 `json:"total"`
}
