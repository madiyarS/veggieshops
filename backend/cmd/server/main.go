package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/veggieshop/backend/internal/api"
	"github.com/veggieshop/backend/internal/app"
	"github.com/veggieshop/backend/internal/config"
	"github.com/veggieshop/backend/internal/config/logger"
	"github.com/veggieshop/backend/internal/jobs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	config.InitLogger()
	log := logger.New("main")

	application, err := app.New(cfg)
	if err != nil {
		log.Error("Failed to init app", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Error("App close", "error", err)
		}
	}()

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()
	if application.OrderSvc != nil && cfg.Jobs.StaleSweepIntervalSec > 0 && cfg.Jobs.PendingOrderTimeoutMin > 0 {
		jobs.StartStalePendingSweep(
			jobCtx,
			application.OrderSvc,
			time.Duration(cfg.Jobs.StaleSweepIntervalSec)*time.Second,
			time.Duration(cfg.Jobs.PendingOrderTimeoutMin)*time.Minute,
		)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	apiCfg := application.SetupRoutes()
	r = api.Setup(apiCfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Info("Server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	jobCancel()

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
