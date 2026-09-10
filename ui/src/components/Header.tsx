import React, { useState } from 'react';
import { Zap, Plus, Trash2, Copy, Check, Radio } from 'lucide-react';
import { Endpoint } from '../types';
import { ConnectionStatus } from '../hooks/useWebSocket';

interface HeaderProps {
  endpoints: Endpoint[];
  selectedEndpointId: string;
  onSelectEndpoint: (id: string) => void;
  onOpenNewEndpointModal: () => void;
  onClearRequests: () => void;
  connectionStatus: ConnectionStatus;
  totalRequests: number;
}

export const Header: React.FC<HeaderProps> = ({
  endpoints,
  selectedEndpointId,
  onSelectEndpoint,
  onOpenNewEndpointModal,
  onClearRequests,
  connectionStatus,
  totalRequests,
}) => {
  const [copied, setCopied] = useState(false);

  const currentEndpoint = endpoints.find((e) => e.id === selectedEndpointId) || endpoints[0];
  const origin = window.location.origin;
  const webhookUrl = currentEndpoint ? `${origin}/in/${currentEndpoint.slug}` : '';

  const copyWebhookUrl = () => {
    if (!webhookUrl) return;
    navigator.clipboard.writeText(webhookUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <header className="h-16 border-b border-slate-800 bg-slate-900/90 backdrop-blur px-6 flex items-center justify-between shrink-0 z-10">
      {/* Logo & Brand */}
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-2.5">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-tr from-brand-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-brand-500/20">
            <Zap className="w-5 h-5 text-white fill-white" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-bold text-lg tracking-tight text-white">ZeroDrop</span>
              <span className="text-[10px] uppercase font-semibold bg-brand-500/20 text-brand-300 px-1.5 py-0.5 rounded border border-brand-500/30">
                Core Engine
              </span>
            </div>
            <p className="text-[11px] text-slate-400">Webhook Replay & Tunnel Studio</p>
          </div>
        </div>

        {/* WebSocket Status Pill */}
        <div className="flex items-center space-x-2 px-2.5 py-1 rounded-full bg-slate-800/80 border border-slate-700/60 text-xs">
          <span
            className={`w-2 h-2 rounded-full ${
              connectionStatus === 'connected'
                ? 'bg-emerald-400 animate-pulse shadow-sm shadow-emerald-400'
                : connectionStatus === 'connecting'
                ? 'bg-amber-400 animate-ping'
                : 'bg-rose-500'
            }`}
          />
          <span className="text-slate-300 text-[11px] font-medium">
            {connectionStatus === 'connected'
              ? 'Temps Réel Actif'
              : connectionStatus === 'connecting'
              ? 'Connexion...'
              : 'Déconnecté'}
          </span>
        </div>
      </div>

      {/* Middle: Active Endpoint & URL Copy */}
      <div className="flex items-center space-x-3 bg-slate-950/60 p-1.5 rounded-lg border border-slate-800">
        <div className="flex items-center space-x-2 px-2">
          <Radio className="w-3.5 h-3.5 text-brand-400" />
          <select
            value={selectedEndpointId}
            onChange={(e) => onSelectEndpoint(e.target.value)}
            className="bg-transparent text-xs font-medium text-slate-200 focus:outline-none cursor-pointer pr-4"
          >
            <option value="" className="bg-slate-900 text-slate-200">
              Tous les Endpoints
            </option>
            {endpoints.map((ep) => (
              <option key={ep.id} value={ep.id} className="bg-slate-900 text-slate-200">
                {ep.name} (/in/{ep.slug})
              </option>
            ))}
          </select>
        </div>

        {webhookUrl && (
          <div className="flex items-center space-x-1.5 bg-slate-900 px-2.5 py-1 rounded border border-slate-700/70 text-xs">
            <code className="text-brand-300 font-mono text-[11px] max-w-[280px] truncate">{webhookUrl}</code>
            <button
              onClick={copyWebhookUrl}
              className="p-1 hover:bg-slate-800 rounded text-slate-400 hover:text-white transition-colors"
              title="Copier l'URL d'ingestion"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
            </button>
          </div>
        )}
      </div>

      {/* Right Actions */}
      <div className="flex items-center space-x-3">
        <button
          onClick={onOpenNewEndpointModal}
          className="flex items-center space-x-1.5 bg-brand-600 hover:bg-brand-500 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition-colors shadow-sm shadow-brand-600/30"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Nouvel Endpoint</span>
        </button>

        <button
          onClick={onClearRequests}
          disabled={totalRequests === 0}
          className="flex items-center space-x-1 text-xs text-slate-400 hover:text-rose-400 disabled:opacity-40 disabled:hover:text-slate-400 px-2.5 py-1.5 rounded-lg hover:bg-slate-800/60 transition-colors"
          title="Purger l'historique"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span>Purger ({totalRequests})</span>
        </button>
      </div>
    </header>
  );
};
