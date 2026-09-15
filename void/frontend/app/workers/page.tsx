'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { api, SimulationSummary, Worker } from '@/lib/api-client';

const HEALTH_COLOR: Record<string, string> = {
  healthy: 'bg-success/15 text-success',
  degraded: 'bg-warning/15 text-warning',
  unhealthy: 'bg-danger/15 text-danger',
  unknown: 'bg-fg-muted/15 text-fg-muted',
};

export default function WorkersPage() {
  const { t } = useI18n();
  const [sims, setSims] = useState<SimulationSummary[]>([]);
  const [simId, setSimId] = useState('');
  const [workers, setWorkers] = useState<Worker[]>([]);

  useEffect(() => {
    api.listSimulations().then((s) => {
      setSims(s || []);
      if (s?.length) setSimId(s[0].id);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    if (!simId) return;
    const load = () => api.listWorkers(simId).then((r) => setWorkers(r.workers || [])).catch(() => setWorkers([]));
    load();
    const id = setInterval(load, 3000);
    return () => clearInterval(id);
  }, [simId]);

  return (
    <div>
      <Topbar title={t('workers.title')} />
      <select value={simId} onChange={(e) => setSimId(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm mb-6">
        {sims.map((s) => (
          <option key={s.id} value={s.id}>
            {s.id}
          </option>
        ))}
      </select>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {workers.map((w) => (
          <div key={w.id} className="void-card p-4 flex flex-col gap-2">
            <div className="flex items-center justify-between">
              <span className="font-mono text-sm">{w.id}</span>
              <span className={`text-xs px-2 py-0.5 rounded-full ${HEALTH_COLOR[w.health] || HEALTH_COLOR.unknown}`}>{w.health}</span>
            </div>
            <div className="text-sm text-fg-muted">{w.address}</div>
            <div className="text-xs text-fg-muted">
              {t('workers.lastHeartbeat')}: {new Date(w.last_heartbeat).toLocaleTimeString()}
            </div>
          </div>
        ))}
        {workers.length === 0 && <p className="text-fg-muted">{t('common.noData')}</p>}
      </div>
    </div>
  );
}
