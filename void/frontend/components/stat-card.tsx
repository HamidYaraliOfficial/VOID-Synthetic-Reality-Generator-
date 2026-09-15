import React from 'react';

export function StatCard({
  label,
  value,
  icon,
  hint,
}: {
  label: string;
  value: string | number;
  icon?: React.ReactNode;
  hint?: string;
}) {
  return (
    <div className="void-card p-4 flex flex-col gap-2 min-w-[180px]">
      <div className="flex items-center justify-between">
        <span className="text-xs uppercase tracking-wide text-fg-muted">{label}</span>
        {icon && <span className="text-accent">{icon}</span>}
      </div>
      <span className="text-2xl font-semibold">{value}</span>
      {hint && <span className="text-xs text-fg-muted">{hint}</span>}
    </div>
  );
}
