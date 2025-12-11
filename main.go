package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"handler/handlers"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client

func main() {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set in environment")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection error: ", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("MongoDB connection failed: ", err)
	}
	log.Println("Connected to MongoDB")

	handlers.InitUserCollection(client)

	clientURI := os.Getenv("TRIVTUNES_CLIENT_URI")
	// debug, _ := strconv.ParseBool(os.Getenv("DEBUG", "true"))

	allowedOrigin := clientURI
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
		log.Println("TRIVTUNES_CLIENT_URI not set, defaulting CORS to", allowedOrigin)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user", handlers.UserHandler)
	mux.HandleFunc("/api/leaderboard", handlers.LeaderboardHandler)

	handler := c.Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Println("Starting server on", addr, "with CORS origin", allowedOrigin)
	log.Fatal(http.ListenAndServe(addr, handler))
}
