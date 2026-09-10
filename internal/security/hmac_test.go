package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
		t.Fatalf("Echec attendu : signature Stripe valide rejetee : %v", err)
	}

	// Cas 2 : Mauvais secret
	ok, err = VerifyStripeSignature(payload, header, "mauvais_secret", 5*time.Minute)
	if ok || err == nil {
		t.Fatalf("Echec attendu : signature Stripe avec mauvais secret acceptee")
	}

	// Cas 3 : Charge utile modifiee
	corruptedPayload := []byte(`{"id": "evt_999", "type": "payment_intent.succeeded"}`)
	ok, err = VerifyStripeSignature(corruptedPayload, header, secret, 5*time.Minute)
	if ok || err == nil {
		t.Fatalf("Echec attendu : payload altere accepte")
	}

	// Cas 4 : Horodatage expire (depassement de tolerance anti-rejeu)
	expiredTimestamp := time.Now().Add(-10 * time.Minute).Unix()
	expiredSigned := fmt.Sprintf("%d.%s", expiredTimestamp, string(payload))
	macExpired := hmac.New(sha256.New, []byte(secret))
	macExpired.Write([]byte(expiredSigned))
	expiredMAC := hex.EncodeToString(macExpired.Sum(nil))
	expiredHeader := fmt.Sprintf("t=%d,v1=%s", expiredTimestamp, expiredMAC)

	ok, err = VerifyStripeSignature(payload, expiredHeader, secret, 5*time.Minute)
	if ok || err == nil {
		t.Fatalf("Echec attendu : horodatage expire accepte sans verification anti-rejeu")
	}
}

func TestVerifyGitHubSignature(t *testing.T) {
	secret := "github_secret_test_98765"
	payload := []byte(`{"action": "opened", "pull_request": {"number": 1}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validHex := hex.EncodeToString(mac.Sum(nil))
	header := "sha256=" + validHex

	// Cas valide
	ok, err := VerifyGitHubSignature(payload, header, secret)
	if !ok || err != nil {
		t.Fatalf("Echec attendu : signature GitHub valide rejetee : %v", err)
	}

	// Cas invalide
	ok, err = VerifyGitHubSignature(payload, "sha256=invalid_hash", secret)
	if ok || err == nil {
		t.Fatalf("Echec attendu : signature GitHub invalide acceptee")
	}

	// Format manquant sha256=
	ok, err = VerifyGitHubSignature(payload, validHex, secret)
	if ok || err == nil {
		t.Fatalf("Echec attendu : signature GitHub sans prefixe sha256= acceptee")
	}
}

func TestVerifyShopifySignature(t *testing.T) {
	secret := "shopify_shared_secret"
	payload := []byte(`{"order_id": 123456, "total_price": "99.00"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validBase64 := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	ok, err := VerifyShopifySignature(payload, validBase64, secret)
	if !ok || err != nil {
		t.Fatalf("Echec attendu : signature Shopify valide rejetee : %v", err)
	}

	ok, err = VerifyShopifySignature(payload, "mauvaise_signature_base64", secret)
	if ok || err == nil {
		t.Fatalf("Echec attendu : signature Shopify invalide acceptee")
	}
}

func TestVerifySlackSignature(t *testing.T) {
	secret := "slack_signing_secret"
	payload := []byte(`command=/deploy&text=prod`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	validSig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	ok, err := VerifySlackSignature(payload, timestamp, validSig, secret)
	if !ok || err != nil {
		t.Fatalf("Echec attendu : signature Slack valide rejetee : %v", err)
	}

	ok, err = VerifySlackSignature(payload, timestamp, "v0=bad_hash", secret)
	if ok || err == nil {
		t.Fatalf("Echec attendu : signature Slack invalide acceptee")
	}
}

func TestVerifySignatureDispatcher(t *testing.T) {
	secret := "shared_key"
	payload := []byte(`{"test": true}`)

	// Test Generic
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validHex := hex.EncodeToString(mac.Sum(nil))

	headers := map[string][]string{
		"X-Signature": {validHex},
	}

	ok, err := VerifySignature("custom", payload, headers, secret)
	if !ok || err != nil {
		t.Fatalf("Echec dispatcher signature generique : %v", err)
	}

	// Test avec secret vide (doit accepter)
	ok, err = VerifySignature("custom", payload, headers, "")
	if !ok || err != nil {
		t.Fatalf("Echec dispatcher secret vide : %v", err)
	}
}
