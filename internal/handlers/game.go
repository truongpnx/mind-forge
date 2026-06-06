package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/truongpnx/mind-forge/internal/models"
	"github.com/truongpnx/mind-forge/internal/templates/game"
)

// resolveGame extracts the game slug from the URL and looks it up in the registry.
// It writes a 404 response and returns false when the game is not found.
func resolveGame(w http.ResponseWriter, r *http.Request) (models.Game, bool) {
	slug := models.GameSlug(chi.URLParam(r, "game"))
	g, ok := models.FindGame(slug)
	if !ok {
		http.NotFound(w, r)
		return models.Game{}, false
	}
	return g, true
}

// GameEntryHandler serves /{game-name} — the game's entry/landing screen.
func GameEntryHandler(w http.ResponseWriter, r *http.Request) {
	g, ok := resolveGame(w, r)
	if !ok {
		return
	}
	user := UserFromContext(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := game.Entry(g, user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// GameHowToHandler serves /{game-name}/how-to.
func GameHowToHandler(w http.ResponseWriter, r *http.Request) {
	g, ok := resolveGame(w, r)
	if !ok {
		return
	}
	user := UserFromContext(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := game.HowTo(g, user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// GameMachineHandler serves /{game-name}/machine — play vs machine.
func GameMachineHandler(w http.ResponseWriter, r *http.Request) {
	g, ok := resolveGame(w, r)
	if !ok {
		return
	}
	user := UserFromContext(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := game.Machine(g, user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// GameMatchHandler serves /{game-name}/match — play vs human.
func GameMatchHandler(w http.ResponseWriter, r *http.Request) {
	g, ok := resolveGame(w, r)
	if !ok {
		return
	}
	user := UserFromContext(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := game.Match(g, user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// GameLeaderboardsHandler serves /{game-name}/leader-boards.
func GameLeaderboardsHandler(w http.ResponseWriter, r *http.Request) {
	g, ok := resolveGame(w, r)
	if !ok {
		return
	}
	user := UserFromContext(r.Context())
	// TODO: fetch real entries from a data store.
	entries := []game.LeaderboardEntry{}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := game.Leaderboards(g, entries, user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
