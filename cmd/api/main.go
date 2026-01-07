package main

import (
	"log"

	"github.com/datmedevil17/gopher-uptime/internal/config"
	"github.com/datmedevil17/gopher-uptime/internal/database"
	"github.com/datmedevil17/gopher-uptime/internal/handlers/user"
	"github.com/datmedevil17/gopher-uptime/internal/handlers/website"
	"github.com/datmedevil17/gopher-uptime/internal/middleware"
	"github.com/datmedevil17/gopher-uptime/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

func main() {
	log.Println("🚀 Starting Uptime Monitor API Server...")

	// Load configuration
	cfg := config.Load()

	// Connect to database with GORM
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}
	log.Println("✅ Database connected")

	// Run auto-migration
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal("❌ RabbitMQ connection failed:", err)
	}
	defer conn.Close()
	log.Println("✅ RabbitMQ connected")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("❌ Failed to open RabbitMQ channel:", err)
	}
	defer ch.Close()

	// Initialize payout worker
	if cfg.PlatformPrivateKey != "" {
		worker, err := services.NewPayoutWorker(db, ch, cfg.PlatformPrivateKey)
		if err != nil {
			log.Fatal("❌ Failed to initialize payout worker:", err)
		}

		// Start worker in background
		go func() {
			if err := worker.Start(); err != nil {
				log.Fatal("❌ Payout worker error:", err)
			}
		}()
	} else {
		log.Println("⚠️  No PLATFORM_PRIVATE_KEY provided, payout worker disabled")
	}

	// Initialize Gin router
	r := gin.Default()

	// CORS middleware
	// CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Initialize handlers
	websiteHandler := website.NewHandler(db)
	userHandler := user.NewHandler(db, ch, cfg)

	// API routes
	api := r.Group("/api/v1")
	{
		// Protected routes (require JWT authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Website management
			protected.POST("/website", websiteHandler.CreateWebsite)
			protected.GET("/websites", websiteHandler.GetWebsites)
			protected.GET("/website/status", websiteHandler.GetWebsiteStatus)
			protected.DELETE("/website", websiteHandler.DeleteWebsite)
		}

		// Public routes (or validator-only)
		api.POST("/payout/:validatorId", userHandler.RequestPayout)
		api.GET("/validator/:validatorId/balance", userHandler.GetValidatorBalance)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/signup", userHandler.Signup)
			auth.POST("/login", userHandler.Login)
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "uptime-monitor-api",
		})
	})

	// Start server
	log.Printf("🚀 API Server running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
