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

// Engine gère l'exécution des rejeux et des transferts de webhooks.
type Engine struct {
	db     *database.DB
	hub    *hub.Hub
	client *http.Client
}

// NewEngine initialise le moteur de rejeu avec un client HTTP configuré pour la haute performance.
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

// ReplayRequest réémet une requête webhook capturée vers une URL cible.
func (e *Engine) ReplayRequest(req *models.WebhookRequest, payload models.TriggerReplayPayload) (*models.ReplayLog, error) {
	targetURL := payload.TargetURL
	if targetURL == "" {
		// Récupération de l'URL de redirection par défaut de l'endpoint
		ep, err := e.db.GetEndpointByID(req.EndpointID)
		if err == nil && ep != nil && ep.ForwardURL != "" {
			targetURL = ep.ForwardURL
		}
	}

	if targetURL == "" {
		return nil, fmt.Errorf("aucune URL cible définie pour le rejeu")
	}

	// Corps de la requête (original ou muté)
	bodyToSend := req.RawBody
	if payload.ModifiedBody != nil {
		bodyToSend = *payload.ModifiedBody
	}

	// Préparation de la requête HTTP
	httpReq, err := http.NewRequest(req.Method, targetURL, bytes.NewBufferString(bodyToSend))
	if err != nil {
		return nil, fmt.Errorf("impossible de construire la requête HTTP de rejeu : %w", err)
	}

	// Transfert des en-têtes d'origine (en excluant les en-têtes hop-by-hop)
	for key, values := range req.Headers {
		lowerKey := strings.ToLower(key)
		if lowerKey == "host" || lowerKey == "content-length" || lowerKey == "transfer-encoding" || lowerKey == "connection" {
			continue
		}
		for _, val := range values {
			httpReq.Header.Add(key, val)
		}
	}

	// Ajout des en-têtes personnalisés additionnels
	for key, val := range payload.CustomHeaders {
		httpReq.Header.Set(key, val)
	}

	// Marqueur d'identification ZeroDrop
	httpReq.Header.Set("X-ZeroDrop-Replay", "true")
	httpReq.Header.Set("X-ZeroDrop-Original-ID", req.ID)

	// Mesure du temps d'exécution
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

		// Lecture du corps de réponse (limité à 1 Mo)
		respBodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if readErr == nil {
			replayLog.ResponseBody = string(respBodyBytes)
		}
	}

	// Sauvegarde dans SQLite
	_ = e.db.SaveReplayLog(replayLog)

	// Diffusion de l'événement de rejeu sur le Hub WebSocket
	if e.hub != nil {
		e.hub.BroadcastEvent(models.EventReplayResult, replayLog)
	}

	return &replayLog, nil
}
