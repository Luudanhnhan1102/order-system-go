package backend

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend-order/database"
	"backend-order/models"
)

// SetupBackendPaymentRoutes sets up the payment-related routes for backend communication
func SetupBackendPaymentRoutes(r *gin.Engine) {
	backendGroup := r.Group("/backend")
	{
		backendGroup.POST("/payment-update", handlePaymentUpdate)
	}
}

type PaymentUpdateRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Status  string  `json:"status" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
}

// @Summary Update order payment status
// @Description Update the payment status of an order (backend communication)
// @Tags Backend
// @Accept json
// @Produce json
// @Param payment body PaymentUpdateRequest true "Payment update details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /backend/payment-update [post]
func handlePaymentUpdate(c *gin.Context) {
	var req PaymentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	collection := db.Collection("orders")

	orderID, err := primitive.ObjectIDFromHex(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	now := time.Now()
	var update bson.M

	if req.Status == "Completed" {
		update = bson.M{
			"$set": bson.M{
				"status":      models.OrderStatusConfirmed,
				"paid_amount": req.Amount,
				"updated_at":  now,
			},
			"$push": bson.M{
				"timeline": models.TimelineEvent{
					Name:      "Payment Completed",
					Timestamp: now,
				},
			},
		}
	} else {
		// If payment failed, don't change the order status
		update = bson.M{
			"$set": bson.M{
				"updated_at": now,
			},
			"$push": bson.M{
				"timeline": models.TimelineEvent{
					Name:      "Payment Failed",
					Timestamp: now,
				},
			},
		}
	}

	result, err := collection.UpdateAll(c, bson.M{"_id": orderID}, update)

	if err != nil {
		log.Printf("Error updating order [%s] payment status: %v", req.OrderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Order [%s] payment status update failed: %v", req.OrderID, err)})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Order [%s] not found", req.OrderID)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order payment status updated successfully"})
}
