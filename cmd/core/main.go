package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cartPort "github.com/fanzru/e-commerce-be/internal/app/cart/port"
	cartRepo "github.com/fanzru/e-commerce-be/internal/app/cart/repo"
	cartUseCase "github.com/fanzru/e-commerce-be/internal/app/cart/usecase"
	checkoutPort "github.com/fanzru/e-commerce-be/internal/app/checkout/port"
	checkoutRepo "github.com/fanzru/e-commerce-be/internal/app/checkout/repo"
	checkoutUseCase "github.com/fanzru/e-commerce-be/internal/app/checkout/usecase"
	productPort "github.com/fanzru/e-commerce-be/internal/app/product/port"
	productRepo "github.com/fanzru/e-commerce-be/internal/app/product/repo"
	productUseCase "github.com/fanzru/e-commerce-be/internal/app/product/usecase"
	promotionPort "github.com/fanzru/e-commerce-be/internal/app/promotion/port"
	promotionRepo "github.com/fanzru/e-commerce-be/internal/app/promotion/repo"
	promotionUseCase "github.com/fanzru/e-commerce-be/internal/app/promotion/usecase"
	userPort "github.com/fanzru/e-commerce-be/internal/app/user/port"
	userRepo "github.com/fanzru/e-commerce-be/internal/app/user/repo"
	userUseCase "github.com/fanzru/e-commerce-be/internal/app/user/usecase"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/config"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/persistence"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize OpenTelemetry
	shutdown, err := middleware.InitOTEL()
	if err != nil {
		log.Printf("Warning: Failed to initialize OpenTelemetry: %v", err)
	} else {
		defer shutdown(context.Background())
	}

	// Initialize database
	db, err := initializeDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	repos, err := initializeRepositories(db)
	if err != nil {
		log.Fatalf("Failed to initialize repositories: %v", err)
	}

	// Initialize use cases
	useCases := initializeUseCases(repos, cfg)

	// Create middleware factory
	middlewareFactory := middleware.NewFactory(cfg)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.Handle("/health", middlewareFactory.WrapFunc(middleware.AuthTypePublic, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Static file server for assets
	assetsFS := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", middlewareFactory.Apply(http.StripPrefix("/assets/", assetsFS), middleware.AuthTypePublic))

	// Static file server for web UI
	webFS := http.FileServer(http.Dir("./web"))
	mux.Handle("/web/", middlewareFactory.Apply(http.StripPrefix("/web/", webFS), middleware.AuthTypePublic))

	// Static file server for Swagger docs
	docFS := http.FileServer(http.Dir("./docs"))
	mux.Handle("/docs/", middlewareFactory.Apply(http.StripPrefix("/docs/", docFS), middleware.AuthTypePublic))

	// Swagger UI handler
	mux.Handle("/swagger/", middlewareFactory.WrapFunc(middleware.AuthTypePublic, func(w http.ResponseWriter, r *http.Request) {
		// Read the HTML template from file
		htmlBytes, err := os.ReadFile("./assets/html/swagger.html")
		if err != nil {
			middleware.Logger.Error("Error reading Swagger UI template", "error", err)
			http.Error(w, "Error reading Swagger UI template", http.StatusInternalServerError)
			return
		}

		// Replace @assets with the actual path
		html := strings.ReplaceAll(string(htmlBytes), "@assets", "/assets")

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))

	// API routes
	apiHandler := createAPIHandler(useCases, middlewareFactory)

	// Register API handler with context path
	mux.Handle("/", apiHandler)

	// Create HTTP server
	serverPort := cfg.ServerPort
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverPort),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Get the hostname from environment or default to localhost
	swaggerHost := os.Getenv("SWAGGER_HOST")
	if swaggerHost == "" {
		swaggerHost = "localhost"
	}

	// Start server in a goroutine
	go func() {
		middleware.Logger.Info("Starting server",
			"port", serverPort,
			"swagger_url", fmt.Sprintf("http://%s:%d/swagger/", swaggerHost, serverPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	middleware.Logger.Info("Shutting down server...")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		middleware.Logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	middleware.Logger.Info("Server gracefully stopped")
}

func initializeDatabase(cfg *config.Config) (*sql.DB, error) {
	// Connect to PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMinutes) * time.Minute)

	return db, nil
}

type repositories struct {
	db            *sql.DB
	productRepo   productRepo.ProductRepository
	cartRepo      cartRepo.CartRepository
	checkoutRepo  checkoutRepo.CheckoutRepository
	promotionRepo promotionRepo.PromotionRepository
	userRepo      userRepo.UserRepository
	tokenRepo     userRepo.TokenRepository
}

func initializeRepositories(db *sql.DB) (*repositories, error) {
	// Initialize repositories from each domain

	return &repositories{
		db:            db,
		productRepo:   productRepo.NewProductRepository(db),
		cartRepo:      cartRepo.NewCartRepository(db),
		checkoutRepo:  checkoutRepo.NewCheckoutRepository(db),
		promotionRepo: promotionRepo.NewPromotionRepository(db),
		userRepo:      userRepo.NewUserRepository(db),
		tokenRepo:     userRepo.NewTokenRepository(db),
	}, nil
}

type useCases struct {
	productUseCase   productUseCase.ProductUseCase
	cartUseCase      cartUseCase.CartUseCase
	checkoutUseCase  checkoutUseCase.CheckoutUseCase
	promotionUseCase promotionUseCase.PromotionUseCase
	userUseCase      userUseCase.UserUseCase
}

func initializeUseCases(repos *repositories, cfg *config.Config) *useCases {
	// Initialize transaction manager
	txManager := persistence.ProvideTransactionManager(repos.db)

	// Initialize use cases with proper dependencies
	productUC := productUseCase.NewProductUseCase(repos.productRepo)
	promotionUC := promotionUseCase.NewPromotionUseCase(repos.promotionRepo)
	cartUC := cartUseCase.NewCartUseCase(repos.cartRepo, repos.productRepo, promotionUC)
	checkoutUC := checkoutUseCase.NewCheckoutUseCase(repos.checkoutRepo, repos.cartRepo, repos.promotionRepo, txManager)

	// Initialize user use case with JWT configuration from config
	userUC := userUseCase.NewUserUseCase(
		repos.userRepo,
		cfg.JWT.SecretKey,       // Get from config
		cfg.JWT.ExpirationHours, // Token expiration in hours
	)

	return &useCases{
		productUseCase:   productUC,
		cartUseCase:      cartUC,
		checkoutUseCase:  checkoutUC,
		promotionUseCase: promotionUC,
		userUseCase:      userUC,
	}
}

func createAPIHandler(useCases *useCases, middlewareFactory *middleware.Factory) http.Handler {
	mux := http.NewServeMux()

	// Mount the generated HTTP servers with RBAC handlers

	// User API with operation-based RBAC
	userBaseHandler := userPort.NewHTTPServer(useCases.userUseCase)
	userRBAC := middleware.NewRBACMiddleware(middlewareFactory).
		// Auth operations
		WithOperation("LoginUser", middleware.AuthTypePublic).
		WithOperation("RegisterUser", middleware.AuthTypePublic).
		WithOperation("RefreshToken", middleware.AuthTypePublic).
		WithOperation("LogoutUser", middleware.AuthTypeBearer).
		// Admin-only user management
		WithOperation("ListUsers", middleware.AuthTypeRoleAdmin).
		WithOperation("GetUser", middleware.AuthTypeRoleAdmin).
		WithOperation("UpdateUser", middleware.AuthTypeRoleAdmin).
		WithOperation("DeleteUser", middleware.AuthTypeRoleAdmin).
		WithOperation("UpdatePassword", middleware.AuthTypeRoleAdmin).
		WithDefaultRoles(middleware.AuthTypeRoleAdmin)

	// Authentication endpoints without api/v1 prefix
	userRBAC.RegisterPathPattern("POST", "/api/v1/auth/login", "LoginUser")
	userRBAC.RegisterPathPattern("POST", "/api/v1/auth/register", "RegisterUser")
	userRBAC.RegisterPathPattern("POST", "/api/v1/auth/refresh", "RefreshToken")
	userRBAC.RegisterPathPattern("POST", "/api/v1/auth/logout", "LogoutUser")

	// Also register non-prefixed paths for backward compatibility
	userRBAC.RegisterPathPattern("GET", "/api/v1/users", "ListUsers")
	userRBAC.RegisterPathPattern("GET", "/api/v1/users/{id}", "GetUser")
	userRBAC.RegisterPathPattern("PATCH", "/api/v1/users/{id}", "UpdateUser")
	userRBAC.RegisterPathPattern("DELETE", "/api/v1/users/{id}", "DeleteUser")
	userRBAC.RegisterPathPattern("PUT", "/api/v1/users/{id}/password", "UpdatePassword")

	// Register user API endpoints
	mux.Handle("/api/v1/auth/", userRBAC.Wrap(userBaseHandler))
	mux.Handle("/api/v1/users", userRBAC.Wrap(userBaseHandler))
	mux.Handle("/api/v1/users/", userRBAC.Wrap(userBaseHandler))

	// Product API with direct RBAC middleware
	productBaseHandler := productPort.NewHTTPServer(useCases.productUseCase)
	productRBAC := middleware.NewRBACMiddleware(middlewareFactory).
		// List and Get operations are public
		WithOperation("ListProducts", middleware.AuthTypePublic, middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("GetProduct", middleware.AuthTypePublic).
		// Write operations require admin role
		WithOperation("CreateProduct", middleware.AuthTypeRoleAdmin).
		WithOperation("UpdateProduct", middleware.AuthTypeRoleAdmin).
		WithOperation("DeleteProduct", middleware.AuthTypeRoleAdmin).
		// Set default access control (restrict by default)
		WithDefaultRoles(middleware.AuthTypeRoleAdmin)

	// Register product path patterns
	productRBAC.RegisterPathPattern("GET", "/api/v1/products", "ListProducts")
	productRBAC.RegisterPathPattern("POST", "/api/v1/products", "CreateProduct")
	productRBAC.RegisterPathPattern("GET", "/api/v1/products/{id}", "GetProduct")
	productRBAC.RegisterPathPattern("PUT", "/api/v1/products/{id}", "UpdateProduct")
	productRBAC.RegisterPathPattern("DELETE", "/api/v1/products/{id}", "DeleteProduct")

	// Register product API endpoints
	mux.Handle("/api/v1/products", productRBAC.Wrap(productBaseHandler))
	mux.Handle("/api/v1/products/", productRBAC.Wrap(productBaseHandler))

	// Cart API with operation-based RBAC
	cartBaseHandler := cartPort.NewHTTPServer(useCases.cartUseCase, useCases.promotionUseCase)
	cartRBAC := middleware.NewRBACMiddleware(middlewareFactory).
		// GetCart require customer role
		WithOperation("GetCart", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		// Other cart operations also require customer role
		WithOperation("CreateCart", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("UpdateCart", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("DeleteCart", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("AddItem", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("RemoveItem", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("GetCurrentUserCart", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("AddItemToCurrentUserCart", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		// Default to customer access
		WithDefaultRoles(middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin)

	// Register cart path patterns - no ListCarts endpoint exists in the API
	cartRBAC.RegisterPathPattern("POST", "/api/v1/carts", "CreateCart")
	cartRBAC.RegisterPathPattern("GET", "/api/v1/carts/{id}", "GetCart")
	cartRBAC.RegisterPathPattern("PUT", "/api/v1/carts/{id}", "UpdateCart")
	cartRBAC.RegisterPathPattern("DELETE", "/api/v1/carts/{id}", "DeleteCart")
	cartRBAC.RegisterPathPattern("POST", "/api/v1/carts/{id}/items", "AddItem")
	cartRBAC.RegisterPathPattern("DELETE", "/api/v1/carts/{id}/items/{item_id}", "RemoveItem")
	cartRBAC.RegisterPathPattern("GET", "/api/v1/carts/me", "GetCurrentUserCart")
	cartRBAC.RegisterPathPattern("POST", "/api/v1/carts/me", "AddItemToCurrentUserCart")

	// Register cart API endpoints
	mux.Handle("/api/v1/carts", cartRBAC.Wrap(cartBaseHandler))
	mux.Handle("/api/v1/carts/", cartRBAC.Wrap(cartBaseHandler))

	// Checkout API with operation-based RBAC
	checkoutBaseHandler := checkoutPort.NewHTTPServer(useCases.checkoutUseCase)
	checkoutRBAC := middleware.NewRBACMiddleware(middlewareFactory).
		// All checkout operations require customer role at minimum
		WithOperation("CreateCheckout", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("GetCheckout", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("ListCheckouts", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("ProcessPayment", middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithDefaultRoles(middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin)

	// Register checkout path patterns
	checkoutRBAC.RegisterPathPattern("GET", "/api/v1/checkouts", "ListCheckouts")
	checkoutRBAC.RegisterPathPattern("POST", "/api/v1/checkouts", "CreateCheckout")
	checkoutRBAC.RegisterPathPattern("GET", "/api/v1/checkouts/{id}", "GetCheckout")
	checkoutRBAC.RegisterPathPattern("POST", "/api/v1/checkouts/{id}/payment", "ProcessPayment")

	// Register checkout API endpoints
	mux.Handle("/api/v1/checkouts", checkoutRBAC.Wrap(checkoutBaseHandler))
	mux.Handle("/api/v1/checkouts/", checkoutRBAC.Wrap(checkoutBaseHandler))

	// Promotion API with operation-based RBAC
	promotionBaseHandler := promotionPort.NewHTTPServer(useCases.promotionUseCase)
	promotionRBAC := middleware.NewRBACMiddleware(middlewareFactory).
		// View promotions is public
		WithOperation("ListPromotions", middleware.AuthTypePublic, middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		WithOperation("GetPromotion", middleware.AuthTypePublic, middleware.AuthTypeRoleCustomer, middleware.AuthTypeRoleAdmin).
		// Modify promotions requires admin
		WithOperation("CreatePromotion", middleware.AuthTypeRoleAdmin).
		WithOperation("UpdatePromotion", middleware.AuthTypeRoleAdmin).
		WithOperation("DeletePromotion", middleware.AuthTypeRoleAdmin).
		WithDefaultRoles(middleware.AuthTypeRoleAdmin)

	// Register promotion path patterns
	promotionRBAC.RegisterPathPattern("GET", "/api/v1/promotions", "ListPromotions")
	promotionRBAC.RegisterPathPattern("POST", "/api/v1/promotions", "CreatePromotion")
	promotionRBAC.RegisterPathPattern("GET", "/api/v1/promotions/{id}", "GetPromotion")
	promotionRBAC.RegisterPathPattern("PUT", "/api/v1/promotions/{id}", "UpdatePromotion")
	promotionRBAC.RegisterPathPattern("DELETE", "/api/v1/promotions/{id}", "DeletePromotion")

	// Register promotion API endpoints
	mux.Handle("/api/v1/promotions", promotionRBAC.Wrap(promotionBaseHandler))
	mux.Handle("/api/v1/promotions/", promotionRBAC.Wrap(promotionBaseHandler))

	return mux
}
