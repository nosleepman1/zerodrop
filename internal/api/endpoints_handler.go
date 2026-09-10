package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/hub"
	"github.com/nosleepman1/zerodrop/internal/models"
)

// EndpointsHandler gère les opérations REST relatives aux boîtes de réception (Endpoints).
type EndpointsHandler struct {
	db  *database.DB
	hub *hub.Hub
}

// NewEndpointsHandler instancie le gestionnaire d'endpoints.
func NewEndpointsHandler(db *database.DB, h *hub.Hub) *EndpointsHandler {
	return &EndpointsHandler{db: db, hub: h}
}

// List retourne la liste complète des endpoints configurés.
func (h *EndpointsHandler) List(w http.ResponseWriter, r *http.Request) {
	endpoints, err := h.db.ListEndpoints()
	if err != nil {
		http.Error(w, `{"error": "impossible de lister les endpoints"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(endpoints)
}

// Create ajoute un nouvel endpoint après vérification d'unicité du slug.
func (h *EndpointsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload models.CreateEndpointPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error": "corps de requete JSON invalide"}`, http.StatusBadRequest)
		return
	}

	if payload.Name == "" || payload.Slug == "" {
		http.Error(w, `{"error": "les champs 'name' et 'slug' sont obligatoires"}`, http.StatusBadRequest)
		return
	}

	existing, _ := h.db.GetEndpointBySlug(payload.Slug)
	if existing != nil {
		http.Error(w, `{"error": "un endpoint portant ce slug existe deja"}`, http.StatusConflict)
		return
	}

	ep, err := h.db.CreateEndpoint(payload)
	if err != nil {
		http.Error(w, `{"error": "erreur lors de la creation de l'endpoint"}`, http.StatusInternalServerError)
		return
	}

	if h.hub != nil {
		h.hub.BroadcastEvent(models.EventEndpointUpdate, ep)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ep)
}

// GetByID retourne les détails d'un endpoint spécifié par son identifiant unique.
func (h *EndpointsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ep, err := h.db.GetEndpointByID(id)
	if err != nil {
		http.Error(w, `{"error": "erreur interne de recherche"}`, http.StatusInternalServerError)
		return
	}
	if ep == nil {
		http.Error(w, `{"error": "endpoint introuvable"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ep)
}

// Delete supprime un endpoint et ses requêtes associées.
func (h *EndpointsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "ep_default" {
		http.Error(w, `{"error": "l'endpoint par defaut ne peut pas etre supprime"}`, http.StatusForbidden)
		return
	}

	if err := h.db.DeleteEndpoint(id); err != nil {
		http.Error(w, `{"error": "impossible de supprimer l'endpoint"}`, http.StatusInternalServerError)
		return
	}

	if h.hub != nil {
		h.hub.BroadcastEvent(models.EventEndpointUpdate, map[string]string{"deleted_id": id})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": id})
}
