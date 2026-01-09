package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/storage"
)

type VoicesHandler struct {
	cfg config.Config
	db  *pgxpool.Pool
}

func NewVoicesHandler(cfg config.Config, db *pgxpool.Pool) *VoicesHandler {
	return &VoicesHandler{cfg: cfg, db: db}
}

type voiceCategory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type voiceRecord struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	IsPreset    bool      `json:"is_preset"`
	AudioURL    string    `json:"audio_file"`
	Description string    `json:"description"`
	Duration    float64   `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
}

var defaultVoiceCategories = []voiceCategory{
	{ID: "male", Name: "男声", Icon: "👨", IsDefault: true},
	{ID: "female", Name: "女声", Icon: "👩", IsDefault: true},
	{ID: "child", Name: "童声", Icon: "👶", IsDefault: true},
	{ID: "narrator", Name: "旁白", Icon: "🎙️", IsDefault: true},
}

func (h *VoicesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	_ = h.ensureDefaultVoiceCategories(r.Context(), userID)

	rows, err := h.db.Query(r.Context(), `
		SELECT id, name, icon, is_default, created_at
		FROM voice_categories
		WHERE user_id=$1
		ORDER BY is_default DESC, created_at ASC
	`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	out := make([]voiceCategory, 0)
	for rows.Next() {
		var c voiceCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.IsDefault, &c.CreatedAt); err == nil {
			out = append(out, c)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"data":      out,
		"message":   "获取分类成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *VoicesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	var payload struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "分类名称不能为空"})
		return
	}
	icon := strings.TrimSpace(payload.Icon)
	if icon == "" {
		icon = "🎤"
	}
	id := uuid.NewString()
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO voice_categories (user_id, id, name, icon, is_default)
		VALUES ($1, $2, $3, $4, FALSE)
	`, userID, id, payload.Name, icon)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":         id,
			"name":       payload.Name,
			"icon":       icon,
			"is_default": false,
		},
		"message": "分类创建成功",
	})
}

func (h *VoicesHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	categoryID := chi.URLParam(r, "categoryID")
	var payload struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	_, err := h.db.Exec(r.Context(), `
		UPDATE voice_categories
		SET name=COALESCE(NULLIF($3, ''), name),
		    icon=COALESCE(NULLIF($4, ''), icon)
		WHERE user_id=$1 AND id=$2
	`, userID, categoryID, strings.TrimSpace(payload.Name), strings.TrimSpace(payload.Icon))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "分类更新成功",
	})
}

