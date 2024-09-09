package routes

import (
	"backend-order/routes/api"
	"backend-order/routes/api/admin"
	"backend-order/routes/api/backend"
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

	r.GET("/health", healthCheckHandler)
	api.SetupAuthRoutes(r)
	api.SetupProductRoutes(r)
	api.SetupOrderRoutes(r)

	admin.SetupAdminProductRoutes(r)
	admin.SetupAdminOrderRoutes(r)
	admin.SetupAdminUserRoutes(r)

	// Add this line to set up the new backend payment routes
	backend.SetupBackendPaymentRoutes(r)
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
