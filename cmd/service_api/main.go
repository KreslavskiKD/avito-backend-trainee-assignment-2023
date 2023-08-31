package main

import (
	"service_api/cmd/internal/config"
	"service_api/cmd/internal/db/db_postgresql"
	"service_api/cmd/internal/handlers"

	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg := config.LoadConfigOrFail()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log.Info("Starting...")

	storage, err := db_postgresql.New(cfg.StoragePath)

	if err != nil {
		log.Error("failed to initialize database", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// base functionality
	router.Post("/", handlers.AddSegment(log, storage))
	router.Delete("/", handlers.DeleteSegment(log, storage))
	router.Get("/", handlers.GetSegments(log, storage))

	// additional functionality
	router.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Status asked")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Success"))
	})

	router.Get("/get_history", handlers.GetUserHistory(log, storage, *cfg))
	router.Get("/report/{id}", func(w http.ResponseWriter, r *http.Request) {
		filePath := chi.URLParam(r, "id") + "_history_report.csv"
		http.ServeFile(w, r, filePath)
	})

	log.Info("starting server", slog.String("address", cfg.Address))
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}
