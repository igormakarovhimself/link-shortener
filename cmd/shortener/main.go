package main

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"link-shortener/internal/handler"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
	"log"
	"net/http"
	"net/url"
)

var urlStorage = make(map[string]string)

func main() {
	// mux := http.NewServeMux()
	// mux.HandleFunc(`/`, handleRoot)

	// err := http.ListenAndServe(`:8080`, mux)
	// if err != nil {
	// 	panic(err)
	// }

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

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == `/` {
		if r.Method != http.MethodPost {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		handlePost(w, r)
	} else {
		if r.Method != http.MethodGet {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		handleGet(w, r)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	body := string(bodyBytes)

	_, err = url.ParseRequestURI(body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	shortenUrl := shortenUrl(body)
	resultUrl := "http://localhost:8080/" + shortenUrl
	urlStorage[shortenUrl] = body

	w.Write([]byte(resultUrl))
	w.WriteHeader(http.StatusCreated)
}

func shortenUrl(url string) string {
	log.Println("=================================================")
	log.Println("Url: ", url)
	hash := sha256.New()
	hash.Write([]byte(url))
	resultHash := hash.Sum(nil)
	result := base64.RawURLEncoding.EncodeToString(resultHash)
	log.Println("Result: ", result)

	return result[:8]
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]

	originalURL, err := urlStorage[id]
	if !err {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
