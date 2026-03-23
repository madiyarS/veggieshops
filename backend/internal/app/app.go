package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/api"
	"github.com/veggieshop/backend/internal/api/handlers"
	"github.com/veggieshop/backend/internal/catalogcache"
	"github.com/veggieshop/backend/internal/config"
	"github.com/veggieshop/backend/internal/database"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/services"
	"gorm.io/gorm"
)

type App struct {
	DB       *gorm.DB
	API      *api.Config
	Catalog  *catalogcache.Store
	OrderSvc *services.OrderService
}

func New(cfg *config.Config) (*App, error) {
	// До ~60 с ждём Postgres (после docker compose restart контейнеры поднимаются параллельно).
	db, err := database.ConnectWithRetry(cfg.Database.DSN(), 30, 2*time.Second)
	if err != nil {
		slog.Warn("Database connection failed, running without DB", "error", err)
		db = nil
	}

	catalog := catalogcache.New(cfg.Redis.Addr)
	app := &App{DB: db, Catalog: catalog}
	if db != nil {
		if err := database.ApplyRuntimePatches(db); err != nil {
			return nil, fmt.Errorf("database patches: %w", err)
		}
		apiCfg, orderSvc := initAPI(cfg, db, catalog)
		app.API = apiCfg
		app.OrderSvc = orderSvc
	} else {
		app.API = initAPIWithoutDB(cfg)
	}
	return app, nil
}

// Close освобождает внешние ресурсы (Redis).
func (a *App) Close() error {
	if a == nil || a.Catalog == nil {
		return nil
	}
	return a.Catalog.Close()
}

func initAPI(cfg *config.Config, db *gorm.DB, catalog *catalogcache.Store) (*api.Config, *services.OrderService) {
	userRepo := repositories.NewUserRepository(db)
	storeRepo := repositories.NewStoreRepository(db)
	districtRepo := repositories.NewDistrictRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	orderItemRepo := repositories.NewOrderItemRepository(db)
	slotRepo := repositories.NewDeliveryTimeSlotRepository(db)
	streetRepo := repositories.NewDistrictStreetRepository(db)
	invRepo := repositories.NewInventoryRepository(db)
	var stockSvc *services.StockService
	stockSvc = services.NewStockService(db, invRepo, catalog, func(ctx context.Context, storeID uuid.UUID) {
		alerts, err := stockSvc.ListReorderAlerts(ctx, storeID)
		if err != nil {
			slog.Warn("reorder_alerts_failed", "store_id", storeID, "error", err)
			return
		}
		for _, row := range alerts {
			slog.Info("reorder_alert",
				"store_id", storeID,
				"product_id", row.ProductID,
				"name", row.Name,
				"available", row.Available,
				"reorder_min_qty", row.ReorderMin,
			)
		}
	})
	workflow := services.NewOrderWorkflow(db, orderRepo, stockSvc)
	courierRepo := repositories.NewCourierRepository(db)
	notifRepo := repositories.NewNotificationRepository(db)

	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.AccessTokenExp, cfg.JWT.RefreshTokenExp)
	storeSvc := services.NewStoreService(storeRepo)
	deliverySvc := services.NewDeliveryService(storeRepo, districtRepo, orderRepo, slotRepo)
	categorySvc := services.NewCategoryService(categoryRepo)
	productSvc := services.NewProductService(productRepo, storeRepo, categoryRepo, invRepo, stockSvc, catalog, db)
	orderSvc := services.NewOrderService(orderRepo, orderItemRepo, productRepo, storeRepo, districtRepo, slotRepo, deliverySvc, invRepo, stockSvc, workflow, db)
	analyticsSvc := services.NewAnalyticsService(orderRepo)
	_ = services.NewNotificationService(notifRepo, cfg.APIKeys.WhatsAppAPIKey)
	courierSvc := services.NewCourierService(courierRepo, orderRepo, workflow)
	v2Handler := handlers.NewV2Handler(productSvc, orderSvc, stockSvc)

	cfgAPI := &api.Config{
		AuthHandler:          handlers.NewAuthHandler(authSvc),
		StoreHandler:         handlers.NewStoreHandler(storeSvc, deliverySvc),
		ProductHandler:       handlers.NewProductHandler(productSvc),
		CategoryHandler:      handlers.NewCategoryHandler(categorySvc),
		OrderHandler:         handlers.NewOrderHandler(orderSvc),
		DeliveryHandler:      handlers.NewDeliveryHandler(deliverySvc),
		AdminStoreHandler:    handlers.NewAdminStoreHandler(storeSvc, productSvc),
		AdminCategoryHandler: handlers.NewAdminCategoryHandler(categorySvc, productSvc),
		AdminProductHandler:   handlers.NewAdminProductHandler(productSvc),
		AdminAnalyticsHandler: handlers.NewAdminAnalyticsHandler(analyticsSvc),
		AdminDeliveryHandler:  handlers.NewAdminDeliveryHandler(storeSvc, districtRepo, slotRepo, orderRepo, streetRepo),
		AdminOrderHandler:     handlers.NewAdminOrderHandler(orderSvc),
		AdminStockHandler:     handlers.NewAdminStockHandler(stockSvc),
		CourierHandler:        handlers.NewCourierHandler(courierSvc),
		V2Handler:             v2Handler,
		JWTSecret:             cfg.JWT.Secret,
		UploadDir:             cfg.UploadDir,
	}
	return cfgAPI, orderSvc
}

func initAPIWithoutDB(cfg *config.Config) *api.Config {
	slog.Info("Initializing API without database - limited functionality")
	return &api.Config{
		AuthHandler:          nil,
		StoreHandler:         nil,
		ProductHandler:       nil,
		CategoryHandler:      nil,
		OrderHandler:         nil,
		DeliveryHandler:      nil,
		AdminStoreHandler:    nil,
		AdminCategoryHandler: nil,
		AdminProductHandler:   nil,
		AdminAnalyticsHandler: nil,
		AdminDeliveryHandler:  nil,
		AdminOrderHandler:     nil,
		AdminStockHandler:     nil,
		CourierHandler:        nil,
		JWTSecret:             cfg.JWT.Secret,
	}
}

func (a *App) SetupRoutes() *api.Config {
	if a.API == nil {
		panic("API config not initialized")
	}
	return a.API
}
