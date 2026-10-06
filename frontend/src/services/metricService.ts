import { api } from "./api";
import type { HistoryPoint, ProcessInfo, Range, ServerEvent } from "../types/metric";

const MOCK = import.meta.env.VITE_MOCK === "true";
const delay = <T,>(v: T) => new Promise<T>((r) => setTimeout(() => r(v), 300));

const SPAN_MS: Record<Range, number> = {
  "1h": 3_600_000,
  "6h": 21_600_000,
  "24h": 86_400_000,
  "7d": 604_800_000,
};
const POINTS = 72;

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v));

function mockHistory(range: Range): HistoryPoint[] {
  const now = Date.now();
  const step = SPAN_MS[range] / POINTS;
  let cpu = 60, memory = 58, disk = 46;
  return Array.from({ length: POINTS }, (_, i) => {
    cpu = clamp(cpu + (Math.random() - 0.5) * 18, 20, 95);
    memory = clamp(memory + (Math.random() - 0.5) * 5, 45, 80);
    disk = clamp(disk + (Math.random() - 0.5) * 0.6, 40, 52);
    return { t: now - SPAN_MS[range] + i * step, cpu, memory, disk };
  });
}

export const metricService = {
  history: (serverId: string, range: Range): Promise<HistoryPoint[]> =>
    MOCK
      ? delay(mockHistory(range))
      : api<HistoryPoint[]>(`/api/servers/${serverId}/metrics?range=${range}`),

  processes: (serverId: string): Promise<ProcessInfo[]> =>
    MOCK
      ? delay([
          { name: "api-server", cpu: 38, memoryMb: 1200 },
          { name: "postgres", cpu: 17, memoryMb: 2100 },
          { name: "redis-server", cpu: 6, memoryMb: 240 },
          { name: "nginx", cpu: 4, memoryMb: 90 },
        ])
      : api<ProcessInfo[]>(`/api/servers/${serverId}/processes`),

  events: (serverId: string): Promise<ServerEvent[]> =>
    MOCK
      ? delay([
          { id: "1", kind: "alert", text: "CPU above 90% for 5 min", at: Date.now() - 3_600_000, resolved: true },
          { id: "2", kind: "ok", text: "Agent reconnected", at: Date.now() - 16 * 3_600_000 },
          { id: "3", kind: "error", text: "Health check failed (HTTP 502)", at: Date.now() - 16.1 * 3_600_000, resolved: true },
        ])
      : api<ServerEvent[]>(`/api/servers/${serverId}/events`),
};