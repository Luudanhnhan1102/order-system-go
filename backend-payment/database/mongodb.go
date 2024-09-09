package database

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/qiniu/qmgo"
)

var (
	client *qmgo.Client
	once   sync.Once
)

func GetClient() *qmgo.Client {
	once.Do(func() {
		var err error
		mongoURI := os.Getenv("MONGODB_URI")
		if mongoURI == "" {
			log.Fatal("MONGODB_URI environment variable is not set")
		}
		client, err = qmgo.NewClient(context.Background(), &qmgo.Config{Uri: mongoURI})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		log.Println("Connected to MongoDB")
	})
	return client
}

// GetDB returns a singleton instance of the database connection
func GetDB() *qmgo.Database {
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "backend-payment" // Default database name
	}
	return GetClient().Database(dbName)
}
