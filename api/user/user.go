package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"trivTunes-backend/handlers"
)

var (
	client   *mongo.Client
	initOnce sync.Once
	initErr  error
)

func ensureMongo() error {
	initOnce.Do(func() {
		_ = godotenv.Load()
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			initErr = fmt.Errorf("MONGODB_URI is not set")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, initErr = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if initErr != nil {
			return
		}
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			initErr = err
			return
		}
		handlers.InitUserCollection(client)
	})
	return initErr
}

// Handler is the Vercel entrypoint for /api/user.
func Handler(w http.ResponseWriter, r *http.Request) {
	if err := ensureMongo(); err != nil {
		http.Error(w, "Server initialization error", http.StatusInternalServerError)
		return
	}
	handlers.UserHandler(w, r)
}
