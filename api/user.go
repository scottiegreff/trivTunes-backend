package handler

import "net/http"

// Handler is the Vercel entrypoint for /api/user.
func Handler(w http.ResponseWriter, r *http.Request) {
	if err := ensureMongo(); err != nil {
		http.Error(w, "Server initialization error", http.StatusInternalServerError)
		return
	}
	handlers.UserHandler(w, r)
}
