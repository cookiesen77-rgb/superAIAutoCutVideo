package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/auth"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	cfg    config.Config
	appDB  *pgxpool.Pool
	userDB *pgxpool.Pool
}

func NewAuthHandler(cfg config.Config, appDB *pgxpool.Pool, userDB *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{cfg: cfg, appDB: appDB, userDB: userDB}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}
	if len(password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password too short"})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "hash error"})
		return
	}

	userID, err := createUser(r.Context(), h.userDB, email, hash)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create user failed"})
		return
	}

	h.issueTokens(w, r.Context(), userID)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := getUserByEmail(r.Context(), h.userDB, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}

	if user.Status != "" && user.Status != "active" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "account disabled"})
		return
	}
	if !auth.CheckPassword(user.PasswordHash, password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	h.issueTokens(w, r.Context(), user.ID)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh token required"})
		return
	}
	claims, err := auth.ParseToken(req.RefreshToken, []byte(h.cfg.JWTSecret))
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}

	tokenHash := auth.HashRefreshToken(req.RefreshToken)
	var tokenID string
	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time
	err = h.appDB.QueryRow(r.Context(), `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash=$1
	`, tokenHash).Scan(&tokenID, &userID, &expiresAt, &revokedAt)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "refresh token not found"})
		return
	}
	if revokedAt != nil || time.Now().After(expiresAt) || userID != claims.UserID {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "refresh token expired"})
		return
	}

	_, _ = h.appDB.Exec(r.Context(), `
		UPDATE refresh_tokens SET revoked_at=NOW() WHERE id=$1
	`, tokenID)

	h.issueTokens(w, r.Context(), userID)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh token required"})
		return
	}
	tokenHash := auth.HashRefreshToken(req.RefreshToken)
	_, _ = h.appDB.Exec(r.Context(), `
		UPDATE refresh_tokens SET revoked_at=NOW() WHERE token_hash=$1
	`, tokenHash)
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	user, err := getUserByID(r.Context(), h.userDB, userID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"user_id": userID})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":  user.ID,
		"email":    user.Email,
		"is_admin": user.IsAdmin,
		"status":   user.Status,
	})
}

func (h *AuthHandler) issueTokens(w http.ResponseWriter, ctx context.Context, userID string) {
	pair, err := auth.GenerateTokenPair(userID, []byte(h.cfg.JWTSecret), h.cfg.AccessTokenTTL, h.cfg.RefreshTokenTTL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
		return
	}
	tokenHash := auth.HashRefreshToken(pair.RefreshToken)
	_, _ = h.appDB.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.NewString(), userID, tokenHash, time.Now().Add(h.cfg.RefreshTokenTTL))

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
		"token_type":    "Bearer",
	})
}

func (h *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")
		claims, err := auth.ParseToken(token, []byte(h.cfg.JWTSecret))
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey{}, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) AdminOnly(next http.Handler) http.Handler {
	return h.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		user, err := getUserByID(r.Context(), h.userDB, userID)
		if err != nil || user.Status == "disabled" || !user.IsAdmin {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	}))
}
