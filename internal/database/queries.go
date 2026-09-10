package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nosleepman1/zerodrop/internal/models"
)

// ==========================================
// Opérations sur les Endpoints
// ==========================================

// CreateEndpoint crée un nouveau point de réception de webhooks.
func (db *DB) CreateEndpoint(payload models.CreateEndpointPayload) (*models.Endpoint, error) {
	id := "ep_" + uuid.New().String()[:12]
	now := time.Now().UTC()

	provider := payload.Provider
	if provider == "" {
		provider = "custom"
	}

	query := `
		INSERT INTO endpoints (id, name, slug, secret, provider, forward_url, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, id, payload.Name, payload.Slug, payload.Secret, provider, payload.ForwardURL, payload.Description, now, now)
	if err != nil {
		return nil, fmt.Errorf("impossible de créer l'endpoint : %w", err)
	}

	return &models.Endpoint{
		ID:          id,
		Name:        payload.Name,
		Slug:        payload.Slug,
		Secret:      payload.Secret,
		Provider:    provider,
		ForwardURL:  payload.ForwardURL,
		Description: payload.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetEndpointBySlug récupère un endpoint à partir de son slug d'URL (ex: /in/mon-slug).
func (db *DB) GetEndpointBySlug(slug string) (*models.Endpoint, error) {
	query := `
		SELECT id, name, slug, secret, provider, forward_url, description, created_at, updated_at
		FROM endpoints
		WHERE slug = ?
	`

	ep := &models.Endpoint{}
	var secret, forwardURL, description sql.NullString

	err := db.QueryRow(query, slug).Scan(
		&ep.ID, &ep.Name, &ep.Slug, &secret, &ep.Provider, &forwardURL, &description, &ep.CreatedAt, &ep.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recherche de l'endpoint : %w", err)
	}

	if secret.Valid {
		ep.Secret = secret.String
	}
	if forwardURL.Valid {
		ep.ForwardURL = forwardURL.String
	}
	if description.Valid {
		ep.Description = description.String
	}

	return ep, nil
}

// GetEndpointByID récupère un endpoint à partir de son identifiant unique.
func (db *DB) GetEndpointByID(id string) (*models.Endpoint, error) {
	query := `
		SELECT id, name, slug, secret, provider, forward_url, description, created_at, updated_at
		FROM endpoints
		WHERE id = ?
	`

	ep := &models.Endpoint{}
	var secret, forwardURL, description sql.NullString

	err := db.QueryRow(query, id).Scan(
		&ep.ID, &ep.Name, &ep.Slug, &secret, &ep.Provider, &forwardURL, &description, &ep.CreatedAt, &ep.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recherche de l'endpoint : %w", err)
	}

	if secret.Valid {
		ep.Secret = secret.String
	}
	if forwardURL.Valid {
		ep.ForwardURL = forwardURL.String
	}
	if description.Valid {
		ep.Description = description.String
	}

	return ep, nil
}

// ListEndpoints retourne la liste de tous les endpoints configurés.
func (db *DB) ListEndpoints() ([]models.Endpoint, error) {
	query := `
		SELECT id, name, slug, secret, provider, forward_url, description, created_at, updated_at
		FROM endpoints
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des endpoints : %w", err)
	}
	defer rows.Close()

	var endpoints []models.Endpoint
	for rows.Next() {
		var ep models.Endpoint
		var secret, forwardURL, description sql.NullString

		if err := rows.Scan(&ep.ID, &ep.Name, &ep.Slug, &secret, &ep.Provider, &forwardURL, &description, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture d'un endpoint : %w", err)
		}

		if secret.Valid {
			ep.Secret = secret.String
		}
		if forwardURL.Valid {
			ep.ForwardURL = forwardURL.String
		}
		if description.Valid {
			ep.Description = description.String
		}

		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

// DeleteEndpoint supprime un endpoint et ses requêtes associées en cascade.
func (db *DB) DeleteEndpoint(id string) error {
	_, err := db.Exec("DELETE FROM endpoints WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("impossible de supprimer l'endpoint : %w", err)
	}
	return nil
}

// ==========================================
// Opérations sur les Requêtes Webhooks
// ==========================================

// SaveWebhookRequest persiste une requête entrante dans la base de données.
func (db *DB) SaveWebhookRequest(req models.WebhookRequest) error {
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		headersJSON = []byte("{}")
	}

	queryParamsJSON, err := json.Marshal(req.QueryParams)
	if err != nil {
		queryParamsJSON = []byte("{}")
	}

	var sigValid sql.NullInt64
	if req.SignatureValid != nil {
		sigValid.Valid = true
		if *req.SignatureValid {
			sigValid.Int64 = 1
		} else {
			sigValid.Int64 = 0
		}
	}

	query := `
		INSERT INTO webhook_requests (
			id, endpoint_id, endpoint_slug, method, path, headers, query_params,
			raw_body, content_type, content_length, ip_address, signature_valid, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(
		query,
		req.ID,
		req.EndpointID,
		req.EndpointSlug,
		req.Method,
		req.Path,
		string(headersJSON),
		string(queryParamsJSON),
		req.RawBody,
		req.ContentType,
		req.ContentLength,
		req.IPAddress,
		sigValid,
		req.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur lors de la persistance de la requête webhook : %w", err)
	}

	return nil
}

// ListWebhookRequests retourne les requêtes selon les filtres fournis (endpoint, recherche, pagination).
func (db *DB) ListWebhookRequests(filter models.RequestListFilter) ([]models.WebhookRequest, error) {
	query := `
		SELECT id, endpoint_id, endpoint_slug, method, path, headers, query_params,
		       raw_body, content_type, content_length, ip_address, signature_valid, created_at
		FROM webhook_requests
		WHERE 1=1
	`
	var args []interface{}

	if filter.EndpointID != "" {
		query += " AND endpoint_id = ?"
		args = append(args, filter.EndpointID)
	}

	if filter.Search != "" {
		query += " AND (raw_body LIKE ? OR path LIKE ? OR id LIKE ?)"
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recherche des requêtes : %w", err)
	}
	defer rows.Close()

	var requests []models.WebhookRequest
	for rows.Next() {
		var req models.WebhookRequest
		var headersStr, queryParamsStr sql.NullString
		var sigValid sql.NullInt64

		err := rows.Scan(
			&req.ID, &req.EndpointID, &req.EndpointSlug, &req.Method, &req.Path,
			&headersStr, &queryParamsStr, &req.RawBody, &req.ContentType,
			&req.ContentLength, &req.IPAddress, &sigValid, &req.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture d'une requête : %w", err)
		}

		if headersStr.Valid {
			_ = json.Unmarshal([]byte(headersStr.String), &req.Headers)
		}
		if queryParamsStr.Valid {
			_ = json.Unmarshal([]byte(queryParamsStr.String), &req.QueryParams)
		}
		if sigValid.Valid {
			valid := sigValid.Int64 == 1
			req.SignatureValid = &valid
		}

		requests = append(requests, req)
	}

	return requests, nil
}

// GetWebhookRequestByID récupère une requête spécifique par son identifiant.
func (db *DB) GetWebhookRequestByID(id string) (*models.WebhookRequest, error) {
	query := `
		SELECT id, endpoint_id, endpoint_slug, method, path, headers, query_params,
		       raw_body, content_type, content_length, ip_address, signature_valid, created_at
		FROM webhook_requests
		WHERE id = ?
	`

	var req models.WebhookRequest
	var headersStr, queryParamsStr sql.NullString
	var sigValid sql.NullInt64

	err := db.QueryRow(query, id).Scan(
		&req.ID, &req.EndpointID, &req.EndpointSlug, &req.Method, &req.Path,
		&headersStr, &queryParamsStr, &req.RawBody, &req.ContentType,
		&req.ContentLength, &req.IPAddress, &sigValid, &req.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recherche de la requête : %w", err)
	}

	if headersStr.Valid {
		_ = json.Unmarshal([]byte(headersStr.String), &req.Headers)
	}
	if queryParamsStr.Valid {
		_ = json.Unmarshal([]byte(queryParamsStr.String), &req.QueryParams)
	}
	if sigValid.Valid {
		valid := sigValid.Int64 == 1
		req.SignatureValid = &valid
	}

	return &req, nil
}

// ClearRequests supprime toutes les requêtes capturées (optionnellement pour un endpoint spécifique).
func (db *DB) ClearRequests(endpointID string) error {
	if endpointID != "" {
		_, err := db.Exec("DELETE FROM webhook_requests WHERE endpoint_id = ?", endpointID)
		return err
	}
	_, err := db.Exec("DELETE FROM webhook_requests")
	return err
}

// ==========================================
// Opérations sur les Replay Logs
// ==========================================

// SaveReplayLog enregistre le résultat d'un rejeu dans la base.
func (db *DB) SaveReplayLog(log models.ReplayLog) error {
	headersJSON, _ := json.Marshal(log.ResponseHeaders)

	query := `
		INSERT INTO replay_logs (
			id, request_id, target_url, status_code, response_headers,
			response_body, duration_ms, error_message, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(
		query,
		log.ID,
		log.RequestID,
		log.TargetURL,
		log.StatusCode,
		string(headersJSON),
		log.ResponseBody,
		log.DurationMs,
		log.ErrorMessage,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur lors de l'enregistrement du log de rejeu : %w", err)
	}
	return nil
}

// GetReplaysForRequest récupère l'historique des rejeux pour une requête donnée.
func (db *DB) GetReplaysForRequest(requestID string) ([]models.ReplayLog, error) {
	query := `
		SELECT id, request_id, target_url, status_code, response_headers,
		       response_body, duration_ms, error_message, created_at
		FROM replay_logs
		WHERE request_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, requestID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des logs de rejeu : %w", err)
	}
	defer rows.Close()

	var logs []models.ReplayLog
	for rows.Next() {
		var log models.ReplayLog
		var headersStr, respBody, errMsg sql.NullString

		err := rows.Scan(
			&log.ID, &log.RequestID, &log.TargetURL, &log.StatusCode,
			&headersStr, &respBody, &log.DurationMs, &errMsg, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture d'un log de rejeu : %w", err)
		}

		if headersStr.Valid {
			_ = json.Unmarshal([]byte(headersStr.String), &log.ResponseHeaders)
		}
		if respBody.Valid {
			log.ResponseBody = respBody.String
		}
		if errMsg.Valid {
			log.ErrorMessage = errMsg.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}
