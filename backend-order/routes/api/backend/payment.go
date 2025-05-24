package backend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qiniu/qmgo"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend-order/database"
	"backend-order/helpers"
	"backend-order/models"
)

// Global circuit breaker for payment operations
var (
	paymentCB = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "payment",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	})

	// In-memory idempotency key storage (replace with Redis in production)
	idempotencyStore = struct {
		sync.RWMutex
		m map[string]bool
	}{
		m: make(map[string]bool),
	}
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
// checkIdempotency checks if the request is a duplicate using the idempotency key
func checkIdempotency(c *gin.Context) (string, bool) {
	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey == "" {
		// Generate a new idempotency key if not provided
		idempotencyKey = uuid.New().String()
		return idempotencyKey, false
	}

	idempotencyStore.RLock()
	_, exists := idempotencyStore.m[idempotencyKey]
	idempotencyStore.RUnlock()

	return idempotencyKey, exists
}

// markIdempotency marks a request as processed
func markIdempotency(key string) {
	idempotencyStore.Lock()
	defer idempotencyStore.Unlock()
	idempotencyStore.m[key] = true
}

// processPaymentUpdate handles the actual payment update logic with transaction
func processPaymentUpdate(c *gin.Context, req PaymentUpdateRequest, orderID primitive.ObjectID) error {
	db := database.GetDB()
	now := time.Now()

	// Define the transaction callback
	callback := func(sessCtx context.Context) (interface{}, error) {
		collection := db.Collection("orders")

		// Find the order first to check its current status
		var order models.Order
		err := collection.Find(sessCtx, bson.M{"_id": orderID}).One(&order)
		if err != nil {
			if err == qmgo.ErrNoSuchDocuments {
				return nil, fmt.Errorf("order not found")
			}
			return nil, fmt.Errorf("failed to find order: %v", err)
		}

		// Prepare update based on payment status
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

		// Update the order
		err = collection.UpdateOne(sessCtx, bson.M{"_id": orderID}, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update order: %v", err)
		}

		return nil, nil
	}

	// Get the client and execute the transaction
	client := database.GetClient()
	_, err := client.DoTransaction(c, callback)
	return err
}

func handlePaymentUpdate(c *gin.Context) {
	// Check idempotency
	idempotencyKey, isDuplicate := checkIdempotency(c)
	if isDuplicate {
		c.JSON(http.StatusConflict, gin.H{
			"error": "This request has already been processed",
			"idempotency_key": idempotencyKey,
		})
		return
	}

	// Read and verify request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if !helpers.VerifySignature(c.Request, body) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse request
	var req PaymentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate order ID
	orderID, err := primitive.ObjectIDFromHex(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Execute with circuit breaker
	_, err = paymentCB.Execute(func() (interface{}, error) {
		err := processPaymentUpdate(c, req, orderID)
		return nil, err
	})

	// Handle the result
	if err != nil {
		switch {
		case err.Error() == "order not found":
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Order [%s] not found", req.OrderID)})
		case paymentCB.State() == gobreaker.StateOpen:
			// Circuit breaker is open
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
				"retry_after": "30s",
			})
		default:
			log.Printf("Error processing payment update for order [%s]: %v", req.OrderID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Failed to process payment update",
				"details": err.Error(),
			})
		}
		return
	}

	// Mark request as processed
	markIdempotency(idempotencyKey)

	// Return success response with idempotency key
	c.Header("X-Idempotency-Key", idempotencyKey)
	c.JSON(http.StatusOK, gin.H{
		"message":        "Order payment status updated successfully",
		"idempotency_key": idempotencyKey,
	})
}
