'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { api, World, SimulationSummary } from '@/lib/api-client';
import { Plus, Play, Pause, RotateCcw, Square, StepForward } from 'lucide-react';

export default function SimulationsPage() {
  const { t } = useI18n();
  const [worlds, setWorlds] = useState<World[]>([]);
  const [sims, setSims] = useState<SimulationSummary[]>([]);
  const [worldId, setWorldId] = useState('');
  const [maxTicks, setMaxTicks] = useState(1000);
  const [workerCount, setWorkerCount] = useState(8);
  const [showForm, setShowForm] = useState(false);

  const load = () => {
    api.listSimulations().then(setSims).catch(() => setSims([]));
  };

  useEffect(() => {
    api.listWorlds().then((w) => {
      setWorlds(w || []);
      if (w?.length) setWorldId(w[0].id);
    }).catch(() => {});
    load();
    const id = setInterval(load, 2000);
    return () => clearInterval(id);
  }, []);

  const create = async () => {
    if (!worldId) return;
    await api.createSimulation({ world_id: worldId, max_ticks: maxTicks, worker_count: workerCount });
    setShowForm(false);
    load();
  };

  const act = async (fn: (id: string) => Promise<unknown>, id: string) => {
    await fn(id);
    load();
  };

  return (
    <div>
      <Topbar title={t('simulations.title')} />

      <button
        onClick={() => setShowForm((v) => !v)}
        className="flex items-center gap-2 rounded-win bg-accent text-white px-4 py-2 text-sm mb-6 hover:bg-accent-hover"
      >
        <Plus size={16} /> {t('simulations.create')}
      </button>

      {showForm && (
        <div className="void-card p-5 mb-6 grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">{t('worlds.title')}</label>
            <select value={worldId} onChange={(e) => setWorldId(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm">
              {worlds.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.config.name}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">Max Ticks</label>
            <input
              type="number"
              value={maxTicks}
              onChange={(e) => setMaxTicks(Number(e.target.value))}
              className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm"
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-fg-muted">Worker Count</label>
            <input
              type="number"
              value={workerCount}
              onChange={(e) => setWorkerCount(Number(e.target.value))}
              className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm"
            />
          </div>
          <button onClick={create} className="rounded-win bg-accent text-white px-4 py-2 text-sm hover:bg-accent-hover w-fit">
            {t('common.save')}
          </button>
        </div>
      )}

      <div className="flex flex-col gap-3">
        {sims.map((s) => (
          <div key={s.id} className="void-card p-4 flex flex-wrap items-center gap-4 justify-between">
            <div>
              <div className="font-mono text-sm">{s.id}</div>
              <div className="text-xs text-fg-muted mt-1">
                {t('simulations.status')}: <span className="text-accent">{s.status}</span> · {t('simulations.tick')}: {s.tick} · {t('simulations.entityCount')}: {s.entity_count}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <IconBtn label={t('simulations.start')} onClick={() => act(api.startSimulation, s.id)} icon={<Play size={16} />} />
              <IconBtn label={t('simulations.pause')} onClick={() => act(api.pauseSimulation, s.id)} icon={<Pause size={16} />} />
              <IconBtn label={t('simulations.resume')} onClick={() => act(api.resumeSimulation, s.id)} icon={<RotateCcw size={16} />} />
              <IconBtn label={t('simulations.step')} onClick={() => act(api.stepSimulation, s.id)} icon={<StepForward size={16} />} />
              <IconBtn label={t('simulations.stop')} onClick={() => act(api.stopSimulation, s.id)} icon={<Square size={16} />} />
            </div>
          </div>
        ))}
        {sims.length === 0 && <p className="text-fg-muted">{t('common.noData')}</p>}
      </div>
    </div>
  );
}

function IconBtn({ label, icon, onClick }: { label: string; icon: React.ReactNode; onClick: () => void }) {
  return (
    <button onClick={onClick} title={label} className="flex items-center gap-1.5 rounded-win border border-border px-3 py-1.5 text-xs hover:bg-surface-3">
      {icon}
      <span className="hidden sm:inline">{label}</span>
    </button>
  );
}
