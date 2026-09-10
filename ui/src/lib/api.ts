import { Endpoint, WebhookRequest, ReplayLog, CreateEndpointPayload, TriggerReplayPayload } from "../types";

const API_BASE = "/api";

export async function fetchEndpoints(): Promise<Endpoint[]> {
  const res = await fetch(`${API_BASE}/endpoints`);
  if (!res.ok) throw new Error("Impossible de charger les endpoints");
  return res.json();
}

export async function createEndpoint(payload: CreateEndpointPayload): Promise<Endpoint> {
  const res = await fetch(`${API_BASE}/endpoints`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Erreur lors de la création de l'endpoint");
  }
  return res.json();
}

export async function deleteEndpoint(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/endpoints/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Impossible de supprimer l'endpoint");
}

export async function fetchRequests(endpointId?: string, search?: string): Promise<WebhookRequest[]> {
  const params = new URLSearchParams();
  if (endpointId) params.append("endpoint_id", endpointId);
  if (search) params.append("search", search);
  params.append("limit", "100");

  const res = await fetch(`${API_BASE}/requests?${params.toString()}`);
  if (!res.ok) throw new Error("Impossible de charger les requêtes");
  return res.json();
}

export async function clearRequests(endpointId?: string): Promise<void> {
  const params = new URLSearchParams();
  if (endpointId) params.append("endpoint_id", endpointId);
  await fetch(`${API_BASE}/requests?${params.toString()}`, { method: "DELETE" });
}

export async function triggerReplay(requestId: string, payload: TriggerReplayPayload): Promise<ReplayLog> {
  const res = await fetch(`${API_BASE}/requests/${requestId}/replay`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Erreur lors du rejeu");
  }
  return res.json();
}

export async function fetchReplays(requestId: string): Promise<ReplayLog[]> {
  const res = await fetch(`${API_BASE}/requests/${requestId}/replays`);
  if (!res.ok) throw new Error("Impossible de charger les logs de rejeu");
  return res.json();
}
