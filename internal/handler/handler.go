package handler

import (
	"io"
	"link-shortener/internal/service"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	service service.ShortenerService
	baseURL string
}

func NewHandler(service service.ShortenerService, baseURL string) *URLHandler {
	return &URLHandler{
		service: service,
		baseURL: baseURL,
	}
}

func (h *URLHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL := string(bodyBytes)

	_, err = url.ParseRequestURI(originalURL)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.ShortenURL(originalURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resultURL := h.baseURL + "/" + shortURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (h *URLHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")

	originalURL, err := h.service.GetOriginalURL(shortURL)

	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
