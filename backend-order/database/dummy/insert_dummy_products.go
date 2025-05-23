package main

import (
	"context"
	"log"

	"backend-order/database"
	"backend-order/models"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	errDB := godotenv.Load()
	if errDB != nil {
		log.Println("Error loading .env file, using environment variables")
	}
	
	ctx := context.Background()
	db := database.GetDB()
	collection := db.Collection("products")

	// Check if products already exist
	count, err := collection.Find(ctx, bson.M{}).Count()
	if err != nil {
		log.Fatalf("Error checking product count: %v", err)
	}

	// If products exist, don't insert dummy data
	if count > 0 {
		log.Println("Products already exist, skipping dummy data insertion")
		return
	}

	dummyProducts := []models.Product{
		{Name: "Laptop", Price: 999.99, Stocks: 50},
		{Name: "Smartphone", Price: 499.99, Stocks: 100},
		{Name: "Headphones", Price: 99.99, Stocks: 200},
		{Name: "Tablet", Price: 299.99, Stocks: 75},
		{Name: "Smartwatch", Price: 199.99, Stocks: 150},
	}

	_, err = collection.InsertMany(ctx, dummyProducts)
	if err != nil {
		log.Fatalf("Error inserting dummy products: %v", err)
	}

	log.Println("Dummy products inserted successfully")
}
