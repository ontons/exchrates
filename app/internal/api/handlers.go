package api

import (
	"encoding/json"
	"exchrates/internal/service"
	"net/http"
)

const (
	Error500 = "500 Internal Server Error"
)

type Handler struct {
	svc *service.RateService
}

func NewHandler(s *service.RateService) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) GetLatest(w http.ResponseWriter, r *http.Request) {
	rates, err := h.svc.GetLatest()
	if err != nil {
		h.svc.Logger.Error(err, "Failed to get latest rates")
		http.Error(w, Error500, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rates); err != nil {
		h.svc.Logger.Error(err, "Failed to encode latest rates")
		http.Error(w, Error500, http.StatusInternalServerError)
	}
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	currency := r.URL.Query().Get("currency")
	if currency == "" {
		http.Error(w, "Missing currency parameter", http.StatusBadRequest)
		return
	}

	history, err := h.svc.GetHistory(currency)
	if err != nil {
		h.svc.Logger.Error(err, "Failed to get history for currency", "currency", currency)
		http.Error(w, Error500, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		h.svc.Logger.Error(err, "Failed to encode history for currency", "currency", currency)
		http.Error(w, Error500, http.StatusInternalServerError)
	}
}
