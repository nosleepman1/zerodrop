package tunnel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TriggerConfig spécifie les paramètres pour injecter une charge utile simulée de test.
type TriggerConfig struct {
	TargetURL string // URL d'ingestion complète (ex: http://localhost:8080/in/default)
	Provider  string // 'stripe', 'github', 'shopify' ou 'custom'
	EventName string // Nom de l'événement (ex: 'payment_intent.succeeded', 'pull_request.opened')
	Secret    string // Clé secrète HMAC pour signer automatiquement la requête
}

// GenerateSamplePayload génère les octets et les en-têtes d'un webhook type en fonction du fournisseur.
func GenerateSamplePayload(cfg TriggerConfig) ([]byte, map[string]string) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = "ZeroDrop-CLI-Generator/0.1.0"

	nowUnix := time.Now().Unix()
	var payload []byte

	switch cfg.Provider {
	case "stripe":
		event := cfg.EventName
		if event == "" {
			event = "payment_intent.succeeded"
		}
		payload = []byte(fmt.Sprintf(`{
  "id": "evt_%d",
  "object": "event",
  "api_version": "2024-06-20",
  "created": %d,
  "type": "%s",
  "data": {
    "object": {
      "id": "pi_%d",
      "object": "payment_intent",
      "amount": 4900,
      "currency": "eur",
      "status": "succeeded",
      "customer": "cus_zerodrop_test"
    }
  }
}`, nowUnix, nowUnix, event, nowUnix))

		if cfg.Secret != "" {
			signedPayload := fmt.Sprintf("%d.%s", nowUnix, string(payload))
			mac := hmac.New(sha256.New, []byte(cfg.Secret))
			mac.Write([]byte(signedPayload))
			sigHex := hex.EncodeToString(mac.Sum(nil))
			headers["Stripe-Signature"] = fmt.Sprintf("t=%d,v1=%s", nowUnix, sigHex)
		}

	case "github":
		event := cfg.EventName
		if event == "" {
			event = "pull_request"
		}
		headers["X-GitHub-Event"] = event
		headers["X-GitHub-Delivery"] = fmt.Sprintf("del_%d", nowUnix)

		payload = []byte(fmt.Sprintf(`{
  "action": "opened",
  "number": 42,
  "pull_request": {
    "id": %d,
    "title": "feat: ultra-fast tunneling engine in Go",
    "user": {
      "login": "nosleepman1"
    },
    "state": "open"
  },
  "repository": {
    "name": "zerodrop",
    "full_name": "nosleepman1/zerodrop"
  }
}`, nowUnix))

		if cfg.Secret != "" {
			mac := hmac.New(sha256.New, []byte(cfg.Secret))
			mac.Write(payload)
			sigHex := hex.EncodeToString(mac.Sum(nil))
			headers["X-Hub-Signature-256"] = "sha256=" + sigHex
		}

	default:
		payload = []byte(fmt.Sprintf(`{
  "event": "%s",
  "timestamp": "%s",
  "message": "Webhook de test simule via ZeroDrop CLI",
  "data": {
    "id": %d,
    "status": "active"
  }
}`, cfg.EventName, time.Now().Format(time.RFC3339), nowUnix))

		if cfg.Secret != "" {
			mac := hmac.New(sha256.New, []byte(cfg.Secret))
			mac.Write(payload)
			sigHex := hex.EncodeToString(mac.Sum(nil))
			headers["X-Signature"] = "sha256=" + sigHex
		}
	}

	return payload, headers
}

// TriggerSampleWebhook expédie un webhook type vers le serveur ZeroDrop et affiche la réponse retournée.
func TriggerSampleWebhook(cfg TriggerConfig) error {
	payload, headers := GenerateSamplePayload(cfg)

	req, err := http.NewRequest("POST", cfg.TargetURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("impossible de construire la requete : %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("echec d'envoi vers %s : %w", cfg.TargetURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Println("================================================================")
	fmt.Printf("[OK] Webhook de test envoye avec succes\n")
	fmt.Printf("[INFO] Destination : %s\n", cfg.TargetURL)
	fmt.Printf("[INFO] Provider    : %s\n", cfg.Provider)
	fmt.Printf("[INFO] Statut HTTP : %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Printf("[INFO] Reponse     : %s\n", string(bodyBytes))
	fmt.Println("================================================================")

	return nil
}
