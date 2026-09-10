package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestVerifyStripeSignature(t *testing.T) {
	secret := "whsec_test_secret_key_12345"
	payload := []byte(`{"id": "evt_123", "type": "payment_intent.succeeded"}`)
	now := time.Now().Unix()

	signedPayload := fmt.Sprintf("%d.%s", now, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	validMAC := hex.EncodeToString(mac.Sum(nil))

	header := fmt.Sprintf("t=%d,v1=%s", now, validMAC)

	// Cas 1 : Signature valide
	ok, err := VerifyStripeSignature(payload, header, secret, 5*time.Minute)
	if !ok || err != nil {
		t.Fatalf("Échec attendu : la signature Stripe aurait dû être valide : %v", err)
	}

	// Cas 2 : Mauvais secret
	ok, err = VerifyStripeSignature(payload, header, "mauvais_secret", 5*time.Minute)
	if ok || err == nil {
		t.Fatalf("Échec attendu : la signature avec mauvais secret aurait dû échouer")
	}

	// Cas 3 : Payload altéré
	corruptedPayload := []byte(`{"id": "evt_999", "type": "payment_intent.succeeded"}`)
	ok, err = VerifyStripeSignature(corruptedPayload, header, secret, 5*time.Minute)
	if ok || err == nil {
		t.Fatalf("Échec attendu : le payload altéré aurait dû être rejeté")
	}
}

func TestVerifyGitHubSignature(t *testing.T) {
	secret := "github_secret_test_98765"
	payload := []byte(`{"action": "opened", "pull_request": {"number": 1}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validHex := hex.EncodeToString(mac.Sum(nil))
	header := "sha256=" + validHex

	ok, err := VerifyGitHubSignature(payload, header, secret)
	if !ok || err != nil {
		t.Fatalf("Échec attendu : la signature GitHub aurait dû être valide : %v", err)
	}

	ok, err = VerifyGitHubSignature(payload, "sha256=invalid_hash", secret)
	if ok || err == nil {
		t.Fatalf("Échec attendu : la signature GitHub invalide aurait dû échouer")
	}
}
