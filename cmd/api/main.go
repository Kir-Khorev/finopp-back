package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Kir-Khorev/finopp-back/internal/advice"
	"github.com/Kir-Khorev/finopp-back/internal/auth"
	"github.com/Kir-Khorev/finopp-back/internal/common"
	"github.com/Kir-Khorev/finopp-back/internal/currency"
	appMiddleware "github.com/Kir-Khorev/finopp-back/internal/middleware"
	"github.com/Kir-Khorev/finopp-back/pkg/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load config
	cfg := config.Load()

	// Initialize database
	db, err := common.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis
	rdb := common.InitRedis(cfg)
	defer rdb.Close()

	// Run migrations
	if err := common.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Setup Echo
	e := echo.New()
	e.HideBanner = true
	
	// Custom error handler
	e.HTTPErrorHandler = appMiddleware.ErrorHandler

	// Global middleware
	e.Use(middleware.Recover())
	e.Use(appMiddleware.RequestLogger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://servify.digital",        // Production (new domain)
			"https://finopp-front.vercel.app", // Production (old domain)
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://localhost:8081", // Vite dev server (alternative port)
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Initialize Auth
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	// Initialize Currency Converter
	currencyService := currency.NewService(cfg.FixerAPIKey, rdb)

	// Initialize Advice
	adviceService := advice.NewService(cfg.GroqAPIKey, currencyService)
	adviceHandler := advice.NewHandler(adviceService)

	// API routes
	api := e.Group("/api/v1")
	
	// Public auth routes
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// Public advice routes (опционально можно защитить через middleware)
	api.POST("/advice", adviceHandler.GetAdvice, appMiddleware.OptionalAuthMiddleware(cfg.JWTSecret))
	api.POST("/advice/structured", adviceHandler.GetStructuredAdvice, appMiddleware.OptionalAuthMiddleware(cfg.JWTSecret))
	api.POST("/analyze", adviceHandler.Analyze, appMiddleware.OptionalAuthMiddleware(cfg.JWTSecret))
	
	// Protected routes example (раскомментировать когда добавятся эндпоинты)
	// protected := api.Group("")
	// protected.Use(appMiddleware.AuthMiddleware(cfg.JWTSecret))
	// protected.GET("/profile", profileHandler.GetProfile)

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil {
			log.Println("Server stopped:", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}

