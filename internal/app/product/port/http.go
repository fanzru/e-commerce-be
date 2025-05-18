package port

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/product/port/genhttp"
	"github.com/fanzru/e-commerce-be/internal/app/product/usecase"
	"github.com/fanzru/e-commerce-be/pkg/errors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productUseCase usecase.ProductUseCase
}

// NewProductHandler creates a new product HTTP handler
func NewProductHandler(productUseCase usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
	}
}

// NewHTTPServer creates a new HTTP server for products
func NewHTTPServer(productUseCase usecase.ProductUseCase) http.Handler {
	handler := NewProductHandler(productUseCase)
	return genhttp.HandlerWithOptions(handler, genhttp.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			handleError(w, err)
		},
	})
}

// ListProducts handles GET /products requests
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request, params genhttp.ListProductsParams) {
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

	// Get filter values if provided
	var sku, name string
	if params.Sku != nil {
		sku = *params.Sku
	}
	if params.Name != nil {
		name = *params.Name
	}

	// Call the use case
	products, total, err := h.productUseCase.List(ctx, page, limit, sku, name)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	productsData := make([]struct {
		Id        *openapi_types.UUID `json:"id,omitempty"`
		Inventory *int                `json:"inventory,omitempty"`
		Name      *string             `json:"name,omitempty"`
		Price     *float32            `json:"price,omitempty"`
		Sku       *string             `json:"sku,omitempty"`
	}, len(products))

	for i, product := range products {
		id := openapi_types.UUID(product.ID)
		price := float32(product.Price)
		productsData[i] = struct {
			Id        *openapi_types.UUID `json:"id,omitempty"`
			Inventory *int                `json:"inventory,omitempty"`
			Name      *string             `json:"name,omitempty"`
			Price     *float32            `json:"price,omitempty"`
			Sku       *string             `json:"sku,omitempty"`
		}{
			Id:        &id,
			Sku:       &product.SKU,
			Name:      &product.Name,
			Price:     &price,
			Inventory: &product.Inventory,
		}
	}

	response := genhttp.ProductListResponse{
		Code: "success",
		Data: struct {
			Products *[]struct {
				Id        *openapi_types.UUID `json:"id,omitempty"`
				Inventory *int                `json:"inventory,omitempty"`
				Name      *string             `json:"name,omitempty"`
				Price     *float32            `json:"price,omitempty"`
				Sku       *string             `json:"sku,omitempty"`
			} `json:"products,omitempty"`
			Total *int `json:"total,omitempty"`
		}{
			Products: &productsData,
			Total:    &total,
		},
		Message:    "Products retrieved successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// GetProduct handles GET /products/{id} requests
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	productID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	product, err := h.productUseCase.GetByID(ctx, productID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	productId := openapi_types.UUID(product.ID)
	price := float32(product.Price)

	response := genhttp.ProductResponse{
		Code: "success",
		Data: struct {
			Id        *openapi_types.UUID `json:"id,omitempty"`
			Inventory *int                `json:"inventory,omitempty"`
			Name      *string             `json:"name,omitempty"`
			Price     *float32            `json:"price,omitempty"`
			Sku       *string             `json:"sku,omitempty"`
		}{
			Id:        &productId,
			Sku:       &product.SKU,
			Name:      &product.Name,
			Price:     &price,
			Inventory: &product.Inventory,
		},
		Message:    "Product retrieved successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateProduct handles POST /products requests
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params genhttp.CreateProductJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	product, err := h.productUseCase.Create(ctx, params.Sku, params.Name, float64(params.Price), params.Inventory)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	productId := openapi_types.UUID(product.ID)
	price := float32(product.Price)

	response := genhttp.ProductResponse{
		Code: "success",
		Data: struct {
			Id        *openapi_types.UUID `json:"id,omitempty"`
			Inventory *int                `json:"inventory,omitempty"`
			Name      *string             `json:"name,omitempty"`
			Price     *float32            `json:"price,omitempty"`
			Sku       *string             `json:"sku,omitempty"`
		}{
			Id:        &productId,
			Sku:       &product.SKU,
			Name:      &product.Name,
			Price:     &price,
			Inventory: &product.Inventory,
		},
		Message:    "Product created successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusCreated, response)
}

// UpdateProduct handles PUT /products/{id} requests
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	productID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	var params genhttp.UpdateProductJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	// Extract values from pointers
	var name string
	var price float64
	var inventory int

	if params.Name != nil {
		name = *params.Name
	}

	if params.Price != nil {
		price = float64(*params.Price)
	}

	if params.Inventory != nil {
		inventory = *params.Inventory
	}

	product, err := h.productUseCase.Update(ctx, productID, name, price, inventory)
	if err != nil {
		handleError(w, err)
		return
	}

	// Map to response
	productId := openapi_types.UUID(product.ID)
	productPrice := float32(product.Price)

	response := genhttp.ProductResponse{
		Code: "success",
		Data: struct {
			Id        *openapi_types.UUID `json:"id,omitempty"`
			Inventory *int                `json:"inventory,omitempty"`
			Name      *string             `json:"name,omitempty"`
			Price     *float32            `json:"price,omitempty"`
			Sku       *string             `json:"sku,omitempty"`
		}{
			Id:        &productId,
			Sku:       &product.SKU,
			Name:      &product.Name,
			Price:     &productPrice,
			Inventory: &product.Inventory,
		},
		Message:    "Product updated successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteProduct handles DELETE /products/{id} requests
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	productID, err := uuid.Parse(id.String())
	if err != nil {
		handleError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	err = h.productUseCase.Delete(ctx, productID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return a standard success response
	response := genhttp.StandardResponse{
		Code:       "success",
		Data:       map[string]interface{}{},
		Message:    "Product deleted successfully",
		ServerTime: time.Now(),
	}

	respondJSON(w, http.StatusOK, response)
}

// Helper functions

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
	var code string
	var message string

	switch {
	case errors.IsNotFound(err):
		status = http.StatusNotFound
		code = "not_found"
		message = err.Error()
	case errors.IsBadRequest(err):
		status = http.StatusBadRequest
		code = "bad_request"
		message = err.Error()
	default:
		status = http.StatusInternalServerError
		code = "internal_error"
		message = "internal server error"
	}

	errorResponse := genhttp.ErrorResponse{
		Code:       code,
		Message:    message,
		ServerTime: time.Now(),
	}

	respondJSON(w, status, errorResponse)
}
