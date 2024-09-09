package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "backend-payment/docs" // This is where Swag will generate its docs.go file
	"backend-payment/routes"
)

// @title Payment API
// @version 1.0
// @description This is a payment service API.
// @host localhost:8081
// @BasePath /
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	// Create a new Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

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
