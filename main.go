package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"

	"context"

	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"trivTunes-backend/handlers" // Import the handlers package
)

var client *mongo.Client

// init function to load .env and setup MongoDB connection
func init() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    // Get MongoDB URI from environment variables
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        log.Fatal("MONGODB_URI is not set in .env file")
    }

    // Setup MongoDB connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatal("MongoDB connection error: ", err)
    }

    // Ping the MongoDB server to check if the connection is alive
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        log.Fatal("MongoDB connection failed: ", err)
    }
    log.Println("Successfully connected to MongoDB")

    // Initialize the user collection in MongoDB
    handlers.InitUserCollection(client)
}
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	allowedOrigin := os.Getenv("TRIVTUNES_CLIENT_URI")
	if allowedOrigin == "" {
		log.Fatal("TRIVTUNES_CLIENT_URI is not set in .env file")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user", handlers.UserHandler)
	handler := c.Handler(mux)
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}