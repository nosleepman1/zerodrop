package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/hub"
	"github.com/nosleepman1/zerodrop/internal/models"
	"github.com/nosleepman1/zerodrop/internal/replay"
)

func setupTestRouter(t *testing.T) (http.Handler, *database.DB, *replay.Engine) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_api.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Impossible d'initialiser SQLite de test : %v", err)
	}

	eventHub := hub.NewHub()
	go eventHub.Run()

	replayEngine := replay.NewEngine(db, eventHub)
	router := NewRouter(db, eventHub, replayEngine)

	return router, db, replayEngine
}

func TestAPIHealthCheck(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Statut attendu 200, obtenu %d", w.Code)
	}

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Errorf("Statut attendu 'healthy', obtenu '%s'", resp["status"])
	}
}

func TestIngestGatewayAndListing(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer db.Close()

	// 1. Ingestion vers l'endpoint 'default'
	payload := `{"event": "payment.succeeded", "amount": 4200}`
	req := httptest.NewRequest("POST", "/in/default", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("Statut d'ingestion attendu 202 Accepted, obtenu %d", w.Code)
	}

	var ingestResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &ingestResp); err != nil {
		t.Fatalf("Erreur decodage reponse ingestion : %v", err)
	}
	reqID, ok := ingestResp["id"].(string)
	if !ok || reqID == "" {
		t.Fatalf("ID de requete manquant dans la reponse d'ingestion")
	}

	// 2. Ingestion vers un slug inconnu (doit retourner 404)
	req404 := httptest.NewRequest("POST", "/in/inconnu-12345", bytes.NewBufferString(payload))
	w404 := httptest.NewRecorder()
	router.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound {
		t.Errorf("Statut attendu 404 pour slug inconnu, obtenu %d", w404.Code)
	}

	// 3. Récupération de la liste des requêtes via GET /api/requests
	listReq := httptest.NewRequest("GET", "/api/requests", nil)
	wList := httptest.NewRecorder()
	router.ServeHTTP(wList, listReq)

	if wList.Code != http.StatusOK {
		t.Fatalf("Statut attendu 200 pour GET /api/requests, obtenu %d", wList.Code)
	}

	var requests []models.WebhookRequest
	_ = json.Unmarshal(wList.Body.Bytes(), &requests)
	if len(requests) != 1 {
		t.Fatalf("Attendu 1 requete dans la liste, obtenu %d", len(requests))
	}
	if requests[0].ID != reqID {
		t.Errorf("ID de requete attendu '%s', obtenu '%s'", reqID, requests[0].ID)
	}

	// 4. Récupération par ID via GET /api/requests/{id}
	getReq := httptest.NewRequest("GET", "/api/requests/"+reqID, nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, getReq)

	if wGet.Code != http.StatusOK {
		t.Errorf("Statut attendu 200 pour GET /api/requests/{id}, obtenu %d", wGet.Code)
	}
}

func TestEndpointsCRUD(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer db.Close()

	// 1. Création d'un endpoint
	newEPPayload := models.CreateEndpointPayload{
		Name:     "GitHub Webhook",
		Slug:     "github-hook",
		Provider: "github",
		Secret:   "secret123",
	}
	body, _ := json.Marshal(newEPPayload)
	req := httptest.NewRequest("POST", "/api/endpoints", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Statut attendu 201 Created, obtenu %d", w.Code)
	}

	var createdEP models.Endpoint
	_ = json.Unmarshal(w.Body.Bytes(), &createdEP)
	if createdEP.Slug != "github-hook" {
		t.Errorf("Slug attendu 'github-hook', obtenu '%s'", createdEP.Slug)
	}

	// 2. Tentative de création avec un slug en doublon (doit retourner 409 Conflict)
	reqDup := httptest.NewRequest("POST", "/api/endpoints", bytes.NewBuffer(body))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	router.ServeHTTP(wDup, reqDup)
	if wDup.Code != http.StatusConflict {
		t.Errorf("Statut attendu 409 Conflict pour doublon, obtenu %d", wDup.Code)
	}

	// 3. Suppression de l'endpoint créé
	delReq := httptest.NewRequest("DELETE", "/api/endpoints/"+createdEP.ID, nil)
	wDel := httptest.NewRecorder()
	router.ServeHTTP(wDel, delReq)
	if wDel.Code != http.StatusOK {
		t.Errorf("Statut attendu 200 pour DELETE /api/endpoints/{id}, obtenu %d", wDel.Code)
	}

	// 4. Tentative de suppression de l'endpoint par défaut (doit retourner 403 Forbidden)
	delDefReq := httptest.NewRequest("DELETE", "/api/endpoints/ep_default", nil)
	wDefDel := httptest.NewRecorder()
	router.ServeHTTP(wDefDel, delDefReq)
	if wDefDel.Code != http.StatusForbidden {
		t.Errorf("Statut attendu 403 Forbidden pour suppression de l'endpoint par defaut, obtenu %d", wDefDel.Code)
	}
}
