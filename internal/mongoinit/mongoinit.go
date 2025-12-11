package mongoinit

import (
	"context"
	"fmt"
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

// EnsureMongo initializes a shared Mongo client once per cold start and wires handler collections.
func EnsureMongo() error {
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
