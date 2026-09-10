package models

// Définition des types d'événements diffusés sur le bus WebSocket.
const (
	// EventNewRequest est émis dès la réception et la persistance d'une nouvelle requête webhook.
	EventNewRequest = "EVENT_NEW_REQUEST"

	// EventReplayResult est émis à la fin de l'exécution d'un rejeu HTTP.
	EventReplayResult = "EVENT_REPLAY_RESULT"

	// EventTunnelForward est envoyé par le Hub au client CLI de tunneling pour lui ordonner de relayer la requête en local.
	EventTunnelForward = "EVENT_TUNNEL_FORWARD"

	// EventTunnelResponse est envoyé par le client CLI au Hub pour lui notifier du résultat de l'exécution locale.
	EventTunnelResponse = "EVENT_TUNNEL_RESPONSE"

	// EventEndpointUpdate est diffusé lors de la création, modification ou suppression d'un endpoint.
	EventEndpointUpdate = "EVENT_ENDPOINT_UPDATE"
)

// WSMessage représente l'enveloppe standard de communication transitant sur les flux WebSockets.
type WSMessage struct {
	// Type identifie la nature de l'événement (ex: EVENT_NEW_REQUEST).
	Type string `json:"type"`

	// Timestamp est l'horodatage Unix en millisecondes de l'émission du message.
	Timestamp int64 `json:"timestamp"`

	// Payload contient la structure de données spécifique associée à l'événement.
	Payload interface{} `json:"payload"`
}
