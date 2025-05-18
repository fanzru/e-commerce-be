package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PromotionType defines the type of promotion
type PromotionType string

const (
	// BuyOneGetOneFree is a promotion where buying one product gets another free
	BuyOneGetOneFree PromotionType = "BUY_ONE_GET_ONE_FREE"
	// Buy3Pay2 is a promotion where buying 3 items pays for only 2
	Buy3Pay2 PromotionType = "BUY_3_PAY_2"
	// BulkDiscount is a promotion with a percentage discount for buying in bulk
	BulkDiscount PromotionType = "BULK_DISCOUNT"
)

// Promotion is the base promotion entity
type Promotion struct {
	ID          uuid.UUID       `json:"id"`
	Type        PromotionType   `json:"type"`
	Description string          `json:"description"`
	Rule        json.RawMessage `json:"rule,omitempty"`
	Active      bool            `json:"active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

// PromotionRule is the interface that all promotion rules must implement
type PromotionRule interface {
	// Apply applies the promotion to the cart items and returns the discount amount
	Apply(items []CartItem) float64
}

// CartItem is a simplified representation of a cart item used for promotion rules
type CartItem struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductSKU  string    `json:"product_sku"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
}

// BuyOneGetOneFreePromotion represents a promotion where buying one product gets another free
type BuyOneGetOneFreePromotion struct {
	Promotion
	TriggerSKU      string `json:"trigger_sku"`
	FreeSKU         string `json:"free_sku"`
	TriggerQuantity int    `json:"trigger_quantity"`
	FreeQuantity    int    `json:"free_quantity"`
}

// Apply implements the PromotionRule interface for BuyOneGetOneFreePromotion
func (p *BuyOneGetOneFreePromotion) Apply(items []CartItem) float64 {
	if !p.Active {
		return 0
	}

	var triggerItem, freeItem *CartItem
	for i := range items {
		if items[i].ProductSKU == p.TriggerSKU {
			triggerItem = &items[i]
		} else if items[i].ProductSKU == p.FreeSKU {
			freeItem = &items[i]
		}
	}

	if triggerItem == nil || freeItem == nil {
		return 0
	}

	// Calculate how many free items can be given
	triggerCount := triggerItem.Quantity / p.TriggerQuantity
	freeCount := triggerCount * p.FreeQuantity

	if freeCount > freeItem.Quantity {
		freeCount = freeItem.Quantity
	}

	discount := float64(freeCount) * freeItem.UnitPrice
	return discount
}

// Buy3Pay2Promotion represents a promotion where buying 3 items pays for only 2
type Buy3Pay2Promotion struct {
	Promotion
	SKU                 string `json:"sku"`
	MinQuantity         int    `json:"min_quantity"`
	PaidQuantityDivisor int    `json:"paid_quantity_divisor"`
	FreeQuantityDivisor int    `json:"free_quantity_divisor"`
}

// Apply implements the PromotionRule interface for Buy3Pay2Promotion
func (p *Buy3Pay2Promotion) Apply(items []CartItem) float64 {
	if !p.Active {
		return 0
	}

	var targetItem *CartItem
	for i := range items {
		if items[i].ProductSKU == p.SKU {
			targetItem = &items[i]
			break
		}
	}

	if targetItem == nil || targetItem.Quantity < p.MinQuantity {
		return 0
	}

	totalItems := targetItem.Quantity
	divisor := p.PaidQuantityDivisor + p.FreeQuantityDivisor
	sets := totalItems / divisor

	// Calculate how many items are free
	freeItems := sets * p.FreeQuantityDivisor

	discount := float64(freeItems) * targetItem.UnitPrice
	return discount
}

// BulkDiscountPromotion represents a promotion with a percentage discount for buying in bulk
type BulkDiscountPromotion struct {
	Promotion
	SKU                string  `json:"sku"`
	MinQuantity        int     `json:"min_quantity"`
	DiscountPercentage float64 `json:"discount_percentage"`
}

// Apply implements the PromotionRule interface for BulkDiscountPromotion
func (p *BulkDiscountPromotion) Apply(items []CartItem) float64 {
	if !p.Active {
		return 0
	}

	var targetItem *CartItem
	for i := range items {
		if items[i].ProductSKU == p.SKU {
			targetItem = &items[i]
			break
		}
	}

	if targetItem == nil || targetItem.Quantity < p.MinQuantity {
		return 0
	}

	totalPrice := targetItem.UnitPrice * float64(targetItem.Quantity)
	discount := totalPrice * (p.DiscountPercentage / 100)
	return discount
}

// NewPromotion creates a new promotion
func NewPromotion(promotionType PromotionType, description string, rule PromotionRule) (*Promotion, error) {
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	promotion := &Promotion{
		ID:          uuid.New(),
		Type:        promotionType,
		Description: description,
		Rule:        ruleJSON,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	return promotion, nil
}

// ParseRule parses the rule from JSON and returns the appropriate rule instance
func (p *Promotion) ParseRule() (PromotionRule, error) {
	var ruleInstance PromotionRule
	var err error

	switch p.Type {
	case BuyOneGetOneFree:
		var rule BuyOneGetOneFreePromotion
		err = json.Unmarshal(p.Rule, &rule)
		if err == nil {
			ruleInstance = &rule
		}
	case Buy3Pay2:
		var rule Buy3Pay2Promotion
		err = json.Unmarshal(p.Rule, &rule)
		if err == nil {
			ruleInstance = &rule
		}
	case BulkDiscount:
		var rule BulkDiscountPromotion
		err = json.Unmarshal(p.Rule, &rule)
		if err == nil {
			ruleInstance = &rule
		}
	default:
		return nil, nil
	}

	return ruleInstance, err
}

// ApplyToCart applies the promotion to a cart and returns the discount
func (p *Promotion) ApplyToCart(items []CartItem) (float64, error) {
	if !p.Active {
		return 0, nil
	}

	rule, err := p.ParseRule()
	if err != nil {
		return 0, err
	}

	if rule == nil {
		return 0, nil
	}

	return rule.Apply(items), nil
}
