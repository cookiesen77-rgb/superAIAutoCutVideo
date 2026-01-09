package httpapi

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadHandler struct {
	cfg config.Config
	db  *pgxpool.Pool
}

func NewUploadHandler(cfg config.Config, db *pgxpool.Pool) *UploadHandler {
	return &UploadHandler{cfg: cfg, db: db}
}

type uploadInitRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

func (h *UploadHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	var req uploadInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.Filename) == "" || strings.TrimSpace(req.ContentType) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "filename and content_type required"})
		return
	}
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	presigner, err := storage.NewS3Presigner(r.Context(), h.cfg.AWSRegion, h.cfg.S3Bucket)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "storage init failed"})
		return
	}

	key := path.Join("uploads", userID, uuid.NewString(), sanitizeFilename(req.Filename))
	url, err := presigner.PresignPut(r.Context(), key, req.ContentType, 15*time.Minute)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "presign failed"})
		return
	}

	fileURL := "https://" + h.cfg.S3Bucket + ".s3." + h.cfg.AWSRegion + ".amazonaws.com/" + key
	writeJSON(w, http.StatusOK, map[string]any{
		"upload_url": url,
		"file_url":   fileURL,
		"expires_in": int64((15 * time.Minute).Seconds()),
	})
}

func sanitizeFilename(name string) string {
	clean := path.Base(strings.ReplaceAll(name, "\\", "/"))
	if clean == "." || clean == "/" || clean == "" {
		return "upload.bin"
	}
	return clean
}
