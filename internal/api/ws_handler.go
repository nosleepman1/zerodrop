package api

import (
	"net/http"

	"github.com/nosleepman1/zerodrop/internal/hub"
)

// WSHandler gère l'attachement des clients WebSockets (UI et Tunnels).
type WSHandler struct {
	hub *hub.Hub
}

// NewWSHandler instancie le handler WebSocket.
func NewWSHandler(h *hub.Hub) *WSHandler {
	return &WSHandler{hub: h}
}

// ServeEvents gère les connexions WebSockets du Dashboard UI (Live Feed).
func (h *WSHandler) ServeEvents(w http.ResponseWriter, r *http.Request) {
	hub.ServeWS(h.hub, w, r, false, "")
}

// ServeTunnel gère les connexions WebSockets des agents CLI de Tunneling.
func (h *WSHandler) ServeTunnel(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		slug = "*" // Tunnel global par défaut
	}
	hub.ServeWS(h.hub, w, r, true, slug)
}
