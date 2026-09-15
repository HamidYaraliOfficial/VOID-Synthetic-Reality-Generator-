'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { api, World } from '@/lib/api-client';
import { Plus } from 'lucide-react';

export default function WorldsPage() {
  const { t } = useI18n();
  const [worlds, setWorlds] = useState<World[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('My Synthetic Reality');
  const [type, setType] = useState('ecommerce');
  const [size, setSize] = useState('medium');
  const [seed, setSeed] = useState<number | ''>('');
  const [busy, setBusy] = useState(false);

  const load = () => api.listWorlds().then(setWorlds).catch(() => setWorlds([]));
  useEffect(() => {
    load();
  }, []);

  const create = async () => {
    setBusy(true);
    try {
      await api.createWorld({ name, type, size, seed: seed === '' ? undefined : Number(seed) });
      setShowForm(false);
      load();
    } finally {
      setBusy(false);
    }
  };

  return (
    <div>
      <Topbar title={t('worlds.title')} />

      <button
        onClick={() => setShowForm((v) => !v)}
        className="flex items-center gap-2 rounded-win bg-accent text-white px-4 py-2 text-sm mb-6 hover:bg-accent-hover"
      >
        <Plus size={16} /> {t('worlds.create')}
      </button>

      {showForm && (
        <div className="void-card p-5 mb-6 grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">{t('worlds.name')}</label>
            <input value={name} onChange={(e) => setName(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">{t('worlds.type')}</label>
            <select value={type} onChange={(e) => setType(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm">
              {['ecommerce', 'banking', 'saas', 'social', 'iot', 'logistics', 'cyber_range', 'smart_city', 'cloud_infra', 'generic'].map((o) => (
                <option key={o} value={o}>
                  {o}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">{t('worlds.size')}</label>
            <select value={size} onChange={(e) => setSize(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm">
              {['small', 'medium', 'large', 'massive'].map((o) => (
                <option key={o} value={o}>
                  {o}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">{t('worlds.seed')}</label>
            <input
              type="number"
              value={seed}
              onChange={(e) => setSeed(e.target.value === '' ? '' : Number(e.target.value))}
              placeholder="random"
              className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm"
            />
          </div>
          <button
            onClick={create}
            disabled={busy}
            className="md:col-span-4 rounded-win bg-accent text-white px-4 py-2 text-sm hover:bg-accent-hover disabled:opacity-60 w-fit"
          >
            {t('common.save')}
          </button>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {worlds.map((w) => (
          <div key={w.id} className="void-card p-4 flex flex-col gap-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold">{w.config.name}</span>
              <span className="text-xs px-2 py-0.5 rounded-full bg-accent/15 text-accent">{w.config.type}</span>
            </div>
            <div className="text-sm text-fg-muted">
              {t('worlds.size')}: {w.config.size} · {t('worlds.seed')}: {w.config.seed}
            </div>
            <div className="text-sm text-fg-muted">
              {t('worlds.regions')}: {w.config.regions?.length ?? 0}
            </div>
            <div className="text-xs text-fg-muted/70">{w.id}</div>
          </div>
        ))}
        {worlds.length === 0 && <p className="text-fg-muted">{t('common.noData')}</p>}
      </div>
    </div>
  );
}
