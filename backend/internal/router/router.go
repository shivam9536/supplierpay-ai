package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/handlers"
	"github.com/supplierpay/backend/internal/middleware"
	"go.uber.org/zap"
)

func Setup(db *sqlx.DB, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "supplierpay-ai"})
	})

	// Initialize handlers
	invoiceHandler := handlers.NewInvoiceHandler(db, cfg, logger)
	vendorHandler := handlers.NewVendorHandler(db, logger)
	poHandler := handlers.NewPurchaseOrderHandler(db, logger)
	paymentHandler := handlers.NewPaymentHandler(db, cfg, logger)
	forecastHandler := handlers.NewForecastHandler(db, logger)
	sseHandler := handlers.NewSSEHandler(logger)

	// ── API v1 Routes ───────────────────────
	v1 := r.Group("/api/v1")
	{
		// Auth (simple JWT for hackathon)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login(cfg))
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Invoices
			invoices := protected.Group("/invoices")
			{
				invoices.POST("/upload", invoiceHandler.Upload)
				invoices.POST("/upload-json", invoiceHandler.UploadJSON)
				invoices.GET("", invoiceHandler.List)
				invoices.GET("/:id", invoiceHandler.GetByID)
				invoices.GET("/:id/audit-log", invoiceHandler.GetAuditLog)
				invoices.POST("/:id/reprocess", invoiceHandler.Reprocess)
			}

			// Vendors
			vendors := protected.Group("/vendors")
			{
				vendors.GET("", vendorHandler.List)
				vendors.GET("/:id", vendorHandler.GetByID)
				vendors.POST("", vendorHandler.Create)
			}

			// Purchase Orders
			pos := protected.Group("/purchase-orders")
			{
				pos.GET("", poHandler.List)
				pos.GET("/:id", poHandler.GetByID)
				pos.POST("", poHandler.Create)
			}

			// Payments
			payments := protected.Group("/payments")
			{
				payments.GET("/schedule", paymentHandler.GetSchedule)
				payments.POST("/run", paymentHandler.TriggerRun)
				payments.GET("/runs", paymentHandler.ListRuns)
			}

			// Cash Flow Forecast
			forecast := protected.Group("/forecast")
			{
				forecast.GET("", forecastHandler.GetForecast)
			}
		}

		// SSE (Server-Sent Events) — separate auth for streaming
		v1.GET("/events/invoices/:id", sseHandler.StreamInvoiceUpdates)

		// Pine Labs Webhook (no auth — validated by signature)
		v1.POST("/webhooks/pinelabs", paymentHandler.PineLabsWebhook)
	}

	return r
}
