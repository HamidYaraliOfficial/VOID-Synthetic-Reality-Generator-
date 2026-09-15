'use client';

import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export interface SeriesPoint {
  label: string | number;
  value: number;
}

export function TickLineChart({ data, dataKeyLabel = 'Value' }: { data: SeriesPoint[]; dataKeyLabel?: string }) {
  return (
    <div className="w-full h-64 void-card p-4">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="label" stroke="var(--fg-muted)" fontSize={12} />
          <YAxis stroke="var(--fg-muted)" fontSize={12} />
          <Tooltip
            contentStyle={{ background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 8, color: 'var(--fg)' }}
          />
          <Line type="monotone" dataKey="value" name={dataKeyLabel} stroke="var(--accent)" strokeWidth={2} dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
