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

// ToleranceStripe correspond au délai maximal autorisé pour la signature Stripe (5 minutes par défaut).
const ToleranceStripe = 5 * time.Minute

// VerifySignature dispatche la vérification selon le type de provider spécifié.
func VerifySignature(provider string, payload []byte, headers map[string][]string, secret string) (bool, error) {
	if secret == "" {
		// Aucun secret configuré, pas de vérification nécessaire.
		return true, nil
	}

	provider = strings.ToLower(provider)
	switch provider {
	case "stripe":
		sigHeader := getHeader(headers, "Stripe-Signature")
		if sigHeader == "" {
			return false, fmt.Errorf("en-tête Stripe-Signature manquant")
		}
		return VerifyStripeSignature(payload, sigHeader, secret, ToleranceStripe)

	case "github":
		sigHeader := getHeader(headers, "X-Hub-Signature-256")
		if sigHeader == "" {
			sigHeader = getHeader(headers, "X-Hub-Signature")
		}
		if sigHeader == "" {
			return false, fmt.Errorf("en-tête X-Hub-Signature-256 manquant")
		}
		return VerifyGitHubSignature(payload, sigHeader, secret)

	case "shopify":
		sigHeader := getHeader(headers, "X-Shopify-Hmac-Sha256")
		if sigHeader == "" {
			return false, fmt.Errorf("en-tête X-Shopify-Hmac-Sha256 manquant")
		}
		return VerifyShopifySignature(payload, sigHeader, secret)

	case "slack":
		sigHeader := getHeader(headers, "X-Slack-Signature")
		timestamp := getHeader(headers, "X-Slack-Request-Timestamp")
		if sigHeader == "" || timestamp == "" {
			return false, fmt.Errorf("en-têtes Slack manquants")
		}
		return VerifySlackSignature(payload, timestamp, sigHeader, secret)

	default:
		// Mode générique : teste X-Signature ou X-Webhook-Signature en HMAC-SHA256 hexadécimal
		sigHeader := getHeader(headers, "X-Signature")
		if sigHeader == "" {
			sigHeader = getHeader(headers, "X-Webhook-Signature")
		}
		if sigHeader == "" {
			sigHeader = getHeader(headers, "Signature")
		}
		if sigHeader == "" {
			return false, fmt.Errorf("en-tête de signature manquant")
		}
		return VerifyGenericHMACSHA256(payload, sigHeader, secret)
	}
}

// VerifyStripeSignature vérifie la signature HMAC-SHA256 envoyée par Stripe (format t=timestamp,v1=hash).
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
		return false, fmt.Errorf("format de signature Stripe invalide")
	}

	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("horodatage Stripe invalide : %w", err)
	}

	// Vérification de la tolérance temporelle anti-replay
	if tolerance > 0 {
		timestamp := time.Unix(timestampInt, 0)
		if time.Since(timestamp) > tolerance || timestamp.After(time.Now().Add(tolerance)) {
			return false, fmt.Errorf("horodatage hors tolérance (délai dépassé)")
		}
	}

	// Concaténation : t + "." + payload
	signedPayload := fmt.Sprintf("%s.%s", timestampStr, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expectedMAC)) {
			return true, nil
		}
	}

	return false, fmt.Errorf("signature Stripe non correspondante")
}

// VerifyGitHubSignature vérifie la signature X-Hub-Signature-256 (format: sha256=hex).
func VerifyGitHubSignature(payload []byte, header string, secret string) (bool, error) {
	prefix := "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false, fmt.Errorf("la signature GitHub doit commencer par sha256=")
	}
	receivedHex := strings.TrimPrefix(header, prefix)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(receivedHex), []byte(expectedHex)) {
		return true, nil
	}
	return false, fmt.Errorf("signature GitHub invalide")
}

// VerifyShopifySignature vérifie la signature X-Shopify-Hmac-Sha256 (Base64).
func VerifyShopifySignature(payload []byte, header string, secret string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedBase64 := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(header), []byte(expectedBase64)) {
		return true, nil
	}
	return false, fmt.Errorf("signature Shopify invalide")
}

// VerifySlackSignature vérifie la signature X-Slack-Signature (format: v0=hex, basé sur v0:timestamp:body).
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

// VerifyGenericHMACSHA256 vérifie un HMAC-SHA256 hexadécimal direct.
func VerifyGenericHMACSHA256(payload []byte, header string, secret string) (bool, error) {
	cleanHeader := strings.TrimPrefix(header, "sha256=")
	cleanHeader = strings.TrimPrefix(cleanHeader, "SHA256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(cleanHeader), []byte(expectedHex)) {
		return true, nil
	}
	return false, fmt.Errorf("signature HMAC générique invalide")
}

func getHeader(headers map[string][]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
