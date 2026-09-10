package models

// Types d'événements diffusés via WebSockets.
const (
	EventNewRequest     = "EVENT_NEW_REQUEST"
	EventReplayResult   = "EVENT_REPLAY_RESULT"
	EventTunnelForward  = "EVENT_TUNNEL_FORWARD"
	EventTunnelResponse = "EVENT_TUNNEL_RESPONSE"
	EventEndpointUpdate = "EVENT_ENDPOINT_UPDATE"
)

// WSMessage encapsule tout message transitant sur le bus WebSocket.
type WSMessage struct {
	Type      string      `json:"type"`                // Type d'événement (ex: EVENT_NEW_REQUEST)
	Timestamp int64       `json:"timestamp"`           // Horodatage Unix en millisecondes
	Payload   interface{} `json:"payload"`             // Données associées à l'événement
}
