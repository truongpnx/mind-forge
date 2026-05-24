package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/truongpnx/mind-forge/internal/handlers"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static assets
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	r.Get("/", handlers.HomeHandler)

	r.Route("/{game}", func(r chi.Router) {
		r.Get("/", handlers.GameEntryHandler)
		r.Get("/how-to", handlers.GameHowToHandler)
		r.Get("/machine", handlers.GameMachineHandler)
		r.Get("/match", handlers.GameMatchHandler)
		r.Get("/leader-boards", handlers.GameLeaderboardsHandler)
	})

	addr := ":8080"
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
