package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"

    "trivTunes-backend/handlers"  // Import the handlers package
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
    // Start HTTP server
    http.HandleFunc("/api/user", handlers.UserHandler)  // Use the imported UserHandler
    log.Println("Starting server on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
