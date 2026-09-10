// Package security implémente les mécanismes de vérification cryptographique des signatures HMAC
// pour valider l'authenticité et l'intégrité des charges utiles reçues de fournisseurs tiers.
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ToleranceStripe définit la fenêtre temporelle maximale d'acceptation des webhooks Stripe (5 minutes).
// Cette tolérance permet de se prémunir contre les attaques par rejeu (replay attacks).
const ToleranceStripe = 5 * time.Minute

// VerifySignature sélectionne et exécute l'algorithme de validation cryptographique adapté
// en fonction du fournisseur déclaré sur l'endpoint.
//
// Paramètres :
//   - provider : Type de fournisseur ('stripe', 'github', 'shopify', 'slack' ou 'custom').
//   - payload  : Corps brut de la requête HTTP en octets (doit être non tronqué).
//   - headers  : En-têtes HTTP de la requête entrante.
//   - secret   : Clé secrète partagée configurée sur l'endpoint.
//
// Retourne :
//   - bool : true si la signature est rigoureusement valide, false sinon.
//   - error: description de l'anomalie cryptographique ou de formatage en cas d'échec.
func VerifySignature(provider string, payload []byte, headers map[string][]string, secret string) (bool, error) {
	if secret == "" {
		// Aucun secret configuré, la validation est ignorée.
		return true, nil
	}

	provider = strings.ToLower(provider)
	switch provider {
	case "stripe":
		sigHeader := getHeader(headers, "Stripe-Signature")
		if sigHeader == "" {
			return false, fmt.Errorf("en-tete 'Stripe-Signature' manquant")
		}
		return VerifyStripeSignature(payload, sigHeader, secret, ToleranceStripe)

	case "github":
		sigHeader := getHeader(headers, "X-Hub-Signature-256")
		if sigHeader == "" {
			sigHeader = getHeader(headers, "X-Hub-Signature")
		}
		if sigHeader == "" {
			return false, fmt.Errorf("en-tete 'X-Hub-Signature-256' manquant")
		}
		return VerifyGitHubSignature(payload, sigHeader, secret)

	case "shopify":
		sigHeader := getHeader(headers, "X-Shopify-Hmac-Sha256")
		if sigHeader == "" {
			return false, fmt.Errorf("en-tete 'X-Shopify-Hmac-Sha256' manquant")
		}
		return VerifyShopifySignature(payload, sigHeader, secret)

	case "slack":
		sigHeader := getHeader(headers, "X-Slack-Signature")
		timestamp := getHeader(headers, "X-Slack-Request-Timestamp")
		if sigHeader == "" || timestamp == "" {
			return false, fmt.Errorf("en-tetes Slack 'X-Slack-Signature' ou 'X-Slack-Request-Timestamp' manquants")
		}
		return VerifySlackSignature(payload, timestamp, sigHeader, secret)

	default:
		// Mode générique : recherche des en-têtes de signature usuels
		sigHeader := getHeader(headers, "X-Signature")
		if sigHeader == "" {
			sigHeader = getHeader(headers, "X-Webhook-Signature")
		}
		if sigHeader == "" {
			sigHeader = getHeader(headers, "Signature")
		}
		if sigHeader == "" {
			return false, fmt.Errorf("aucun en-tete de signature valide detecte (X-Signature, X-Webhook-Signature, Signature)")
		}
		return VerifyGenericHMACSHA256(payload, sigHeader, secret)
	}
}

// VerifyStripeSignature valide une signature au format Stripe v1 (t=timestamp,v1=hash).
//
// L'algorithme calcule : HMAC-SHA256(secret, timestamp + "." + payload).
// La comparaison utilise hmac.Equal afin de garantir un temps d'exécution constant
// et d'éliminer les vulnérabilités aux attaques temporelles.
func VerifyStripeSignature(payload []byte, header string, secret string, tolerance time.Duration) (bool, error) {
	var timestampStr string
	var signatures []string

	pairs := strings.Split(header, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := parts[1]
			if key == "t" {
				timestampStr = val
			} else if key == "v1" {
				signatures = append(signatures, val)
			}
		}
	}

	if timestampStr == "" || len(signatures) == 0 {
		return false, fmt.Errorf("format de signature Stripe invalide (attendu t=...,v1=...)")
	}

	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("horodatage Stripe invalide : %w", err)
	}

	// Contrôle de la fenêtre d'anti-rejeu
	if tolerance > 0 {
		timestamp := time.Unix(timestampInt, 0)
		if time.Since(timestamp) > tolerance || timestamp.After(time.Now().Add(tolerance)) {
			return false, fmt.Errorf("horodatage hors tolerance : risque d'attaque par rejeu")
		}
	}

	signedPayload := fmt.Sprintf("%s.%s", timestampStr, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expectedMAC)) {
			return true, nil
		}
	}

	return false, fmt.Errorf("la signature Stripe ne correspond pas au secret configure")
}

// VerifyGitHubSignature valide une signature GitHub au format HMAC-SHA256 préfixé (sha256=hash).
func VerifyGitHubSignature(payload []byte, header string, secret string) (bool, error) {
	prefix := "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false, fmt.Errorf("la signature GitHub doit imperativement debuter par 'sha256='")
	}
	receivedHex := strings.TrimPrefix(header, prefix)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(receivedHex), []byte(expectedHex)) {
		return true, nil
	}
	return false, fmt.Errorf("signature GitHub invalide ou cle secrete incorrecte")
}

// VerifyShopifySignature valide une signature Shopify encodée en Base64 standard.
func VerifyShopifySignature(payload []byte, header string, secret string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedBase64 := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(header), []byte(expectedBase64)) {
		return true, nil
	}
	return false, fmt.Errorf("signature Shopify invalide")
}

// VerifySlackSignature valide une signature Slack (v0=hash, basé sur 'v0:timestamp:body').
func VerifySlackSignature(payload []byte, timestamp string, header string, secret string) (bool, error) {
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(header), []byte(expectedSignature)) {
		return true, nil
	}
	return false, fmt.Errorf("signature Slack invalide")
}

// VerifyGenericHMACSHA256 valide un HMAC-SHA256 direct au format hexadécimal.
func VerifyGenericHMACSHA256(payload []byte, header string, secret string) (bool, error) {
	cleanHeader := strings.TrimPrefix(header, "sha256=")
	cleanHeader = strings.TrimPrefix(cleanHeader, "SHA256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(cleanHeader), []byte(expectedHex)) {
		return true, nil
	}
	return false, fmt.Errorf("signature HMAC generique non valide")
}

// getHeader recherche un en-tête de manière insensible à la casse.
func getHeader(headers map[string][]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
