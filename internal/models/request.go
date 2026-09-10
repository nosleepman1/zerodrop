package models

import "time"

// WebhookRequest représente l'état immuable d'une requête HTTP capturée par la passerelle d'ingestion ZeroDrop.
// Elle stocke les données brutes sans altération afin de permettre une inspection fidèle et des rejeux ultérieurs.
type WebhookRequest struct {
	// ID est l'identifiant unique généré lors de la réception (ex: "req_01HZX...").
	ID string `json:"id"`

	// EndpointID est la clé étrangère liant cette requête à son endpoint récepteur.
	EndpointID string `json:"endpoint_id"`

	// EndpointSlug est le slug d'URL de l'endpoint récepteur, dénormalisé pour optimiser les requêtes de lecture.
	EndpointSlug string `json:"endpoint_slug"`

	// Method est le verbe HTTP de la requête originale (généralement POST, PUT, PATCH ou DELETE).
	Method string `json:"method"`

	// Path est le chemin d'URL complet reçu (ex: "/in/stripe-dev/v1/events").
	Path string `json:"path"`

	// Headers contient l'intégralité des en-têtes HTTP de la requête sous forme de dictionnaire clé -> liste de valeurs.
	Headers map[string][]string `json:"headers"`

	// QueryParams contient les paramètres de chaîne de requête d'URL (query string) extraits de l'URL.
	QueryParams map[string][]string `json:"query_params,omitempty"`

	// RawBody est le corps textuel brut et exact de la requête HTTP sans aucune modification ni troncature.
	RawBody string `json:"raw_body"`

	// ContentType est le type MIME déclaré dans l'en-tête Content-Type (ex: "application/json").
	ContentType string `json:"content_type"`

	// ContentLength est la taille en octets du corps de la charge utile.
	ContentLength int64 `json:"content_length"`

	// IPAddress est l'adresse IP de l'émetteur (extraite de X-Forwarded-For, X-Real-IP ou RemoteAddr).
	IPAddress string `json:"ip_address"`

	// SignatureValid indique le résultat de la validation cryptographique :
	// - true  : Signature valide et conforme au secret configuré.
	// - false : Signature invalide ou altérée.
	// - nil   : Aucun secret configuré sur l'endpoint récepteur (requête non vérifiée).
	SignatureValid *bool `json:"signature_valid"`

	// CreatedAt est l'horodatage UTC exact auquel la requête a été interceptée par la passerelle.
	CreatedAt time.Time `json:"created_at"`
}

// RequestListFilter spécifie les critères de filtrage et de pagination pour la consultation des requêtes capturées.
type RequestListFilter struct {
	// EndpointID filtre les requêtes pour un endpoint spécifique (laisser vide pour tous les endpoints).
	EndpointID string `json:"endpoint_id"`

	// Search applique un filtre textuel sur le corps de la requête, le chemin d'accès ou l'identifiant.
	Search string `json:"search"`

	// Limit définit le nombre maximal de requêtes retournées (valeur par défaut : 50, maximum : 100).
	Limit int `json:"limit"`

	// Offset définit le décalage pour la pagination des résultats.
	Offset int `json:"offset"`
}
