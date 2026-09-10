// Package hub implémente le gestionnaire centralisé de connexions WebSockets de ZeroDrop.
// Il gère deux catégories de clients :
//  1. Les navigateurs (Dashboard Web UI) recevant le flux temps réel de tous les webhooks capturés.
//  2. Les agents CLI de tunneling recevant les requêtes spécifiques à relayer en local (localhost).
package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/nosleepman1/zerodrop/internal/models"
)

// Hub maintient l'ensemble des connexions WebSockets actives et coordonne la diffusion
// asynchrone des messages via des canaux Go concurrents.
type Hub struct {
	// clients stocke l'ensemble des connexions du Dashboard UI abonnées au flux global.
	clients map[*Client]bool

	// tunnels associe chaque slug d'endpoint au client CLI de tunneling correspondant.
	tunnels map[string]*Client

	// broadcast est le canal tamponné recevant les messages destinés à tous les clients UI.
	broadcast chan []byte

	// register gère l'inscription asynchrone d'une nouvelle connexion.
	register chan *Client

	// unregister gère le désabonnement et la fermeture propre d'une connexion.
	unregister chan *Client

	// mu protège l'accès concurrent aux dictionnaires clients et tunnels.
	mu sync.RWMutex
}

// NewHub instancie et initialise une nouvelle structure Hub prête à être démarrée.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		tunnels:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run démarre la boucle événementielle du Hub. Cette méthode doit être exécutée
// dans une Goroutine dédiée pour assurer un traitement non bloquant des événements.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if client.isTunnel {
				h.tunnels[client.endpointSlug] = client
				log.Printf("[TUNNEL] Connexion de l'agent CLI pour l'endpoint '%s'", client.endpointSlug)
			} else {
				h.clients[client] = true
				log.Printf("[WS] Connexion d'un client Dashboard UI (Total actif: %d)", len(h.clients))
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if client.isTunnel {
				if _, ok := h.tunnels[client.endpointSlug]; ok {
					delete(h.tunnels, client.endpointSlug)
					close(client.send)
					log.Printf("[TUNNEL] Deconnexion de l'agent CLI pour l'endpoint '%s'", client.endpointSlug)
				}
			} else {
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					log.Printf("[WS] Deconnexion d'un client Dashboard UI (Total actif: %d)", len(h.clients))
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// En cas de saturation du tampon d'un client lent, fermeture de la connexion
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent sérialise une charge utile et la diffuse à l'ensemble des clients UI connectés.
func (h *Hub) BroadcastEvent(eventType string, payload interface{}) {
	msg := models.WSMessage{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ERROR] Echec de serialisation du message WebSocket : %v", err)
		return
	}

	h.broadcast <- bytes
}

// ForwardToTunnel achemine une requête webhook capturée vers l'agent CLI de tunneling correspondant.
// Retourne true si un tunnel était connecté et a accepté le message, false sinon.
func (h *Hub) ForwardToTunnel(endpointSlug string, req models.WebhookRequest) bool {
	h.mu.RLock()
	tunnelClient, exists := h.tunnels[endpointSlug]
	if !exists {
		// Vérification de la présence d'un tunnel global ('*')
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
