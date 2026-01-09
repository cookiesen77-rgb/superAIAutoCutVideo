package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/ai"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectsHandler struct {
	cfg config.Config
	db  *pgxpool.Pool
	hub *Hub
	mu  sync.Mutex
}

func NewProjectsHandler(cfg config.Config, db *pgxpool.Pool, hub *Hub) *ProjectsHandler {
	return &ProjectsHandler{cfg: cfg, db: db, hub: hub}
}

type projectRecord struct {
	ID              string            `json:"id"`
	UserID          string            `json:"-"`
	Name            string            `json:"name"`
	Description     *string           `json:"description,omitempty"`
	NarrationType   string            `json:"narration_type"`
	Status          string            `json:"status"`
	VideoPath       *string           `json:"video_path,omitempty"`
	VideoPaths      []string          `json:"video_paths"`
	VideoNames      map[string]string `json:"video_names"`
	VideoCurrent    *string           `json:"video_current_name,omitempty"`
	MergedVideoPath *string           `json:"merged_video_path,omitempty"`
	SubtitlePath    *string           `json:"subtitle_path,omitempty"`
	OutputVideoPath *string           `json:"output_video_path,omitempty"`
	Script          json.RawMessage   `json:"script,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type createProjectRequest struct {
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	NarrationType *string `json:"narration_type"`
}

type updateProjectRequest struct {
	Name            *string           `json:"name"`
	Description     *string           `json:"description"`
	NarrationType   *string           `json:"narration_type"`
	Status          *string           `json:"status"`
	VideoPath       *string           `json:"video_path"`
	VideoPaths      []string          `json:"video_paths"`
	MergedVideoPath *string           `json:"merged_video_path"`
	SubtitlePath    *string           `json:"subtitle_path"`
	OutputVideoPath *string           `json:"output_video_path"`
	Script          json.RawMessage   `json:"script"`
	VideoNames      map[string]string `json:"video_names"`
	VideoCurrent    *string           `json:"video_current_name"`
}

type mergeStatus struct {
	TaskID   string  `json:"task_id"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Message  string  `json:"message"`
	FilePath string  `json:"file_path,omitempty"`
}

type mergeTask struct {
	status  mergeStatus
	updated time.Time
	project string
}

