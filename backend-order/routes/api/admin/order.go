package admin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"

	"backend-order/database"
	"backend-order/middleware"
	"backend-order/models"
)

// SetupAdminRoutes sets up the admin-related routes
func SetupAdminOrderRoutes(r *gin.Engine) {
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware()) // Ensure this middleware checks for admin role
	{
		adminGroup.GET("/orders", GetAllOrders)
		// Add other admin routes here
	}
}

// GetAllOrders godoc
// @Summary Get all orders
// @Description Retrieve all orders from all users (admin only)
// @Tags admin,orders
// @Accept json
// @Produce json
// @Success 200 {array} models.Order
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/orders [get]
func GetAllOrders(c *gin.Context) {
	ctx := context.Background()
	db := database.GetDB()

	// Set up options for sorting by creation date, newest first
	var orders []models.Order
	err := db.Collection("orders").Find(ctx, qmgo.M{}).Sort("-created_at").All(&orders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	// If orders is nil, initialize it as an empty slice
	if orders == nil {
		orders = []models.Order{}
	}

	c.JSON(http.StatusOK, orders)
}
