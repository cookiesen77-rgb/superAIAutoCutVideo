package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/db"
	httpapi "github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	appPool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer appPool.Close()

	userPool, err := db.Connect(ctx, cfg.UserDatabaseURL)
	if err != nil {
		log.Fatalf("user db connect error: %v", err)
	}
	defer userPool.Close()

	server := httpapi.NewServer(cfg, appPool, userPool)
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: server.Router,
	}
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go httpapi.StartJobWorker(workerCtx, cfg, appPool)

	go func() {
		log.Printf("API listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	workerCancel()
	_ = srv.Shutdown(ctxShutdown)
	log.Println("shutdown complete")
}
