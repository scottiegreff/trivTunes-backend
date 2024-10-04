package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Score int32 `json:"score"`
}

var userCollection *mongo.Collection

// Initialize the user collection (MongoDB) and create index on the "score" field
func InitUserCollection(client *mongo.Client) {
    userCollection = client.Database("trivTunes").Collection("users")

    // Create an index on the "score" field in descending order
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Define the index model for the "score" field
    indexModel := mongo.IndexModel{
        Keys: bson.D{{Key: "score", Value: -1}},  // Create an index on the "score" field (descending)
        Options: options.Index().SetName("score_index"),
    }

    // Create the index on the userCollection
    indexName, err := userCollection.Indexes().CreateOne(ctx, indexModel)
    if err != nil {
        log.Fatal("Error creating index on score field: ", err)
    }

    log.Println("Index created: ", indexName)
}

// UserHandler handles both GET and POST requests
func UserHandler(w http.ResponseWriter, r *http.Request) {
    // Enable CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    // Handle different HTTP methods
    switch r.Method {
    case http.MethodGet:
        handleGetUsers(w, r) // Handle GET
    case http.MethodPost:
        handlePostUser(w, r) // Handle POST
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// Handle GET request to retrieve the top 50 users with the highest score from MongoDB
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
    var users []User

    // Set a context for the MongoDB query
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // MongoDB query: Find all users, sort by score in descending order, and limit to 50
    options := options.Find()
    options.SetSort(bson.D{{Key: "score", Value: -1}}) // Sort by score descending
    options.SetLimit(50)                   // Limit to 50 users

    cursor, err := userCollection.Find(ctx, bson.D{}, options) // Empty filter (bson.D{}) to get all users
    if err != nil {
        log.Println("Error fetching users:", err)
        http.Error(w, "Error fetching users from the database", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    // Iterate through the cursor to decode each user document
    for cursor.Next(ctx) {
        var user User
        if err := cursor.Decode(&user); err != nil {
            log.Println("Error decoding user:", err)
            http.Error(w, "Error decoding user data", http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    // Check if the cursor encountered any errors
    if err := cursor.Err(); err != nil {
        log.Println("Cursor error:", err)
        http.Error(w, "Cursor error", http.StatusInternalServerError)
        return
    }

    // Return the list of users as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}


// Handle POST request to create a new user
func handlePostUser(w http.ResponseWriter, r *http.Request) {
    var newUser User

    // Read the request body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Unable to read request body", http.StatusBadRequest)
        return
    }

    // Unmarshal JSON to User struct
    if err := json.Unmarshal(body, &newUser); err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    // Insert new user into MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = userCollection.InsertOne(ctx, bson.M{
        "name":  newUser.Name,
        "email": newUser.Email,
        "score": newUser.Score,
    })
    if err != nil {
        log.Println("Error inserting user into MongoDB:", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Respond with success message
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}
