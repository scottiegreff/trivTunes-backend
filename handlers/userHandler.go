package handlers

import (
	"context"
	"encoding/json"
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
	switch r.Method {
	case http.MethodGet:
		// Check if we're fetching a single user or multiple users
		if email := r.URL.Query().Get("email"); email != "" {
			handleGetUser(w, r)
		} else {
			handleGetUsers(w, r)
		}
	case http.MethodPost:
		handlePostUser(w, r)
    case http.MethodPatch:
		handleUpdateUserScore(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    // Get the email from the query parameters
    email := r.URL.Query().Get("email")
    if email == "" {
        http.Error(w, "Email is required", http.StatusBadRequest)
        return
    }
    
    // Create a context with a timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Assume userCollection is your MongoDB collection for users
    var user User
    err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
        } else {
            log.Printf("Error fetching user: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    // Set the content type to JSON
    w.Header().Set("Content-Type", "application/json")

    // Encode the user to JSON and send the response
    if err := json.NewEncoder(w).Encode(user); err != nil {
        log.Printf("Error encoding user to JSON: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
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

func handlePostUser(w http.ResponseWriter, r *http.Request) {
    var newUser User

    // Read and parse the JSON body
    if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Check if user already exists
    var existingUser User
    err := userCollection.FindOne(ctx, bson.M{"email": newUser.Email}).Decode(&existingUser)

    if err == nil {
        // User already exists, return the existing user
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(existingUser)
        return
    } else if err != mongo.ErrNoDocuments {
        // An error occurred that wasn't "document not found"
        log.Println("Error checking for existing user:", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // User doesn't exist, insert new user into MongoDB
    _, err = userCollection.InsertOne(ctx, newUser)
    if err != nil {
        log.Println("Error inserting user into MongoDB:", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Respond with the newly created user
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newUser)
}

func handleUpdateUserScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateReq User

	// Read and parse the JSON body
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user and update their score
	filter := bson.M{"email": updateReq.Email}
	update := bson.M{"$set": bson.M{"score": updateReq.Score}}

	var updatedUser User
	err := userCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Error updating user score:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Respond with the updated user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

