'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n';
import { api, OperatingHoursWindow, OperatingHoursStatus } from '@/lib/api-client';
import { Plus, Trash2, Clock } from 'lucide-react';

function emptyWindow(): OperatingHoursWindow {
  return { weekday: 1, start_hour: 9, start_min: 0, end_hour: 18, end_min: 0 };
}

function formatDuration(ns?: number): string {
  if (!ns || ns <= 0) return '—';
  const totalSeconds = Math.floor(ns / 1e9);
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  const days = Math.floor(h / 24);
  const hh = h % 24;
  if (days > 0) return `${days}d ${hh}h ${m}m`;
  return `${String(hh).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

export function SchedulerPanel() {
  const { t } = useI18n();
  const weekdays: string[] = t('weekdays');

  const [timezone, setTimezone] = useState('UTC');
  const [windows, setWindows] = useState<OperatingHoursWindow[]>([]);
  const [status, setStatus] = useState<OperatingHoursStatus | null>(null);
  const [saving, setSaving] = useState(false);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    api
      .getOperatingHours()
      .then((res) => {
        setTimezone(res.timezone || 'UTC');
        setWindows(res.windows || []);
      })
      .catch(() => {})
      .finally(() => setLoaded(true));
  }, []);

  useEffect(() => {
    const poll = () => api.operatingHoursStatus().then(setStatus).catch(() => {});
    poll();
    const id = setInterval(poll, 1000);
    return () => clearInterval(id);
  }, []);

  const addWindow = () => setWindows((w) => [...w, emptyWindow()]);
  const removeWindow = (idx: number) => setWindows((w) => w.filter((_, i) => i !== idx));
  const updateWindow = (idx: number, patch: Partial<OperatingHoursWindow>) =>
    setWindows((w) => w.map((win, i) => (i === idx ? { ...win, ...patch } : win)));

  const save = async () => {
    setSaving(true);
    try {
      await api.setOperatingHours({ timezone, windows });
      const s = await api.operatingHoursStatus();
      setStatus(s);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="void-card p-5 flex flex-col gap-5">
      <div>
        <h3 className="font-semibold text-lg flex items-center gap-2">
          <Clock size={18} className="text-accent" /> {t('settings.operatingHours')}
        </h3>
        <p className="text-sm text-fg-muted mt-1">{t('settings.operatingHoursDesc')}</p>
      </div>

      {status && (
        <div
          className={`rounded-win px-4 py-3 text-sm flex items-center justify-between border
          ${status.open ? 'border-success/40 bg-success/10 text-success' : 'border-warning/40 bg-warning/10 text-warning'}`}
        >
          <span className="font-medium">{status.open ? t('settings.statusOpen') : t('settings.statusClosed')}</span>
          <span>
            {status.open ? t('settings.nextTransitionOpen') : t('settings.nextTransitionClosed')}:{' '}
            <strong>{formatDuration(status.time_until_next)}</strong>
          </span>
        </div>
      )}

      <div className="flex flex-col gap-2 max-w-xs">
        <label className="text-sm text-fg-muted">{t('settings.timezone')}</label>
        <input
          value={timezone}
          onChange={(e) => setTimezone(e.target.value)}
          placeholder="e.g. Asia/Baku, UTC, America/New_York"
          className="void-card bg-surface-2 px-3 py-2 rounded-win outline-none text-sm"
        />
      </div>

      <div className="flex flex-col gap-3">
        {loaded && windows.length === 0 && <p className="text-sm text-fg-muted">{t('common.noData')}</p>}
        {windows.map((w, idx) => (
          <div key={idx} className="flex flex-wrap items-center gap-3 void-card p-3 bg-surface-3">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-fg-muted">{t('settings.weekday')}</label>
              <select
                value={w.weekday}
                onChange={(e) => updateWindow(idx, { weekday: Number(e.target.value) })}
                className="void-card bg-surface-2 px-2 py-1.5 rounded-win text-sm"
              >
                {weekdays.map((d, i) => (
                  <option key={i} value={i}>
                    {d}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-fg-muted">{t('settings.startTime')}</label>
              <input
                type="time"
                value={`${String(w.start_hour).padStart(2, '0')}:${String(w.start_min).padStart(2, '0')}`}
                onChange={(e) => {
                  const [h, m] = e.target.value.split(':').map(Number);
                  updateWindow(idx, { start_hour: h, start_min: m });
                }}
                className="void-card bg-surface-2 px-2 py-1.5 rounded-win text-sm"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-fg-muted">{t('settings.endTime')}</label>
              <input
                type="time"
                value={`${String(w.end_hour).padStart(2, '0')}:${String(w.end_min).padStart(2, '0')}`}
                onChange={(e) => {
                  const [h, m] = e.target.value.split(':').map(Number);
                  updateWindow(idx, { end_hour: h, end_min: m });
                }}
                className="void-card bg-surface-2 px-2 py-1.5 rounded-win text-sm"
              />
            </div>
            <button
              onClick={() => removeWindow(idx)}
              className="ms-auto flex items-center gap-1 text-sm text-danger hover:opacity-80 self-end pb-1.5"
            >
              <Trash2 size={16} /> {t('settings.remove')}
            </button>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <button
          onClick={addWindow}
          className="flex items-center gap-2 rounded-win border border-border px-3 py-2 text-sm hover:bg-surface-3"
        >
          <Plus size={16} /> {t('settings.addWindow')}
        </button>
        <button
          onClick={save}
          disabled={saving}
          className="flex items-center gap-2 rounded-win bg-accent text-white px-4 py-2 text-sm hover:bg-accent-hover disabled:opacity-60"
        >
          {t('settings.save')}
        </button>
      </div>
    </div>
  );
}
