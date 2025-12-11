package handler

import (
	"net/http"

	"trivTunes-backend/internal/mongoinit"
)

// Handler is the Vercel entrypoint for /api/user.
func Handler(w http.ResponseWriter, r *http.Request) {
	if err := mongoinit.EnsureMongo(); err != nil {
		http.Error(w, "Server initialization error", http.StatusInternalServerError)
		return
	}
	handlers.UserHandler(w, r)
}
