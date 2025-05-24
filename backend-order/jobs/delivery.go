package jobs

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"backend-order/database"
	"backend-order/models"
)

func DeliverConfirmedOrders() {
	ctx := context.Background()
	db := database.GetDB()
	collection := db.Collection("orders")

	now := time.Now()

	// Find orders that are in "Confirmed" status and older than 3 seconds
	filter := bson.M{
		"status":     models.OrderStatusConfirmed,
		"updated_at": bson.M{"$lt": now.Add(-3 * time.Second)},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.OrderStatusDelivered,
			"updated_at": now,
		},
		"$push": bson.M{
			"timeline": models.TimelineEvent{
				Name:      "Delivered",
				Timestamp: now,
			},
		},
	}

	result, err := collection.UpdateAll(ctx, filter, update)
	if err != nil {
		log.Printf("Error delivering confirmed orders: %v", err)
		return
	}

	log.Printf("Delivered %d confirmed orders", result.ModifiedCount)
}
