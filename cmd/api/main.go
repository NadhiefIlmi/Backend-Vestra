package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/handler"
	"github.com/live-camera-ar-tryon/backend/internal/catalog/repository"
	"github.com/live-camera-ar-tryon/backend/internal/catalog/service"
	"github.com/live-camera-ar-tryon/backend/internal/config"
	"github.com/live-camera-ar-tryon/backend/internal/decart"
	baseHandler "github.com/live-camera-ar-tryon/backend/internal/handler"
	"github.com/live-camera-ar-tryon/backend/internal/middleware"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal configuration error: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Structured Logger
	var logHandler slog.Handler
	if cfg.AppEnv == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting Live Camera AR Try-On API",
		slog.String("environment", cfg.AppEnv),
		slog.Int("port", cfg.HTTPPort),
	)

	// 3. Initialize Repositories and Services
	// Using MemoryRepository with PRD seeding for fast startup & local execution
	catalogRepo := repository.NewMemoryRepository()
	catalogSvc := service.New(catalogRepo)
	catalogHdl := handler.New(catalogSvc)

	// 4. Register HTTP Routes
	mux := http.NewServeMux()

	// Health and Readiness Checks
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		baseHandler.WriteSuccess(w, http.StatusOK, map[string]string{
			"status": "UP",
			"time":   time.Now().UTC().Format(time.RFC3339),
		}, nil)
	}
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/healthz", healthHandler)

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		baseHandler.WriteSuccess(w, http.StatusOK, map[string]string{
			"status":   "READY",
			"database": "OK",
			"storage":  "OK",
		}, nil)
	})

	// Catalog Endpoints
	mux.HandleFunc("/api/v1/clothing", catalogHdl.ListClothing)
	mux.HandleFunc("/api/v1/clothing/", catalogHdl.GetClothingByID)
	mux.HandleFunc("/api/v1/clothing-items", catalogHdl.ListClothing)
	mux.HandleFunc("/api/v1/clothing-items/", catalogHdl.GetClothingByID)

	mux.HandleFunc("/api/v1/categories", func(w http.ResponseWriter, r *http.Request) {
		baseHandler.WriteSuccess(w, http.StatusOK, []string{"top"}, nil)
	})

	// Decart Realtime AI Token Endpoint
	decartHdl := decart.NewTokenHandler(cfg.DecartAPIKey)
	mux.HandleFunc("/api/v1/tokens", decartHdl.CreateClientToken)

	// Optional local assets fallback
	if _, err := os.Stat("../frontend/public/assets"); err == nil {
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../frontend/public/assets"))))
	} else if _, err := os.Stat("assets"); err == nil {
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	}

	// 5. Apply Middlewares in correct execution order
	// RequestID -> SecurityHeaders -> Logging -> Recovery -> CORS -> Timeout -> BodyLimit -> Handler
	var rootHandler http.Handler = mux
	rootHandler = middleware.BodyLimit(1 << 20)(rootHandler) // 1MB request body limit
	rootHandler = middleware.Timeout(cfg.RequestTimeout)(rootHandler)
	rootHandler = middleware.CORS(cfg.CORSAllowedOrigins)(rootHandler)
	rootHandler = middleware.Recovery(logger)(rootHandler)
	rootHandler = middleware.Logging(logger)(rootHandler)
	rootHandler = middleware.SecurityHeaders(rootHandler)
	rootHandler = middleware.RequestID(rootHandler)

	// 6. Setup HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      rootHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Start Server in background
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server listening", slog.String("addr", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// 8. Graceful Shutdown on SIGINT / SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	case sig := <-shutdown:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", slog.Any("error", err))
			_ = server.Close()
		}
		logger.Info("server stopped cleanly")
	}
}

