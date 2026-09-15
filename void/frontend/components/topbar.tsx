'use client';

import { ThemeSwitcher } from './theme-switcher';
import { LanguageSwitcher } from './language-switcher';

export function Topbar({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <div className="flex items-start justify-between gap-4 mb-6 flex-wrap">
      <div>
        <h1 className="text-2xl font-semibold">{title}</h1>
        {subtitle && <p className="text-fg-muted mt-1 max-w-2xl">{subtitle}</p>}
      </div>
      <div className="flex items-center gap-3">
        <LanguageSwitcher />
        <ThemeSwitcher compact />
      </div>
    </div>
  );
}
