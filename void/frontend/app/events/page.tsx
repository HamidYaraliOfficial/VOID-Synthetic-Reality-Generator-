'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { api, VoidEvent } from '@/lib/api-client';

const SEVERITY_COLOR: Record<string, string> = {
  info: 'bg-accent/15 text-accent',
  warning: 'bg-warning/15 text-warning',
  critical: 'bg-danger/15 text-danger',
};

export default function EventsPage() {
  const { t } = useI18n();
  const [events, setEvents] = useState<VoidEvent[]>([]);
  const [typeFilter, setTypeFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');

  useEffect(() => {
    const load = () => {
      const params: Record<string, string> = {};
      if (typeFilter) params.type = typeFilter;
      if (severityFilter) params.severity = severityFilter;
      api.listEvents(params).then(setEvents).catch(() => setEvents([]));
    };
    load();
    const id = setInterval(load, 3000);
    return () => clearInterval(id);
  }, [typeFilter, severityFilter]);

  return (
    <div>
      <Topbar title={t('events.title')} />

      <div className="flex flex-wrap gap-3 mb-6">
        <input
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          placeholder={t('events.type')}
          className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm"
        />
        <select value={severityFilter} onChange={(e) => setSeverityFilter(e.target.value)} className="void-card bg-surface-2 px-3 py-2 rounded-win text-sm">
          <option value="">{t('events.severity')}</option>
          <option value="info">info</option>
          <option value="warning">warning</option>
          <option value="critical">critical</option>
        </select>
      </div>

      <div className="void-card overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-surface-3 text-fg-muted">
            <tr>
              <th className="text-start p-3">{t('events.tick')}</th>
              <th className="text-start p-3">{t('events.type')}</th>
              <th className="text-start p-3">{t('events.source')}</th>
              <th className="text-start p-3">{t('events.target')}</th>
              <th className="text-start p-3">{t('events.severity')}</th>
            </tr>
          </thead>
          <tbody>
            {events.map((e) => (
              <tr key={e.id} className="border-t border-border">
                <td className="p-3 font-mono">{e.tick}</td>
                <td className="p-3">{e.type}</td>
                <td className="p-3 font-mono text-xs">{e.source_id || '—'}</td>
                <td className="p-3 font-mono text-xs">{e.target_id || '—'}</td>
                <td className="p-3">
                  <span className={`text-xs px-2 py-0.5 rounded-full ${SEVERITY_COLOR[e.severity] || SEVERITY_COLOR.info}`}>{e.severity}</span>
                </td>
              </tr>
            ))}
            {events.length === 0 && (
              <tr>
                <td colSpan={5} className="p-4 text-fg-muted">
                  {t('common.noData')}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
