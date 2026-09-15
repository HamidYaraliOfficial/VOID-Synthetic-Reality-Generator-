'use client';

import { useI18n } from '@/lib/i18n';
import { Topbar } from '@/components/topbar';
import { ThemeSwitcher } from '@/components/theme-switcher';
import { LanguageSwitcher } from '@/components/language-switcher';
import { SchedulerPanel } from '@/components/scheduler-panel';
import { API_BASE } from '@/lib/api-client';

export default function SettingsPage() {
  const { t } = useI18n();

  return (
    <div className="flex flex-col gap-8">
      <Topbar title={t('settings.title')} />

      <div className="void-card p-5 flex flex-col gap-5">
        <h3 className="font-semibold text-lg">{t('settings.appearance')}</h3>
        <div className="flex flex-col gap-2">
          <label className="text-sm text-fg-muted">{t('settings.theme')}</label>
          <ThemeSwitcher />
        </div>
        <div className="flex flex-col gap-2 max-w-xs">
          <label className="text-sm text-fg-muted">{t('settings.language')}</label>
          <LanguageSwitcher />
        </div>
        <div className="flex flex-col gap-1 max-w-md">
          <label className="text-sm text-fg-muted">{t('settings.apiBase')}</label>
          <div className="void-card bg-surface-3 px-3 py-2 rounded-win text-sm font-mono">{API_BASE}</div>
          <p className="text-xs text-fg-muted">Set via NEXT_PUBLIC_API_BASE at build/deploy time.</p>
        </div>
      </div>

      <SchedulerPanel />
    </div>
  );
}
