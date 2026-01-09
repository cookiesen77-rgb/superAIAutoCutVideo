package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingHandler struct {
	db *pgxpool.Pool
}

func NewBillingHandler(db *pgxpool.Pool) *BillingHandler {
	return &BillingHandler{db: db}
}

type planInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	MonthlyLLMQuota int    `json:"monthly_llm_quota"`
	MonthlyTTSQuota int    `json:"monthly_tts_quota"`
	PriceCents      int    `json:"price_cents"`
	Currency        string `json:"currency"`
}

type subscriptionInfo struct {
	PlanID             string    `json:"plan_id"`
	Status             string    `json:"status"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CancelAtPeriodEnd  bool      `json:"cancel_at_period_end"`
}

type usageSummary struct {
	PeriodStart  time.Time        `json:"period_start"`
	PeriodEnd    time.Time        `json:"period_end"`
	Used         map[string]int   `json:"used"`
	Quota        map[string]int   `json:"quota"`
	Plan         planInfo         `json:"plan"`
	Subscription subscriptionInfo `json:"subscription"`
}

func (h *BillingHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	plan, sub, err := ensureSubscription(r.Context(), h.db, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "billing unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"plan":         plan,
			"subscription": sub,
		},
		"message": "获取计费信息成功",
	})
}

func (h *BillingHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	plan, sub, err := ensureSubscription(r.Context(), h.db, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "billing unavailable"})
		return
	}
	used, err := usageTotals(r.Context(), h.db, userID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	quota := map[string]int{
		"llm_chars": plan.MonthlyLLMQuota,
		"tts_chars": plan.MonthlyTTSQuota,
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": usageSummary{
			PeriodStart:  sub.CurrentPeriodStart,
			PeriodEnd:    sub.CurrentPeriodEnd,
			Used:         used,
			Quota:        quota,
			Plan:         plan,
			Subscription: sub,
		},
		"message": "获取用量成功",
	})
}

func ensureSubscription(ctx context.Context, db *pgxpool.Pool, userID string) (planInfo, subscriptionInfo, error) {
	plan, err := getPlan(ctx, db, "free")
	if err != nil {
		return planInfo{}, subscriptionInfo{}, err
	}
	sub, err := getSubscription(ctx, db, userID)
	if errors.Is(err, errSubscriptionNotFound) {
		sub, err = createSubscription(ctx, db, userID, plan.ID)
	}
	if err != nil {
		return planInfo{}, subscriptionInfo{}, err
	}
	now := time.Now().UTC()
	if !now.Before(sub.CurrentPeriodEnd) {
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		sub.CurrentPeriodStart = start
		sub.CurrentPeriodEnd = end
		_, err = db.Exec(ctx, `
			UPDATE subscriptions
			SET current_period_start=$2, current_period_end=$3, updated_at=NOW()
			WHERE user_id=$1
		`, userID, start, end)
		if err != nil {
			return planInfo{}, subscriptionInfo{}, err
		}
	}
	return plan, sub, nil
}

func getPlan(ctx context.Context, db *pgxpool.Pool, planID string) (planInfo, error) {
	var p planInfo
	err := db.QueryRow(ctx, `
		SELECT id, name, monthly_llm_quota, monthly_tts_quota, price_cents, currency
		FROM plans
		WHERE id=$1
	`, planID).Scan(&p.ID, &p.Name, &p.MonthlyLLMQuota, &p.MonthlyTTSQuota, &p.PriceCents, &p.Currency)
	return p, err
}

var errSubscriptionNotFound = errors.New("subscription not found")

func getSubscription(ctx context.Context, db *pgxpool.Pool, userID string) (subscriptionInfo, error) {
	var s subscriptionInfo
	err := db.QueryRow(ctx, `
		SELECT plan_id, status, current_period_start, current_period_end, cancel_at_period_end
		FROM subscriptions
		WHERE user_id=$1
	`, userID).Scan(&s.PlanID, &s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CancelAtPeriodEnd)
	if err != nil {
		return subscriptionInfo{}, errSubscriptionNotFound
	}
	return s, nil
}

func createSubscription(ctx context.Context, db *pgxpool.Pool, userID, planID string) (subscriptionInfo, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	_, err := db.Exec(ctx, `
		INSERT INTO subscriptions (user_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end)
		VALUES ($1, $2, 'active', $3, $4, FALSE)
		ON CONFLICT (user_id) DO NOTHING
	`, userID, planID, start, end)
	if err != nil {
		return subscriptionInfo{}, err
	}
	return subscriptionInfo{
		PlanID:             planID,
		Status:             "active",
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   end,
		CancelAtPeriodEnd:  false,
	}, nil
}

func usageTotals(ctx context.Context, db *pgxpool.Pool, userID string, start, end time.Time) (map[string]int, error) {
	rows, err := db.Query(ctx, `
		SELECT category, COALESCE(SUM(units), 0) as total
		FROM usage_events
		WHERE user_id=$1 AND created_at >= $2 AND created_at < $3
		GROUP BY category
	`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{
		"llm_chars": 0,
		"tts_chars": 0,
	}
	for rows.Next() {
		var category string
		var total int
		if err := rows.Scan(&category, &total); err == nil {
			out[category] = total
		}
	}
	return out, nil
}

func consumeUsage(ctx context.Context, db *pgxpool.Pool, userID, category string, units int, meta map[string]any) error {
	if units <= 0 {
		return nil
	}
	plan, sub, err := ensureSubscription(ctx, db, userID)
	if err != nil {
		return err
	}
	used, err := usageTotals(ctx, db, userID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		return err
	}
	quota := map[string]int{
		"llm_chars": plan.MonthlyLLMQuota,
		"tts_chars": plan.MonthlyTTSQuota,
	}
	if used[category]+units > quota[category] {
		return errors.New("quota exceeded")
	}
	if meta == nil {
		meta = map[string]any{}
	}
	raw, _ := json.Marshal(meta)
	_, err = db.Exec(ctx, `
		INSERT INTO usage_events (user_id, category, units, meta, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, userID, category, units, raw)
	return err
}
