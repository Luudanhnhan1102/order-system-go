package main

import (
	"context"
	"fmt"
	"log"

	"backend-order/database"
	"backend-order/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/joho/godotenv"
)

func main() {
	// Get MongoDB client
	errDB := godotenv.Load()
	if errDB != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	client := database.GetClient()
	defer client.Close(context.Background())

	// Get the database
	db := database.GetDB()

	// Create admin user
	adminEmail := "admin@sotatek.com"
	adminPassword := "admin123"

	// Create the admin user
	adminUser := models.User{
		Email:   adminEmail,
		IsAdmin: true,
	}

	// Set the password (this will hash it)
	err := adminUser.SetPassword(adminPassword)
	if err != nil {
		log.Fatal("Failed to set password:", err)
	}

	// Check if the user already exists
	collection := db.Collection("users")
	var existingUser models.User
	err = collection.Find(context.Background(), bson.M{"email": adminEmail}).One(&existingUser)
	if err == nil {
		fmt.Println("Admin user already exists")
		return
	} else if err != mongo.ErrNoDocuments {
		log.Fatal("Error checking for existing user:", err)
	}

	// Save the new admin user
	_, err = collection.InsertOne(context.Background(), adminUser)
	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Println("Admin user created successfully")
}
