package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Kir-Khorev/finopp-back/internal/common"
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

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := e.Group("/api/v1")
	
	// Auth routes (to be implemented)
	api.POST("/auth/register", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "register endpoint"})
	})
	api.POST("/auth/login", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "login endpoint"})
	})

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

