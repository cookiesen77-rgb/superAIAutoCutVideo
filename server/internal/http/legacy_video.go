package httpapi

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LegacyVideoHandler struct {
	hub *Hub
}

func NewLegacyVideoHandler(hub *Hub) *LegacyVideoHandler {
	return &LegacyVideoHandler{hub: hub}
}

type legacyTaskStatus struct {
	TaskID   string    `json:"task_id"`
	Status   string    `json:"status"`
	Progress float64   `json:"progress"`
	Message  string    `json:"message"`
	Updated  time.Time `json:"-"`
}

var legacyTasks = struct {
	sync.Mutex
	m map[string]*legacyTaskStatus
}{
	m: make(map[string]*legacyTaskStatus),
}

func (h *LegacyVideoHandler) GetVideoInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"duration": 0,
		"format":   "",
	})
}

func (h *LegacyVideoHandler) ProcessVideo(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		VideoPath  string `json:"video_path"`
		OutputPath string `json:"output_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.VideoPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "video_path required"})
		return
	}

	taskID := uuid.NewString()
	status := &legacyTaskStatus{
		TaskID:   taskID,
		Status:   "processing",
		Progress: 0,
		Message:  "任务已提交",
		Updated:  time.Now().UTC(),
	}
	cleanupLegacyTasks(time.Now().UTC())
	legacyTasks.Lock()
	legacyTasks.m[taskID] = status
	legacyTasks.Unlock()

	go h.simulateLegacyProgress(taskID)

	writeJSON(w, http.StatusOK, map[string]any{
		"task_id": taskID,
	})
}

func (h *LegacyVideoHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	cleanupLegacyTasks(time.Now().UTC())
	legacyTasks.Lock()
	task := legacyTasks.m[taskID]
	legacyTasks.Unlock()
	if task == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *LegacyVideoHandler) simulateLegacyProgress(taskID string) {
	stages := []struct {
		Progress float64
		Message  string
		Status   string
	}{
		{10, "任务排队中", "pending"},
		{50, "云端处理中", "processing"},
		{100, "处理完成", "completed"},
	}
	for _, stage := range stages {
		time.Sleep(400 * time.Millisecond)
		legacyTasks.Lock()
		task := legacyTasks.m[taskID]
		if task != nil {
			task.Progress = stage.Progress
			task.Message = stage.Message
			task.Status = stage.Status
			task.Updated = time.Now().UTC()
		}
		legacyTasks.Unlock()
		if h.hub != nil {
			msg, _ := json.Marshal(map[string]any{
				"type":      "progress",
				"task_id":   taskID,
				"progress":  stage.Progress,
				"message":   stage.Message,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			if stage.Progress >= 100 {
				msg, _ = json.Marshal(map[string]any{
					"type":      "completed",
					"task_id":   taskID,
					"progress":  100,
					"message":   stage.Message,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
			}
			h.hub.Broadcast(msg)
		}
	}
}

func cleanupLegacyTasks(now time.Time) {
	const ttl = 30 * time.Minute
	legacyTasks.Lock()
	for id, task := range legacyTasks.m {
		if task == nil {
			delete(legacyTasks.m, id)
			continue
		}
		if now.Sub(task.Updated) > ttl {
			delete(legacyTasks.m, id)
		}
	}
	legacyTasks.Unlock()
}
