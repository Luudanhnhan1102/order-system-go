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
	"github.com/qiniu/qmgo"
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
	OrderID       string  `json:"order_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
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
	log.Println("createPaymentHandler: Payment request received")
	
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("createPaymentHandler: Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	log.Printf("createPaymentHandler: Processing payment for order %s, amount: %.2f", req.OrderID, req.Amount)

	db := database.GetDB()
	collection := db.Collection("transactions")

	// Check for duplicate request using order_id
	existingTxn := models.Transaction{}
	err := collection.Find(context.Background(), qmgo.M{"order_id": req.OrderID}).One(&existingTxn)

	if err == nil {
		log.Printf("createPaymentHandler: Duplicate request detected for order %s, returning existing transaction", req.OrderID)
		c.JSON(http.StatusOK, existingTxn)
		return
	} else if err != qmgo.ErrNoSuchDocuments {
		log.Printf("createPaymentHandler: Error checking for existing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing transaction"})
		return
	}

	// Create new transaction
	transaction := models.Transaction{
		ID:             primitive.NewObjectID(),
		OrderID:        req.OrderID,
		Amount:         req.Amount,
		Status:         models.TransactionStatusPending,
		IdempotencyKey: req.IdempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Insert the new transaction
	_, err = collection.InsertOne(context.Background(), &transaction)
	if err != nil {
		log.Printf("createPaymentHandler: Failed to create transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Simulate payment processing
	log.Println("createPaymentHandler: Processing payment...")
	time.Sleep(1 * time.Second) // Simulate processing time

	// Mock payment gateway response (80% success, 20% failure)
	if rand.Float32() < 0.8 {
		transaction.Status = models.TransactionStatusCompleted
		log.Printf("createPaymentHandler: Payment for order %s completed successfully", req.OrderID)
	} else {
		transaction.Status = models.TransactionStatusFailed
		log.Printf("createPaymentHandler: Payment for order %s failed", req.OrderID)
	}

	transaction.UpdatedAt = time.Now()

	// Update transaction status
	err = collection.UpdateOne(
		context.Background(),
		qmgo.M{"_id": transaction.ID},
		qmgo.M{"$set": qmgo.M{
			"status":     transaction.Status,
			"updated_at": transaction.UpdatedAt,
		}},
	)

	if err != nil {
		log.Printf("createPaymentHandler: Failed to update transaction status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
		return
	}

	// Notify order service in the background
	go func() {
		if err := notifyOrderService(transaction); err != nil {
			log.Printf("createPaymentHandler: Failed to notify order service: %v", err)
		} else {
			log.Printf("createPaymentHandler: Successfully notified order service for order %s", req.OrderID)
		}
	}()

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
