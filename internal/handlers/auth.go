package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/truongpnx/mind-forge/internal/models"
	"github.com/truongpnx/mind-forge/internal/store"
	authtmpl "github.com/truongpnx/mind-forge/internal/templates/auth"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie = "session_token"
	sessionTTL    = 24 * time.Hour
	bcryptCost    = 12
)

// contextKey is an unexported type for context keys in this package.
type contextKey int

const userContextKey contextKey = iota

// AuthHandlers groups the HTTP handlers that depend on auth stores.
type AuthHandlers struct {
	users    store.UserStore
	sessions store.SessionStore
}

// NewAuthHandlers constructs an AuthHandlers.
func NewAuthHandlers(users store.UserStore, sessions store.SessionStore) *AuthHandlers {
	return &AuthHandlers{users: users, sessions: sessions}
}

// ── Register ─────────────────────────────────────────────────────────────────

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterHandler handles POST /auth/register.
func (a *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "username, email, and password are required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err := a.users.CreateUser(r.Context(), req.Username, req.Email, string(hash))
	if err != nil {
		// Surface duplicate-key violations as a 409 so the client can react.
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "username or email already taken", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := a.createSession(r.Context(), w, user.ID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
		"id":       user.ID,
		"username": user.Username,
	})
}

// ── Login ─────────────────────────────────────────────────────────────────────

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginHandler handles POST /auth/login.
func (a *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	user, err := a.users.GetUserByEmail(r.Context(), strings.TrimSpace(req.Email))
	if err != nil {
		// Return the same message for "not found" and "wrong password" to
		// avoid disclosing whether the account exists.
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := a.createSession(r.Context(), w, user.ID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
		"id":       user.ID,
		"username": user.Username,
	})
}

// ── Logout ────────────────────────────────────────────────────────────────────

// LogoutHandler handles POST /auth/logout.
func (a *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err == nil {
		_ = a.sessions.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// ── Middleware ────────────────────────────────────────────────────────────────

// AuthMiddleware reads the session_token cookie, resolves it against Redis,
// and injects the authenticated *models.User into the request context.
// Unauthenticated requests are passed through with a nil user — no hard block.
func (a *AuthHandlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := a.sessions.GetSession(r.Context(), cookie.Value)
		if err != nil || userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Load the full user so Username is available in every template (e.g. navbar).
		user, err := a.users.GetUserByID(r.Context(), userID)
		if err != nil {
			// Session points to a deleted user — clear cookie and continue as guest.
			_ = a.sessions.DeleteSession(r.Context(), cookie.Value)
			http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserFromContext returns the authenticated user from ctx, or nil.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(userContextKey).(*models.User)
	return u
}

// ── helpers ───────────────────────────────────────────────────────────────────

// createSession generates a new token, stores it in Redis, and sets the cookie.
func (a *AuthHandlers) createSession(ctx context.Context, w http.ResponseWriter, userID string) error {
	token := uuid.New().String()
	if err := a.sessions.SetSession(ctx, token, userID, sessionTTL); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// ── Page handlers ─────────────────────────────────────────────────────────────

// LoginPageHandler handles GET /auth/login — renders the login form.
func (a *AuthHandlers) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	if UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := authtmpl.Login("").Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// RegisterPageHandler handles GET /auth/register — renders the registration form.
func (a *AuthHandlers) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	if UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := authtmpl.Register().Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// ProfilePageHandler handles GET /auth/profile — renders the profile management page.
func (a *AuthHandlers) ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	ctxUser := UserFromContext(r.Context())
	if ctxUser == nil {
		http.Redirect(w, r, "/auth/login?next=/auth/profile", http.StatusSeeOther)
		return
	}
	// Fetch full user record (middleware stores ID only).
	user, err := a.users.GetUserByID(r.Context(), ctxUser.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := authtmpl.Profile(user).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// ── Profile update handlers ───────────────────────────────────────────────────

type updateUsernameRequest struct {
	Username string `json:"username"`
}

// UpdateUsernameHandler handles POST /auth/profile/username.
func (a *AuthHandlers) UpdateUsernameHandler(w http.ResponseWriter, r *http.Request) {
	ctxUser := UserFromContext(r.Context())
	if ctxUser == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req updateUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}
	if err := a.users.UpdateUsername(r.Context(), ctxUser.ID, req.Username); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UpdatePasswordHandler handles POST /auth/profile/password.
func (a *AuthHandlers) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctxUser := UserFromContext(r.Context())
	if ctxUser == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req updatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 8 {
		http.Error(w, "new password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	// Verify current password.
	user, err := a.users.GetUserByID(r.Context(), ctxUser.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "current password is incorrect", http.StatusUnauthorized)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := a.users.UpdatePassword(r.Context(), ctxUser.ID, string(hash)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
