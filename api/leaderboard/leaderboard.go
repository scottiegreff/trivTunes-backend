package handler

import (
	"net/http"

	"trivTunes-backend/handlers"
	"trivTunes-backend/internal/mongoinit"
)

// Handler is the Vercel entrypoint for /api/leaderboard.
func Handler(w http.ResponseWriter, r *http.Request) {
	if err := mongoinit.EnsureMongo(); err != nil {
		http.Error(w, "Server initialization error", http.StatusInternalServerError)
		return
	}
	handlers.LeaderboardHandler(w, r)
}
