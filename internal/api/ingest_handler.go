package api

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/hub"
	"github.com/nosleepman1/zerodrop/internal/models"
	"github.com/nosleepman1/zerodrop/internal/security"
)

// IngestHandler gère la réception des requêtes webhooks externes sur la route /in/{slug}.
type IngestHandler struct {
	db  *database.DB
	hub *hub.Hub
}

// NewIngestHandler instancie le gestionnaire d'ingestion.
func NewIngestHandler(db *database.DB, h *hub.Hub) *IngestHandler {
	return &IngestHandler{db: db, hub: h}
}

// HandleIngest intercepte, valide cryptographiquement, persiste et diffuse la requête webhook entrante.
// Répond immédiatement avec un code HTTP 202 Accepted.
func (h *IngestHandler) HandleIngest(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, `{"error": "slug d'endpoint manquant"}`, http.StatusBadRequest)
		return
	}

	ep, err := h.db.GetEndpointBySlug(slug)
	if err != nil {
		http.Error(w, `{"error": "erreur interne lors de la verification de l'endpoint"}`, http.StatusInternalServerError)
		return
	}
	if ep == nil {
		http.Error(w, `{"error": "endpoint introuvable pour ce slug"}`, http.StatusNotFound)
		return
	}

	// Lecture intégrale du flux d'octets sans troncature (limité à 10 Mo)
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	if err != nil {
		http.Error(w, `{"error": "impossible de lire le corps de la requete"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	} else {
		host, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr == nil {
			clientIP = host
		}
	}

	var signatureValid *bool
	if ep.Secret != "" {
		valid, _ := security.VerifySignature(ep.Provider, bodyBytes, r.Header, ep.Secret)
		signatureValid = &valid
	}

	reqID := "req_" + uuid.New().String()[:12]
	now := time.Now().UTC()

	webhookReq := models.WebhookRequest{
		ID:             reqID,
		EndpointID:     ep.ID,
		EndpointSlug:   ep.Slug,
		Method:         r.Method,
		Path:           r.URL.Path,
		Headers:        r.Header,
		QueryParams:    r.URL.Query(),
		RawBody:        string(bodyBytes),
		ContentType:    r.Header.Get("Content-Type"),
		ContentLength:  int64(len(bodyBytes)),
		IPAddress:      clientIP,
		SignatureValid: signatureValid,
		CreatedAt:      now,
	}

	if err := h.db.SaveWebhookRequest(webhookReq); err != nil {
		http.Error(w, `{"error": "echec de persistance de la requete"}`, http.StatusInternalServerError)
		return
	}

	if h.hub != nil {
		h.hub.BroadcastEvent(models.EventNewRequest, webhookReq)
		_ = h.hub.ForwardToTunnel(ep.Slug, webhookReq)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "accepted",
		"id":              reqID,
		"endpoint":        ep.Slug,
		"signature_valid": signatureValid,
		"received_at":     now.Format(time.RFC3339),
	})
}
