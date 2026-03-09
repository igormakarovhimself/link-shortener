// Package handler содержит HTTP-обработчики сервиса сокращения URL.
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"link-shortener/internal/audit"
	"link-shortener/internal/middleware"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

// ShortenRequest — тело запроса для эндпоинта POST /api/shorten.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse — тело ответа для эндпоинта POST /api/shorten.
type ShortenResponse struct {
	Result string `json:"result"`
}

// BatchShortenRequest — один элемент запроса для сокращения батчами.
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortenResponse — один элемент ответа для сокращения батчами.
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// URLHandler обрабатывает HTTP-запросы сервиса сокращения URL.
type URLHandler struct {
	service   service.ShortenerService
	baseURL   string
	publisher *audit.Publisher
}

// NewHandler создает URLHandler с переданным сервисом, базовым URL и publisher-ом аудита.
func NewHandler(service service.ShortenerService, baseURL string, publisher *audit.Publisher) *URLHandler {
	return &URLHandler{
		service:   service,
		baseURL:   baseURL,
		publisher: publisher,
	}
}

func (h *URLHandler) handleConflictError(w http.ResponseWriter, conflictErr *repository.ConflictError) {
	resultURL := h.baseURL + "/" + conflictErr.ShortURL
	w.WriteHeader(http.StatusConflict)
	_, _ = w.Write([]byte(resultURL))
}

// HandlePost принимает оригинальный URL в теле запроса и возвращает короткий URL.
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
	_, _ = w.Write([]byte(resultURL))

	if h.publisher != nil {
		event := audit.NewAuditEvent("shorten", userID, originalURL)
		h.publisher.Notify(event)
	}
}

// HandleGet ищет оригинальный URL по короткому идентификатору и делает редирект 307.
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

	if h.publisher != nil {
		userID := middleware.GetUserID(r.Context())
		event := audit.NewAuditEvent("follow", userID, originalURL)
		h.publisher.Notify(event)
	}
}

// HandleAPIShorten принимает URL в JSON и возвращает короткий URL в JSON.
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
			_ = json.NewEncoder(w).Encode(resp)
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
	_ = json.NewEncoder(w).Encode(resp)

	if h.publisher != nil {
		event := audit.NewAuditEvent("shorten", userID, req.URL)
		h.publisher.Notify(event)
	}
}

// HandlePing проверяет доступность хранилища.
func (h *URLHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleAPIBatch принимает массив URL в JSON и сокращает их пакетом.
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
	_ = json.NewEncoder(w).Encode(responses)
}

// HandleGetUserURLs возвращает все URL текущего пользователя в JSON.
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
	_ = json.NewEncoder(w).Encode(responses)
}

// HandleDeleteUserURLs принимает список коротких URL в JSON и удаляет их.
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
