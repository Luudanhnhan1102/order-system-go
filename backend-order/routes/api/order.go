package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend-order/database"
	"backend-order/middleware"
	"backend-order/models"
)

// SetupOrderRoutes sets up the order-related routes
func SetupOrderRoutes(r *gin.Engine) {
	orderGroup := r.Group("/orders")
	orderGroup.Use(middleware.AuthMiddleware())
	{
		orderGroup.GET("", getOrdersHandler)
		orderGroup.POST("", createOrderHandler)
		orderGroup.POST("/:id/cancel", cancelOrderHandler) // Add this line
	}
}

// @Summary Get my orders
// @Description Get the list of orders for the authenticated user
// @Tags Orders
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} models.Order
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [get]
func getOrdersHandler(c *gin.Context) {
	ctx := context.Background()
	db := database.GetDB()
	collection := db.Collection("orders")

	// Get the authenticated user from the context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	authenticatedUser, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	var orders []models.Order
	err := collection.Find(ctx, bson.M{"customer_id": authenticatedUser.Email}).Sort("-created_at").All(&orders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching orders"})
		return
	}

	// If orders is nil, initialize it as an empty slice
	if orders == nil {
		orders = []models.Order{}
	}

	c.JSON(http.StatusOK, orders)
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ProductID  string `json:"product_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
}

// @Summary Create a new order
// @Description Create a new order (requires authentication)
// @Tags Orders
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order details"
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [post]
func createOrderHandler(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert product_id string to ObjectID
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var newOrder models.Order

	callback := func(sessCtx context.Context) (interface{}, error) {
		db := database.GetDB()
		if db == nil {
			return nil, errors.New("database connection is nil")
		}

		// Fetch the product to ensure it exists and get its details
		var product models.Product
		err := db.Collection("products").Find(sessCtx, bson.M{"_id": productID}).One(&product)
		if err != nil {
			if err == qmgo.ErrNoSuchDocuments {
				return nil, errors.New("product not found")
			}
			return nil, fmt.Errorf("error fetching product: %v", err)
		}

		// Check if there's enough stock
		if product.Stocks < req.Quantity {
			return nil, errors.New("insufficient stock")
		}

		// Calculate the total amount
		totalAmount := float64(req.Quantity) * product.Price

		newOrder = models.Order{
			ID:         primitive.NewObjectID(),
			CustomerID: req.CustomerID,
			Product: models.OrderProduct{
				ID:    product.ID.Hex(),
				Name:  product.Name,
				Price: product.Price,
			},
			Quantity:    req.Quantity,
			TotalAmount: totalAmount,
			Status:      models.OrderStatusCreated,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Timeline: []models.TimelineEvent{
				{
					Name:      "Created",
					Timestamp: time.Now(),
				},
			},
		}

		// Insert the new order
		_, err = db.Collection("orders").InsertOne(sessCtx, newOrder)
		if err != nil {
			return nil, err
		}

		// Update the product stock
		err = db.Collection("products").UpdateOne(sessCtx, bson.M{"_id": productID}, bson.M{
			"$inc": bson.M{"stocks": -req.Quantity},
		})
		if err != nil {
			return nil, err
		}

		return newOrder, nil
	}

	_, err = database.GetClient().DoTransaction(c, callback)

	if err != nil {
		switch {
		case err.Error() == "product not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		case err == qmgo.ErrTransactionNotSupported:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
		default:
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing order: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order": newOrder})
}

// @Summary Cancel an order
// @Description Cancel an existing order (requires authentication)
// @Tags Orders
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{id}/cancel [post]
func cancelOrderHandler(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	ctx := context.Background()
	db := database.GetDB()

	var order models.Order
	err = db.Collection("orders").Find(ctx, bson.M{"_id": orderID}).One(&order)
	if err != nil {
		if err == qmgo.ErrNoSuchDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order"})
		}
		return
	}

	if order.Status != models.OrderStatusCreated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order cannot be cancelled"})
		return
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":     models.OrderStatusCancelled,
			"updated_at": now,
		},
		"$push": bson.M{
			"timeline": models.TimelineEvent{
				Name:      "Cancelled",
				Timestamp: now,
			},
		},
	}

	err = db.Collection("orders").UpdateOne(ctx, bson.M{"_id": orderID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cancelling order"})
		return
	}

	// Convert the product ID string to ObjectID
	productID, err := primitive.ObjectIDFromHex(order.Product.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid product ID in order"})
		return
	}

	// Restore the product stock
	err = db.Collection("products").UpdateOne(ctx, bson.M{"_id": productID}, bson.M{
		"$inc": bson.M{"stocks": order.Quantity},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error restoring product stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}
