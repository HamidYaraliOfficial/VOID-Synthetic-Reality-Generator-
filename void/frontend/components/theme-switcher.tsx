'use client';

import { useTheme, ThemeName } from './theme-provider';
import { useI18n } from '@/lib/i18n';
import { Sun, Moon, Monitor, Flame, Droplet } from 'lucide-react';

const THEMES: { key: ThemeName; icon: React.ReactNode; labelKey: string }[] = [
  { key: 'windows', icon: <Monitor size={16} />, labelKey: 'settings.themeWindows' },
  { key: 'light', icon: <Sun size={16} />, labelKey: 'settings.themeLight' },
  { key: 'dark', icon: <Moon size={16} />, labelKey: 'settings.themeDark' },
  { key: 'red', icon: <Flame size={16} />, labelKey: 'settings.themeRed' },
  { key: 'blue', icon: <Droplet size={16} />, labelKey: 'settings.themeBlue' },
];

export function ThemeSwitcher({ compact = false }: { compact?: boolean }) {
  const { theme, setTheme } = useTheme();
  const { t } = useI18n();

  return (
    <div className={`flex ${compact ? 'gap-1' : 'gap-2 flex-wrap'}`}>
      {THEMES.map((opt) => (
        <button
          key={opt.key}
          onClick={() => setTheme(opt.key)}
          className={`flex items-center gap-2 rounded-win border px-3 py-1.5 text-sm transition-colors
            ${theme === opt.key ? 'border-accent bg-accent/10 text-accent' : 'border-border text-fg-muted hover:text-fg'}`}
          aria-pressed={theme === opt.key}
        >
          {opt.icon}
          {!compact && <span>{t(opt.labelKey)}</span>}
        </button>
      ))}
    </div>
  );
}
