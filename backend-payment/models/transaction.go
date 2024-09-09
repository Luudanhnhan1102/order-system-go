package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderID   string             `json:"order_id" bson:"order_id"`
	Amount    float64            `json:"amount" bson:"amount"`
	Status    string             `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

const (
	TransactionStatusPending   = "Pending"
	TransactionStatusCompleted = "Completed"
	TransactionStatusFailed    = "Failed"
)
