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
		t.Fatalf("Échec de l'initialisation de SQLite : %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// 2. Vérification de l'endpoint par défaut
	defaultEP, err := db.GetEndpointBySlug("default")
	if err != nil || defaultEP == nil {
		t.Fatalf("L'endpoint par défaut 'default' aurait dû être créé automatiquement : %v", err)
	}

	// 3. Création d'un nouvel endpoint
	newEP, err := db.CreateEndpoint(models.CreateEndpointPayload{
		Name:       "Stripe Test",
		Slug:       "stripe-test",
		Secret:     "whsec_123",
		Provider:   "stripe",
		ForwardURL: "http://localhost:3000/stripe",
	})
	if err != nil {
		t.Fatalf("Erreur lors de la création de l'endpoint : %v", err)
	}

	if newEP.Slug != "stripe-test" {
		t.Errorf("Slug attendu 'stripe-test', obtenu '%s'", newEP.Slug)
	}

	// 4. Sauvegarde d'une requête webhook
	now := time.Now().UTC()
	sigValid := true
	req := models.WebhookRequest{
		ID:             "req_test_001",
		EndpointID:     newEP.ID,
		EndpointSlug:   newEP.Slug,
		Method:         "POST",
		Path:           "/in/stripe-test",
		Headers:        map[string][]string{"Content-Type": {"application/json"}},
		QueryParams:    map[string][]string{"ref": {"abc"}},
		RawBody:        `{"event": "charge.succeeded"}`,
		ContentType:    "application/json",
		ContentLength:  30,
		IPAddress:      "127.0.0.1",
		SignatureValid: &sigValid,
		CreatedAt:      now,
	}

	if err := db.SaveWebhookRequest(req); err != nil {
		t.Fatalf("Erreur lors de la sauvegarde de la requête : %v", err)
	}

	// 5. Récupération de la requête par ID
	fetchedReq, err := db.GetWebhookRequestByID("req_test_001")
	if err != nil || fetchedReq == nil {
		t.Fatalf("Erreur lors de la récupération de la requête : %v", err)
	}
	if fetchedReq.RawBody != req.RawBody {
		t.Errorf("RawBody attendu '%s', obtenu '%s'", req.RawBody, fetchedReq.RawBody)
	}
	if fetchedReq.SignatureValid == nil || !*fetchedReq.SignatureValid {
		t.Errorf("SignatureValid attendu true, obtenu %v", fetchedReq.SignatureValid)
	}

	// 6. Test de Replay Log
	replayLog := models.ReplayLog{
		ID:              "rep_test_001",
		RequestID:       "req_test_001",
		TargetURL:       "http://localhost:3000/stripe",
		StatusCode:      200,
		ResponseHeaders: map[string][]string{"Server": {"TestServer"}},
		ResponseBody:    `{"status": "ok"}`,
		DurationMs:      42,
		CreatedAt:       time.Now().UTC(),
	}

	if err := db.SaveReplayLog(replayLog); err != nil {
		t.Fatalf("Erreur lors de la sauvegarde du log de rejeu : %v", err)
	}

	replays, err := db.GetReplaysForRequest("req_test_001")
	if err != nil || len(replays) != 1 {
		t.Fatalf("Attendu 1 log de rejeu, obtenu %d (err: %v)", len(replays), err)
	}
	if replays[0].StatusCode != 200 {
		t.Errorf("StatusCode attendu 200, obtenu %d", replays[0].StatusCode)
	}
}
