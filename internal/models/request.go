package models

import "time"

// WebhookRequest représente une requête HTTP entrante capturée par ZeroDrop.
type WebhookRequest struct {
	ID             string              `json:"id"`                       // Identifiant unique (ex: req_01HZX...)
	EndpointID     string              `json:"endpoint_id"`             // ID de l'endpoint récepteur
	EndpointSlug   string              `json:"endpoint_slug"`           // Slug pour affichage rapide
	Method         string              `json:"method"`                  // POST, GET, PUT, PATCH, DELETE
	Path           string              `json:"path"`                    // Chemin complet reçu
	Headers        map[string][]string `json:"headers"`                 // En-têtes HTTP bruts
	QueryParams    map[string][]string `json:"query_params,omitempty"`   // Paramètres d'URL
	RawBody        string              `json:"raw_body"`                // Corps de la requête tel quel
	ContentType    string              `json:"content_type"`            // Type MIME (ex: application/json)
	ContentLength  int64               `json:"content_length"`          // Taille en octets
	IPAddress      string              `json:"ip_address"`              // Adresse IP de l'émetteur
	SignatureValid *bool               `json:"signature_valid"`         // true, false, ou null (si pas de secret)
	CreatedAt      time.Time           `json:"created_at"`              // Date de capture
}

// RequestListFilter contient les paramètres de filtrage pour lister les requêtes.
type RequestListFilter struct {
	EndpointID string `json:"endpoint_id"`
	Search     string `json:"search"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}
