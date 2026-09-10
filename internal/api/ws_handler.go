package api

import (
	"net/http"

	"github.com/nosleepman1/zerodrop/internal/hub"
)

// WSHandler gère l'attachement et la mise à niveau des clients WebSockets.
type WSHandler struct {
	hub *hub.Hub
}

// NewWSHandler instancie le gestionnaire WebSocket.
func NewWSHandler(h *hub.Hub) *WSHandler {
	return &WSHandler{hub: h}
}

// ServeEvents gère les flux WebSockets du Dashboard Web UI.
func (h *WSHandler) ServeEvents(w http.ResponseWriter, r *http.Request) {
	hub.ServeWS(h.hub, w, r, false, "")
}

// ServeTunnel gère les flux WebSockets des agents de tunneling CLI.
func (h *WSHandler) ServeTunnel(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		slug = "*"
	}
	hub.ServeWS(h.hub, w, r, true, slug)
}
