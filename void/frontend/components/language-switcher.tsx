'use client';

import { useI18n, Locale } from '@/lib/i18n';
import { Languages } from 'lucide-react';

const LANGS: { key: Locale; label: string }[] = [
  { key: 'en', label: 'English' },
  { key: 'fa', label: 'فارسی' },
  { key: 'zh', label: '中文' },
];

export function LanguageSwitcher() {
  const { locale, setLocale } = useI18n();
  return (
    <div className="flex items-center gap-2">
      <Languages size={16} className="text-fg-muted" />
      <select
        value={locale}
        onChange={(e) => setLocale(e.target.value as Locale)}
        className="void-card bg-surface-2 text-fg text-sm rounded-win px-2 py-1.5 outline-none"
      >
        {LANGS.map((l) => (
          <option key={l.key} value={l.key}>
            {l.label}
          </option>
        ))}
      </select>
    </div>
  );
}
