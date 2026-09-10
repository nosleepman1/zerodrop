package models

import "time"

// ReplayLog conserve l'historique de chaque tentative de transfert ou de rejeu manuel.
type ReplayLog struct {
	ID              string              `json:"id"`                         // Identifiant unique du log (ex: rep_01HZX...)
	RequestID       string              `json:"request_id"`                 // ID de la requête réémise
	TargetURL       string              `json:"target_url"`                 // URL cible vers laquelle la requête a été envoyée
	StatusCode      int                 `json:"status_code"`                // Code de statut HTTP reçu (ex: 200, 500)
	ResponseHeaders map[string][]string `json:"response_headers,omitempty"` // En-têtes HTTP retournés par la cible
	ResponseBody    string              `json:"response_body,omitempty"`    // Corps de la réponse retournée
	DurationMs      int64               `json:"duration_ms"`                // Temps d'exécution en millisecondes
	ErrorMessage    string              `json:"error_message,omitempty"`    // Message d'erreur (timeout, connexion refusée...)
	CreatedAt       time.Time           `json:"created_at"`                 // Date du rejeu
}

// TriggerReplayPayload contient les options pour relancer une requête.
type TriggerReplayPayload struct {
	TargetURL      string              `json:"target_url,omitempty"`      // URL cible (si vide, utilise ForwardURL de l'endpoint)
	ModifiedBody   *string             `json:"modified_body,omitempty"`   // Corps modifié si mutation voulue
	CustomHeaders  map[string]string   `json:"custom_headers,omitempty"`  // En-têtes personnalisés additionnels
}
