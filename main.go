package main

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/truongpnx/mind-forge/internal/config"
	"github.com/truongpnx/mind-forge/internal/handlers"
	pgstore "github.com/truongpnx/mind-forge/internal/store/postgres"
	redisstore "github.com/truongpnx/mind-forge/internal/store/redis"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load("/etc/mindforge/config.toml")
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.PostgreSQL.DSN())
	if err != nil {
		slog.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("postgres ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("connected to postgres")

	// ── Migrations ────────────────────────────────────────────────────────────
	src, err := iofs.New(migrationsFS, "db/migrations")
	if err != nil {
		slog.Error("failed to load migrations source", "err", err)
		os.Exit(1)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.PostgreSQL.DSN())
	if err != nil {
		slog.Error("failed to initialise migrator", "err", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := goredis.NewClient(cfg.Redis.Options())
	if err := redisClient.Ping(ctx).Err(); err != nil {
		slog.Error("redis ping failed", "err", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("connected to redis")

	// ── Stores ────────────────────────────────────────────────────────────────
	userStore := pgstore.NewUserStore(pool)
	sessionStore := redisstore.NewSessionStore(redisClient)

	// ── Auth handlers ─────────────────────────────────────────────────────────
	auth := handlers.NewAuthHandlers(userStore, sessionStore)

	// ── Router ────────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(auth.AuthMiddleware)

	// Static assets
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth routes
	r.Get("/auth/login", auth.LoginPageHandler)
	r.Get("/auth/register", auth.RegisterPageHandler)
	r.Get("/auth/profile", auth.ProfilePageHandler)
	r.Post("/auth/register", auth.RegisterHandler)
	r.Post("/auth/login", auth.LoginHandler)
	r.Post("/auth/logout", auth.LogoutHandler)
	r.Post("/auth/profile/username", auth.UpdateUsernameHandler)
	r.Post("/auth/profile/password", auth.UpdatePasswordHandler)

	// Game routes
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
