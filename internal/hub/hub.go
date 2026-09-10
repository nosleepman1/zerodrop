package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/nosleepman1/zerodrop/internal/models"
)

// Hub gère l'ensemble des connexions WebSockets (clients UI et tunnels CLI).
type Hub struct {
	// Clients UI abonnés au flux en direct
	clients map[*Client]bool

	// Tunnels CLI connectés par endpointSlug (ou global)
	tunnels map[string]*Client

	// Canal pour diffuser un message à tous les clients UI
	broadcast chan []byte

	// Canaux d'enregistrement / désenregistrement
	register   chan *Client
	unregister chan *Client

	mu sync.RWMutex
}

// NewHub instancie un nouveau gestionnaire de WebSockets.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		tunnels:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run démarre la boucle événementielle du Hub dans une Goroutine dédiée.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if client.isTunnel {
				h.tunnels[client.endpointSlug] = client
				log.Printf("🔌 Tunnel CLI connecté pour l'endpoint '%s'", client.endpointSlug)
			} else {
				h.clients[client] = true
				log.Printf("💻 Client UI connecté (Total UI: %d)", len(h.clients))
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if client.isTunnel {
				if _, ok := h.tunnels[client.endpointSlug]; ok {
					delete(h.tunnels, client.endpointSlug)
					close(client.send)
					log.Printf("🔌 Tunnel CLI déconnecté pour l'endpoint '%s'", client.endpointSlug)
				}
			} else {
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					log.Printf("💻 Client UI déconnecté (Total UI: %d)", len(h.clients))
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent formate et diffuse un événement typé à tous les clients UI.
func (h *Hub) BroadcastEvent(eventType string, payload interface{}) {
	msg := models.WSMessage{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation du message WS : %v", err)
		return
	}

	h.broadcast <- bytes
}

// ForwardToTunnel transmet une requête webhook capturée à un agent CLI de tunneling actif.
func (h *Hub) ForwardToTunnel(endpointSlug string, req models.WebhookRequest) bool {
	h.mu.RLock()
	tunnelClient, exists := h.tunnels[endpointSlug]
	if !exists {
		// Vérification si un tunnel global '*' existe
		tunnelClient, exists = h.tunnels["*"]
	}
	h.mu.RUnlock()

	if !exists || tunnelClient == nil {
		return false
	}

	msg := models.WSMessage{
		Type:      models.EventTunnelForward,
		Timestamp: time.Now().UnixMilli(),
		Payload:   req,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return false
	}

	select {
	case tunnelClient.send <- bytes:
		return true
	default:
		return false
	}
}
