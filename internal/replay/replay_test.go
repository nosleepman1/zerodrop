package replay

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nosleepman1/zerodrop/internal/models"
)

func TestReplayEngineExecution(t *testing.T) {
	var receivedBody string
	var receivedHeader string
	var receivedReplayMarker string

	// Serveur HTTP factice pour tester la réception du rejeu
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		receivedHeader = r.Header.Get("X-Custom-Test")
		receivedReplayMarker = r.Header.Get("X-ZeroDrop-Replay")

		w.Header().Set("X-Response-From", "MockServer")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "replayed"}`))
	}))
	defer ts.Close()

	engine := NewEngine(nil, nil)

	req := &models.WebhookRequest{
		ID:          "req_original_001",
		Method:      "POST",
		RawBody:     `{"original": true}`,
		Headers:     map[string][]string{"Content-Type": {"application/json"}},
		ContentType: "application/json",
	}

	// Cas 1 : Rejeu sans mutation
	log, err := engine.ReplayRequest(req, models.TriggerReplayPayload{
		TargetURL: ts.URL,
		CustomHeaders: map[string]string{
			"X-Custom-Test": "Value123",
		},
	})
	if err != nil {
		t.Fatalf("Echec du rejeu vers le serveur mock : %v", err)
	}

	if log.StatusCode != http.StatusOK {
		t.Errorf("Statut attendu 200, obtenu %d", log.StatusCode)
	}
	if receivedBody != `{"original": true}` {
		t.Errorf("Corps recu attendu '{\"original\": true}', obtenu '%s'", receivedBody)
	}
	if receivedHeader != "Value123" {
		t.Errorf("En-tete custom attendu 'Value123', obtenu '%s'", receivedHeader)
	}
	if receivedReplayMarker != "true" {
		t.Errorf("Marqueur X-ZeroDrop-Replay manquant")
	}

	// Cas 2 : Rejeu avec mutation de charge utile
	mutated := `{"mutated": true, "amount": 9900}`
	logMutated, err := engine.ReplayRequest(req, models.TriggerReplayPayload{
		TargetURL:    ts.URL,
		ModifiedBody: &mutated,
	})
	if err != nil {
		t.Fatalf("Echec du rejeu mute : %v", err)
	}

	if logMutated.StatusCode != http.StatusOK {
		t.Errorf("Statut attendu 200, obtenu %d", logMutated.StatusCode)
	}
	if receivedBody != mutated {
		t.Errorf("Corps mute non transmis, obtenu '%s'", receivedBody)
	}

	// Cas 3 : Rejeu vers une URL invalide / injoignable
	logErr, err := engine.ReplayRequest(req, models.TriggerReplayPayload{
		TargetURL: "http://127.0.0.1:59999/unreachable",
	})
	if err != nil {
		t.Fatalf("L'appel ReplayRequest ne doit pas paniquer sur erreur reseau : %v", err)
	}
	if logErr.StatusCode != 0 || logErr.ErrorMessage == "" {
		t.Errorf("Attendu StatusCode 0 et message d'erreur renseigne pour hote injoignable")
	}
}
