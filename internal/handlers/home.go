package handlers

import (
	"net/http"

	"github.com/truongpnx/mind-forge/internal/models"
	"github.com/truongpnx/mind-forge/internal/templates/home"
)

// HomeHandler serves the root page listing all games.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := home.Index(models.AllGames, user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
