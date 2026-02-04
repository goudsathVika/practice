package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"practice/internal/shortener"
)

type Handler struct {
	service *shortener.Service
	baseURL string
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

type metricsResponse struct {
	Domains []shortener.DomainCount `json:"domains"`
}

func NewHandler(service *shortener.Service, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/shorten", h.handleShorten)
	mux.HandleFunc("/metrics", h.handleMetrics)
	mux.HandleFunc("/", h.handleRedirect)
}

func (h *Handler) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid url"})
		return
	}

	code, err := h.service.Shorten(req.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid url"})
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURLForRequest(r), code)
	writeJSON(w, http.StatusOK, shortenResponse{ShortURL: shortURL})
}

func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := metricsResponse{Domains: h.service.Metrics(3)}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || path == "shorten" || path == "metrics" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if target, ok := h.service.Resolve(path); ok {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) baseURLForRequest(r *http.Request) string {
	if h.baseURL != "" {
		return strings.TrimSuffix(h.baseURL, "/")
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
