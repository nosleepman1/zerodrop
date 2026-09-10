package database

import (
	"fmt"
	"time"
)

// Migrate applique l'ensemble des définitions de tables et d'index SQL de manière idempotente.
// Cette méthode garantit que l'état du schéma de base de données est aligné lors du démarrage du serveur.
func (db *DB) Migrate() error {
	schema := `
	-- Table des points d'écoute (Endpoints)
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

	-- Table des requêtes webhooks capturées
	CREATE TABLE IF NOT EXISTS webhook_requests (
		id TEXT PRIMARY KEY,
		endpoint_id TEXT NOT NULL,
		endpoint_slug TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		headers TEXT NOT NULL,          -- JSON sérialisé des en-têtes HTTP
		query_params TEXT,              -- JSON sérialisé des paramètres de requête d'URL
		raw_body TEXT NOT NULL,         -- Corps brut intégral non modifié
		content_type TEXT,
		content_length INTEGER NOT NULL DEFAULT 0,
		ip_address TEXT,
		signature_valid INTEGER,        -- 1 = valide, 0 = invalide, NULL = non vérifié
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_requests_endpoint_id ON webhook_requests(endpoint_id);
	CREATE INDEX IF NOT EXISTS idx_requests_created_at ON webhook_requests(created_at DESC);

	-- Table des journaux de rejeu et de transfert
	CREATE TABLE IF NOT EXISTS replay_logs (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		target_url TEXT NOT NULL,
		status_code INTEGER,
		response_headers TEXT,          -- JSON sérialisé des en-têtes de réponse
		response_body TEXT,             -- Corps textuel de la réponse
		duration_ms INTEGER NOT NULL DEFAULT 0,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(request_id) REFERENCES webhook_requests(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_replay_request_id ON replay_logs(request_id);

	-- Table des clés d'API (Authentification du CLI et des intégrations distantes)
	CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("erreur lors de l'execution du schema de base : %w", err)
	}

	// Initialisation automatique d'un endpoint 'default' si la table est vide
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM endpoints").Scan(&count)
	if err != nil {
		return fmt.Errorf("impossible de compter les endpoints existants : %w", err)
	}

	if count == 0 {
		now := time.Now().UTC()
		_, err := db.Exec(`
			INSERT INTO endpoints (id, name, slug, secret, provider, forward_url, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, "ep_default", "Endpoint Principal", "default", "", "custom", "http://localhost:3000/webhook", "Endpoint de test pret a l'emploi", now, now)
		if err != nil {
			return fmt.Errorf("erreur lors de l'initialisation de l'endpoint par defaut : %w", err)
		}
	}

	return nil
}
