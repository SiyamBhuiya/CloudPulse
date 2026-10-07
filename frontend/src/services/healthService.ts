import { api } from "./api";
import type { DayStatus, HealthCheck, HealthCheckInput } from "../types/healthcheck";

const MOCK = import.meta.env.VITE_MOCK === "true";
const delay = <T,>(v: T) => new Promise<T>((r) => setTimeout(() => r(v), 300));
const ago = (s: number) => Date.now() - s * 1000;

// 30 days, all fine except the listed day numbers (0 = oldest)
const days = (bad: Record<number, DayStatus> = {}): DayStatus[] =>
  Array.from({ length: 30 }, (_, i) => bad[i] ?? "ok");

// Fake data, only used when VITE_MOCK=true
let mockChecks: HealthCheck[] = [
  { id: "1", name: "Main API", url: "https://api.example.com/health", status: "up", responseMs: 142, uptime: 99.98, intervalSec: 60, expectedStatus: 200, lastChecked: ago(12), enabled: true, days: days() },
  { id: "2", name: "Marketing site", url: "https://example.com", status: "slow", responseMs: 1480, uptime: 99.41, intervalSec: 60, expectedStatus: 200, lastChecked: ago(30), enabled: true, days: days({ 12: "degraded", 13: "degraded" }) },
  { id: "3", name: "Checkout service", url: "https://pay.example.com/ping", status: "up", responseMs: 96, uptime: 99.99, intervalSec: 30, expectedStatus: 200, lastChecked: ago(8), enabled: true, days: days() },
  { id: "4", name: "Legacy admin", url: "https://admin.example.com/login", status: "down", responseMs: null, uptime: 97.1, intervalSec: 300, expectedStatus: 200, lastChecked: ago(120), enabled: true, days: days({ 27: "down", 28: "down", 29: "down" }) },
];

export const healthService = {
  list: (): Promise<HealthCheck[]> =>
    MOCK ? delay([...mockChecks]) : api<HealthCheck[]>("/api/health-checks"),

  create: (input: HealthCheckInput): Promise<HealthCheck> => {
    if (MOCK) {
      const check: HealthCheck = {
        ...input,
        id: String(Date.now()),
        status: "up",
        responseMs: 120,
        uptime: 100,
        lastChecked: Date.now(),
        enabled: true,
        days: days(),
      };
      mockChecks = [...mockChecks, check];
      return delay(check);
    }
    return api<HealthCheck>("/api/health-checks", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  setEnabled: (id: string, enabled: boolean): Promise<void> => {
    if (MOCK) {
      mockChecks = mockChecks.map((c) => (c.id === id ? { ...c, enabled } : c));
      return delay(undefined);
    }
    return api<void>(`/api/health-checks/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ enabled }),
    });
  },

  remove: (id: string): Promise<void> => {
    if (MOCK) {
      mockChecks = mockChecks.filter((c) => c.id !== id);
      return delay(undefined);
    }
    return api<void>(`/api/health-checks/${id}`, { method: "DELETE" });
  },
};