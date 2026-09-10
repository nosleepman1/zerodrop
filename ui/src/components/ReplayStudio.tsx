import React, { useState, useEffect } from 'react';
import { Play, RotateCw, Check, Copy, Clock, AlertCircle, ArrowUpRight } from 'lucide-react';
import { WebhookRequest, ReplayLog } from '../types';
import { fetchReplays, triggerReplay } from '../lib/api';
import { formatDate } from '../lib/utils';

interface ReplayStudioProps {
  request: WebhookRequest;
  onReplayTriggered?: (log: ReplayLog) => void;
  modifiedBody?: string;
}

export const ReplayStudio: React.FC<ReplayStudioProps> = ({ request, onReplayTriggered, modifiedBody }) => {
  const [targetUrl, setTargetUrl] = useState('');
  const [replays, setReplays] = useState<ReplayLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  useEffect(() => {
    // Charge l'historique des rejeux pour cette requête
    fetchReplays(request.id)
      .then(setReplays)
      .catch((err) => console.error("Erreur chargement rejeux :", err));
  }, [request.id]);

  const handleReplay = async () => {
    setError(null);
    setLoading(true);

    try {
      const log = await triggerReplay(request.id, {
        target_url: targetUrl.trim() || undefined,
        modified_body: modifiedBody !== request.raw_body ? modifiedBody : undefined,
      });

      setReplays((prev) => [log, ...prev]);
      onReplayTriggered?.(log);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Erreur lors du rejeu");
    } finally {
      setLoading(false);
    }
  };

  const copyResponse = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="p-6 space-y-6">
      {/* Control Bar */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <RotateCw className="w-4 h-4 text-brand-400" />
            <h4 className="text-xs font-semibold text-white uppercase tracking-wider">
              Déclencher un Rejeu HTTP
            </h4>
          </div>
          {modifiedBody && modifiedBody !== request.raw_body && (
            <span className="text-[11px] bg-amber-500/10 text-amber-400 border border-amber-500/30 px-2 py-0.5 rounded-full font-medium">
              Payload Modifié
            </span>
          )}
        </div>

        <div className="flex items-center space-x-2">
          <input
            type="url"
            value={targetUrl}
            onChange={(e) => setTargetUrl(e.target.value)}
            placeholder="URL cible (ex: http://localhost:3000/api/webhook)"
            className="flex-1 bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-brand-500"
          />

          <button
            onClick={handleReplay}
            disabled={loading}
            className="flex items-center space-x-1.5 bg-brand-600 hover:bg-brand-500 disabled:opacity-50 text-white text-xs font-medium px-4 py-2 rounded-lg transition-colors shadow-sm shadow-brand-600/30"
          >
            {loading ? <RotateCw className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5 fill-white" />}
            <span>{loading ? 'Envoi...' : 'Rejouer'}</span>
          </button>
        </div>

        {error && (
          <div className="flex items-center space-x-2 text-xs text-rose-400 bg-rose-500/10 p-2.5 rounded-lg border border-rose-500/20">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}
      </div>

      {/* History Timeline */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
            Historique des Tentatives ({replays.length})
          </h4>
        </div>

        {replays.length === 0 ? (
          <div className="p-8 text-center border border-dashed border-slate-800 rounded-xl">
            <p className="text-xs text-slate-500">Aucun rejeu exécuté pour cette requête.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {replays.map((rep) => {
              const isSuccess = rep.status_code >= 200 && rep.status_code < 300;
              return (
                <div
                  key={rep.id}
                  className="bg-slate-900/70 border border-slate-800 rounded-xl p-4 space-y-2.5 hover:border-slate-700 transition-colors"
                >
                  {/* Top info */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2.5">
                      <span
                        className={`text-xs font-bold px-2 py-0.5 rounded border ${
                          isSuccess
                            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                            : 'bg-rose-500/10 text-rose-400 border-rose-500/30'
                        }`}
                      >
                        {rep.status_code || 'ERREUR'}
                      </span>

                      <span className="text-xs font-mono text-slate-300 flex items-center space-x-1">
                        <ArrowUpRight className="w-3 h-3 text-slate-500" />
                        <span>{rep.target_url}</span>
                      </span>
                    </div>

                    <div className="flex items-center space-x-3 text-[11px] text-slate-400">
                      <span className="flex items-center space-x-1 text-slate-300 font-mono">
                        <Clock className="w-3 h-3 text-brand-400" />
                        <span>{rep.duration_ms} ms</span>
                      </span>
                      <span>•</span>
                      <span>{formatDate(rep.created_at)}</span>
                    </div>
                  </div>

                  {/* Error or Response Body */}
                  {rep.error_message ? (
                    <div className="p-2.5 bg-rose-500/10 border border-rose-500/20 rounded-lg text-xs text-rose-300 font-mono">
                      {rep.error_message}
                    </div>
                  ) : (
                    rep.response_body && (
                      <div className="relative group bg-slate-950 rounded-lg p-3 border border-slate-800">
                        <button
                          onClick={() => copyResponse(rep.id, rep.response_body || '')}
                          className="absolute right-2 top-2 p-1 bg-slate-800 hover:bg-slate-700 rounded text-slate-400 hover:text-white transition-colors"
                          title="Copier la réponse"
                        >
                          {copiedId === rep.id ? (
                            <Check className="w-3 h-3 text-emerald-400" />
                          ) : (
                            <Copy className="w-3 h-3" />
                          )}
                        </button>
                        <pre className="text-xs font-mono text-slate-300 whitespace-pre-wrap max-h-40 overflow-y-auto pr-8">
                          {rep.response_body}
                        </pre>
                      </div>
                    )
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};
