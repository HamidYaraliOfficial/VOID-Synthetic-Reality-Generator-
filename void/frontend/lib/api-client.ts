const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...(options?.headers || {}) },
    cache: 'no-store',
  });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`API ${res.status} ${path}: ${text}`);
  }
  if (res.status === 204) return undefined as unknown as T;
  return res.json() as Promise<T>;
}

export interface World {
  id: string;
  config: {
    name: string;
    type: string;
    seed: number;
    size: string;
    regions: { id: string; name: string; population: number; wealth_index: number }[];
    economy: { base_currency: string };
    time_scale: number;
    tick_unit: string;
  };
  created_at: string;
}

export interface SimulationSummary {
  id: string;
  status: string;
  tick: number;
  entity_count: number;
}

export interface Archetype {
  name: string;
  description: string;
}

export interface Scenario {
  id: string;
  name: string;
  category: string;
  description: string;
}

export interface EntitySnapshot {
  id: string;
  archetype: string;
  attrs: Record<string, number>;
  state: Record<string, string>;
  resources: Record<string, number>;
  goals: { name: string; priority: number; progress: number }[];
  relationships: { target_id: string; type: string; strength: number }[];
  memory: { tick: number; kind: string; payload?: Record<string, unknown> }[];
}

export interface VoidEvent {
  id: string;
  type: string;
  tick: number;
  timestamp: string;
  source_id?: string;
  target_id?: string;
  severity: string;
  scenario_id?: string;
}

export interface Worker {
  id: string;
  address: string;
  health: string;
  assigned_partition?: string[];
  last_heartbeat: string;
}

export interface MetricsSnapshot {
  counters: Record<string, number>;
  gauges: Record<string, number>;
  goroutines: number;
  heap_alloc_mb: number;
  uptime_seconds: number;
}

export interface OperatingHoursWindow {
  weekday: number;
  start_hour: number;
  start_min: number;
  end_hour: number;
  end_min: number;
}

export interface OperatingHoursStatus {
  open: boolean;
  now: string;
  next_transition?: string;
  time_until_next?: number; // nanoseconds
  active_window_label?: string;
}

export const api = {
  health: () => request<{ status: string }>('/healthz'),

  listWorlds: () => request<World[]>('/api/v1/worlds'),
  getWorld: (id: string) => request<World>(`/api/v1/worlds/${id}`),
  createWorld: (body: Partial<{ name: string; type: string; seed: number; size: string; description: string }>) =>
    request<World>('/api/v1/worlds', { method: 'POST', body: JSON.stringify(body) }),

  listArchetypes: () => request<Archetype[]>('/api/v1/archetypes'),

  generatePopulation: (worldId: string, body: unknown) =>
    request<{ generated: number }>(`/api/v1/worlds/${worldId}/population`, { method: 'POST', body: JSON.stringify(body) }),

  listSimulations: () => request<SimulationSummary[]>('/api/v1/simulations'),
  getSimulation: (id: string) => request<SimulationSummary>(`/api/v1/simulations/${id}`),
  createSimulation: (body: { world_id: string; max_ticks?: number; worker_count?: number; seed?: number }) =>
    request<{ id: string }>('/api/v1/simulations', { method: 'POST', body: JSON.stringify(body) }),
  startSimulation: (id: string) => request(`/api/v1/simulations/${id}/start`, { method: 'POST' }),
  pauseSimulation: (id: string) => request(`/api/v1/simulations/${id}/pause`, { method: 'POST' }),
  resumeSimulation: (id: string) => request(`/api/v1/simulations/${id}/resume`, { method: 'POST' }),
  stopSimulation: (id: string) => request(`/api/v1/simulations/${id}/stop`, { method: 'POST' }),
  stepSimulation: (id: string) => request(`/api/v1/simulations/${id}/step`, { method: 'POST' }),

  listEntities: (simId: string, limit = 200) => request<EntitySnapshot[]>(`/api/v1/simulations/${simId}/entities?limit=${limit}`),
  getEntity: (simId: string, entityId: string) => request<EntitySnapshot>(`/api/v1/simulations/${simId}/entities/${entityId}`),

  listEvents: (params: Record<string, string> = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request<VoidEvent[]>(`/api/v1/events${qs ? `?${qs}` : ''}`);
  },

  listScenarios: () => request<Scenario[]>('/api/v1/scenarios'),
  injectScenario: (simId: string, scenarioId: string) =>
    request(`/api/v1/simulations/${simId}/scenarios/${scenarioId}/inject`, { method: 'POST' }),

  injectFault: (body: unknown) => request('/api/v1/chaos/faults', { method: 'POST', body: JSON.stringify(body) }),
  chaosImpact: () => request<Record<string, number>>('/api/v1/chaos/impact'),

  listWorkers: (simId: string) => request<{ workers: Worker[] }>(`/api/v1/simulations/${simId}/workers`),

  getMetrics: () => request<MetricsSnapshot>('/api/v1/metrics'),

  listJobs: () => request<unknown[]>('/api/v1/scheduler/jobs'),
  enqueueJob: (body: unknown) => request('/api/v1/scheduler/jobs', { method: 'POST', body: JSON.stringify(body) }),

  getOperatingHours: () => request<{ timezone: string; windows: OperatingHoursWindow[] }>('/api/v1/scheduler/operating-hours'),
  setOperatingHours: (body: { timezone: string; windows: OperatingHoursWindow[] }) =>
    request('/api/v1/scheduler/operating-hours', { method: 'PUT', body: JSON.stringify(body) }),
  operatingHoursStatus: () => request<OperatingHoursStatus>('/api/v1/scheduler/operating-hours/status'),

  exportSimulation: (simId: string, format: string) =>
    request<{ path: string; count: number }>(`/api/v1/simulations/${simId}/export`, { method: 'POST', body: JSON.stringify({ format }) }),
  snapshotSimulation: (simId: string) => request(`/api/v1/simulations/${simId}/snapshot`, { method: 'POST' }),

  wsURL: (simId: string) => `${API_BASE.replace(/^http/, 'ws')}/ws/simulations/${simId}`,
};

export { API_BASE };
