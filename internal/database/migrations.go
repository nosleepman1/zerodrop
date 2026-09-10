package database

import (
	"fmt"
	"time"
)

// Migrate applique les schémas SQL nécessaires de manière idempotente.
func (db *DB) Migrate() error {
	schema := `
	-- Table 1 : Endpoints (Boîtes de réception des webhooks)
	CREATE TABLE IF NOT EXISTS endpoints (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		secret TEXT,
		provider TEXT NOT NULL DEFAULT 'custom',
		forward_url TEXT,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_endpoints_slug ON endpoints(slug);

	-- Table 2 : Webhook Requests (Chaque requête entrante capturée)
	CREATE TABLE IF NOT EXISTS webhook_requests (
		id TEXT PRIMARY KEY,
		endpoint_id TEXT NOT NULL,
		endpoint_slug TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		headers TEXT NOT NULL,          -- JSON sérialisé des en-têtes
		query_params TEXT,              -- JSON sérialisé des query params
		raw_body TEXT NOT NULL,         -- Corps brut intégral
		content_type TEXT,
		content_length INTEGER NOT NULL DEFAULT 0,
		ip_address TEXT,
		signature_valid INTEGER,        -- 1 = valide, 0 = invalide, NULL = pas de vérification
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_requests_endpoint_id ON webhook_requests(endpoint_id);
	CREATE INDEX IF NOT EXISTS idx_requests_created_at ON webhook_requests(created_at DESC);

	-- Table 3 : Replay Logs (Historique des tentatives de rejeu)
	CREATE TABLE IF NOT EXISTS replay_logs (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		target_url TEXT NOT NULL,
		status_code INTEGER,
		response_headers TEXT,          -- JSON sérialisé
		response_body TEXT,
		duration_ms INTEGER NOT NULL DEFAULT 0,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(request_id) REFERENCES webhook_requests(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_replay_request_id ON replay_logs(request_id);

	-- Table 4 : API Keys (Pour sécuriser le CLI et l'API distante)
	CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("erreur lors de l'exécution du schéma de base : %w", err)
	}

	// Initialisation automatique d'un endpoint par défaut si la table est vide
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM endpoints").Scan(&count)
	if err != nil {
		return fmt.Errorf("impossible de vérifier les endpoints existants : %w", err)
	}

	if count == 0 {
		now := time.Now().UTC()
		_, err := db.Exec(`
			INSERT INTO endpoints (id, name, slug, secret, provider, forward_url, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, "ep_default", "Endpoint Principal", "default", "", "custom", "http://localhost:3000/webhook", "Endpoint de test par défaut prêt à l'emploi", now, now)
		if err != nil {
			return fmt.Errorf("erreur lors de la création de l'endpoint par défaut : %w", err)
		}
	}

	return nil
}
