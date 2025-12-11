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
	Score int32  `json:"score"`
	D1950 int32  `json:"d1950" bson:"d1950"`
	D1960 int32  `json:"d1960" bson:"d1960"`
	D1970 int32  `json:"d1970" bson:"d1970"`
	D1980 int32  `json:"d1980" bson:"d1980"`
	D1990 int32  `json:"d1990" bson:"d1990"`
	D2000 int32  `json:"d2000" bson:"d2000"`
	D2010 int32  `json:"d2010" bson:"d2010"`
	D2020 int32  `json:"d2020" bson:"d2020"`
}

var userCollection *mongo.Collection

// Initialize the user collection and indexes.
func InitUserCollection(client *mongo.Client) {
	userCollection = client.Database("trivTunes").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Indexes for overall score and per-decade ranking
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "score", Value: -1}}, Options: options.Index().SetName("score_desc")},
		{Keys: bson.D{{Key: "d1950", Value: -1}}},
		{Keys: bson.D{{Key: "d1960", Value: -1}}},
		{Keys: bson.D{{Key: "d1970", Value: -1}}},
		{Keys: bson.D{{Key: "d1980", Value: -1}}},
		{Keys: bson.D{{Key: "d1990", Value: -1}}},
		{Keys: bson.D{{Key: "d2000", Value: -1}}},
		{Keys: bson.D{{Key: "d2010", Value: -1}}},
		{Keys: bson.D{{Key: "d2020", Value: -1}}},
	}

	if _, err := userCollection.Indexes().CreateMany(ctx, indexes); err != nil {
		log.Fatal("Error creating indexes: ", err)
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
		} else {
			log.Printf("Error fetching user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding user to JSON: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// Handle GET request to retrieve the top 50 users with the highest score from MongoDB
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	var users []User

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: -1}}).
		SetLimit(50)

	cursor, err := userCollection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println("Error fetching users:", err)
		http.Error(w, "Error fetching users from the database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			log.Println("Error decoding user:", err)
			http.Error(w, "Error decoding user data", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		log.Println("Cursor error:", err)
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

func handlePostUser(w http.ResponseWriter, r *http.Request) {
	var newUser User

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingUser User
	err := userCollection.FindOne(ctx, bson.M{"email": newUser.Email}).Decode(&existingUser)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(existingUser)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Println("Error checking for existing user:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = userCollection.InsertOne(ctx, bson.M{
		"name":  newUser.Name,
		"email": newUser.Email,
		"score": newUser.Score,
		"d1950": 0,
		"d1960": 0,
		"d1970": 0,
		"d1980": 0,
		"d1990": 0,
		"d2000": 0,
		"d2010": 0,
		"d2020": 0,
	})
	if err != nil {
		log.Println("Error inserting user:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newUser)
}

func handleUpdateUserScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateReq struct {
		Email  string `json:"email"`
		Score  int32  `json:"score"`
		Decade string `json:"decade,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	decadeCol := mapDecadeToColumn(updateReq.Decade)
	updateOps := bson.D{{Key: "$set", Value: bson.M{"score": updateReq.Score}}}
	if decadeCol != "" {
		updateOps = append(updateOps, bson.E{Key: "$inc", Value: bson.M{decadeCol: 1}})
	}

	var updatedUser User
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := userCollection.FindOneAndUpdate(ctx, bson.M{"email": updateReq.Email}, updateOps, opts).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Error updating user score:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updatedUser)
}

func mapDecadeToColumn(decade string) string {
	switch decade {
	case "1950s":
		return "d1950"
	case "1960s":
		return "d1960"
	case "1970s":
		return "d1970"
	case "1980s":
		return "d1980"
	case "1990s":
		return "d1990"
	case "2000s":
		return "d2000"
	case "2010s":
		return "d2010"
	case "2020s":
		return "d2020"
	default:
		return ""
	}
}

// LeaderboardHandler returns top players overall and per-decade.
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	decades := []string{"1950s", "1960s", "1970s", "1980s", "1990s", "2000s", "2010s", "2020s"}
	decadeCols := []string{"d1950", "d1960", "d1970", "d1980", "d1990", "d2000", "d2010", "d2020"}

	resp := struct {
		Overall  []User            `json:"overall"`
		ByDecade map[string][]User `json:"byDecade"`
	}{
		Overall:  []User{},
		ByDecade: make(map[string][]User),
	}

	// Overall top 5 by score
	if cur, err := userCollection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "score", Value: -1}}).SetLimit(5)); err == nil {
		for cur.Next(ctx) {
			var u User
			if err := cur.Decode(&u); err == nil {
				resp.Overall = append(resp.Overall, u)
			}
		}
		cur.Close(ctx)
	}

	// Top 5 per decade
	for i, dec := range decades {
		col := decadeCols[i]
		filter := bson.M{col: bson.M{"$gt": 0}}
		opts := options.Find().
			SetSort(bson.D{{Key: col, Value: -1}, {Key: "score", Value: -1}}).
			SetLimit(5)

		if cur, err := userCollection.Find(ctx, filter, opts); err == nil {
			var list []User
			for cur.Next(ctx) {
				var u User
				if err := cur.Decode(&u); err == nil {
					list = append(list, u)
				}
			}
			cur.Close(ctx)
			resp.ByDecade[dec] = list
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
