package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/models"
	"github.com/nosleepman1/zerodrop/internal/replay"
)

// RequestsHandler gère la consultation, le rejeu et la purge des requêtes webhooks.
type RequestsHandler struct {
	db           *database.DB
	replayEngine *replay.Engine
}

// NewRequestsHandler instancie le gestionnaire de requêtes.
func NewRequestsHandler(db *database.DB, re *replay.Engine) *RequestsHandler {
	return &RequestsHandler{
		db:           db,
		replayEngine: re,
	}
}

// List retourne la liste filtrée et paginée des requêtes.
func (h *RequestsHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	filter := models.RequestListFilter{
		EndpointID: query.Get("endpoint_id"),
		Search:     query.Get("search"),
		Limit:      limit,
		Offset:     offset,
	}

	requests, err := h.db.ListWebhookRequests(filter)
	if err != nil {
		http.Error(w, `{"error": "impossible de récupérer les requêtes"}`, http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []models.WebhookRequest{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(requests)
}

// GetByID retourne les détails complets d'une requête avec son body brut.
func (h *RequestsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	req, err := h.db.GetWebhookRequestByID(id)
	if err != nil {
		http.Error(w, `{"error": "erreur interne"}`, http.StatusInternalServerError)
		return
	}
	if req == nil {
		http.Error(w, `{"error": "requête introuvable"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(req)
}

// Replay déclenche le rejeu d'une requête spécifique.
func (h *RequestsHandler) Replay(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	req, err := h.db.GetWebhookRequestByID(id)
	if err != nil || req == nil {
		http.Error(w, `{"error": "requête introuvable pour le rejeu"}`, http.StatusNotFound)
		return
	}

	var payload models.TriggerReplayPayload
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	}

	replayLog, err := h.replayEngine.ReplayRequest(req, payload)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(replayLog)
}

// GetReplays retourne l'historique des rejeux d'une requête.
func (h *RequestsHandler) GetReplays(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	replays, err := h.db.GetReplaysForRequest(id)
	if err != nil {
		http.Error(w, `{"error": "impossible de récupérer les logs de rejeu"}`, http.StatusInternalServerError)
		return
	}

	if replays == nil {
		replays = []models.ReplayLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(replays)
}

// Clear purge les requêtes de la base de données.
func (h *RequestsHandler) Clear(w http.ResponseWriter, r *http.Request) {
	endpointID := r.URL.Query().Get("endpoint_id")
	if err := h.db.ClearRequests(endpointID); err != nil {
		http.Error(w, `{"error": "impossible de purger les requêtes"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
