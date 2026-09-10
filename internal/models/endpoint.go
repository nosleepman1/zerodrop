package models

import "time"

// Endpoint représente un point d'écoute (boîte de réception) pour les webhooks.
// Chaque endpoint dispose d'un slug unique accessible via /in/{slug}.
type Endpoint struct {
	ID          string    `json:"id"`                     // Identifiant unique (ex: ep_01HZX...)
	Name        string    `json:"name"`                   // Nom lisible (ex: "Stripe Production")
	Slug        string    `json:"slug"`                   // Segment d'URL unique (ex: "stripe-prod")
	Secret      string    `json:"secret,omitempty"`      // Secret HMAC partagé pour validation
	Provider    string    `json:"provider"`               // Fournisseur : 'stripe', 'github', 'shopify', 'custom'
	ForwardURL  string    `json:"forward_url,omitempty"`  // URL de redirection locale ou distante par défaut
	Description string    `json:"description,omitempty"`  // Description optionnelle
	CreatedAt   time.Time `json:"created_at"`             // Date de création
	UpdatedAt   time.Time `json:"updated_at"`             // Date de dernière modification
}

// CreateEndpointPayload représente le corps de la requête pour créer un endpoint.
type CreateEndpointPayload struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Secret      string `json:"secret,omitempty"`
	Provider    string `json:"provider,omitempty"`
	ForwardURL  string `json:"forward_url,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateEndpointPayload représente le corps de la requête pour modifier un endpoint.
type UpdateEndpointPayload struct {
	Name        *string `json:"name,omitempty"`
	Secret      *string `json:"secret,omitempty"`
	Provider    *string `json:"provider,omitempty"`
	ForwardURL  *string `json:"forward_url,omitempty"`
	Description *string `json:"description,omitempty"`
}
