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

// IngestHandler gère la réception des requêtes webhooks externes sur /in/{slug}.
type IngestHandler struct {
	db  *database.DB
	hub *hub.Hub
}

// NewIngestHandler instancie le handler d'ingestion.
func NewIngestHandler(db *database.DB, h *hub.Hub) *IngestHandler {
	return &IngestHandler{
		db:  db,
		hub: h,
	}
}

// HandleIngest intercepte, vérifie, sauvegarde et retransmet la requête webhook entrante.
func (h *IngestHandler) HandleIngest(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, `{"error": "slug d'endpoint manquant"}`, http.StatusBadRequest)
		return
	}

	// 1. Recherche de l'endpoint correspondant
	ep, err := h.db.GetEndpointBySlug(slug)
	if err != nil {
		http.Error(w, `{"error": "erreur interne lors de la vérification de l'endpoint"}`, http.StatusInternalServerError)
		return
	}
	if ep == nil {
		http.Error(w, `{"error": "endpoint introuvable pour ce slug"}`, http.StatusNotFound)
		return
	}

	// 2. Lecture intégrale du corps brut (limité à 10 Mo par sécurité)
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	if err != nil {
		http.Error(w, `{"error": "impossible de lire le corps de la requête"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 3. Extraction de l'adresse IP client
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			clientIP = host
		}
	}

	// 4. Vérification cryptographique de la signature si un secret est configuré
	var signatureValid *bool
	if ep.Secret != "" {
		valid, _ := security.VerifySignature(ep.Provider, bodyBytes, r.Header, ep.Secret)
		signatureValid = &valid
	}

	// 5. Construction de l'objet WebhookRequest
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

	// 6. Sauvegarde immédiate dans la base SQLite
	if err := h.db.SaveWebhookRequest(webhookReq); err != nil {
		http.Error(w, `{"error": "impossible de persister la requête"}`, http.StatusInternalServerError)
		return
	}

	// 7. Diffusion temps réel
	if h.hub != nil {
		// Broadcast aux dashboards connectés
		h.hub.BroadcastEvent(models.EventNewRequest, webhookReq)

		// Forward automatique au tunnel CLI s'il est actif
		_ = h.hub.ForwardToTunnel(ep.Slug, webhookReq)
	}

	// 8. Réponse immédiate 202 Accepted
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
