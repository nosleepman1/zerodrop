package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nosleepman1/zerodrop/internal/models"
)

func TestDatabaseOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_zerodrop.db")

	// 1. Initialisation de la base
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Echec de l'initialisation de SQLite : %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// 2. Vérification de l'endpoint par défaut
	defaultEP, err := db.GetEndpointBySlug("default")
	if err != nil || defaultEP == nil {
		t.Fatalf("L'endpoint par defaut 'default' aurait du etre cree automatiquement : %v", err)
	}

	// 3. Création d'un nouvel endpoint
	newEP, err := db.CreateEndpoint(models.CreateEndpointPayload{
		Name:        "Stripe Production",
		Slug:        "stripe-prod",
		Secret:      "whsec_prod_123",
		Provider:    "stripe",
		ForwardURL:  "http://localhost:3000/stripe",
		Description: "Endpoint de production pour tester",
	})
	if err != nil {
		t.Fatalf("Erreur lors de la creation de l'endpoint : %v", err)
	}

	if newEP.Slug != "stripe-prod" {
		t.Errorf("Slug attendu 'stripe-prod', obtenu '%s'", newEP.Slug)
	}

	// 4. Liste des endpoints
	endpoints, err := db.ListEndpoints()
	if err != nil || len(endpoints) < 2 {
		t.Fatalf("Attendu au moins 2 endpoints, obtenu %d", len(endpoints))
	}

	// 5. Sauvegarde de plusieurs requêtes webhook
	now := time.Now().UTC()
	sigValid := true
	for i := 1; i <= 5; i++ {
		req := models.WebhookRequest{
			ID:             "req_test_00" + string(rune('0'+i)),
			EndpointID:     newEP.ID,
			EndpointSlug:   newEP.Slug,
			Method:         "POST",
			Path:           "/in/stripe-prod",
			Headers:        map[string][]string{"Content-Type": {"application/json"}},
			QueryParams:    map[string][]string{"batch": {"true"}},
			RawBody:        `{"event": "payment.succeeded", "index": ` + string(rune('0'+i)) + `}`,
			ContentType:    "application/json",
			ContentLength:  40,
			IPAddress:      "127.0.0.1",
			SignatureValid: &sigValid,
			CreatedAt:      now.Add(time.Duration(i) * time.Second),
		}

		if err := db.SaveWebhookRequest(req); err != nil {
			t.Fatalf("Erreur lors de la sauvegarde de la requete %d : %v", i, err)
		}
	}

	// 6. Test de pagination et filtrage par recherche
	reqs, err := db.ListWebhookRequests(models.RequestListFilter{
		EndpointID: newEP.ID,
		Limit:      3,
		Offset:     0,
	})
	if err != nil || len(reqs) != 3 {
		t.Fatalf("Attendu 3 requetes avec limit=3, obtenu %d (err: %v)", len(reqs), err)
	}

	// Test de recherche
	searchReqs, err := db.ListWebhookRequests(models.RequestListFilter{
		Search: "payment.succeeded",
	})
	if err != nil || len(searchReqs) != 5 {
		t.Fatalf("Attendu 5 requetes matchees par la recherche, obtenu %d", len(searchReqs))
	}

	// 7. Test de Replay Log
	replayLog := models.ReplayLog{
		ID:              "rep_test_001",
		RequestID:       "req_test_001",
		TargetURL:       "http://localhost:3000/stripe",
		StatusCode:      200,
		ResponseHeaders: map[string][]string{"Server": {"ZeroDrop-Mock"}},
		ResponseBody:    `{"status": "ok"}`,
		DurationMs:      24,
		CreatedAt:       time.Now().UTC(),
	}

	if err := db.SaveReplayLog(replayLog); err != nil {
		t.Fatalf("Erreur lors de la sauvegarde du journal de rejeu : %v", err)
	}

	replays, err := db.GetReplaysForRequest("req_test_001")
	if err != nil || len(replays) != 1 {
		t.Fatalf("Attendu 1 log de rejeu, obtenu %d", len(replays))
	}

	// 8. Test de suppression en cascade de l'endpoint
	if err := db.DeleteEndpoint(newEP.ID); err != nil {
		t.Fatalf("Erreur lors de la suppression de l'endpoint : %v", err)
	}

	// Vérification de la suppression de l'endpoint
	deletedEP, _ := db.GetEndpointByID(newEP.ID)
	if deletedEP != nil {
		t.Errorf("L'endpoint aurait du etre supprime")
	}

	// Vérification de la suppression en cascade des requêtes
	remainingReqs, _ := db.ListWebhookRequests(models.RequestListFilter{EndpointID: newEP.ID})
	if len(remainingReqs) != 0 {
		t.Errorf("Les requetes liees auraient du etre supprimees en cascade (obtenu %d)", len(remainingReqs))
	}
}
