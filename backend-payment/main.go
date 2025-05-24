package main

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "backend-payment/docs"
	"backend-payment/middleware"
	"backend-payment/routes"
)

// @title Payment API
// @version 1.0
// @description This is a payment service API.
// @BasePath /

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	// Create a new Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add request logging middleware
	r.Use(middleware.LoggerMiddleware())

	// Setup routes
	routes.SetupRoutes(r)

	// Get API_URL from environment and parse the host
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8081" // Default value if not set
	}

	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		log.Printf("Error parsing API_URL: %v. Using default.", err)
		docs.SwaggerInfo.Host = "localhost:8081"
	} else {
		// Remove port if it's the default port for the scheme
		host := parsedURL.Host
		if (parsedURL.Scheme == "http" && strings.HasSuffix(host, ":80")) ||
			(parsedURL.Scheme == "https" && strings.HasSuffix(host, ":443")) {
			host = strings.Split(host, ":")[0]
		}
		docs.SwaggerInfo.Host = host
	}

	// Add Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
