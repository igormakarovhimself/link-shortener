package main

import (
	"link-shortener/internal/handler"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
	"log"
	"net/http"
)

var urlStorage = make(map[string]string)

func main() {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.HandleRoot)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}

}
