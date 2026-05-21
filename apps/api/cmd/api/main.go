package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gatewarden/api/internal/config"
	"github.com/gatewarden/api/internal/handler"
	"github.com/gatewarden/api/internal/logger"
	"github.com/gatewarden/api/internal/repository"
	"github.com/gatewarden/api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer log.Sync()

	db, err := repository.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalw("failed to connect to database", "error", err)
	}
	defer db.Close()

	svc := service.New(db, log, cfg.JWTSecret)
	if err := svc.Auth.EnsureSeedUser(context.Background()); err != nil {
		log.Warnw("failed to seed user", "error", err)
	}

	r := handler.NewRouter(svc, log, cfg, db)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Infow("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("server error", "error", err)
		}
	}()

	<-ctx.Done()
	log.Infow("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalw("server forced shutdown", "error", err)
	}

	log.Infow("server stopped")
}
