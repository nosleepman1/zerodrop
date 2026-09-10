// Package models définit les structures de données fondamentales utilisées à travers
// les différentes couches de l'application ZeroDrop (persistance, API, hub et moteur de rejeu).
package models

import "time"

// Endpoint représente un point d'écoute (boîte de réception) dédié à la capture de webhooks.
// Chaque endpoint est identifié de façon unique par un slug d'URL, accessible publiquement
// via la route /in/{slug}.
//
// Un endpoint peut être associé à un fournisseur spécifique (ex: 'stripe', 'github', 'shopify')
// afin d'activer la vérification cryptographique automatique de la signature des requêtes entrantes
// à l'aide du secret HMAC partagé.
type Endpoint struct {
	// ID est l'identifiant unique interne de l'endpoint (ex: "ep_a1b2c3d4e5f6").
	ID string `json:"id"`

	// Name est le libellé lisible par l'utilisateur (ex: "Stripe Production").
	Name string `json:"name"`

	// Slug est le segment d'URL unique pour l'ingestion (ex: "stripe-prod" accessible sur /in/stripe-prod).
	Slug string `json:"slug"`

	// Secret est la clé secrète partagée utilisée pour valider l'intégrité et l'authenticité des signatures HMAC.
	Secret string `json:"secret,omitempty"`

	// Provider indique le format de signature attendu : 'stripe', 'github', 'shopify', 'slack' ou 'custom'.
	Provider string `json:"provider"`

	// ForwardURL est l'URL de redirection par défaut vers laquelle relayer automatiquement les requêtes reçues.
	ForwardURL string `json:"forward_url,omitempty"`

	// Description est une note descriptive facultative relative à l'usage de cet endpoint.
	Description string `json:"description,omitempty"`

	// CreatedAt correspond à l'horodatage UTC de création de l'enregistrement.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt correspond à l'horodatage UTC de la dernière mise à jour de l'enregistrement.
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateEndpointPayload encapsule les données requises lors de la création d'un nouvel endpoint via l'API REST.
type CreateEndpointPayload struct {
	// Name est le nom attribué à l'endpoint (obligatoire).
	Name string `json:"name"`

	// Slug est le chemin d'accès unique d'ingestion (obligatoire, format alphanumérique avec tirets).
	Slug string `json:"slug"`

	// Secret est la clé secrète HMAC optionnelle pour valider les requêtes.
	Secret string `json:"secret,omitempty"`

	// Provider définit l'algorithme de vérification (valeur par défaut : 'custom').
	Provider string `json:"provider,omitempty"`

	// ForwardURL est l'URL de destination locale ou distante par défaut.
	ForwardURL string `json:"forward_url,omitempty"`

	// Description est un texte d'accompagnement contextuel.
	Description string `json:"description,omitempty"`
}

// UpdateEndpointPayload contient les champs modifiables d'un endpoint existant.
// Les pointeurs permettent de distinguer l'absence de modification d'une valeur explicitement vide.
type UpdateEndpointPayload struct {
	Name        *string `json:"name,omitempty"`
	Secret      *string `json:"secret,omitempty"`
	Provider    *string `json:"provider,omitempty"`
	ForwardURL  *string `json:"forward_url,omitempty"`
	Description *string `json:"description,omitempty"`
}
