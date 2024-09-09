package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderProduct struct {
	ID    string  `json:"id" bson:"id"`
	Name  string  `json:"name" bson:"name"`
	Price float64 `json:"price" bson:"price"`
}

type TimelineEvent struct {
	Name      string    `json:"name" bson:"name"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type Order struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerID  string             `json:"customer_id" bson:"customer_id"`
	Product     OrderProduct       `json:"product" bson:"product"`
	Quantity    int                `json:"quantity" bson:"quantity"`
	TotalAmount float64            `json:"total_amount" bson:"total_amount"`
	Status      string             `json:"status" bson:"status"`
	PaymentID   string             `json:"payment_id,omitempty" bson:"payment_id,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	Timeline    []TimelineEvent    `json:"timeline" bson:"timeline"`
}

const (
	OrderStatusCreated   = "Created"
	OrderStatusConfirmed = "Confirmed"
	OrderStatusCancelled = "Cancelled"
	OrderStatusDelivered = "Delivered"
)
