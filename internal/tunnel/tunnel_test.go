package tunnel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateSamplePayloadStripe(t *testing.T) {
	secret := "whsec_test_secret"
	cfg := TriggerConfig{
		Provider:  "stripe",
		EventName: "payment_intent.succeeded",
		Secret:    secret,
	}

	payload, headers := GenerateSamplePayload(cfg)
	if len(payload) == 0 {
		t.Fatalf("Le payload genere ne doit pas etre vide")
	}

	sigHeader, exists := headers["Stripe-Signature"]
	if !exists || !strings.Contains(sigHeader, "t=") || !strings.Contains(sigHeader, "v1=") {
		t.Errorf("En-tete Stripe-Signature manquant ou invalide : %s", sigHeader)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(payload, &jsonMap); err != nil {
		t.Errorf("Le payload Stripe genere n'est pas un JSON valide : %v", err)
	}
}

func TestGenerateSamplePayloadGitHub(t *testing.T) {
	secret := "github_secret"
	cfg := TriggerConfig{
		Provider:  "github",
		EventName: "pull_request",
		Secret:    secret,
	}

	payload, headers := GenerateSamplePayload(cfg)
	sigHeader, exists := headers["X-Hub-Signature-256"]
	if !exists || !strings.HasPrefix(sigHeader, "sha256=") {
		t.Errorf("En-tete X-Hub-Signature-256 GitHub manquant : %s", sigHeader)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(payload, &jsonMap); err != nil {
		t.Errorf("Le payload GitHub genere n'est pas un JSON valide : %v", err)
	}
}

func TestTriggerSampleWebhookExecution(t *testing.T) {
	receivedMethod := ""
	receivedContentType := ""

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status": "accepted", "id": "req_123"}`))
	}))
	defer ts.Close()

	err := TriggerSampleWebhook(TriggerConfig{
		TargetURL: ts.URL,
		Provider:  "custom",
		EventName: "test.event",
	})
	if err != nil {
		t.Fatalf("Echec execution TriggerSampleWebhook : %v", err)
	}

	if receivedMethod != "POST" {
		t.Errorf("Methode attendue POST, obtenu %s", receivedMethod)
	}
	if receivedContentType != "application/json" {
		t.Errorf("Content-Type attendu application/json, obtenu %s", receivedContentType)
	}
}
