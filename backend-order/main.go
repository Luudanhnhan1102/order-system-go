package main

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "backend-order/docs"
	"backend-order/jobs"
	"backend-order/middleware"
	_ "backend-order/models"
	"backend-order/routes"
)

// @title Order API
// @version 1.0
// @description This is a simple backend server using Go and Gin framework.
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	r := gin.Default()

	r.Use(middleware.LoggerMiddleware())

	// Setup routes
	routes.SetupRoutes(r)

	// Get API_URL from environment and parse the host
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080" // Default value if not set
	}

	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		log.Printf("Error parsing API_URL: %v. Using default.", err)
		docs.SwaggerInfo.Host = "localhost:8080"
	} else {
		// Remove port if it's the default port for the scheme
		host := parsedURL.Host
		if (parsedURL.Scheme == "http" && strings.HasSuffix(host, ":80")) ||
			(parsedURL.Scheme == "https" && strings.HasSuffix(host, ":443")) {
			host = strings.Split(host, ":")[0]
		}
		docs.SwaggerInfo.Host = host
	}

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start background job
	go runBackgroundJob()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func runBackgroundJob() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// concurrency in go
	for range ticker.C {
		jobs.DeliverConfirmedOrders()
	}
}
