package hub

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait définit le délai d'attente maximal pour l'écriture d'un message sur la socket.
	writeWait = 10 * time.Second

	// pongWait définit le délai d'attente maximal pour la réception d'un battement de cœur Pong.
	pongWait = 60 * time.Second

	// pingPeriod définit l'intervalle d'envoi périodique des pings vers le client.
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize est la taille maximale autorisée d'un message entrant (512 Ko).
	maxMessageSize = 512 * 1024
)

// upgrader configure les paramètres de mise à niveau de connexion HTTP vers WebSocket.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Autorise les connexions multi-origines pour le développement local et les tunnels distants
		return true
	},
}

// Client représente une connexion WebSocket active, encapsulant les buffers de lecture et d'écriture.
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan []byte
	isTunnel     bool
	endpointSlug string
}

// ReadPump écoute en continu les trames entrantes pour maintenir la connexion et gérer les pings/pongs.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Fermeture inattendue de connexion : %v", err)
			}
			break
		}
	}
}

// WritePump assure l'acheminement des messages sérialisés du canal send vers la socket WebSocket.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Le canal a été fermé par le Hub
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Vidage des messages supplémentaires en attente dans le même paquet
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS orchestre la mise à niveau HTTP -> WebSocket et associe le Client au Hub.
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, isTunnel bool, endpointSlug string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] Echec de mise a niveau WebSocket : %v", err)
		return
	}

	client := &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, 256),
		isTunnel:     isTunnel,
		endpointSlug: endpointSlug,
	}

	client.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
