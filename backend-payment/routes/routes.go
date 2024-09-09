package routes

import (
	"backend-payment/routes/api"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the routes for the application
func SetupRoutes(r *gin.Engine) {
	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api.SetupPaymentRoutes(r)

	r.GET("/health", healthCheckHandler)
}

// @Summary Health check
// @Description Get a health check message
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func healthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Service is healthy",
	})
}
