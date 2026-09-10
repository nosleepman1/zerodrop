// Package tunnel fournit l'agent client de tunneling local permettant aux développeurs
// de recevoir instantanément sur leur machine (localhost) les requêtes webhooks capturées dans le cloud.
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

// Config regroupe les paramètres de configuration pour l'agent de tunneling local.
type Config struct {
	// ServerURL est l'URL du hub ZeroDrop (ex: ws://localhost:8080 ou wss://webhooks.domaine.com).
	ServerURL string

	// EndpointSlug est le slug d'endpoint spécifique à écouter (ex: "default", "stripe-dev", ou "*").
	EndpointSlug string

	// ForwardTo est l'URL locale vers laquelle acheminer les requêtes (ex: "http://localhost:3000/api/webhook").
	ForwardTo string

	// CustomHeaders contient les en-têtes HTTP personnalisés à injecter lors de la redirection locale.
	CustomHeaders map[string]string
}

// Agent représente l'agent de tunneling local gérant la liaison WebSocket persistante.
type Agent struct {
	config     Config
	httpClient *http.Client
}

// NewAgent instancie un nouvel agent de tunneling local.
func NewAgent(cfg Config) *Agent {
	return &Agent{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Start établit la connexion au Hub WebSocket et boucle avec reconnexion automatique en cas de coupure réseau.
func (a *Agent) Start() error {
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
	fmt.Printf("[INFO] ZeroDrop Tunnel Agent - Relai Local de Webhooks\n")
	fmt.Printf("[INFO] Hub distant      : %s\n", serverURL)
	fmt.Printf("[INFO] Endpoint ecoute  : %s\n", slug)
	fmt.Printf("[INFO] Redirection vers : %s\n", a.config.ForwardTo)
	fmt.Println("================================================================")
	fmt.Println("[INFO] En attente de requetes webhooks... (Ctrl+C pour quitter)")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	for {
		err := a.connectAndListen(wsURL, interrupt)
		if err != nil {
			log.Printf("[WARN] Connexion WebSocket interrompue : %v. Reconnexion automatique dans 3s...", err)
		}

		select {
		case <-interrupt:
			fmt.Println("\n[INFO] Arret de l'agent de tunneling local.")
			return nil
		case <-time.After(3 * time.Second):
			// Nouvelle tentative de reconnexion
		}
	}
}

func (a *Agent) connectAndListen(wsURL string, interrupt chan os.Signal) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("echec de connexion au Hub distant : %w", err)
	}
	defer conn.Close()

	log.Printf("[OK] Connexion etablie avec succes au Hub. Relai actif vers %s", a.config.ForwardTo)

	done := make(chan struct{})

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
		return fmt.Errorf("connexion fermee par le serveur distant")
	case <-interrupt:
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return nil
	}
}

// forwardLocally réémet la requête reçue vers l'URL locale cible et affiche les métriques dans le terminal.
func (a *Agent) forwardLocally(req models.WebhookRequest) {
	startTime := time.Now()

	targetURL := a.config.ForwardTo
	httpReq, err := http.NewRequest(req.Method, targetURL, bytes.NewBufferString(req.RawBody))
	if err != nil {
		log.Printf("[ERROR] [%s] Impossible de creer la requete vers %s : %v", req.Method, targetURL, err)
		return
	}

	for k, values := range req.Headers {
		lower := strings.ToLower(k)
		if lower == "host" || lower == "content-length" || lower == "transfer-encoding" || lower == "connection" {
			continue
		}
		for _, v := range values {
			httpReq.Header.Add(k, v)
		}
	}

	for k, v := range a.config.CustomHeaders {
		httpReq.Header.Set(k, v)
	}

	httpReq.Header.Set("X-ZeroDrop-Forwarded", "true")
	httpReq.Header.Set("X-ZeroDrop-Request-ID", req.ID)

	resp, err := a.httpClient.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	timeStr := time.Now().Format("15:04:05")

	if err != nil {
		fmt.Printf("[%s] [ERROR] %s /in/%s -> %s (ERR: %v) [%dms]\n",
			timeStr, req.Method, req.EndpointSlug, targetURL, err, duration,
		)
		return
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	statusTag := "[OK]"
	if resp.StatusCode >= 400 {
		statusTag = "[WARN]"
	}
	if resp.StatusCode >= 500 {
		statusTag = "[ERROR]"
	}

	fmt.Printf("[%s] %s %d %s | %s /in/%s -> %s [%dms]\n",
		timeStr, statusTag, resp.StatusCode, http.StatusText(resp.StatusCode), req.Method, req.EndpointSlug, targetURL, duration,
	)
}
