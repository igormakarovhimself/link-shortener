package handler

import (
	"io"
	"link-shortener/internal/service"
	"net/http"
	"net/url"
)

type URLHandler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService) *URLHandler {
	return &URLHandler{
		service: service,
	}
}

func (h *URLHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == `/` {
		if r.Method != http.MethodPost {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		h.HandlePost(w, r)
	} else {
		if r.Method != http.MethodGet {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		h.handleGet(w, r)
	}
}

func (h *URLHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	originalURL := string(bodyBytes)

	_, err = url.ParseRequestURI(originalURL)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	shortURL, err := h.service.ShortenURL(originalURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resultURL := "http://localhost:8080/" + shortURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (h *URLHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]

	originalURL, err := h.service.GetOriginalURL(shortURL)

	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
