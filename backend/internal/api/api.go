package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/veggieshop/backend/internal/api/handlers"
	"github.com/veggieshop/backend/internal/api/middleware"
)

// Config holds all dependencies for the API
type Config struct {
	AuthHandler          *handlers.AuthHandler
	StoreHandler         *handlers.StoreHandler
	ProductHandler       *handlers.ProductHandler
	CategoryHandler      *handlers.CategoryHandler
	OrderHandler         *handlers.OrderHandler
	DeliveryHandler      *handlers.DeliveryHandler
	AdminStoreHandler    *handlers.AdminStoreHandler
	AdminCategoryHandler *handlers.AdminCategoryHandler
	AdminProductHandler   *handlers.AdminProductHandler
	AdminAnalyticsHandler *handlers.AdminAnalyticsHandler
	AdminDeliveryHandler  *handlers.AdminDeliveryHandler
	AdminOrderHandler     *handlers.AdminOrderHandler
	AdminStockHandler     *handlers.AdminStockHandler
	CourierHandler        *handlers.CourierHandler
	V2Handler             *handlers.V2Handler
	JWTSecret             string
	UploadDir             string // файлы товаров; при непустом — Static /uploads и POST .../upload/product-image
}

func Setup(cfg *Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if cfg.UploadDir != "" {
		_ = os.MkdirAll(cfg.UploadDir, 0755)
		r.Static("/uploads", cfg.UploadDir)
	}

	v1 := r.Group("/api/v1")
	v1.GET("/health", handlers.Health)

	if cfg.AuthHandler != nil {
		auth := v1.Group("/auth")
		auth.POST("/register", cfg.AuthHandler.Register)
		auth.POST("/login", cfg.AuthHandler.Login)
		auth.POST("/refresh", cfg.AuthHandler.Refresh)
	}
	if cfg.StoreHandler != nil {
		v1.GET("/stores", cfg.StoreHandler.GetStores)
		v1.GET("/stores/:id", cfg.StoreHandler.GetStoreByID)
		v1.GET("/stores/:id/districts", cfg.StoreHandler.GetDistricts)
	}
	if cfg.DeliveryHandler != nil {
		v1.GET("/stores/:id/time-slots", cfg.DeliveryHandler.GetTimeSlots)
	}
	if cfg.ProductHandler != nil {
		v1.GET("/products/availability", cfg.ProductHandler.GetProductAvailability)
		v1.GET("/products", cfg.ProductHandler.GetProducts)
		v1.GET("/products/:id", cfg.ProductHandler.GetProductByID)
	}
	if cfg.CategoryHandler != nil {
		v1.GET("/categories", cfg.CategoryHandler.GetCategories)
	}
	if cfg.OrderHandler != nil {
		v1.POST("/orders/check-delivery", cfg.DeliveryHandler.CheckDelivery)
		if cfg.JWTSecret != "" {
			ordersCustomer := v1.Group("/orders")
			ordersCustomer.Use(middleware.JWTAuth(cfg.JWTSecret))
			ordersCustomer.Use(middleware.CustomerOnly())
			ordersCustomer.POST("", cfg.OrderHandler.CreateOrder)
			ordersCustomer.GET("/mine", cfg.OrderHandler.ListMyOrders)
			ordersCustomer.GET("/:id", cfg.OrderHandler.GetMyOrder)
		}
	}

	hasAdmin := cfg.AdminStoreHandler != nil || cfg.AdminCategoryHandler != nil || cfg.AdminProductHandler != nil || cfg.AdminAnalyticsHandler != nil || cfg.AdminDeliveryHandler != nil || cfg.AdminOrderHandler != nil || cfg.AdminStockHandler != nil
	if hasAdmin {
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWTSecret))
		admin.Use(middleware.AdminOnly())
		if cfg.AdminStoreHandler != nil {
			admin.GET("/stores", cfg.AdminStoreHandler.GetStores)
			admin.POST("/stores", cfg.AdminStoreHandler.CreateStore)
			admin.PATCH("/stores/:id", cfg.AdminStoreHandler.UpdateStore)
		}
		if cfg.AdminCategoryHandler != nil {
			admin.GET("/categories", cfg.AdminCategoryHandler.List)
			admin.POST("/categories", cfg.AdminCategoryHandler.Create)
			admin.PATCH("/categories/:id", cfg.AdminCategoryHandler.Patch)
			admin.DELETE("/categories/:id", cfg.AdminCategoryHandler.Delete)
		}
		if cfg.AdminProductHandler != nil {
			admin.GET("/stores/:storeId/products", cfg.AdminProductHandler.ListByStore)
			admin.POST("/stores/:storeId/products", cfg.AdminProductHandler.Create)
			admin.PATCH("/products/:id", cfg.AdminProductHandler.Patch)
			admin.DELETE("/products/:id", cfg.AdminProductHandler.Deactivate)
		}
		if cfg.AdminAnalyticsHandler != nil {
			admin.GET("/analytics/revenue", cfg.AdminAnalyticsHandler.GetRevenueReport)
		}
		if cfg.AdminOrderHandler != nil {
			admin.PATCH("/orders/:id", cfg.AdminOrderHandler.PatchOrder)
		}
		if cfg.AdminStockHandler != nil {
			admin.GET("/stores/:storeId/stock/zones", cfg.AdminStockHandler.ListZones)
			admin.GET("/stores/:storeId/stock/expiring", cfg.AdminStockHandler.ListExpiring)
			admin.GET("/stores/:storeId/stock/reorder-alerts", cfg.AdminStockHandler.ListReorderAlerts)
			admin.GET("/stores/:storeId/stock/moves-journal", cfg.AdminStockHandler.ListMovesJournal)
			admin.POST("/stores/:storeId/stock/receive-simple", cfg.AdminStockHandler.SimpleReceive)
			admin.POST("/stores/:storeId/stock/set-actual", cfg.AdminStockHandler.SetInventoryActual)
			admin.POST("/stores/:storeId/stock/write-off", cfg.AdminStockHandler.WriteOff)
			admin.POST("/stores/:storeId/stock/receipt", cfg.AdminStockHandler.ApplyReceipt)
			admin.POST("/stores/:storeId/stock/audit/complete", cfg.AdminStockHandler.CompleteAudit)
			admin.GET("/stores/:storeId/suppliers", cfg.AdminStockHandler.ListSuppliers)
			admin.POST("/stores/:storeId/suppliers", cfg.AdminStockHandler.CreateSupplier)
			admin.GET("/stores/:storeId/products/:productId/batches", cfg.AdminStockHandler.ListBatches)
		}
		if cfg.AdminDeliveryHandler != nil {
			admin.GET("/stores/:storeId/districts", cfg.AdminDeliveryHandler.ListDistricts)
			admin.POST("/stores/:storeId/districts", cfg.AdminDeliveryHandler.CreateDistrict)
			admin.PATCH("/districts/:id", cfg.AdminDeliveryHandler.PatchDistrict)
			admin.DELETE("/districts/:id", cfg.AdminDeliveryHandler.DeleteDistrict)
			admin.GET("/stores/:storeId/time-slots", cfg.AdminDeliveryHandler.ListTimeSlots)
			admin.POST("/stores/:storeId/time-slots", cfg.AdminDeliveryHandler.CreateTimeSlot)
			admin.PATCH("/time-slots/:id", cfg.AdminDeliveryHandler.PatchTimeSlot)
			admin.DELETE("/time-slots/:id", cfg.AdminDeliveryHandler.DeleteTimeSlot)
		}
		if cfg.UploadDir != "" {
			uh := handlers.NewAdminUploadHandler(cfg.UploadDir)
			admin.POST("/upload/product-image", uh.UploadProductImage)
		}
	}

	if cfg.CourierHandler != nil && cfg.JWTSecret != "" {
		cr := v1.Group("/courier")
		cr.Use(middleware.JWTAuth(cfg.JWTSecret))
		cr.Use(middleware.CourierOnly())
		cr.GET("/orders", cfg.CourierHandler.ListOrders)
		cr.POST("/orders/:id/accept", cfg.CourierHandler.AcceptOrder)
		cr.POST("/orders/:id/complete", cfg.CourierHandler.CompleteDelivery)
	}

	if cfg.V2Handler != nil {
		v2 := r.Group("/api/v2")
		v2.GET("/products", cfg.V2Handler.ListProductsPaged)
		adm := v2.Group("/admin")
		adm.Use(middleware.JWTAuth(cfg.JWTSecret))
		adm.Use(middleware.AdminOnly())
		adm.GET("/stores/:storeId/orders", cfg.V2Handler.ListOrdersPaged)
		adm.GET("/stores/:storeId/stock/movements", cfg.V2Handler.ListStockMovementsPaged)
	}

	return r
}
