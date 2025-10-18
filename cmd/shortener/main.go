package main

import (
	"link-shortener/internal/handler"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var urlStorage = make(map[string]string)

func main() {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo)
	h := handler.NewHandler(svc)

	r := chi.NewRouter()
	r.Post("/", h.HandlePost)
	r.Get("/{id}", h.HandleGet)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}
