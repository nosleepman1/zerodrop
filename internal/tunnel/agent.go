package tunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nosleepman1/zerodrop/internal/models"
)

// Config regroupe les paramètres pour l'agent de tunneling local.
type Config struct {
	ServerURL    string            // ex: ws://localhost:8080 ou wss://zerodrop.mon-domaine.com
	EndpointSlug string            // Slug de l'endpoint à écouter (ex: "default", "stripe-dev", ou "*")
	ForwardTo    string            // URL locale de destination (ex: "http://localhost:3000/api/webhook")
	CustomHeaders map[string]string // En-têtes personnalisés additionnels à injecter
}

// Agent représente le client WebSocket de tunneling local.
type Agent struct {
	config     Config
	httpClient *http.Client
}

// NewAgent instancie un nouvel agent de tunneling.
func NewAgent(cfg Config) *Agent {
	return &Agent{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Start démarre la boucle de connexion persistante du tunnel avec reconnexion automatique.
func (a *Agent) Start() error {
	// Préparation de l'URL WebSocket
	serverURL := a.config.ServerURL
	if !strings.HasPrefix(serverURL, "ws://") && !strings.HasPrefix(serverURL, "wss://") {
		if strings.HasPrefix(serverURL, "http://") {
			serverURL = "ws://" + strings.TrimPrefix(serverURL, "http://")
		} else if strings.HasPrefix(serverURL, "https://") {
			serverURL = "wss://" + strings.TrimPrefix(serverURL, "https://")
		} else {
			serverURL = "ws://" + serverURL
		}
	}

	slug := a.config.EndpointSlug
	if slug == "" {
		slug = "*"
	}

	wsURL := fmt.Sprintf("%s/ws/tunnel?slug=%s", strings.TrimSuffix(serverURL, "/"), url.QueryEscape(slug))

	fmt.Println("================================================================")
	fmt.Printf("⚡ ZeroDrop Tunnel Agent — Relai Local Haute Performance\n")
	fmt.Printf("📡 Hub distant      : %s\n", serverURL)
	fmt.Printf("🎯 Endpoint écouté  : %s\n", slug)
	fmt.Printf("🔀 Redirection vers : %s\n", a.config.ForwardTo)
	fmt.Println("================================================================")
	fmt.Println("🟢 En attente de webhooks entrants... (Ctrl+C pour quitter)")

	// Gestion de l'interruption (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	for {
		err := a.connectAndListen(wsURL, interrupt)
		if err != nil {
			log.Printf("⚠️ Connexion au Hub interrompue : %v. Reconnexion dans 3 secondes...", err)
		}

		select {
		case <-interrupt:
			fmt.Println("\n🛑 Arrêt du tunnel local...")
			return nil
		case <-time.After(3 * time.Second):
			// Nouvelle tentative de connexion
		}
	}
}

func (a *Agent) connectAndListen(wsURL string, interrupt chan os.Signal) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("impossible d'établir la connexion WebSocket : %w", err)
	}
	defer conn.Close()

	log.Printf("✅ Connecté avec succès au Hub ZeroDrop ! Prêt à relayer vers %s", a.config.ForwardTo)

	done := make(chan struct{})

	// Goroutine de lecture des messages envoyés par le Hub
	go func() {
		defer close(done)
		for {
			_, message, readErr := conn.ReadMessage()
			if readErr != nil {
				return
			}

			var wsMsg models.WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				continue
			}

			if wsMsg.Type == models.EventTunnelForward {
				// Désérialisation de la requête webhook à transférer
				reqBytes, _ := json.Marshal(wsMsg.Payload)
				var req models.WebhookRequest
				if err := json.Unmarshal(reqBytes, &req); err == nil {
					go a.forwardLocally(req)
				}
			}
		}
	}()

	select {
	case <-done:
		return fmt.Errorf("connexion fermée par le serveur")
	case <-interrupt:
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return nil
	}
}

// forwardLocally retransmet la requête HTTP reçue vers l'URL locale du développeur.
func (a *Agent) forwardLocally(req models.WebhookRequest) {
	startTime := time.Now()

	targetURL := a.config.ForwardTo
	httpReq, err := http.NewRequest(req.Method, targetURL, bytes.NewBufferString(req.RawBody))
	if err != nil {
		log.Printf("❌ [%s] Erreur création requête vers %s : %v", req.Method, targetURL, err)
		return
	}

	// Copie des en-têtes d'origine (sauf hop-by-hop)
	for k, values := range req.Headers {
		lower := strings.ToLower(k)
		if lower == "host" || lower == "content-length" || lower == "transfer-encoding" || lower == "connection" {
			continue
		}
		for _, v := range values {
			httpReq.Header.Add(k, v)
		}
	}

	// En-têtes personnalisés additionnels
	for k, v := range a.config.CustomHeaders {
		httpReq.Header.Set(k, v)
	}

	httpReq.Header.Set("X-ZeroDrop-Forwarded", "true")
	httpReq.Header.Set("X-ZeroDrop-Request-ID", req.ID)

	// Exécution HTTP locale
	resp, err := a.httpClient.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	timeStr := time.Now().Format("15:04:05")

	if err != nil {
		fmt.Printf("🔴 [%s] %s | %s /in/%s -> %s (ERR: %v) [%dms]\n",
			timeStr, "CONN_FAIL", req.Method, req.EndpointSlug, targetURL, err, duration,
		)
		return
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body) // Épuise le body pour libérer la connexion

	statusEmoji := "🟢"
	if resp.StatusCode >= 400 {
		statusEmoji = "🟡"
	}
	if resp.StatusCode >= 500 {
		statusEmoji = "🔴"
	}

	fmt.Printf("%s [%s] %d %s | %s /in/%s -> %s [%dms]\n",
		statusEmoji, timeStr, resp.StatusCode, http.StatusText(resp.StatusCode), req.Method, req.EndpointSlug, targetURL, duration,
	)
}
