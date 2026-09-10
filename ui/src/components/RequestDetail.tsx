import React, { useState, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import {
  Code,
  FileText,
  RotateCw,
  Copy,
  Check,
  ShieldCheck,
  ShieldAlert,
  Terminal,
  FileCode,
  Clock,
  HardDrive,
  Globe,
} from 'lucide-react';
import { WebhookRequest, ReplayLog } from '../types';
import { ReplayStudio } from './ReplayStudio';
import { formatDate, formatBytes } from '../lib/utils';

interface RequestDetailProps {
  request: WebhookRequest | null;
  onReplayTriggered?: (log: ReplayLog) => void;
}

export const RequestDetail: React.FC<RequestDetailProps> = ({ request, onReplayTriggered }) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'payload' | 'replay'>('payload');
  const [payloadCode, setPayloadCode] = useState('');
  const [copiedType, setCopiedType] = useState<'curl' | 'ts' | 'body' | null>(null);

  useEffect(() => {
    if (request) {
      // Formate le JSON si possible
      try {
        const parsed = JSON.parse(request.raw_body);
        setPayloadCode(JSON.stringify(parsed, null, 2));
      } catch {
        setPayloadCode(request.raw_body);
      }
    }
  }, [request]);

  if (!request) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-12 text-center text-slate-500 bg-slate-950">
        <FileText className="w-12 h-12 stroke-1 mb-3 text-slate-700" />
        <h3 className="text-sm font-medium text-slate-400">Aucune requête sélectionnée</h3>
        <p className="text-xs text-slate-600 mt-1 max-w-sm">
          Sélectionnez un webhook dans la liste de gauche pour inspecter ses en-têtes, son payload et déclencher des rejeux.
        </p>
      </div>
    );
  }

  // Génération de la commande cURL
  const generateCurl = () => {
    let curl = `curl -X ${request.method} "${window.location.origin}${request.path}"`;
    for (const [key, values] of Object.entries(request.headers)) {
      if (['host', 'content-length'].includes(key.toLowerCase())) continue;
      curl += ` \\\n  -H "${key}: ${values.join('; ')}"`;
    }
    if (request.raw_body) {
      curl += ` \\\n  -d '${request.raw_body.replace(/'/g, "'\\''")}'`;
    }
    return curl;
  };

  // Génération de types TypeScript simplifiés à partir du JSON
  const generateTypeScriptTypes = () => {
    try {
      const obj = JSON.parse(request.raw_body);
      let ts = `export interface WebhookPayload {\n`;
      for (const [k, v] of Object.entries(obj)) {
        const type = Array.isArray(v) ? 'any[]' : typeof v;
        ts += `  ${k}: ${type};\n`;
      }
      ts += `}`;
      return ts;
    } catch {
      return `export interface WebhookPayload {\n  [key: string]: any;\n}`;
    }
  };

  const handleCopy = (type: 'curl' | 'ts' | 'body') => {
    let text = '';
    if (type === 'curl') text = generateCurl();
    if (type === 'ts') text = generateTypeScriptTypes();
    if (type === 'body') text = payloadCode;

    navigator.clipboard.writeText(text);
    setCopiedType(type);
    setTimeout(() => setCopiedType(null), 2000);
  };

  return (
    <div className="flex-1 flex flex-col h-full bg-slate-950 overflow-hidden">
      {/* Top Banner Bar */}
      <div className="p-4 border-b border-slate-800 bg-slate-900/60 flex items-center justify-between shrink-0">
        <div className="space-y-1">
          <div className="flex items-center space-x-3">
            <span className="text-xs font-bold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 uppercase">
              {request.method}
            </span>
            <span className="text-sm font-mono font-medium text-white">{request.path}</span>
            <span className="text-xs text-slate-500 font-mono">({request.id})</span>
          </div>

          <div className="flex items-center space-x-4 text-xs text-slate-400">
            <span className="flex items-center space-x-1.5">
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              <span>{formatDate(request.created_at)}</span>
            </span>
            <span className="flex items-center space-x-1.5">
              <Globe className="w-3.5 h-3.5 text-slate-500" />
              <span>IP: {request.ip_address}</span>
            </span>
            <span className="flex items-center space-x-1.5">
              <HardDrive className="w-3.5 h-3.5 text-slate-500" />
              <span>{formatBytes(request.content_length)}</span>
            </span>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center space-x-2">
          <button
            onClick={() => handleCopy('curl')}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 hover:border-slate-700 text-xs text-slate-300 hover:text-white transition-colors"
            title="Copier en commande cURL"
          >
            {copiedType === 'curl' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Terminal className="w-3.5 h-3.5" />}
            <span>cURL</span>
          </button>

          <button
            onClick={() => handleCopy('ts')}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 hover:border-slate-700 text-xs text-slate-300 hover:text-white transition-colors"
            title="Générer interface TypeScript"
          >
            {copiedType === 'ts' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <FileCode className="w-3.5 h-3.5" />}
            <span>Types TS</span>
          </button>

          <button
            onClick={() => setActiveTab('replay')}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-brand-600 hover:bg-brand-500 text-xs font-medium text-white transition-colors shadow-sm shadow-brand-600/30"
          >
            <RotateCw className="w-3.5 h-3.5" />
            <span>Rejouer</span>
          </button>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex items-center space-x-1 px-4 border-b border-slate-800 bg-slate-900/30 shrink-0">
        <button
          onClick={() => setActiveTab('payload')}
          className={`flex items-center space-x-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-colors ${
            activeTab === 'payload'
              ? 'border-brand-500 text-brand-300'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          <Code className="w-3.5 h-3.5" />
          <span>Payload JSON (Monaco)</span>
        </button>

        <button
          onClick={() => setActiveTab('overview')}
          className={`flex items-center space-x-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-colors ${
            activeTab === 'overview'
              ? 'border-brand-500 text-brand-300'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          <FileText className="w-3.5 h-3.5" />
          <span>En-têtes & Paramètres</span>
        </button>

        <button
          onClick={() => setActiveTab('replay')}
          className={`flex items-center space-x-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-colors ${
            activeTab === 'replay'
              ? 'border-brand-500 text-brand-300'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          <RotateCw className="w-3.5 h-3.5" />
          <span>Replay Studio</span>
        </button>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'payload' && (
          <div className="h-full flex flex-col">
            <div className="p-2 px-4 bg-slate-900/40 border-b border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
              <span>Éditeur JSON interactif (modifications actives pour le rejeu)</span>
              <button
                onClick={() => handleCopy('body')}
                className="flex items-center space-x-1 hover:text-white transition-colors"
              >
                {copiedType === 'body' ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                <span>Copier</span>
              </button>
            </div>
            <div className="flex-1 min-h-[350px]">
              <Editor
                height="100%"
                defaultLanguage="json"
                theme="vs-dark"
                value={payloadCode}
                onChange={(val) => setPayloadCode(val || '')}
                options={{
                  minimap: { enabled: false },
                  fontSize: 12,
                  scrollBeyondLastLine: false,
                  wordWrap: 'on',
                  formatOnPaste: true,
                  automaticLayout: true,
                }}
              />
            </div>
          </div>
        )}

        {activeTab === 'overview' && (
          <div className="p-6 space-y-6">
            {/* Signature Status */}
            <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 flex items-center justify-between">
              <div className="flex items-center space-x-3">
                {request.signature_valid === true ? (
                  <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
                    <ShieldCheck className="w-5 h-5" />
                  </div>
                ) : request.signature_valid === false ? (
                  <div className="p-2 rounded-lg bg-rose-500/10 text-rose-400">
                    <ShieldAlert className="w-5 h-5" />
                  </div>
                ) : (
                  <div className="p-2 rounded-lg bg-slate-800 text-slate-400">
                    <Globe className="w-5 h-5" />
                  </div>
                )}
                <div>
                  <h4 className="text-xs font-semibold text-white">
                    {request.signature_valid === true
                      ? 'Signature HMAC Valide'
                      : request.signature_valid === false
                      ? 'Signature HMAC Invalide / Échouée'
                      : 'Aucun secret configuré'}
                  </h4>
                  <p className="text-[11px] text-slate-400">
                    {request.signature_valid === true
                      ? "L'authenticité et l'intégrité de la charge utile sont garanties."
                      : request.signature_valid === false
                      ? 'La clé secrète ne correspond pas à la signature transmise.'
                      : "La requête a été acceptée sans vérification d'empreinte HMAC."}
                  </p>
                </div>
              </div>
            </div>

            {/* Headers Table */}
            <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
              <div className="px-4 py-3 border-b border-slate-800 bg-slate-900/80 flex items-center justify-between">
                <h4 className="text-xs font-semibold text-white uppercase tracking-wider">
                  En-têtes HTTP Reçus ({Object.keys(request.headers).length})
                </h4>
              </div>
              <div className="divide-y divide-slate-800/60 font-mono text-xs">
                {Object.entries(request.headers).map(([key, vals]) => (
                  <div key={key} className="p-3 flex items-start space-x-4 hover:bg-slate-800/30">
                    <span className="text-brand-300 font-semibold w-1/3 shrink-0 break-all">{key}</span>
                    <span className="text-slate-300 w-2/3 break-all">{vals.join(', ')}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Query Params Table if any */}
            {request.query_params && Object.keys(request.query_params).length > 0 && (
              <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
                <div className="px-4 py-3 border-b border-slate-800 bg-slate-900/80">
                  <h4 className="text-xs font-semibold text-white uppercase tracking-wider">Paramètres d'URL (Query)</h4>
                </div>
                <div className="divide-y divide-slate-800/60 font-mono text-xs">
                  {Object.entries(request.query_params).map(([key, vals]) => (
                    <div key={key} className="p-3 flex items-start space-x-4 hover:bg-slate-800/30">
                      <span className="text-amber-300 font-semibold w-1/3 shrink-0">{key}</span>
                      <span className="text-slate-300 w-2/3">{vals.join(', ')}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'replay' && (
          <ReplayStudio
            request={request}
            modifiedBody={payloadCode}
            onReplayTriggered={onReplayTriggered}
          />
        )}
      </div>
    </div>
  );
};
