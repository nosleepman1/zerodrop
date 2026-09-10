import { useEffect, useRef, useState, useCallback } from "react";
import { WSMessage, WebhookRequest, ReplayLog, Endpoint } from "../types";

export type ConnectionStatus = "connecting" | "connected" | "disconnected";

interface UseWebSocketProps {
  onNewRequest?: (req: WebhookRequest) => void;
  onReplayResult?: (replay: ReplayLog) => void;
  onEndpointUpdate?: (endpoint: Endpoint | { deleted_id: string }) => void;
}

export function useWebSocket({ onNewRequest, onReplayResult, onEndpointUpdate }: UseWebSocketProps = {}) {
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);

  const connect = useCallback(() => {
    // Calcul de l'URL WebSocket selon l'environnement
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws/events`;

    setStatus("connecting");
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus("connected");
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };

    ws.onmessage = (event) => {
      try {
        const data: WSMessage = JSON.parse(event.data);
        switch (data.type) {
          case "EVENT_NEW_REQUEST":
            onNewRequest?.(data.payload as WebhookRequest);
            break;
          case "EVENT_REPLAY_RESULT":
            onReplayResult?.(data.payload as ReplayLog);
            break;
          case "EVENT_ENDPOINT_UPDATE":
            onEndpointUpdate?.(data.payload as Endpoint | { deleted_id: string });
            break;
        }
      } catch (err) {
        console.error("Erreur de parsing WS :", err);
      }
    };

    ws.onclose = () => {
      setStatus("disconnected");
      // Tentative de reconnexion automatique après 3 secondes
      reconnectTimeoutRef.current = window.setTimeout(() => {
        connect();
      }, 3000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [onNewRequest, onReplayResult, onEndpointUpdate]);

  useEffect(() => {
    connect();
    return () => {
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      if (wsRef.current) wsRef.current.close();
    };
  }, [connect]);

  return { status };
}
