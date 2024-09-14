package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend-payment/database"
	"backend-payment/helpers"
	"backend-payment/models"
)

// SetupPaymentRoutes sets up the payment-related routes
func SetupPaymentRoutes(r *gin.Engine) {
	paymentGroup := r.Group("/payments")
	{
		paymentGroup.POST("", createPaymentHandler)
	}
}

type CreatePaymentRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

// @Summary Create a new payment
// @Description Create a new payment transaction
// @Tags Payments
// @Accept json
// @Produce json
// @Param payment body CreatePaymentRequest true "Payment details"
// @Success 201 {object} models.Transaction
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments [post]
func createPaymentHandler(c *gin.Context) {
	fmt.Println("createPaymentHandler called")
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction := models.Transaction{
		ID:        primitive.NewObjectID(),
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Status:    models.TransactionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db := database.GetDB()
	collection := db.Collection("transactions")

	_, err := collection.InsertOne(context.Background(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Mock payment gateway integration
	if rand.Float32() < 0.8 {
		// 80% chance of success
		transaction.Status = models.TransactionStatusCompleted
	} else {
		// 20% chance of failure
		transaction.Status = models.TransactionStatusFailed
	}
	transaction.UpdatedAt = time.Now()

	err = collection.UpdateOne(
		context.Background(),
		primitive.M{"_id": transaction.ID},
		primitive.M{"$set": primitive.M{"status": transaction.Status, "updated_at": transaction.UpdatedAt}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
		return
	}

	// After updating the transaction status
	if err = notifyOrderService(transaction); err != nil {
		log.Printf("Error notifying order service: %v", err)
	}

	c.JSON(http.StatusCreated, transaction)
}

func notifyOrderService(transaction models.Transaction) error {
	payload := map[string]interface{}{
		"order_id": transaction.OrderID,
		"status":   transaction.Status,
		"amount":   transaction.Amount,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	orderServiceURL := os.Getenv("API_ORDER_URL")
	if orderServiceURL == "" {
		return fmt.Errorf("API_ORDER_URL environment variable is not set")
	}

	req, err := http.NewRequest("POST", orderServiceURL+"/backend/payment-update", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Sign the request
	timestamp := time.Now()
	signature := helpers.SignRequest(req.Method, req.URL.Path, jsonPayload, timestamp)
	req.Header.Set(helpers.SignatureHeader, signature)
	req.Header.Set(helpers.TimestampHeader, timestamp.Format(time.RFC3339))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to order service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("order service responded with status code: %d", resp.StatusCode)
	}

	return nil
}
