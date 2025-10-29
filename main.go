package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/adii2ma/dbms-backend/database"
	"github.com/adii2ma/dbms-backend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file loaded: %v", err)
	}

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	if shouldMigrate() {
		if err := database.RunMigrations(context.Background()); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
	}

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins, or specify your frontend URL like "http://localhost:3000"
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/signup", routes.SignUp)
			auth.POST("/signin", routes.SignIn)
		}

		requests := api.Group("/requests")
		{
			requests.POST("", routes.CreateRequest)
			requests.GET("/active", routes.GetActiveRequest)
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func shouldMigrate() bool {
	value := os.Getenv("SHOULD_MIGRATE")
	if value == "" {
		return false
	}
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "true" || value == "1" || value == "yes" || value == "y"
}
