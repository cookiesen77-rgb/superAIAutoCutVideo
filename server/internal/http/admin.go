package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/auth"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminHandler struct {
	cfg    config.Config
	appDB  *pgxpool.Pool
	userDB *pgxpool.Pool
}

func NewAdminHandler(cfg config.Config, appDB *pgxpool.Pool, userDB *pgxpool.Pool) *AdminHandler {
	return &AdminHandler{cfg: cfg, appDB: appDB, userDB: userDB}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	users, err := listUsers(r.Context(), h.userDB, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	resp := make([]map[string]any, 0, len(users))
	for _, user := range users {
		resp = append(resp, map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"is_admin":   user.IsAdmin,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      resp,
		"message":   "获取用户成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	user, err := getUserByID(r.Context(), h.userDB, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"is_admin":   user.IsAdmin,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		},
		"message":   "获取用户成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var payload struct {
		Email   *string `json:"email"`
		IsAdmin *bool   `json:"is_admin"`
		Status  *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if payload.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*payload.Status))
		if status != "active" && status != "disabled" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
			return
		}
		payload.Status = &status
	}
	user, err := updateUser(r.Context(), h.userDB, userID, payload.Email, payload.IsAdmin, payload.Status)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if payload.Status != nil && strings.ToLower(strings.TrimSpace(*payload.Status)) == "disabled" {
		_ = h.revokeUserTokens(r.Context(), userID)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"is_admin":   user.IsAdmin,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		},
		"message":   "更新用户成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *AdminHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var payload struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Password) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password required"})
		return
	}
	if len(payload.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password too short"})
		return
	}
	if err := ensureUserExists(r.Context(), h.userDB, userID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	hash, err := auth.HashPassword(payload.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "hash error"})
		return
	}
	if err := updateUserPassword(r.Context(), h.userDB, userID, hash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}
	_ = h.revokeUserTokens(r.Context(), userID)
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "密码已重置",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *AdminHandler) GetUserUsage(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	plan, sub, err := ensureSubscription(r.Context(), h.appDB, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "billing unavailable"})
		return
	}
	used, err := usageTotals(r.Context(), h.appDB, userID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	quota := map[string]int{
		"llm_chars": plan.MonthlyLLMQuota,
		"tts_chars": plan.MonthlyTTSQuota,
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": usageSummary{
			PeriodStart:  sub.CurrentPeriodStart,
			PeriodEnd:    sub.CurrentPeriodEnd,
			Used:         used,
			Quota:        quota,
			Plan:         plan,
			Subscription: sub,
		},
		"message":   "获取用量成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *AdminHandler) UpdateUserPlan(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var payload struct {
		PlanID string `json:"plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.PlanID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "plan_id required"})
		return
	}
	plan, err := getPlan(r.Context(), h.appDB, payload.PlanID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "plan not found"})
		return
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	_, err = h.appDB.Exec(r.Context(), `
		INSERT INTO subscriptions (user_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end)
		VALUES ($1, $2, 'active', $3, $4, FALSE)
		ON CONFLICT (user_id)
		DO UPDATE SET plan_id=EXCLUDED.plan_id, status='active', current_period_start=$3, current_period_end=$4, updated_at=NOW()
	`, userID, plan.ID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "套餐已更新",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *AdminHandler) revokeUserTokens(ctx context.Context, userID string) error {
	_, err := h.appDB.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at=NOW() WHERE user_id=$1 AND revoked_at IS NULL
	`, userID)
	return err
}