var mergeTasks = struct {
	sync.Mutex
	m map[string]*mergeTask
}{m: make(map[string]*mergeTask)}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	rows, err := h.db.Query(r.Context(), `
		SELECT id, user_id, name, description, narration_type, status, video_path, video_paths,
		       video_names, video_current_name, merged_video_path, subtitle_path, output_video_path,
		       script, created_at, updated_at
		FROM projects
		WHERE user_id=$1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	out := make([]projectRecord, 0)
	for rows.Next() {
		rec, err := scanProject(rows)
		if err == nil {
			out = append(out, rec)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":      out,
		"message":   fmt.Sprintf("获取到 %d 个项目", len(out)),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      rec,
		"message":   "获取项目成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "项目名称不能为空"})
		return
	}

	narration := "短剧解说"
	if req.NarrationType != nil && strings.TrimSpace(*req.NarrationType) != "" {
		narration = *req.NarrationType
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO projects (id, user_id, name, description, narration_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, userID, req.Name, req.Description, narration, "draft", now, now)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建项目失败"})
		return
	}

	rec, _ := h.getProject(r.Context(), userID, id)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      rec,
		"message":   "项目创建成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var videoNamesJSON []byte
	if req.VideoNames != nil {
		videoNamesJSON, _ = json.Marshal(req.VideoNames)
	}
	scriptJSON := req.Script
	if len(scriptJSON) == 0 {
		scriptJSON = nil
	}

	_, err := h.db.Exec(r.Context(), `
		UPDATE projects
		SET name = COALESCE($3, name),
		    description = COALESCE($4, description),
		    narration_type = COALESCE($5, narration_type),
		    status = COALESCE($6, status),
		    video_path = COALESCE($7, video_path),
		    video_paths = CASE WHEN $8 IS NULL THEN video_paths ELSE $8 END,
		    merged_video_path = COALESCE($9, merged_video_path),
		    subtitle_path = COALESCE($10, subtitle_path),
		    output_video_path = COALESCE($11, output_video_path),
		    script = CASE WHEN $12 IS NULL THEN script ELSE $12 END,
		    video_names = CASE WHEN $13 IS NULL THEN video_names ELSE $13 END,
		    video_current_name = COALESCE($14, video_current_name),
		    updated_at = NOW()
		WHERE id=$1 AND user_id=$2
	`, projectID, userID, req.Name, req.Description, req.NarrationType, req.Status, req.VideoPath,
		req.VideoPaths, req.MergedVideoPath, req.SubtitlePath, req.OutputVideoPath,
		scriptJSON, videoNamesJSON, req.VideoCurrent)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新项目失败"})
		return
	}

	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      rec,
		"message":   "项目更新成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	_, err := h.db.Exec(r.Context(), `DELETE FROM projects WHERE id=$1 AND user_id=$2`, projectID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"message":   "项目删除成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "视频文件缺失"})
		return
	}
	defer file.Close()

	uploader, err := storage.NewS3Uploader(r.Context(), h.cfg.AWSRegion, h.cfg.S3Bucket)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "存储初始化失败"})
		return
	}

	key := path.Join("projects", projectID, "videos", uuid.NewString()+"_"+sanitizeFilename(header.Filename))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	url, err := uploader.Upload(r.Context(), key, contentType, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "上传失败"})
		return
	}

	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}

	paths := append(rec.VideoPaths, url)
	names := rec.VideoNames
	if names == nil {
		names = map[string]string{}
	}
	names[url] = header.Filename
	videoPath := url
	videoCurrent := header.Filename
	if rec.MergedVideoPath != nil && *rec.MergedVideoPath != "" {
		videoPath = *rec.MergedVideoPath
		if rec.VideoCurrent != nil {
			videoCurrent = *rec.VideoCurrent
		}
	}

	namesJSON, _ := json.Marshal(names)
	_, err = h.db.Exec(r.Context(), `
		UPDATE projects
		SET video_paths=$3, video_names=$4, video_path=$5, video_current_name=$6, updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, projectID, userID, paths, namesJSON, videoPath, videoCurrent)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新项目失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"file_path":   url,
			"file_name":   header.Filename,
			"file_size":   header.Size,
			"upload_time": time.Now().UTC().Format(time.RFC3339),
		},
		"message":   "视频上传成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) UploadSubtitle(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字幕文件缺失"})
		return
	}
	defer file.Close()

	uploader, err := storage.NewS3Uploader(r.Context(), h.cfg.AWSRegion, h.cfg.S3Bucket)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "存储初始化失败"})
		return
	}

	key := path.Join("projects", projectID, "subtitles", uuid.NewString()+"_"+sanitizeFilename(header.Filename))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	url, err := uploader.Upload(r.Context(), key, contentType, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "上传失败"})
		return
	}

	_, err = h.db.Exec(r.Context(), `
		UPDATE projects
		SET subtitle_path=$3, updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, projectID, userID, url)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新项目失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"file_path":   url,
			"file_name":   header.Filename,
			"file_size":   header.Size,
			"upload_time": time.Now().UTC().Format(time.RFC3339),
		},
		"message":   "字幕上传成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	var payload struct {
		FilePath string `json:"file_path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}

	target := strings.TrimSpace(payload.FilePath)
	if target == "" && rec.VideoPath != nil {
		target = *rec.VideoPath
	}

	paths := make([]string, 0, len(rec.VideoPaths))
	for _, p := range rec.VideoPaths {
		if p != target {
			paths = append(paths, p)
		}
	}
	if rec.VideoNames != nil && target != "" {
		delete(rec.VideoNames, target)
	}

	var newVideoPath *string
	var newCurrent *string
	if len(paths) > 0 {
		newVideoPath = &paths[0]
		if rec.VideoNames != nil {
			if name, ok := rec.VideoNames[paths[0]]; ok {
				newCurrent = &name
			}
		}
	}
	if rec.MergedVideoPath != nil && *rec.MergedVideoPath != "" {
		newVideoPath = rec.MergedVideoPath
		if rec.VideoCurrent != nil {
			newCurrent = rec.VideoCurrent
		}
	}

	namesJSON, _ := json.Marshal(rec.VideoNames)
	_, err = h.db.Exec(r.Context(), `
		UPDATE projects
		SET video_paths=$3, video_path=$4, video_current_name=$5, video_names=$6, updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, projectID, userID, paths, newVideoPath, newCurrent, namesJSON)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"removed":      true,
			"removed_path": target,
		},
		"message":   "视频删除成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) DeleteSubtitle(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	_, err := h.db.Exec(r.Context(), `
		UPDATE projects SET subtitle_path=NULL, updated_at=NOW() WHERE id=$1 AND user_id=$2
	`, projectID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      map[string]any{"removed": true},
		"message":   "字幕删除成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) UpdateVideoOrder(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	var payload struct {
		OrderedPaths []string `json:"ordered_paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}
	if len(payload.OrderedPaths) != len(rec.VideoPaths) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "排序列表与现有视频不匹配"})
		return
	}

	newVideoPath := ""
	if len(payload.OrderedPaths) > 0 {
		newVideoPath = payload.OrderedPaths[0]
	}
	if rec.MergedVideoPath != nil && *rec.MergedVideoPath != "" {
		newVideoPath = *rec.MergedVideoPath
	}
	var currentName *string
	if rec.VideoNames != nil && newVideoPath != "" {
		if name, ok := rec.VideoNames[newVideoPath]; ok {
			currentName = &name
		}
	}

	_, err = h.db.Exec(r.Context(), `
		UPDATE projects
		SET video_paths=$3, video_path=$4, video_current_name=$5, updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, projectID, userID, payload.OrderedPaths, newVideoPath, currentName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}

	rec, _ = h.getProject(r.Context(), userID, projectID)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      rec,
		"message":   "更新排序成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) GenerateScript(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProjectID     string `json:"project_id"`
		VideoPath     string `json:"video_path"`
		SubtitlePath  string `json:"subtitle_path"`
		NarrationType string `json:"narration_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	userID := UserIDFromContext(r.Context())
	if payload.ProjectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project_id required"})
		return
	}

	rec, err := h.getProject(r.Context(), userID, payload.ProjectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}

	if payload.VideoPath == "" && rec.VideoPath != nil {
		payload.VideoPath = *rec.VideoPath
	}
	if payload.VideoPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "video_path required"})
		return
	}
	if payload.SubtitlePath == "" && rec.SubtitlePath != nil {
		payload.SubtitlePath = *rec.SubtitlePath
	}

	configID, cfgMap, ok := loadActiveModelConfig(r.Context(), h.db, userID, modelTypeContent)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先配置文案生成模型"})
		return
	}
	if _, ok := cfgMap["provider"]; !ok {
		if provider := inferProviderFromConfigID(configID); provider != "" {
			cfgMap["provider"] = provider
		}
	}
	modelCfg, err := modelConfigFromMap(cfgMap, inferProviderFromConfigID(configID))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	prompt, err := resolveScriptPrompt(r.Context(), h.db, userID, payload.ProjectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "提示词配置异常"})
		return
	}

	subtitleText, err := fetchSubtitleContent(r.Context(), payload.SubtitlePath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字幕读取失败"})
		return
	}

	plotAnalysis := ""
	if rec.Description != nil {
		plotAnalysis = strings.TrimSpace(*rec.Description)
	}
	vars := map[string]any{
		"drama_name":       rec.Name,
		"plot_analysis":    plotAnalysis,
		"subtitle_content": subtitleText,
	}
	messages := buildScriptMessages(prompt, vars)
	req := buildChatRequest(modelCfg, cfgMap, messages, true)
	units := 0
	for _, msg := range messages {
		units += runeCount(msg.Content)
	}
	if err := consumeUsage(r.Context(), h.db, userID, "llm_chars", units, map[string]any{
		"provider": modelCfg.Provider,
		"model":    modelCfg.ModelName,
	}); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "配额不足"})
		return
	}
	raw, err := ai.GenerateChat(r.Context(), modelCfg, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "脚本生成失败"})
		return
	}
	data, _, err := parseJSONFromText(raw)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "模型返回格式解析失败"})
		return
	}
	items, err := extractScriptItems(data)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "模型返回格式错误"})
		return
	}
	script := buildVideoScriptFromItems(items)
	scriptJSON, _ := json.Marshal(script)

	_, err = h.db.Exec(r.Context(), `
		UPDATE projects
		SET script=$3, updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, payload.ProjectID, userID, scriptJSON)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "脚本保存失败"})
		return
	}

	go h.emitProgress(payload.ProjectID, "generate_script")

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"script":        script,
			"plot_analysis": plotAnalysis,
		},
		"message":   "解说脚本生成成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) SaveScript(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	var payload struct {
		Script json.RawMessage `json:"script"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Script) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "script required"})
		return
	}
	_, err := h.db.Exec(r.Context(), `
		UPDATE projects
		SET script=$3, updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, projectID, userID, payload.Script)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      json.RawMessage(payload.Script),
		"message":   "脚本保存成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) GenerateVideo(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}

	output := ""
	if rec.MergedVideoPath != nil && *rec.MergedVideoPath != "" {
		output = *rec.MergedVideoPath
	} else if rec.VideoPath != nil {
		output = *rec.VideoPath
	}

	if output != "" {
		_, _ = h.db.Exec(r.Context(), `
			UPDATE projects SET output_video_path=$3, updated_at=NOW() WHERE id=$1 AND user_id=$2
		`, projectID, userID, output)
	}

	go h.emitProgress(projectID, "generate_video")

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"output_path":    output,
			"segments_count": 1,
		},
		"message":   "视频生成完成",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) OutputVideo(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil || rec.OutputVideoPath == nil || *rec.OutputVideoPath == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "输出视频不存在"})
		return
	}
	h.redirectToFile(w, r, *rec.OutputVideoPath)
}

func (h *ProjectsHandler) MergedVideo(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil || rec.MergedVideoPath == nil || *rec.MergedVideoPath == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "合并视频不存在"})
		return
	}
	h.redirectToFile(w, r, *rec.MergedVideoPath)
}

func (h *ProjectsHandler) MergeVideos(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	rec, err := h.getProject(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
		return
	}
	taskID := fmt.Sprintf("merge_%s_%d", projectID, time.Now().UnixNano())

	cleanupMergeTasks(time.Now().UTC())
	mergeTasks.Lock()
	mergeTasks.m[taskID] = &mergeTask{
		status: mergeStatus{
			TaskID:   taskID,
			Status:   "processing",
			Progress: 0,
			Message:  "开始合并",
		},
		updated: time.Now(),
		project: projectID,
	}
	mergeTasks.Unlock()

	go func() {
		for i := 0; i <= 100; i += 20 {
			time.Sleep(300 * time.Millisecond)
			h.emitMergeProgress(projectID, taskID, float64(i), "合并中")
		}
		mergedPath := ""
		if len(rec.VideoPaths) > 0 {
			mergedPath = rec.VideoPaths[0]
		} else if rec.VideoPath != nil {
			mergedPath = *rec.VideoPath
		}
		if mergedPath != "" {
			_, _ = h.db.Exec(context.Background(), `
				UPDATE projects SET merged_video_path=$3, video_path=$3, updated_at=NOW() WHERE id=$1 AND user_id=$2
			`, projectID, userID, mergedPath)
		}
		h.emitMergeCompleted(projectID, taskID, mergedPath)
	}()

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"task_id": taskID,
		},
		"message":   "开始合并",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) MergeStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	cleanupMergeTasks(time.Now().UTC())
	mergeTasks.Lock()
	task := mergeTasks.m[taskID]
	mergeTasks.Unlock()
	if task == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      task.status,
		"message":   "获取合并进度",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ProjectsHandler) emitProgress(projectID string, scope string) {
	stages := []struct {
		Progress float64
		Message  string
	}{
		{10, "任务已提交"},
		{50, "云端处理中"},
		{100, "处理完成"},
	}
	for _, stage := range stages {
		time.Sleep(300 * time.Millisecond)
		payload := map[string]any{
			"type":       "progress",
			"scope":      scope,
			"project_id": projectID,
			"progress":   stage.Progress,
			"message":    stage.Message,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}
		if stage.Progress >= 100 {
			payload["type"] = "completed"
		}
		msg, _ := json.Marshal(payload)
		h.hub.Broadcast(msg)
	}
}

func (h *ProjectsHandler) emitMergeProgress(projectID, taskID string, progress float64, message string) {
	mergeTasks.Lock()
	task := mergeTasks.m[taskID]
	if task != nil {
		task.status.Progress = progress
		task.status.Message = message
		task.status.Status = "processing"
		task.updated = time.Now()
	}
	mergeTasks.Unlock()

	payload := map[string]any{
		"type":       "progress",
		"scope":      "merge_videos",
		"project_id": projectID,
		"task_id":    taskID,
		"progress":   progress,
		"message":    message,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	msg, _ := json.Marshal(payload)
	h.hub.Broadcast(msg)
}

func (h *ProjectsHandler) emitMergeCompleted(projectID, taskID, filePath string) {
	mergeTasks.Lock()
	task := mergeTasks.m[taskID]
	if task != nil {
		task.status.Progress = 100
		task.status.Status = "completed"
		task.status.Message = "合并完成"
		task.status.FilePath = filePath
		task.updated = time.Now()
	}
	mergeTasks.Unlock()

	payload := map[string]any{
		"type":       "completed",
		"scope":      "merge_videos",
		"project_id": projectID,
		"task_id":    taskID,
		"progress":   100,
		"message":    "合并完成",
		"file_path":  filePath,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	msg, _ := json.Marshal(payload)
	h.hub.Broadcast(msg)
}

func (h *ProjectsHandler) getProject(ctx context.Context, userID, projectID string) (projectRecord, error) {
	row := h.db.QueryRow(ctx, `
		SELECT id, user_id, name, description, narration_type, status, video_path, video_paths,
		       video_names, video_current_name, merged_video_path, subtitle_path, output_video_path,
		       script, created_at, updated_at
		FROM projects
		WHERE id=$1 AND user_id=$2
	`, projectID, userID)
	return scanProject(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProject(row scanner) (projectRecord, error) {
	var rec projectRecord
	var videoPaths []string
	var namesJSON []byte
	var scriptJSON []byte
	err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.Name,
		&rec.Description,
		&rec.NarrationType,
		&rec.Status,
		&rec.VideoPath,
		&videoPaths,
		&namesJSON,
		&rec.VideoCurrent,
		&rec.MergedVideoPath,
		&rec.SubtitlePath,
		&rec.OutputVideoPath,
		&scriptJSON,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return rec, err
	}
	rec.VideoPaths = videoPaths
	if rec.VideoPaths == nil {
		rec.VideoPaths = []string{}
	}
	rec.VideoNames = map[string]string{}
	if len(namesJSON) > 0 {
		_ = json.Unmarshal(namesJSON, &rec.VideoNames)
	}
	if len(scriptJSON) > 0 {
		rec.Script = scriptJSON
	}
	return rec, nil
}

func (h *ProjectsHandler) redirectToFile(w http.ResponseWriter, r *http.Request, url string) {
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

func extractS3Key(url, bucket, region string) string {
	prefix := "https://" + bucket + ".s3." + region + ".amazonaws.com/"
	if strings.HasPrefix(url, prefix) {
		return strings.TrimPrefix(url, prefix)
	}
	return ""
}

func cleanupMergeTasks(now time.Time) {
	const ttl = 30 * time.Minute
	mergeTasks.Lock()
	for id, task := range mergeTasks.m {
		if task == nil {
			delete(mergeTasks.m, id)
			continue
		}
		if now.Sub(task.updated) > ttl {
			delete(mergeTasks.m, id)
		}
	}
	mergeTasks.Unlock()
}
