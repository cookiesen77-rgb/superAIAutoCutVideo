package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
)

type Server struct {
	Router *chi.Mux
}

func NewServer(cfg config.Config, appDB *pgxpool.Pool, userDB *pgxpool.Pool) *Server {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CorsAllowedOrigin,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	hub := NewHub()
	health := NewHealthHandler()
	auth := NewAuthHandler(cfg, appDB, userDB)
	admin := NewAdminHandler(cfg, appDB, userDB)
	uploads := NewUploadHandler(cfg, appDB)
	jobs := NewJobHandler(appDB)
	projects := NewProjectsHandler(cfg, appDB, hub)
	models := NewModelConfigHandler(appDB)
	tts := NewTTSHandler(cfg, appDB)
	voices := NewVoicesHandler(cfg, appDB)
	prompts := NewPromptsHandler(appDB)
	billing := NewBillingHandler(appDB)
	legacyVideo := NewLegacyVideoHandler(hub)
	ws := NewWSHandler(hub)

	r.Get("/api/health", health.Health)
	r.Get("/api/status", health.Health)
	r.Post("/api/health/test-integrations", health.TestIntegrations)
	r.Get("/ws", ws.ServeHTTP)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/refresh", auth.Refresh)
		r.Post("/logout", auth.Logout)
		r.With(auth.Middleware).Get("/me", auth.Me)
	})

	r.Route("/api/uploads", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Post("/init", uploads.InitUpload)
	})

	r.Route("/api/video", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Post("/info", legacyVideo.GetVideoInfo)
		r.Post("/process", legacyVideo.ProcessVideo)
	})
	r.Route("/api/task", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/{taskID}", legacyVideo.GetTaskStatus)
	})

	r.Route("/api/jobs", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Post("/", jobs.CreateJob)
		r.Get("/", jobs.ListJobs)
		r.Get("/{jobID}", jobs.GetJob)
	})

	r.Route("/api/billing", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/plan", billing.GetPlan)
		r.Get("/usage", billing.GetUsage)
	})

	r.Route("/api/projects", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/", projects.List)
		r.Get("/{projectID}", projects.Get)
		r.Post("/", projects.Create)
		r.Post("/{projectID}", projects.Update)
		r.Post("/{projectID}/delete", projects.Delete)
		r.Post("/{projectID}/upload/video", projects.UploadVideo)
		r.Post("/{projectID}/upload/subtitle", projects.UploadSubtitle)
		r.Post("/{projectID}/delete/video", projects.DeleteVideo)
		r.Post("/{projectID}/delete/subtitle", projects.DeleteSubtitle)
		r.Post("/{projectID}/merge/videos", projects.MergeVideos)
		r.Get("/{projectID}/merge/videos/status/{taskID}", projects.MergeStatus)
		r.Post("/{projectID}/videos/order", projects.UpdateVideoOrder)
		r.Post("/generate-script", projects.GenerateScript)
		r.Post("/{projectID}/script", projects.SaveScript)
		r.Post("/{projectID}/generate-video", projects.GenerateVideo)
		r.Get("/{projectID}/output-video", projects.OutputVideo)
		r.Get("/{projectID}/merged-video", projects.MergedVideo)
	})

	r.Route("/api/models", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Route("/video-analysis", func(r chi.Router) {
			r.Get("/configs", models.ListVideoConfigs)
			r.Put("/configs/{configID}", models.UpdateVideoConfig)
			r.Post("/configs/{configID}/activate", models.ActivateVideoConfig)
			r.Post("/test/{configID}", models.TestVideoConfig)
		})
		r.Route("/content-generation", func(r chi.Router) {
			r.Get("/configs", models.ListContentConfigs)
			r.Put("/configs/{configID}", models.UpdateContentConfig)
			r.Post("/test/{configID}", models.TestContentConfig)
		})
	})

	r.Route("/api/tts", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/engines", tts.GetEngines)
		r.Get("/voices", tts.GetVoices)
		r.Get("/configs", tts.GetConfigs)
		r.Patch("/configs/{configID}", tts.PatchConfig)
		r.Post("/configs/{configID}/activate", tts.ActivateConfig)
		r.Post("/configs/{configID}/test", tts.TestConfig)
		r.Post("/voices/{voiceID}/preview", tts.PreviewVoice)
		r.Get("/emotions", tts.GetEmotions)
		r.Get("/index-tts/status", tts.GetIndexTtsStatus)
		r.Post("/index-tts/preload", tts.PreloadIndexTts)
		r.Post("/index-tts/test", tts.TestIndexTts)
	})

	r.Route("/api/voices", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/categories", voices.ListCategories)
		r.Post("/categories", voices.CreateCategory)
		r.Patch("/categories/{categoryID}", voices.UpdateCategory)
		r.Delete("/categories/{categoryID}", voices.DeleteCategory)
		r.Get("/", voices.ListVoices)
		r.Get("/{voiceID}", voices.GetVoice)
		r.Post("/upload", voices.UploadVoice)
		r.Post("/record", voices.UploadRecording)
		r.Patch("/{voiceID}", voices.UpdateVoice)
		r.Delete("/{voiceID}", voices.DeleteVoice)
		r.Get("/{voiceID}/audio", voices.GetAudio)
		r.Post("/{voiceID}/preview", voices.PreviewVoice)
	})

	r.Route("/api/prompts", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/categories", prompts.ListCategories)
		r.Get("/", prompts.ListPrompts)
		r.Get("/{promptID}", prompts.GetPrompt)
		r.Post("/", prompts.CreatePrompt)
		r.Put("/{promptID}", prompts.UpdatePrompt)
		r.Delete("/{promptID}", prompts.DeletePrompt)
		r.Post("/{promptID}/render-preview", prompts.RenderPreview)
		r.Post("/validate", prompts.ValidateTemplate)
		r.Post("/projects/{projectID}/prompts/select", prompts.SetProjectSelection)
		r.Get("/projects/{projectID}/prompts/selection", prompts.GetProjectSelection)
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(auth.AdminOnly)
		r.Get("/users", admin.ListUsers)
		r.Get("/users/{userID}", admin.GetUser)
		r.Patch("/users/{userID}", admin.UpdateUser)
		r.Post("/users/{userID}/reset-password", admin.ResetPassword)
		r.Get("/users/{userID}/usage", admin.GetUserUsage)
		r.Post("/users/{userID}/plan", admin.UpdateUserPlan)
	})

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	return &Server{Router: r}
}
