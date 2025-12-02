package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"link-shortener/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type URLHandler struct {
	service service.ShortenerService
	baseURL string
	db      *sql.DB
}

func NewHandler(service service.ShortenerService, baseURL string, db *sql.DB) *URLHandler {
	return &URLHandler{
		service: service,
		baseURL: baseURL,
		db:      db,
	}
}

func (h *URLHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL := string(bodyBytes)

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

func (h *URLHandler) HandleAPIShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.ShortenURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resultURL := h.baseURL + "/" + shortURL
	resp := ShortenResponse{
		Result: resultURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *URLHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := h.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