func (h *VoicesHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	categoryID := chi.URLParam(r, "categoryID")
	var isDefault bool
	_ = h.db.QueryRow(r.Context(), `
		SELECT is_default FROM voice_categories WHERE user_id=$1 AND id=$2
	`, userID, categoryID).Scan(&isDefault)
	if isDefault {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "预置分类不可删除"})
		return
	}
	var count int
	_ = h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM voices WHERE user_id=$1 AND category=$2
	`, userID, categoryID).Scan(&count)
	if count > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "该分类下仍有音色"})
		return
	}
	_, err := h.db.Exec(r.Context(), `
		DELETE FROM voice_categories WHERE user_id=$1 AND id=$2
	`, userID, categoryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "删除成功",
	})
}

func (h *VoicesHandler) ListVoices(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	category := strings.TrimSpace(r.URL.Query().Get("category"))

	query := `
		SELECT id, name, category, is_preset, audio_url, description, duration, created_at
		FROM voices
		WHERE user_id=$1
	`
	args := []any{userID}
	if category != "" {
		query += " AND category=$2"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	out := make([]voiceRecord, 0)
	for rows.Next() {
		var v voiceRecord
		if err := rows.Scan(&v.ID, &v.Name, &v.Category, &v.IsPreset, &v.AudioURL, &v.Description, &v.Duration, &v.CreatedAt); err == nil {
			out = append(out, v)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"data":      out,
		"message":   "获取音色成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *VoicesHandler) GetVoice(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	voiceID := chi.URLParam(r, "voiceID")
	var v voiceRecord
	err := h.db.QueryRow(r.Context(), `
		SELECT id, name, category, is_preset, audio_url, description, duration, created_at
		FROM voices
		WHERE user_id=$1 AND id=$2
	`, userID, voiceID).Scan(&v.ID, &v.Name, &v.Category, &v.IsPreset, &v.AudioURL, &v.Description, &v.Duration, &v.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "音色不存在"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    v,
		"message": "获取音色成功",
	})
}

func (h *VoicesHandler) UploadVoice(w http.ResponseWriter, r *http.Request) {
	h.handleUpload(w, r)
}

func (h *VoicesHandler) UploadRecording(w http.ResponseWriter, r *http.Request) {
	h.handleUpload(w, r)
}

func (h *VoicesHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	category := strings.TrimSpace(r.FormValue("category"))
	description := strings.TrimSpace(r.FormValue("description"))
	if name == "" || category == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and category required"})
		return
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "audio required"})
		return
	}
	defer file.Close()

	uploader, err := storage.NewS3Uploader(r.Context(), h.cfg.AWSRegion, h.cfg.S3Bucket)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "storage init failed"})
		return
	}

	key := path.Join("voices", userID, uuid.NewString()+"_"+sanitizeFilename(header.Filename))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	url, err := uploader.Upload(r.Context(), key, contentType, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "上传失败"})
		return
	}

	var id string
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO voices (user_id, name, category, is_preset, audio_url, description, duration, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, $4, $5, 0, NOW(), NOW())
		RETURNING id
	`, userID, name, category, url, description).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":          id,
			"name":        name,
			"category":    category,
			"is_preset":   false,
			"audio_file":  url,
			"description": description,
			"duration":    0,
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		},
		"message":   "音色上传成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *VoicesHandler) UpdateVoice(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	voiceID := chi.URLParam(r, "voiceID")
	var payload struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	_, err := h.db.Exec(r.Context(), `
		UPDATE voices
		SET name=COALESCE(NULLIF($3, ''), name),
		    category=COALESCE(NULLIF($4, ''), category),
		    description=COALESCE($5, description),
		    updated_at=NOW()
		WHERE user_id=$1 AND id=$2
	`, userID, voiceID, strings.TrimSpace(payload.Name), strings.TrimSpace(payload.Category), payload.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "音色更新成功",
	})
}

func (h *VoicesHandler) DeleteVoice(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	voiceID := chi.URLParam(r, "voiceID")
	_, err := h.db.Exec(r.Context(), `
		DELETE FROM voices WHERE user_id=$1 AND id=$2
	`, userID, voiceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "删除成功",
	})
}

func (h *VoicesHandler) GetAudio(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	voiceID := chi.URLParam(r, "voiceID")
	var url string
	err := h.db.QueryRow(r.Context(), `
		SELECT audio_url FROM voices WHERE user_id=$1 AND id=$2
	`, userID, voiceID).Scan(&url)
	if err != nil || url == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "音频不存在"})
		return
	}
	h.redirectToFile(w, r, url)
}

func (h *VoicesHandler) PreviewVoice(w http.ResponseWriter, r *http.Request) {
	h.GetAudio(w, r)
}

func (h *VoicesHandler) redirectToFile(w http.ResponseWriter, r *http.Request, url string) {
	key := extractS3Key(url, h.cfg.S3Bucket, h.cfg.AWSRegion)
	if key == "" {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	presigner, err := storage.NewS3Presigner(r.Context(), h.cfg.AWSRegion, h.cfg.S3Bucket)
	if err != nil {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	signed, err := presigner.PresignGet(r.Context(), key, 15*time.Minute)
	if err != nil {
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	http.Redirect(w, r, signed, http.StatusFound)
}

func (h *VoicesHandler) ensureDefaultVoiceCategories(ctx context.Context, userID string) error {
	var count int
	if err := h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM voice_categories WHERE user_id=$1
	`, userID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, cat := range defaultVoiceCategories {
		_, _ = h.db.Exec(ctx, `
			INSERT INTO voice_categories (user_id, id, name, icon, is_default)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
		`, userID, cat.ID, cat.Name, cat.Icon, cat.IsDefault)
	}
	return nil
}
