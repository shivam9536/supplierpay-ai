package main

import (
	"log"

	"github.com/supplierpay/backend/internal/agent"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/database"
	"github.com/supplierpay/backend/internal/events"
	"github.com/supplierpay/backend/internal/router"
	"github.com/supplierpay/backend/internal/scheduler"
	"github.com/supplierpay/backend/internal/services"

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

	// Services (mock or real based on config)
	var llm services.LLMService
	var storage services.StorageService
	var email services.EmailService
	if cfg.MockMode {
		llm = services.NewMockLLMClient(logger)
		storage = services.NewMockStorageClient(logger)
		email = services.NewMockEmailClient(logger)
	} else {
		llm = services.NewBedrockClient(cfg, logger)
		storage = services.NewS3Client(cfg, logger)
		email = services.NewSESClient(cfg, logger)
	}

	// Agent orchestrator and event broadcaster
	orch := agent.NewOrchestrator(db, cfg, logger, llm, storage, email)
	broadcaster := events.NewBroadcaster(orch.GetEventChannel())
	broadcaster.Start()
	defer broadcaster.Stop()
	logger.Info("Agent orchestrator and event broadcaster started")

	// Start payment scheduler (cron) + invoice poll loop
	cron := scheduler.NewPaymentScheduler(db, cfg, logger, orch)
	cron.Start()
	defer cron.Stop()
	logger.Info("Payment scheduler started")

	// Setup and start HTTP server
	r := router.Setup(db, cfg, logger, orch, broadcaster, storage)

	port := cfg.AppPort
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
