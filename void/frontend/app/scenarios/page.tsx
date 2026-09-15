'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { api, Scenario, SimulationSummary } from '@/lib/api-client';
import { Zap } from 'lucide-react';

const CATEGORY_COLORS: Record<string, string> = {
  growth: 'bg-success/15 text-success',
  chaos: 'bg-danger/15 text-danger',
  fraud: 'bg-warning/15 text-warning',
  market: 'bg-accent/15 text-accent',
  infra: 'bg-fg-muted/15 text-fg-muted',
  template: 'bg-accent/10 text-accent',
};

export default function ScenariosPage() {
  const { t } = useI18n();
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [sims, setSims] = useState<SimulationSummary[]>([]);
  const [simId, setSimId] = useState('');
  const [message, setMessage] = useState('');

  useEffect(() => {
    api.listScenarios().then(setScenarios).catch(() => setScenarios([]));
    api.listSimulations().then((s) => {
      setSims(s || []);
      if (s?.length) setSimId(s[0].id);
    }).catch(() => {});
  }, []);

  const inject = async (scenarioId: string) => {
    if (!simId) {
      setMessage('Create a simulation first.');
      return;
    }
    await api.injectScenario(simId, scenarioId);
    setMessage(`Injected "${scenarioId}" into ${simId}`);
    setTimeout(() => setMessage(''), 3000);
  };

  return (
    <div>
      <Topbar title={t('scenarios.title')} />

      <div className="flex flex-wrap items-center gap-3 mb-6">
        <select value={simId} onChange={(e) => setSimId(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm">
          {sims.map((s) => (
            <option key={s.id} value={s.id}>
              {s.id}
            </option>
          ))}
        </select>
        {message && <span className="text-sm text-accent">{message}</span>}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {scenarios.map((s) => (
          <div key={s.id} className="void-card p-4 flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <span className="font-semibold">{s.name}</span>
              <span className={`text-xs px-2 py-0.5 rounded-full ${CATEGORY_COLORS[s.category] || 'bg-fg-muted/15'}`}>{s.category}</span>
            </div>
            <p className="text-sm text-fg-muted flex-1">{s.description}</p>
            <button
              onClick={() => inject(s.id)}
              className="flex items-center justify-center gap-2 rounded-win bg-accent text-white px-3 py-2 text-sm hover:bg-accent-hover"
            >
              <Zap size={16} /> {t('scenarios.inject')}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
