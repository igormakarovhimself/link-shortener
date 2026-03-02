package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"link-shortener/internal/audit"
	"link-shortener/internal/config"
	"link-shortener/internal/handler"
	"link-shortener/internal/middleware"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
)

func main() {
	cfg := config.SetupConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	var repo repository.URLRepository
	var db *sql.DB

	if cfg.DatabaseDSN != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			log.Fatal(err)
		}

		repo, err = repository.NewDBRepository(db)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfg.FileStoragePath != "" {
		fileRepo, err := repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		defer fileRepo.Close()
		repo = fileRepo
	} else {
		repo = repository.NewLocalRepository()
	}

	publisher := audit.NewPublisher()

	if cfg.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			log.Printf("Failed to create file observer: %v", err)
		} else if fileObserver != nil {
			publisher.Register("file", fileObserver)
			defer fileObserver.Close()
		}
	}

	if cfg.AuditURL != "" {
		urlObserver, err := audit.NewURLObserver(cfg.AuditURL)
		if err != nil {
			log.Printf("Failed to create URL observer: %v", err)
		} else if urlObserver != nil {
			publisher.Register("url", urlObserver)
		}
	}

	svc := service.NewShortenerService(repo, sugar)
	h := handler.NewHandler(svc, cfg.BaseURL, publisher)

	r := chi.NewRouter()
	r.Use(middleware.WithLogging(sugar))
	r.Use(middleware.WithGzip())
	r.Use(middleware.WithAuth())
	r.Mount("/debug", http.DefaultServeMux)
	r.Post("/", h.HandlePost)
	r.Get("/{id}", h.HandleGet)
	r.Get("/ping", h.HandlePing)
	r.Post("/api/shorten", h.HandleAPIShorten)
	r.Post("/api/shorten/batch", h.HandleAPIBatch)
	r.Get("/api/user/urls", h.HandleGetUserURLs)
	r.Delete("/api/user/urls", h.HandleDeleteUserURLs)

	log.Println("Starting server on", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatal(err)
	}

}
