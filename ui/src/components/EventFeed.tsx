import React from 'react';
import { Search, ShieldCheck, ShieldAlert, Clock, Inbox, ChevronRight } from 'lucide-react';
import { WebhookRequest } from '../types';
import { formatDate, formatBytes } from '../lib/utils';

interface EventFeedProps {
  requests: WebhookRequest[];
  selectedRequestId: string | null;
  onSelectRequest: (id: string) => void;
  searchQuery: string;
  onSearchChange: (q: string) => void;
}

export const EventFeed: React.FC<EventFeedProps> = ({
  requests,
  selectedRequestId,
  onSelectRequest,
  searchQuery,
  onSearchChange,
}) => {
  const getMethodBadgeClass = (method: string) => {
    switch (method.toUpperCase()) {
      case 'POST':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
      case 'GET':
        return 'bg-sky-500/10 text-sky-400 border-sky-500/30';
      case 'PUT':
      case 'PATCH':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
      case 'DELETE':
        return 'bg-rose-500/10 text-rose-400 border-rose-500/30';
      default:
        return 'bg-slate-500/10 text-slate-400 border-slate-500/30';
    }
  };

  return (
    <div className="w-80 md:w-96 border-r border-slate-800 bg-slate-900/50 flex flex-col h-full shrink-0">
      {/* Search Header */}
      <div className="p-3 border-b border-slate-800 bg-slate-900/70">
        <div className="relative">
          <Search className="w-3.5 h-3.5 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Rechercher payload, ID, route..."
            className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-8 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-brand-500 transition-colors"
          />
        </div>
      </div>

      {/* Requests List */}
      <div className="flex-1 overflow-y-auto divide-y divide-slate-800/60">
        {requests.length === 0 ? (
          <div className="p-8 text-center flex flex-col items-center justify-center space-y-3">
            <div className="w-12 h-12 rounded-full bg-slate-800/60 flex items-center justify-center text-slate-500">
              <Inbox className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs font-semibold text-slate-300">En attente de requêtes</p>
              <p className="text-[11px] text-slate-500 mt-0.5">
                Envoyez un webhook vers votre endpoint pour le voir apparaître en direct.
              </p>
            </div>
          </div>
        ) : (
          requests.map((req) => {
            const isSelected = req.id === selectedRequestId;
            return (
              <button
                key={req.id}
                onClick={() => onSelectRequest(req.id)}
                className={`w-full text-left p-3.5 transition-all flex items-start justify-between group ${
                  isSelected
                    ? 'bg-brand-600/10 border-l-2 border-brand-500'
                    : 'hover:bg-slate-800/40 border-l-2 border-transparent'
                }`}
              >
                <div className="space-y-1.5 min-w-0 pr-2">
                  {/* Top line : Method + Slug + Signature status */}
                  <div className="flex items-center space-x-2">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.5 rounded border uppercase tracking-wider ${getMethodBadgeClass(
                        req.method
                      )}`}
                    >
                      {req.method}
                    </span>

                    <span className="text-xs font-medium text-slate-300 truncate max-w-[140px]">
                      /{req.endpoint_slug}
                    </span>

                    {/* Signature Status */}
                    {req.signature_valid === true && (
                      <span className="text-emerald-400" title="Signature HMAC vérifiée avec succès">
                        <ShieldCheck className="w-3.5 h-3.5" />
                      </span>
                    )}
                    {req.signature_valid === false && (
                      <span className="text-rose-400" title="Échec de validation de signature HMAC">
                        <ShieldAlert className="w-3.5 h-3.5" />
                      </span>
                    )}
                  </div>

                  {/* Snippet / preview */}
                  <p className="text-[11px] text-slate-400 font-mono truncate max-w-[220px]">
                    {req.raw_body ? req.raw_body.replace(/\s+/g, ' ').slice(0, 50) : '(Payload vide)'}
                  </p>

                  {/* Metadata line: Time + Size */}
                  <div className="flex items-center space-x-3 text-[10px] text-slate-500">
                    <span className="flex items-center space-x-1">
                      <Clock className="w-3 h-3" />
                      <span>{formatDate(req.created_at)}</span>
                    </span>
                    <span>•</span>
                    <span>{formatBytes(req.content_length)}</span>
                  </div>
                </div>

                <ChevronRight
                  className={`w-4 h-4 text-slate-600 transition-transform group-hover:translate-x-0.5 ${
                    isSelected ? 'text-brand-400' : ''
                  }`}
                />
              </button>
            );
          })
        )}
      </div>
    </div>
  );
};
