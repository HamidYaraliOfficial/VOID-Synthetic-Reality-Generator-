'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { StatCard } from '@/components/stat-card';
import { TickLineChart, SeriesPoint } from '@/components/charts/tick-chart';
import { api, MetricsSnapshot } from '@/lib/api-client';

export default function MetricsPage() {
  const { t } = useI18n();
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null);
  const [heapHistory, setHeapHistory] = useState<SeriesPoint[]>([]);

  useEffect(() => {
    const load = () =>
      api.getMetrics().then((m) => {
        setMetrics(m);
        setHeapHistory((h) => [...h, { label: new Date().toLocaleTimeString(), value: m.heap_alloc_mb }].slice(-30));
      }).catch(() => {});
    load();
    const id = setInterval(load, 2000);
    return () => clearInterval(id);
  }, []);

  return (
    <div>
      <Topbar title={t('metrics.title')} />

      <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-8">
        <StatCard label={t('metrics.heap')} value={metrics?.heap_alloc_mb.toFixed(1) ?? '—'} />
        <StatCard label={t('metrics.goroutines')} value={metrics?.goroutines ?? '—'} />
        <StatCard label={t('metrics.uptime')} value={metrics ? `${Math.floor(metrics.uptime_seconds)}s` : '—'} />
      </div>

      <TickLineChart data={heapHistory} dataKeyLabel={t('metrics.heap')} />

      <div className="void-card p-4 mt-6">
        <h3 className="font-semibold mb-3">Counters &amp; Gauges</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-1 text-sm font-mono">
          {metrics &&
            Object.entries({ ...metrics.counters, ...metrics.gauges }).map(([k, v]) => (
              <div key={k} className="flex justify-between border-b border-border/50 py-1">
                <span className="text-fg-muted">{k}</span>
                <span>{typeof v === 'number' ? v.toFixed(2) : String(v)}</span>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}
