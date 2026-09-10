import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Header } from './components/Header';
import { EventFeed } from './components/EventFeed';
import { RequestDetail } from './components/RequestDetail';
import { NewEndpointModal } from './components/NewEndpointModal';
import { useWebSocket } from './hooks/useWebSocket';
import { fetchEndpoints, fetchRequests, createEndpoint, clearRequests } from './lib/api';
import { Endpoint, WebhookRequest, CreateEndpointPayload } from './types';

export const App: React.FC = () => {
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [selectedEndpointId, setSelectedEndpointId] = useState<string>('');
  const [requests, setRequests] = useState<WebhookRequest[]>([]);
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);

  // 1. Chargement initial des endpoints et des requêtes
  const loadData = useCallback(async () => {
    try {
      const eps = await fetchEndpoints();
      setEndpoints(eps);

      const reqs = await fetchRequests(selectedEndpointId || undefined);
      setRequests(reqs);
      if (reqs.length > 0 && !selectedRequestId) {
        setSelectedRequestId(reqs[0].id);
      }
    } catch (err) {
      console.error("Erreur de chargement des données :", err);
    }
  }, [selectedEndpointId, selectedRequestId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // 2. Écoute du flux WebSocket temps réel
  const handleNewRequest = useCallback((newReq: WebhookRequest) => {
    setRequests((prev) => {
      if (prev.some((r) => r.id === newReq.id)) return prev;
      return [newReq, ...prev];
    });

    setSelectedRequestId((curr) => curr || newReq.id);
  }, []);

  const handleEndpointUpdate = useCallback(() => {
    fetchEndpoints().then(setEndpoints).catch(console.error);
  }, []);

  const { status: wsStatus } = useWebSocket({
    onNewRequest: handleNewRequest,
    onEndpointUpdate: handleEndpointUpdate,
  });

  // 3. Filtrage côté client des requêtes selon la recherche et l'endpoint
  const filteredRequests = useMemo(() => {
    return requests.filter((r) => {
      const matchEndpoint = !selectedEndpointId || r.endpoint_id === selectedEndpointId;
      const q = searchQuery.toLowerCase();
      const matchSearch =
        !q ||
        r.id.toLowerCase().includes(q) ||
        r.path.toLowerCase().includes(q) ||
        r.raw_body.toLowerCase().includes(q) ||
        r.endpoint_slug.toLowerCase().includes(q);
      return matchEndpoint && matchSearch;
    });
  }, [requests, selectedEndpointId, searchQuery]);

  const selectedRequest = useMemo(() => {
    return requests.find((r) => r.id === selectedRequestId) || null;
  }, [requests, selectedRequestId]);

  // Handlers
  const handleCreateEndpoint = async (payload: CreateEndpointPayload) => {
    const created = await createEndpoint(payload);
    setEndpoints((prev) => [created, ...prev]);
    setSelectedEndpointId(created.id);
  };

  const handleClear = async () => {
    if (confirm("Êtes-vous sûr de vouloir purger toutes les requêtes capturées ?")) {
      await clearRequests(selectedEndpointId || undefined);
      setRequests([]);
      setSelectedRequestId(null);
    }
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden bg-slate-950 text-slate-100 font-sans">
      {/* Header */}
      <Header
        endpoints={endpoints}
        selectedEndpointId={selectedEndpointId}
        onSelectEndpoint={setSelectedEndpointId}
        onOpenNewEndpointModal={() => setIsModalOpen(true)}
        onClearRequests={handleClear}
        connectionStatus={wsStatus}
        totalRequests={requests.length}
      />

      {/* Main Studio Body */}
      <div className="flex flex-1 overflow-hidden">
        <EventFeed
          requests={filteredRequests}
          selectedRequestId={selectedRequestId}
          onSelectRequest={setSelectedRequestId}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
        />

        <RequestDetail
          request={selectedRequest}
        />
      </div>

      {/* Modal Création Endpoint */}
      <NewEndpointModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateEndpoint}
      />
    </div>
  );
};

export default App;
