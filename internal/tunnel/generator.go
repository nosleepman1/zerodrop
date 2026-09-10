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

// TriggerConfig contient les paramètres pour émettre un webhook de test.
type TriggerConfig struct {
	TargetURL string // URL complète vers l'ingestion (ex: http://localhost:8080/in/default)
	Provider  string // 'stripe', 'github', 'shopify', 'custom'
	EventName string // ex: 'payment_intent.succeeded', 'pull_request.opened'
	Secret    string // Secret HMAC pour signature automatique
}

// TriggerSampleWebhook génère et envoie un webhook type au serveur ZeroDrop.
func TriggerSampleWebhook(cfg TriggerConfig) error {
	var payload []byte
	var contentType = "application/json"
	var headers = make(map[string]string)

	nowUnix := time.Now().Unix()

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
  "message": "Webhook de test simulé via ZeroDrop CLI",
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

	// Envoi de la requête HTTP
	req, err := http.NewRequest("POST", cfg.TargetURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("impossible de construire la requête : %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "ZeroDrop-CLI-Generator/0.1.0")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("échec de l'envoi vers %s : %w", cfg.TargetURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Println("================================================================")
	fmt.Printf("🚀 Webhook de test envoyé avec succès !\n")
	fmt.Printf("🎯 Destination : %s\n", cfg.TargetURL)
	fmt.Printf("📦 Provider    : %s\n", cfg.Provider)
	fmt.Printf("🏷️ Statut HTTP : %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Printf("📄 Réponse     : %s\n", string(bodyBytes))
	fmt.Println("================================================================")

	return nil
}
