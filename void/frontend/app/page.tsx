'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { StatCard } from '@/components/stat-card';
import { TickLineChart, SeriesPoint } from '@/components/charts/tick-chart';
import { api, SimulationSummary, World, MetricsSnapshot } from '@/lib/api-client';
import { Globe2, PlayCircle, Users, Activity, Gauge, ServerCog } from 'lucide-react';

export default function DashboardPage() {
  const { t } = useI18n();
  const [worlds, setWorlds] = useState<World[]>([]);
  const [sims, setSims] = useState<SimulationSummary[]>([]);
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null);
  const [history, setHistory] = useState<SeriesPoint[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const [w, s, m] = await Promise.all([api.listWorlds(), api.listSimulations(), api.getMetrics()]);
        setWorlds(w || []);
        setSims(s || []);
        setMetrics(m);
        setError(null);
        setHistory((h) => {
          const eventsProcessed = Object.entries(m.counters || {})
            .filter(([k]) => k.endsWith('.events_processed'))
            .reduce((sum, [, v]) => sum + v, 0);
          const next = [...h, { label: new Date().toLocaleTimeString(), value: eventsProcessed }];
          return next.slice(-30);
        });
      } catch (e: any) {
        setError(e.message || 'Failed to reach the VOID API');
      }
    };
    load();
    const id = setInterval(load, 3000);
    return () => clearInterval(id);
  }, []);

  const totalEntities = sims.reduce((sum, s) => sum + (s.entity_count || 0), 0);
  const runningSims = sims.filter((s) => s.status === 'running').length;

  return (
    <div>
      <Topbar title={t('dashboard.title')} subtitle={t('dashboard.subtitle')} />

      {error && (
        <div className="void-card border-danger/40 bg-danger/10 text-danger p-4 mb-6 text-sm">
          {error} — {t('common.noData')} ({process.env.NEXT_PUBLIC_API_BASE})
        </div>
      )}

      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4 mb-8">
        <StatCard label={t('dashboard.activeWorlds')} value={worlds.length} icon={<Globe2 size={18} />} />
        <StatCard label={t('dashboard.runningSimulations')} value={runningSims} icon={<PlayCircle size={18} />} />
        <StatCard label={t('dashboard.totalEntities')} value={totalEntities.toLocaleString()} icon={<Users size={18} />} />
        <StatCard
          label={t('dashboard.eventsPerSecond')}
          value={history.length ? history[history.length - 1].value : 0}
          icon={<Activity size={18} />}
        />
        <StatCard label={t('dashboard.tickRate')} value={sims[0]?.tick ?? 0} icon={<Gauge size={18} />} />
        <StatCard
          label={t('dashboard.healthyWorkers')}
          value={metrics ? metrics.goroutines : 0}
          icon={<ServerCog size={18} />}
          hint="goroutines"
        />
      </div>

      <TickLineChart data={history} dataKeyLabel={t('dashboard.eventsPerSecond')} />
    </div>
  );
}
