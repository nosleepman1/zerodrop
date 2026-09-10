export interface Endpoint {
  id: string;
  name: string;
  slug: string;
  secret?: string;
  provider: 'stripe' | 'github' | 'shopify' | 'slack' | 'custom';
  forward_url?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface WebhookRequest {
  id: string;
  endpoint_id: string;
  endpoint_slug: string;
  method: string;
  path: string;
  headers: Record<string, string[]>;
  query_params?: Record<string, string[]>;
  raw_body: string;
  content_type: string;
  content_length: number;
  ip_address: string;
  signature_valid: boolean | null;
  created_at: string;
}

export interface ReplayLog {
  id: string;
  request_id: string;
  target_url: string;
  status_code: number;
  response_headers?: Record<string, string[]>;
  response_body?: string;
  duration_ms: number;
  error_message?: string;
  created_at: string;
}

export interface WSMessage<T = unknown> {
  type: string;
  timestamp: number;
  payload: T;
}

export interface CreateEndpointPayload {
  name: string;
  slug: string;
  secret?: string;
  provider?: string;
  forward_url?: string;
  description?: string;
}

export interface TriggerReplayPayload {
  target_url?: string;
  modified_body?: string;
  custom_headers?: Record<string, string>;
}
