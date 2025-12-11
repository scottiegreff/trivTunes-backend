package handler

import "net/http"

// Handler is the Vercel entrypoint for /api/leaderboard.
func Handler(w http.ResponseWriter, r *http.Request) {
	if err := ensureMongo(); err != nil {
		http.Error(w, "Server initialization error", http.StatusInternalServerError)
		return
	}
	handlers.LeaderboardHandler(w, r)
}
