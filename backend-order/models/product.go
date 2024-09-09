package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a product in the store
type Product struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name   string             `json:"name" bson:"name"`
	Price  float64            `json:"price" bson:"price"`
	Stocks int                `json:"stocks" bson:"stocks"`
}
