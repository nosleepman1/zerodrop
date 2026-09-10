// Package replay implémente le moteur d'exécution des rejeux de requêtes HTTP.
// Il permet de réémettre fidèlement des webhooks capturés vers des cibles locales ou distantes,
// en conservant les en-têtes d'origine et en supportant la mutation de charge utile.
package replay

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/hub"
	"github.com/nosleepman1/zerodrop/internal/models"
)

// Engine encapsule le client HTTP haut débit et les dépendances vers la persistance et le hub d'événements.
type Engine struct {
	db     *database.DB
	hub    *hub.Hub
	client *http.Client
}

// NewEngine initialise un nouveau moteur de rejeu doté d'un pool de connexions HTTP réutilisables.
func NewEngine(db *database.DB, hub *hub.Hub) *Engine {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	return &Engine{
		db:     db,
		hub:    hub,
		client: client,
	}
}

// ReplayRequest réémet une requête webhook capturée vers une URL cible spécifiée ou par défaut.
//
// Processus :
//  1. Détermination de l'URL cible (priorité à TargetURL du payload, puis ForwardURL de l'endpoint).
//  2. Sélection du corps de requête (corps original ou corps modifié fourni dans payload.ModifiedBody).
//  3. Recréation de la requête HTTP en filtrant les en-têtes de transport ("hop-by-hop").
//  4. Injection des en-têtes de traçabilité ZeroDrop (X-ZeroDrop-Replay, X-ZeroDrop-Original-ID).
//  5. Mesure de la latence, capture du code de réponse et sauvegarde du journal dans SQLite.
//  6. Diffusion de l'événement EVENT_REPLAY_RESULT sur le flux WebSocket.
func (e *Engine) ReplayRequest(req *models.WebhookRequest, payload models.TriggerReplayPayload) (*models.ReplayLog, error) {
	targetURL := payload.TargetURL
	if targetURL == "" && e.db != nil {
		ep, err := e.db.GetEndpointByID(req.EndpointID)
		if err == nil && ep != nil && ep.ForwardURL != "" {
			targetURL = ep.ForwardURL
		}
	}

	if targetURL == "" {
		return nil, fmt.Errorf("aucune URL cible valide specifiee pour executer le rejeu")
	}

	bodyToSend := req.RawBody
	if payload.ModifiedBody != nil {
		bodyToSend = *payload.ModifiedBody
	}

	httpReq, err := http.NewRequest(req.Method, targetURL, bytes.NewBufferString(bodyToSend))
	if err != nil {
		return nil, fmt.Errorf("erreur de construction de la requete HTTP de rejeu : %w", err)
	}

	// Filtrage des en-têtes hop-by-hop pour laisser la couche transport Go les recalculer
	for key, values := range req.Headers {
		lowerKey := strings.ToLower(key)
		if lowerKey == "host" || lowerKey == "content-length" || lowerKey == "transfer-encoding" || lowerKey == "connection" {
			continue
		}
		for _, val := range values {
			httpReq.Header.Add(key, val)
		}
	}

	// Injection des en-têtes additionnels personnalisés
	for key, val := range payload.CustomHeaders {
		httpReq.Header.Set(key, val)
	}

	// En-têtes de métadonnées ZeroDrop
	httpReq.Header.Set("X-ZeroDrop-Replay", "true")
	httpReq.Header.Set("X-ZeroDrop-Original-ID", req.ID)

	startTime := time.Now()
	resp, err := e.client.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	logID := "rep_" + uuid.New().String()[:12]
	replayLog := models.ReplayLog{
		ID:         logID,
		RequestID:  req.ID,
		TargetURL:  targetURL,
		DurationMs: duration,
		CreatedAt:  time.Now().UTC(),
	}

	if err != nil {
		replayLog.StatusCode = 0
		replayLog.ErrorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		replayLog.StatusCode = resp.StatusCode
		replayLog.ResponseHeaders = resp.Header

		respBodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if readErr == nil {
			replayLog.ResponseBody = string(respBodyBytes)
		}
	}

	if e.db != nil {
		_ = e.db.SaveReplayLog(replayLog)
	}

	if e.hub != nil {
		e.hub.BroadcastEvent(models.EventReplayResult, replayLog)
	}

	return &replayLog, nil
}
