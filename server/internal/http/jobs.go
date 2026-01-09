package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobHandler struct {
	db *pgxpool.Pool
}

func NewJobHandler(db *pgxpool.Pool) *JobHandler {
	return &JobHandler{db: db}
}

type createJobRequest struct {
	Type     string         `json:"type"`
	InputURL string         `json:"input_url"`
	Payload  map[string]any `json:"payload"`
}

type jobResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	InputURL  string         `json:"input_url"`
	Progress  int            `json:"progress"`
	Payload   map[string]any `json:"payload,omitempty"`
	Result    map[string]any `json:"result,omitempty"`
	OutputURL *string        `json:"output_url,omitempty"`
	Error     *string        `json:"error,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Type == "" || (req.InputURL == "" && len(req.Payload) == 0) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type and input_url or payload required"})
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	payloadRaw, _ := json.Marshal(req.Payload)
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO jobs (id, user_id, type, status, input_url, payload, progress, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, userID, req.Type, "queued", req.InputURL, payloadRaw, 0, now, now)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create job failed"})
		return
	}

	writeJSON(w, http.StatusAccepted, jobResponse{
		ID:        id,
		Type:      req.Type,
		Status:    "queued",
		InputURL:  req.InputURL,
		Progress:  0,
		Payload:   req.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	rows, err := h.db.Query(r.Context(), `
		SELECT id, type, status, input_url, output_url, error, created_at, updated_at, payload, result, progress
		FROM jobs
		WHERE user_id=$1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	out := make([]jobResponse, 0)
	for rows.Next() {
		var j jobResponse
		var output sql.NullString
		var errMsg sql.NullString
		var payloadRaw []byte
		var resultRaw []byte
		_ = rows.Scan(&j.ID, &j.Type, &j.Status, &j.InputURL, &output, &errMsg, &j.CreatedAt, &j.UpdatedAt, &payloadRaw, &resultRaw, &j.Progress)
		if output.Valid {
			j.OutputURL = &output.String
		}
		if errMsg.Valid {
			j.Error = &errMsg.String
		}
		if len(payloadRaw) > 0 {
			_ = json.Unmarshal(payloadRaw, &j.Payload)
		}
		if len(resultRaw) > 0 {
			_ = json.Unmarshal(resultRaw, &j.Result)
		}
		out = append(out, j)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job id required"})
		return
	}

	var j jobResponse
	var output sql.NullString
	var errMsg sql.NullString
	var payloadRaw []byte
	var resultRaw []byte
	err := h.db.QueryRow(r.Context(), `
		SELECT id, type, status, input_url, output_url, error, created_at, updated_at, payload, result, progress
		FROM jobs
		WHERE id=$1 AND user_id=$2
	`, jobID, userID).Scan(&j.ID, &j.Type, &j.Status, &j.InputURL, &output, &errMsg, &j.CreatedAt, &j.UpdatedAt, &payloadRaw, &resultRaw, &j.Progress)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	if output.Valid {
		j.OutputURL = &output.String
	}
	if errMsg.Valid {
		j.Error = &errMsg.String
	}
	if len(payloadRaw) > 0 {
		_ = json.Unmarshal(payloadRaw, &j.Payload)
	}
	if len(resultRaw) > 0 {
		_ = json.Unmarshal(resultRaw, &j.Result)
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": j})
}
