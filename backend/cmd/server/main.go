package main

import (
	"log"
	"os"

	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/database"
	"github.com/supplierpay/backend/internal/router"
	"github.com/supplierpay/backend/internal/scheduler"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	logger.Info("Starting SupplierPay AI",
		zap.String("env", cfg.AppEnv),
		zap.String("port", cfg.AppPort),
		zap.Bool("mock_mode", cfg.MockMode),
	)

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("Connected to database")

	// Start payment scheduler (cron)
	cron := scheduler.NewPaymentScheduler(db, cfg, logger)
	cron.Start()
	defer cron.Stop()
	logger.Info("Payment scheduler started")

	// Setup and start HTTP server
	r := router.Setup(db, cfg, logger)

	port := cfg.AppPort
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
