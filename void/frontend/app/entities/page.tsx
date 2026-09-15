'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { api, SimulationSummary, EntitySnapshot } from '@/lib/api-client';
import { Search } from 'lucide-react';

export default function EntitiesPage() {
  const { t } = useI18n();
  const [sims, setSims] = useState<SimulationSummary[]>([]);
  const [simId, setSimId] = useState('');
  const [entities, setEntities] = useState<EntitySnapshot[]>([]);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<EntitySnapshot | null>(null);

  useEffect(() => {
    api.listSimulations().then((s) => {
      setSims(s || []);
      if (s?.length) setSimId(s[0].id);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    if (!simId) return;
    api.listEntities(simId, 300).then(setEntities).catch(() => setEntities([]));
  }, [simId]);

  const filtered = entities.filter((e) => e.id.toLowerCase().includes(query.toLowerCase()) || e.archetype.includes(query));

  return (
    <div>
      <Topbar title={t('entities.title')} />

      <div className="flex flex-wrap gap-3 mb-6 items-center">
        <select value={simId} onChange={(e) => setSimId(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm">
          {sims.map((s) => (
            <option key={s.id} value={s.id}>
              {s.id}
            </option>
          ))}
        </select>
        <div className="flex items-center gap-2 void-card bg-surface-2 px-3 py-2 rounded-win flex-1 min-w-[200px]">
          <Search size={16} className="text-fg-muted" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('entities.search')}
            className="bg-transparent outline-none text-sm w-full"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 void-card overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-surface-3 text-fg-muted">
              <tr>
                <th className="text-start p-3">ID</th>
                <th className="text-start p-3">{t('entities.archetype')}</th>
                <th className="text-start p-3">Status</th>
              </tr>
            </thead>
            <tbody>
              {filtered.slice(0, 100).map((e) => (
                <tr
                  key={e.id}
                  onClick={() => setSelected(e)}
                  className={`cursor-pointer border-t border-border hover:bg-surface-3 ${selected?.id === e.id ? 'bg-accent/10' : ''}`}
                >
                  <td className="p-3 font-mono text-xs">{e.id}</td>
                  <td className="p-3">{e.archetype}</td>
                  <td className="p-3">{e.state?.status}</td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={3} className="p-4 text-fg-muted">
                    {t('common.noData')}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="void-card p-4">
          {!selected && <p className="text-fg-muted text-sm">{t('common.noData')}</p>}
          {selected && (
            <div className="flex flex-col gap-4 text-sm">
              <div>
                <div className="font-mono text-xs text-fg-muted">{selected.id}</div>
                <div className="font-semibold text-base">{selected.archetype}</div>
              </div>
              <Section title={t('entities.attributes')}>
                {Object.entries(selected.attrs || {}).map(([k, v]) => (
                  <Row key={k} k={k} v={typeof v === 'number' ? v.toFixed(3) : String(v)} />
                ))}
              </Section>
              <Section title={t('entities.goals')}>
                {(selected.goals || []).map((g, i) => (
                  <Row key={i} k={g.name} v={`${(g.progress * 100).toFixed(0)}%`} />
                ))}
              </Section>
              <Section title={t('entities.relationships')}>
                {(selected.relationships || []).map((r, i) => (
                  <Row key={i} k={r.type} v={r.target_id} />
                ))}
              </Section>
              <Section title={t('entities.memory')}>
                {(selected.memory || []).slice(-5).reverse().map((m, i) => (
                  <Row key={i} k={`t${m.tick}`} v={m.kind} />
                ))}
              </Section>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="text-xs uppercase text-fg-muted mb-1">{title}</div>
      <div className="flex flex-col gap-1">{children}</div>
    </div>
  );
}
function Row({ k, v }: { k: string; v: string }) {
  return (
    <div className="flex items-center justify-between border-b border-border/50 py-1">
      <span className="text-fg-muted">{k}</span>
      <span>{v}</span>
    </div>
  );
}
