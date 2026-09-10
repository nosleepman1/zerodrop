package hub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nosleepman1/zerodrop/internal/models"
)

func TestHubRegistrationAndBroadcast(t *testing.T) {
	h := NewHub()
	go h.Run()

	// Création d'un faux client UI
	clientChan := make(chan []byte, 10)
	client := &Client{
		hub:      h,
		send:     clientChan,
		isTunnel: false,
	}

	// 1. Inscription du client UI
	h.register <- client
	time.Sleep(20 * time.Millisecond)

	h.mu.RLock()
	if len(h.clients) != 1 {
		t.Fatalf("Attendu 1 client UI inscrit, obtenu %d", len(h.clients))
	}
	h.mu.RUnlock()

	// 2. Test de diffusion Broadcast
	testReq := models.WebhookRequest{
		ID:           "req_ws_test",
		EndpointSlug: "default",
		Method:       "POST",
		RawBody:      `{"event": "ping"}`,
	}
	h.BroadcastEvent(models.EventNewRequest, testReq)

	select {
	case msg := <-clientChan:
		var wsMsg models.WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			t.Fatalf("Erreur de decodage du message WS diffuse : %v", err)
		}
		if wsMsg.Type != models.EventNewRequest {
			t.Errorf("Type attendu '%s', obtenu '%s'", models.EventNewRequest, wsMsg.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Timeout : le message diffuse n'a pas ete recu par le client")
	}

	// 3. Test de désinscription
	h.unregister <- client
	time.Sleep(20 * time.Millisecond)

	h.mu.RLock()
	if len(h.clients) != 0 {
		t.Errorf("Attendu 0 client UI apres desinscription, obtenu %d", len(h.clients))
	}
	h.mu.RUnlock()
}

func TestHubTunnelForwarding(t *testing.T) {
	h := NewHub()
	go h.Run()

	tunnelChan := make(chan []byte, 10)
	tunnelClient := &Client{
		hub:          h,
		send:         tunnelChan,
		isTunnel:     true,
		endpointSlug: "stripe-dev",
	}

	h.register <- tunnelClient
	time.Sleep(20 * time.Millisecond)

	// Forward vers l'endpoint écouté
	req := models.WebhookRequest{
		ID:           "req_tunnel_01",
		EndpointSlug: "stripe-dev",
		Method:       "POST",
		RawBody:      `{"data": "secret"}`,
	}

	forwarded := h.ForwardToTunnel("stripe-dev", req)
	if !forwarded {
		t.Fatalf("Le forward vers le tunnel 'stripe-dev' aurait du reussir")
	}

	select {
	case msg := <-tunnelChan:
		var wsMsg models.WSMessage
		_ = json.Unmarshal(msg, &wsMsg)
		if wsMsg.Type != models.EventTunnelForward {
			t.Errorf("Type attendu EVENT_TUNNEL_FORWARD, obtenu %s", wsMsg.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Le tunnel n'a pas recu le message relaye")
	}

	// Forward vers un slug inexistant sans tunnel global
	forwarded = h.ForwardToTunnel("inexistant", req)
	if forwarded {
		t.Errorf("Le forward vers un endpoint non ecoute aurait du echouer")
	}
}
