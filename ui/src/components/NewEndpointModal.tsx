import React, { useState } from 'react';
import { X, Globe, Key, CheckCircle2 } from 'lucide-react';
import { CreateEndpointPayload } from '../types';

interface NewEndpointModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (payload: CreateEndpointPayload) => Promise<void>;
}

export const NewEndpointModal: React.FC<NewEndpointModalProps> = ({ isOpen, onClose, onSubmit }) => {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [provider, setProvider] = useState<'custom' | 'stripe' | 'github' | 'shopify' | 'slack'>('custom');
  const [secret, setSecret] = useState('');
  const [forwardUrl, setForwardUrl] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setName(val);
    if (!slug || slug === name.toLowerCase().replace(/[^a-z0-9]/g, '-')) {
      setSlug(val.toLowerCase().replace(/[^a-z0-9]/g, '-'));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!name || !slug) {
      setError("Le nom et le slug d'URL sont obligatoires.");
      return;
    }

    try {
      setLoading(true);
      await onSubmit({
        name,
        slug,
        provider,
        secret: secret.trim() || undefined,
        forward_url: forwardUrl.trim() || undefined,
        description: description.trim() || undefined,
      });
      onClose();
      setName('');
      setSlug('');
      setSecret('');
      setForwardUrl('');
      setDescription('');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Erreur lors de la création");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center space-x-2.5">
            <div className="p-2 rounded-lg bg-brand-500/10 text-brand-400">
              <Globe className="w-5 h-5" />
            </div>
            <div>
              <h3 className="font-semibold text-white text-base">Créer un nouvel Endpoint</h3>
              <p className="text-xs text-slate-400">Générez une URL de réception avec validation de signature</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-xs text-rose-400">
              {error}
            </div>
          )}

          {/* Name & Slug */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Nom de l'Endpoint *</label>
              <input
                type="text"
                value={name}
                onChange={handleNameChange}
                placeholder="ex: Stripe Webhook"
                required
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Slug URL (/in/...) *</label>
              <input
                type="text"
                value={slug}
                onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                placeholder="stripe-webhook"
                required
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-brand-500"
              />
            </div>
          </div>

          {/* Provider */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Type de Fournisseur</label>
            <div className="grid grid-cols-5 gap-2">
              {[
                { id: 'custom', label: 'Personnalisé' },
                { id: 'stripe', label: 'Stripe' },
                { id: 'github', label: 'GitHub' },
                { id: 'shopify', label: 'Shopify' },
                { id: 'slack', label: 'Slack' },
              ].map((p) => (
                <button
                  type="button"
                  key={p.id}
                  onClick={() => setProvider(p.id as typeof provider)}
                  className={`px-2 py-1.5 rounded-lg text-xs font-medium border text-center transition-all ${
                    provider === p.id
                      ? 'bg-brand-600/20 border-brand-500 text-brand-300'
                      : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700'
                  }`}
                >
                  {p.label}
                </button>
              ))}
            </div>
          </div>

          {/* Secret HMAC */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-xs font-medium text-slate-300 flex items-center space-x-1.5">
                <Key className="w-3.5 h-3.5 text-brand-400" />
                <span>Secret de Signature HMAC (Optionnel)</span>
              </label>
              <span className="text-[10px] text-slate-500">Pour vérifier l'authenticité</span>
            </div>
            <input
              type="password"
              value={secret}
              onChange={(e) => setSecret(e.target.value)}
              placeholder={provider === 'stripe' ? 'whsec_...' : provider === 'github' ? 'secret_github' : 'secret_hmac'}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-brand-500"
            />
          </div>

          {/* Forward URL */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">
              URL de Redirection / Tunnel par Défaut (Optionnel)
            </label>
            <input
              type="url"
              value={forwardUrl}
              onChange={(e) => setForwardUrl(e.target.value)}
              placeholder="http://localhost:3000/api/webhook"
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-brand-500"
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Description (Optionnel)</label>
            <textarea
              rows={2}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Ex: Écoute des paiements réussis en environnement de test"
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-brand-500"
            />
          </div>

          {/* Buttons */}
          <div className="pt-3 border-t border-slate-800 flex items-center justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
            >
              Annuler
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center space-x-1.5 bg-brand-600 hover:bg-brand-500 disabled:opacity-50 text-white text-xs font-medium px-4 py-2 rounded-lg transition-colors shadow-sm shadow-brand-600/30"
            >
              <CheckCircle2 className="w-4 h-4" />
              <span>{loading ? 'Création...' : 'Créer l\'Endpoint'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
