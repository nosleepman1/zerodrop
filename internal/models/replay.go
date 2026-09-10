package models

import "time"

// ReplayLog consigne le résultat d'une tentative de rejeu ou de retransmission d'une requête webhook.
// Il enregistre la réponse HTTP complète retournée par le serveur cible ainsi que les métriques de performance.
type ReplayLog struct {
	// ID est l'identifiant unique du journal de rejeu (ex: "rep_01HZX...").
	ID string `json:"id"`

	// RequestID fait référence à l'identifiant de la WebhookRequest originale réémise.
	RequestID string `json:"request_id"`

	// TargetURL est l'URL de destination vers laquelle la requête a été expédiée lors de ce rejeu.
	TargetURL string `json:"target_url"`

	// StatusCode est le code de statut HTTP retourné par le serveur distant (ex: 200, 404, 500), ou 0 en cas d'échec réseau.
	StatusCode int `json:"status_code"`

	// ResponseHeaders contient les en-têtes HTTP renvoyés par le serveur distant.
	ResponseHeaders map[string][]string `json:"response_headers,omitempty"`

	// ResponseBody est le corps de réponse textuel retourné par la cible (limité à 1 Mo).
	ResponseBody string `json:"response_body,omitempty"`

	// DurationMs est la durée totale d'aller-retour réseau en millisecondes.
	DurationMs int64 `json:"duration_ms"`

	// ErrorMessage contient la description de l'erreur réseau (ex: "connexion refusée", "délai d'attente dépassé").
	ErrorMessage string `json:"error_message,omitempty"`

	// CreatedAt est l'horodatage UTC du déclenchement du rejeu.
	CreatedAt time.Time `json:"created_at"`
}

// TriggerReplayPayload contient les paramètres personnalisables lors du déclenchement d'un rejeu.
type TriggerReplayPayload struct {
	// TargetURL permet de spécifier une URL cible alternative. Si omis, l'URL par défaut de l'endpoint est utilisée.
	TargetURL string `json:"target_url,omitempty"`

	// ModifiedBody permet de modifier la charge utile avant expédition (mutation de payload).
	ModifiedBody *string `json:"modified_body,omitempty"`

	// CustomHeaders permet d'injecter ou d'écraser des en-têtes HTTP spécifiques pour ce rejeu.
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}
