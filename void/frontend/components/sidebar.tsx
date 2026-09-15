'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useI18n } from '@/lib/i18n';
import {
  LayoutDashboard, Globe2, Users, Sparkles, PlayCircle,
  Server, Activity, ListTree, Settings,
} from 'lucide-react';

const ITEMS = [
  { href: '/', icon: LayoutDashboard, key: 'nav.dashboard' },
  { href: '/worlds', icon: Globe2, key: 'nav.worlds' },
  { href: '/entities', icon: Users, key: 'nav.entities' },
  { href: '/scenarios', icon: Sparkles, key: 'nav.scenarios' },
  { href: '/simulations', icon: PlayCircle, key: 'nav.simulations' },
  { href: '/workers', icon: Server, key: 'nav.workers' },
  { href: '/metrics', icon: Activity, key: 'nav.metrics' },
  { href: '/events', icon: ListTree, key: 'nav.events' },
  { href: '/settings', icon: Settings, key: 'nav.settings' },
];

export function Sidebar() {
  const pathname = usePathname();
  const { t } = useI18n();

  return (
    <aside className="acrylic w-64 shrink-0 h-screen sticky top-0 flex flex-col p-4 gap-1">
      <div className="flex items-center gap-2 px-2 py-3 mb-2">
        <div className="w-8 h-8 rounded-win bg-accent flex items-center justify-center text-white font-bold">V</div>
        <div>
          <div className="font-semibold leading-none">{t('appName')}</div>
          <div className="text-xs text-fg-muted">{t('tagline')}</div>
        </div>
      </div>
      <nav className="flex flex-col gap-1">
        {ITEMS.map(({ href, icon: Icon, key }) => {
          const active = pathname === href;
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-3 rounded-win px-3 py-2 text-sm transition-colors
                ${active ? 'bg-accent/15 text-accent' : 'text-fg-muted hover:text-fg hover:bg-surface-3'}`}
            >
              <Icon size={18} />
              <span>{t(key)}</span>
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
