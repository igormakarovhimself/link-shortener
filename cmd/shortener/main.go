package main

import (
	"link-shortener/internal/config"
	"link-shortener/internal/handler"
	"link-shortener/internal/middleware"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var urlStorage = make(map[string]string)

func main() {
	cfg := config.SetupConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo)
	h := handler.NewHandler(svc, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.WithLogging(sugar))
	r.Post("/", h.HandlePost)
	r.Get("/{id}", h.HandleGet)

	log.Println("Starting server on", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatal(err)
	}

}
