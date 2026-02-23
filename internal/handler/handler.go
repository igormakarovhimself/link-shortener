package handler

import (
	"encoding/json"
	"errors"
	"io"
	"link-shortener/internal/middleware"
	"link-shortener/internal/repository"
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

type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

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

func (h *URLHandler) handleConflictError(w http.ResponseWriter, conflictErr *repository.ConflictError) {
	resultURL := h.baseURL + "/" + conflictErr.ShortURL
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte(resultURL))
}

func (h *URLHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL := string(bodyBytes)
	userID := middleware.GetUserID(r.Context())

	shortURL, err := h.service.ShortenURL(r.Context(), originalURL, userID)

	if err != nil {
		var conflictErr *repository.ConflictError
		if errors.As(err, &conflictErr) {
			h.handleConflictError(w, conflictErr)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resultURL := h.baseURL + "/" + shortURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (h *URLHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")

	originalURL, err := h.service.GetOriginalURL(r.Context(), shortURL)

	if err != nil {
		if errors.Is(err, repository.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
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

	userID := middleware.GetUserID(r.Context())

	shortURL, err := h.service.ShortenURL(r.Context(), req.URL, userID)
	if err != nil {
		var conflictErr *repository.ConflictError
		if errors.As(err, &conflictErr) {
			resultURL := h.baseURL + "/" + conflictErr.ShortURL
			resp := ShortenResponse{
				Result: resultURL,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
			return
		}
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
	if err := h.service.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) HandleAPIBatch(w http.ResponseWriter, r *http.Request) {
	var requests []BatchShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	shortURLs := make([]string, len(requests))
	originalURLs := make([]string, len(requests))

	for i, req := range requests {
		shortURL, err := h.service.GenerateShortURL(req.OriginalURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shortURLs[i] = shortURL
		originalURLs[i] = req.OriginalURL
	}

	if err := h.service.SaveBatch(r.Context(), shortURLs, originalURLs, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]BatchShortenResponse, len(requests))
	for i, req := range requests {
		responses[i] = BatchShortenResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      h.baseURL + "/" + shortURLs[i],
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}

func (h *URLHandler) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.service.GetURLsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type UserURLResponse struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	responses := make([]UserURLResponse, len(urls))
	for i, url := range urls {
		responses[i] = UserURLResponse{
			ShortURL:    h.baseURL + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

func (h *URLHandler) HandleDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.service.DeleteURLsAsync(r.Context(), shortURLs, userID)

	w.WriteHeader(http.StatusAccepted)
}
